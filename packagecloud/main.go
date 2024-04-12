package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	PackagecloudAPIURL = "https://packagecloud.io/api/v1/repos/"
	PackagecloudDEBAny = "35"
	PackagecloudRPMAny = "227"
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
	switch dist {
	case "deb":
		log.Printf("Uploading DEB package: %s\n", file)
		err := uploadPackage(config, PackagecloudDEBAny, file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		log.Println("DEB package uploaded")
	case "rpm":
		log.Printf("Uploading RPM package: %s\n", file)
		err := uploadPackage(config, PackagecloudRPMAny, file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		log.Println("RPM package uploaded")
	default:
		fmt.Println("Unknown distro")
		os.Exit(1)
	}
}

func uploadPackage(config *Config, distroVersionId string, path string) error {
	var body bytes.Buffer
	f, err := os.Open(path)
	if err != nil {
		return err
	}
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
	req, err := http.NewRequest("POST", PackagecloudAPIURL+config.PackagecloudUser+"/"+config.PackagecloudRepo+"/packages.json", &body)
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
