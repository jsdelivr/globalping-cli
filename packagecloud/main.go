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
)

const (
	PackagecloudAPIURL = "https://packagecloud.io/api/v1"
)

type Config struct {
	PackagecloudUser   string
	PackagecloudRepo   string
	PackagecloudAPIKey string
}

func main() {
	file := os.Args[1]
	dist := os.Args[2]
	if file == "" || dist == "" {
		fmt.Println("Usage: upload <path> <dist>")
		os.Exit(1)
	}
	config := &Config{
		PackagecloudUser:   os.Getenv("PACKAGECLOUD_USER"),
		PackagecloudRepo:   os.Getenv("PACKAGECLOUD_REPO"),
		PackagecloudAPIKey: os.Getenv("PACKAGECLOUD_APIKEY"),
	}
	log.Println("Fetching distros")
	distros, err := fetchDistros(config)
	if err != nil {
		log.Println("Failed to fetch distros: ", err)
		os.Exit(1)
	}
	log.Println("Opening file")
	f, err := os.Open(file)
	if err != nil {
		log.Println("Failed to open file: ", err)
		os.Exit(1)
	}
	defer f.Close()
	switch dist {
	case "deb":
		for i := range distros.Deb {
			l := len(distros.Deb[i].Versions)
			for j := l - 1; j >= 0 && l-j <= 10; j-- {
				log.Printf("Uploading to distro: %s, version: %d\n", distros.Deb[i].IndexName, distros.Deb[i].Versions[j].ID)
				err := uploadPackage(config, strconv.Itoa(distros.Deb[i].Versions[j].ID), f)
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
		log.Println("DEB package uploaded")
	case "rpm":
		for i := range distros.Rpm {
			l := len(distros.Rpm[i].Versions)
			for j := l - 1; j >= 0 && l-j <= 10; j-- {
				log.Printf("Uploading to distro: %s, version: %d\n", distros.Rpm[i].IndexName, distros.Rpm[i].Versions[j].ID)
				err := uploadPackage(config, strconv.Itoa(distros.Rpm[i].Versions[j].ID), f)
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
		log.Println("RPM package uploaded")
	default:
		fmt.Println("Unknown distro")
		os.Exit(1)
	}
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
	req, err := http.NewRequest("POST", PackagecloudAPIURL+"/repos/"+config.PackagecloudUser+"/"+config.PackagecloudRepo+"/packages.json", &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth(config.PackagecloudAPIKey, "")
	client := &http.Client{}
	resp, err := client.Do(req)
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
	ID int `json:"id"`
}

type Distro struct {
	IndexName string `json:"index_name"`
	Versions  []DistroVersion
}

type Distros struct {
	Deb []Distro `json:"deb"`
	Rpm []Distro `json:"rpm"`
}

func fetchDistros(config *Config) (*Distros, error) {
	req, err := http.NewRequest("GET", PackagecloudAPIURL+"/distributions.json", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(config.PackagecloudAPIKey, "")
	client := &http.Client{}
	resp, err := client.Do(req)
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
