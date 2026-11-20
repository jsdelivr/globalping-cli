package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
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

		MaxDistroVersionsToSupport: 10,
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

	if dist != "deb" && dist != "rpm" {
		fmt.Println("Unknown distro")
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

	defer func() {
		_ = f.Close()
	}()

	switch dist {
	case "deb":
		uploadToAllDistros(config, f, dist, distros.Deb)
	case "rpm":
		uploadToAllDistros(config, f, dist, distros.Rpm)
	}
}

func uploadToAllDistros(config *Config, f *os.File, packageType string, distros []*Distro) {
	totalDistros := 0

	for _, distro := range distros {
		l := len(distro.Versions)

		for j := l - 1; (j >= 0) && (l-j <= config.MaxDistroVersionsToSupport); j-- {
			version := distro.Versions[j]
			totalDistros++
			log.Printf("Uploading to %s/%s (%d)\n", distro.IndexName, version.IndexName, version.ID)
			err := uploadPackage(config, packageType, distro.IndexName, version, f)

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
			path := versionGroupsPath(config, t, distro.IndexName, distro.Versions[j].IndexName)
			var versionGroups []*PackageVersionGroup
			_, err := getPackagecloudJSON(context.Background(), config, path, &versionGroups)

			if err != nil {
				log.Printf("Failed to fetch version groups for %s %s/%s: %s\n", t, distro.IndexName, distro.Versions[j].IndexName, err)
				os.Exit(1)
			}

			noLongerSupported := l-j > config.MaxDistroVersionsToSupport

			for _, versionGroup := range versionGroups {
				var versions []*PackageVersion
				_, err := getPackagecloudJSON(context.Background(), config, versionGroup.VersionsURL, &versions)

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

				for i := range destroyTo {
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

func uploadPackage(config *Config, packageType, distro string, version *DistroVersion, f *os.File) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("package[package_file]", f.Name())

	if err != nil {
		return err
	}

	hash := sha256.New()
	_, err = io.Copy(io.MultiWriter(part, hash), f)

	if err != nil {
		return err
	}

	err = writer.WriteField("package[distro_version_id]", strconv.Itoa(version.ID))

	if err != nil {
		return err
	}

	err = writer.Close()

	if err != nil {
		return err
	}

	const maxAttempts = 3

	for attempt := 1; ; attempt++ {
		status, response, err := postPackage(config, writer.FormDataContentType(), body.Bytes())

		if err != nil {
			return err
		}

		if status == http.StatusCreated {
			return nil
		}

		isTimeout := status == http.StatusGatewayTimeout
		var filename string
		var isDuplicate bool

		if status == http.StatusUnprocessableEntity {
			filename, isDuplicate = duplicatePackage(response)
		}

		if isTimeout || isDuplicate {
			matches, err := packageMatches(config, packageType, distro, version.IndexName, filename, hex.EncodeToString(hash.Sum(nil)))

			if err != nil {
				return fmt.Errorf("checking package after upload returned HTTP %d: %w", status, err)
			}

			if matches {
				log.Printf("Package already uploaded to %s/%s (SHA-256 verified)\n", distro, version.IndexName)

				return nil
			}

			if isTimeout && attempt < maxAttempts {
				log.Printf("Upload to %s/%s timed out, retrying (%d/%d)\n", distro, version.IndexName, attempt+1, maxAttempts)
				time.Sleep(time.Duration(attempt) * time.Second)

				continue
			}
		}

		if isTimeout {
			return fmt.Errorf("upload returned HTTP %d after %d attempts", status, attempt)
		}

		return fmt.Errorf("unexpected status code: %d, body: %s", status, response)
	}
}

func postPackage(config *Config, contentType string, body []byte) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodPost, PackagecloudAPIURL+"/api/v1/repos/"+config.PackagecloudUser+"/"+config.PackagecloudRepo+"/packages.json", bytes.NewReader(body))

	if err != nil {
		return 0, nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(config.PackagecloudAPIKey, "")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return 0, nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// The 504 status is enough to verify whether the upload landed.
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusGatewayTimeout {
		return resp.StatusCode, nil, nil
	}

	response, err := io.ReadAll(resp.Body)

	return resp.StatusCode, response, err
}

func duplicatePackage(response []byte) (string, bool) {
	var validationErrors map[string]json.RawMessage

	if err := json.Unmarshal(response, &validationErrors); err != nil || len(validationErrors) != 1 {
		return "", false
	}

	filenameError := validationErrors["filename"]
	var message string

	if err := json.Unmarshal(filenameError, &message); err != nil {
		var messages []string

		if err := json.Unmarshal(filenameError, &messages); err != nil || len(messages) != 1 {
			return "", false
		}

		message = messages[0]
	}

	const duplicateMessage = "has already been taken"

	if message == duplicateMessage {
		return "", true
	}

	filename, ok := strings.CutSuffix(message, "' "+duplicateMessage)

	if !ok {
		return "", false
	}

	filename, _ = strings.CutPrefix(filename, "'")

	return filename, filename != ""
}

func packageMatches(config *Config, packageType, distro, version, filename, checksum string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	groups, err := packagecloudPages[PackageVersionGroup](ctx, config, versionGroupsPath(config, packageType, distro, version))

	if err != nil {
		return false, err
	}

	for _, group := range groups {
		packages, err := packagecloudPages[PackageVersion](ctx, config, group.VersionsURL)

		if err != nil {
			return false, err
		}

		for _, pkg := range packages {
			if filename != "" && filename != pkg.Filename {
				continue
			}

			var metadata struct {
				SHA256 string `json:"sha256sum"`
			}
			_, err := getPackagecloudJSON(ctx, config, pkg.PackageURL, &metadata)

			if err != nil {
				return false, err
			}

			if strings.EqualFold(checksum, metadata.SHA256) {
				return true, nil
			}
		}
	}

	return false, nil
}

func getPackagecloudJSON(ctx context.Context, config *Config, path string, result any) (string, error) {
	if !strings.HasPrefix(path, "/api/v1/") {
		return "", fmt.Errorf("unexpected Packagecloud API path: %q", path)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, PackagecloudAPIURL+path, nil)

	if err != nil {
		return "", err
	}

	req.SetBasicAuth(config.PackagecloudAPIKey, "")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)

		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, b)
	}

	return resp.Header.Get("Link"), json.NewDecoder(resp.Body).Decode(result)
}

func packagecloudPages[T any](ctx context.Context, config *Config, path string) ([]T, error) {
	var all []T

	for path != "" {
		var page []T
		links, err := getPackagecloudJSON(ctx, config, path, &page)

		if err != nil {
			return nil, err
		}

		all = append(all, page...)
		path = ""

		for link := range strings.SplitSeq(links, ",") {
			target, found := strings.CutSuffix(strings.TrimSpace(link), `>; rel="next"`)

			if !found {
				continue
			}

			next, ok := strings.CutPrefix(target, "<"+PackagecloudAPIURL)

			if !ok || !strings.HasPrefix(next, "/api/v1/") {
				return nil, fmt.Errorf("unexpected Packagecloud pagination link: %q", link)
			}

			path = next

			break
		}
	}

	return all, nil
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
	req, err := http.NewRequest(http.MethodGet, PackagecloudAPIURL+"/api/v1/distributions.json", nil)

	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(config.PackagecloudAPIKey, "")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(b))
	}

	distros := &Distros{}

	if err := json.NewDecoder(resp.Body).Decode(distros); err != nil {
		return nil, err
	}

	return distros, nil
}

type PackageVersion struct {
	CreatedAt  time.Time `json:"created_at"`
	DestroyURL string    `json:"destroy_url"`
	Version    string    `json:"version"`
	Filename   string    `json:"filename"`
	PackageURL string    `json:"package_url"`
}

type PackageVersionGroup struct {
	VersionsURL string `json:"versions_url"`
}

func versionGroupsPath(config *Config, packageType, distro, version string) string {
	return "/api/v1/repos/" + config.PackagecloudUser + "/" + config.PackagecloudRepo + "/packages/" + packageType + "/" + distro + "/" + version + ".json"
}

func deleteVersion(config *Config, url string) error {
	req, err := http.NewRequest(http.MethodDelete, PackagecloudAPIURL+url, nil)

	if err != nil {
		return err
	}

	req.SetBasicAuth(config.PackagecloudAPIKey, "")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(b))
	}

	return nil
}
