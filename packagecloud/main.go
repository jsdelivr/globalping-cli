package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	PackagecloudAPIURL = "https://packagecloud.io"
)

type Config struct {
	PackagecloudUser   string
	PackagecloudRepo   string
	PackagecloudAPIKey string

	MaxDistroVersionsToSupport int
	MaxPackageVersionsToKeep   int
}

func main() {
	cmd := os.Args[1]
	config := &Config{
		PackagecloudUser:   os.Getenv("PACKAGECLOUD_USER"),
		PackagecloudRepo:   os.Getenv("PACKAGECLOUD_REPO"),
		PackagecloudAPIKey: os.Getenv("PACKAGECLOUD_APIKEY"),

		MaxDistroVersionsToSupport: 5,
		MaxPackageVersionsToKeep:   3,
	}
	maxDistrosVersionsToSupport, _ := strconv.Atoi(os.Getenv("PACKAGECLOUD_MAX_DISTRO_VERSIONS_TO_SUPPORT"))
	if maxDistrosVersionsToSupport != 0 {
		config.MaxDistroVersionsToSupport = maxDistrosVersionsToSupport
	}
	maxPackageVersionsToKeep, _ := strconv.Atoi(os.Getenv("PACKAGECLOUD_MAX_PACKAGE_VERSIONS_TO_KEEP"))
	if maxPackageVersionsToKeep != 0 {
		config.MaxPackageVersionsToKeep = maxPackageVersionsToKeep
	}
	switch cmd {
	case "upload":
		upload(os.Args[2], os.Args[3], config)
	case "cleanup":
		cleanup(config)
	default:
		fmt.Println("Unknown command")
		os.Exit(1)
	}
}

func upload(file string, dist string, config *Config) {
	if file == "" || dist == "" {
		fmt.Println("Usage: upload <path> <dist>")
		os.Exit(1)
	}
	log.Println("Fetching distros")
	distros, err := fetchDistros(config)
	if err != nil {
		log.Println("Failed to fetch distros: ", err)
		os.Exit(1)
	}
	f, err := os.Open(file)
	if err != nil {
		log.Println("Failed to open file: ", err)
		os.Exit(1)
	}
	defer f.Close()
	switch dist {
	case "deb":
		uploadToAllDistros(config, f, distros.Deb)
	case "rpm":
		uploadToAllDistros(config, f, distros.Rpm)

	default:
		fmt.Println("Unknown distro")
		os.Exit(1)
	}
}

func uploadToAllDistros(config *Config, f *os.File, distros []*Distro) {
	totalDistros := 0
	for _, distro := range distros {
		l := len(distro.Versions)
		for j := l - 1; (j >= 0) && (l-j <= config.MaxDistroVersionsToSupport); j-- {
			totalDistros++
			log.Printf("Uploading to %s/%s (%d)\n", distro.IndexName, distro.Versions[j].IndexName, distro.Versions[j].ID)
			err := uploadPackage(config, strconv.Itoa(distro.Versions[j].ID), f)
			if err != nil {
				log.Println("Failed to upload package: ", err)
				os.Exit(1)
			}
			_, err = f.Seek(0, 0)
			if err != nil {
				log.Println("Failed to seek file: ", err)
				os.Exit(1)
			}
		}
	}
	log.Printf("Uploaded to %d distros\n", totalDistros)
}

func cleanup(config *Config) {
	log.Println("Fetching distros")
	distros, err := fetchDistros(config)
	if err != nil {
		log.Println("Failed to fetch distros: ", err)
		os.Exit(1)
	}
	destroyOlderVersions(config, "deb", distros.Deb)
	destroyOlderVersions(config, "rpm", distros.Rpm)
	log.Println("Cleanup completed")
}

func destroyOlderVersions(config *Config, t string, distros []*Distro) {
	totalDestroyed := 0
	for _, distro := range distros {
		l := len(distro.Versions)
		for j := l - 1; j >= 0; j-- {
			versionGroups, err := fetchVersionGroups(config, t, distro.IndexName, distro.Versions[j].IndexName)
			if err != nil {
				log.Printf("Failed to fetch version groups for %s %s/%s: %s\n", t, distro.IndexName, distro.Versions[j].IndexName, err)
				os.Exit(1)
			}
			noLongerSupported := l-j > config.MaxDistroVersionsToSupport
			for _, versionGroup := range versionGroups {
				versions, err := fetchVersions(config, versionGroup.VersionsURL)
				if err != nil {
					log.Printf("Failed to fetch versions for %s/%s: %s\n", distro.IndexName, distro.Versions[j].IndexName, err)
					os.Exit(1)
				}
				destroyTo := len(versions) - config.MaxPackageVersionsToKeep
				if noLongerSupported {
					// If we no longer support this distro version, delete all versions
					log.Printf("%s %s/%s is no longer supported, deleting all versions\n", t, distro.IndexName, distro.Versions[j].IndexName)
					destroyTo = len(versions)
				}
				for i := 0; i < destroyTo; i++ {
					v := versions[i]
					log.Printf("Destroying version %s %s/%s: %s %s\n", t, distro.IndexName, distro.Versions[j].IndexName, v.Version, v.Filename)
					err = deleteVersion(config, v.DestroyURL)
					if err != nil {
						log.Println("Failed to destroy version: ", err)
						os.Exit(1)
					}
					totalDestroyed++
				}
			}
		}
	}
	log.Printf("%s: %d versions destroyed\n", t, totalDestroyed)
}

func uploadPackage(config *Config, distroVersionId string, f *os.File) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("package[package_file]", f.Name())
	if err != nil {
		return err
	}
	_, err = io.Copy(part, f)
	if err != nil {
		return err
	}
	err = writer.WriteField("package[distro_version_id]", distroVersionId)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", PackagecloudAPIURL+"/api/v1/repos/"+config.PackagecloudUser+"/"+config.PackagecloudRepo+"/packages.json", &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth(config.PackagecloudAPIKey, "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(b))
	}
	return nil
}

type DistroVersion struct {
	ID        int    `json:"id"`
	IndexName string `json:"index_name"`
}

type Distro struct {
	IndexName string `json:"index_name"`
	Versions  []*DistroVersion
}

type Distros struct {
	Deb []*Distro `json:"deb"`
	Rpm []*Distro `json:"rpm"`
}

func fetchDistros(config *Config) (*Distros, error) {
	req, err := http.NewRequest("GET", PackagecloudAPIURL+"/api/v1/distributions.json", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(config.PackagecloudAPIKey, "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(b))
	}
	distros := &Distros{}
	json.NewDecoder(resp.Body).Decode(distros)
	return distros, nil
}

type PackageVersion struct {
	CreatedAt  time.Time `json:"created_at"`
	DestroyURL string    `json:"destroy_url"`
	Version    string    `json:"version"`
	Filename   string    `json:"filename"`
}

func fetchVersions(config *Config, url string) ([]*PackageVersion, error) {
	req, err := http.NewRequest("GET", PackagecloudAPIURL+url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(config.PackagecloudAPIKey, "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(b))
	}
	versions := []*PackageVersion{}
	json.NewDecoder(resp.Body).Decode(&versions)
	return versions, nil
}

type PackageVersionGroup struct {
	VersionsURL string `json:"versions_url"`
}

func fetchVersionGroups(config *Config, t string, distro string, name string) ([]*PackageVersionGroup, error) {
	req, err := http.NewRequest("GET", PackagecloudAPIURL+"/api/v1/repos/"+config.PackagecloudUser+"/"+config.PackagecloudRepo+"/packages/"+t+"/"+distro+"/"+name+".json", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(config.PackagecloudAPIKey, "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(b))
	}
	versions := []*PackageVersionGroup{}
	json.NewDecoder(resp.Body).Decode(&versions)
	return versions, nil
}

func deleteVersion(config *Config, url string) error {
	req, err := http.NewRequest("DELETE", PackagecloudAPIURL+url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(config.PackagecloudAPIKey, "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(b))
	}
	return nil
}
