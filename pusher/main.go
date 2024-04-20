package main

import (
    "bufio"
    "bytes"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strings"
    "time"

    "gopkg.in/yaml.v2"
)

type Config struct {
    APIKey        string `yaml:"apiToken"`
    KVDBURL       string `yaml:"kvDBURL"`
    UpstreamURL   string `yaml:"upstreamURL"`
    FetchInterval int    `yaml:"fetchInterval"`
}

type URLRequest struct {
    URL       string
    UserAgent string
}

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

func LoadKeys(filename string) ([]URLRequest, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var requests []URLRequest
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        parts := strings.Split(scanner.Text(), ",")
        if len(parts) == 2 {
            requests = append(requests, URLRequest{URL: parts[0], UserAgent: parts[1]})
        } else {
            log.Printf("Skipping invalid line: %s", scanner.Text())
        }
    }
    return requests, scanner.Err()
}

func FetchContent(url, userAgent string) (string, error) {
    client := &http.Client{}
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return "", err
    }
    req.Header.Set("User-Agent", userAgent)
    resp, err := client.Do(req)
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

func PushToKVDB(apiToken, kvDBURL, key, value, userAgent string) error {
    url := fmt.Sprintf("%s/%s", kvDBURL, key)
    request, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(value)))
    if err != nil {
        return err
    }
    request.Header.Set("X-API-Key", apiToken)
    request.Header.Set("Content-Type", "text/plain")
    request.Header.Set("User-Agent", userAgent)

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

    requests, err := LoadKeys("/config/keys.txt")
    if err != nil {
        log.Fatalf("Failed to load keys: %v", err)
    }

    log.Println("Starting to process keys...")
    for {
        for _, request := range requests {
            fullURL := fmt.Sprintf("%s/%s", config.UpstreamURL, request.URL)
            content, err := FetchContent(fullURL, request.UserAgent)
            if err != nil {
                log.Printf("Failed to fetch content from %s with UA %s: %v", fullURL, request.UserAgent, err)
                continue
            }

            err = PushToKVDB(config.APIKey, config.KVDBURL, request.URL, content, request.UserAgent)
            if err != nil {
                log.Printf("Failed to push content to KV DB for %s with UA %s: %v", request.URL, request.UserAgent, err)
                continue
            }
            log.Printf("Successfully pushed content for %s with UA %s", request.URL, request.UserAgent)
        }

        log.Printf("Completed one cycle of processing keys. Waiting %d minute(s) for next cycle...", config.FetchInterval)
        time.Sleep(time.Duration(config.FetchInterval) * time.Minute)
    }
}
