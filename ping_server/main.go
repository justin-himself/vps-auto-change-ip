package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/exec"
    "sync"
    "time"
)

var validAPIKeys map[string]struct{}
var cache = make(map[string]CacheEntry)
var mutex sync.Mutex

type Config struct {
    Ping struct {
        AutoPingInterval        time.Duration `yaml:"autoPingInterval"`
        CacheValidDuration time.Duration `yaml:"cacheValidDuration"`
        CacheTTL   time.Duration `yaml:"cacheTTL"`
    } `yaml:"ping"`
    API struct {
        Port int `yaml:"port"`
    } `yaml:"api"`
}

var config Config

type CacheEntry struct {
    Result         bool
    LastRequested  time.Time
    LastAutoPinged time.Time
}

type PingResponse struct {
    IP     string `json:"ip"`
    Alive  bool   `json:"alive"`
    Cached bool   `json:"cached"`
}

type DebugResponse struct {
    Config       Config                `json:"config"`
    Cache        map[string]CacheEntry `json:"cache"`
    ValidAPIKeys []string              `json:"validAPIKeys"`
}

func loadConfig(path string) error {
    bytes, err := ioutil.ReadFile(path)
    if err != nil {
        return err
    }
    err = yaml.Unmarshal(bytes, &config)
    return err
}

func loadAPIKeys(filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    validAPIKeys = make(map[string]struct{})
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        key := scanner.Text()
        validAPIKeys[key] = struct{}{}
    }

    return scanner.Err()
}

func isAPIKeyValid(key string) bool {
    _, exists := validAPIKeys[key]
    return exists
}

func pingAddress(ip string) bool {
    cmd := exec.Command("ping", "-c", "1", ip)
    err := cmd.Run()
    return err == nil
}

func pingAddresses(addresses []string, useCache bool) []PingResponse {
    var wg sync.WaitGroup
    results := make([]PingResponse, len(addresses))

    for i, addr := range addresses {
        wg.Add(1)
        go func(i int, addr string) {
            defer wg.Done()

            now := time.Now()
            cached := false
            result := false

            mutex.Lock()
            entry, exists := cache[addr]
            if useCache && exists && (now.Sub(entry.LastRequested) < config.Ping.CacheValidDuration || now.Sub(entry.LastAutoPinged) < config.Ping.CacheValidDuration) {
                result = entry.Result
                cached = true
            } else {
                result = pingAddress(addr)
                cache[addr] = CacheEntry{
                    Result:         result,
                    LastRequested:  now,
                    LastAutoPinged: now,
                }
            }
            mutex.Unlock()

            results[i] = PingResponse{
                IP:     addr,
                Alive:  result,
                Cached: cached,
            }
        }(i, addr)
    }

    wg.Wait()
    return results
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
    apiKey := r.URL.Query().Get("token")
    if !isAPIKeyValid(apiKey) {
        http.Error(w, "Invalid or missing API key", http.StatusUnauthorized)
        return
    }

    addresses := r.URL.Query()["addr"]
    if len(addresses) == 0 {
        http.Error(w, "No addresses provided", http.StatusBadRequest)
        return
    }

    useCache := true
    if cacheParam, ok := r.URL.Query()["cache"]; ok && cacheParam[0] == "false" {
        useCache = false
    }

    results := pingAddresses(addresses, useCache)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(results)
}

type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
    lrw.statusCode = code
    lrw.ResponseWriter.WriteHeader(code)
}

func logRequest(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        next.ServeHTTP(lrw, r)

        log.Printf("Method: %s, URI: %s, RemoteAddr: %s, StatusCode: %d, Duration: %s\n", r.Method, r.RequestURI, r.RemoteAddr, lrw.statusCode, time.Since(start))
    }
}

func autoPing() {
    ticker := time.NewTicker(config.Ping.AutoPingInterval)
    for range ticker.C {
        mutex.Lock()
        // Copy keys to avoid holding the lock while pinging
        ips := make([]string, 0, len(cache))
        for ip := range cache {
            ips = append(ips, ip)
        }
        mutex.Unlock()

        for _, ip := range ips {
            now := time.Now()

            // Perform the ping without holding the mutex
            result := pingAddress(ip)

            mutex.Lock()
            entry, exists := cache[ip]
            if exists {
                entry.Result = result
                entry.LastAutoPinged = now
                cache[ip] = entry

                log.Printf("Auto-pinged %s with result %t", ip, result)

                if now.Sub(entry.LastRequested) > config.Ping.CacheTTL {
                    log.Printf("Removing %s from cache due to inactivity", ip)
                    delete(cache, ip)
                }
            }
            mutex.Unlock()
        }
    }
}


func debugHandler(w http.ResponseWriter, r *http.Request) {
    token := r.URL.Query().Get("token")
    if !isAPIKeyValid(token) {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Convert validAPIKeys map to a slice of strings for JSON encoding
    var keys []string
    for key := range validAPIKeys {
        keys = append(keys, key)
    }

    response := DebugResponse{
        Config:       config,
        Cache:        cache,
        ValidAPIKeys: keys,
    }

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}

func main() {
    err := loadConfig("config.yaml")
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    if err := loadAPIKeys("api_tokens.txt"); err != nil {
        log.Fatalf("Failed to load API keys: %v", err)
    }

    go autoPing()

    http.HandleFunc("/ping", logRequest(pingHandler))
    http.HandleFunc("/debug", logRequest(debugHandler))


    log.Printf("Server starting on port %d...", config.API.Port)
    if err := http.ListenAndServe(fmt.Sprintf(":%d", config.API.Port), nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

