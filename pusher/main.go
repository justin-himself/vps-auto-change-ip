package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config structure for YAML configuration
type Config struct {
	APIKey       string `yaml:"apiToken"`
	KVDBURL      string `yaml:"kvDBURL"`
	FetchInterval int    `yaml:"fetchInterval"`
	UpstreamURL  string `yaml:"upstreamURL"`
}

// LoadConfig reads and unmarshals the YAML configuration file.
func LoadConfig(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// LoadKeys loads keys from a file, each key on a new line.
func LoadKeys(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var keys []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		keys = append(keys, scanner.Text())
	}
	return keys, scanner.Err()
}

// FetchContent makes a GET request to the provided URL and returns the response body as a string.
func FetchContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// PushToKVDB sends a POST request to the KV database to store the fetched content.
func PushToKVDB(apiToken, kvDBURL, key, value string) error {
	url := fmt.Sprintf("%s/%s", kvDBURL, key)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(value)))
	if err != nil {
		return err
	}

	request.Header.Set("X-API-Key", apiToken)
	request.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to push data, status code: %d", response.StatusCode)
	}
	return nil
}

func main() {
	config, err := LoadConfig("/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	keys, err := LoadKeys("/config/keys.txt")
	if err != nil {
		log.Fatalf("Failed to load keys: %v", err)
	}

	log.Println("Starting to process keys...")
	for {
		for _, key := range keys {
			fetchURL := fmt.Sprintf("%s/%s", config.UpstreamURL, key)
			content, err := FetchContent(fetchURL)
			if err != nil {
				log.Printf("Failed to fetch content from %s: %v", fetchURL, err)
				continue
			}

			err = PushToKVDB(config.APIKey, config.KVDBURL, key, content)
			if err != nil {
				log.Printf("Failed to push content to KV DB for %s: %v", key, err)
				continue
			}
			log.Printf("Successfully pushed content for %s", key)
		}

		log.Printf("Completed one cycle of processing keys. Waiting %d minute(s) for next cycle...", config.FetchInterval)
		time.Sleep(time.Duration(config.FetchInterval) * time.Minute)
	}
}

