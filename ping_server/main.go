package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/exec"
    "strings"
    "sync"
    "time"
)

var validAPIKeys map[string]struct{}
var cache = make(map[string]CacheEntry)

type CacheEntry struct {
    Timestamp time.Time
    Result    bool
}

type PingResponse struct {
    IP     string `json:"ip"`
    Alive  bool   `json:"alive"`
    Cached bool   `json:"cached"`
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
    mutex := sync.Mutex{}

    for i, addr := range addresses {
        wg.Add(1)
        go func(i int, addr string) {
            defer wg.Done()

            cached := false
            result := false

            mutex.Lock()
            entry, exists := cache[addr]
            mutex.Unlock()

            if useCache && exists && time.Since(entry.Timestamp) < 5*time.Minute {
                result = entry.Result
                cached = true
            } else {
                result = pingAddress(addr)
                if useCache {
                    mutex.Lock()
                    cache[addr] = CacheEntry{Timestamp: time.Now(), Result: result}
                    mutex.Unlock()
                }
            }

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

    // Use cache by default; bypass cache if cache=false is specified
    useCache := true
    if cacheParam, ok := r.URL.Query()["cache"]; ok && len(cacheParam) > 0 {
        useCache = !(cacheParam[0] == "false")
    }

    results := pingAddresses(addresses, useCache)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(results)
}

func logRequest(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        next.ServeHTTP(lrw, r)

        duration := time.Since(start)
        logMessage := fmt.Sprintf("Method: %s, URI: %s, RemoteAddr: %s, StatusCode: %d, Duration: %s", r.Method, r.RequestURI, r.RemoteAddr, lrw.statusCode, duration)

        if token := r.URL.Query().Get("token"); token != "" && isAPIKeyValid(token) {
            redactedURI := strings.Replace(r.RequestURI, "token="+token, "token=[redacted]", 1)
            logMessage = fmt.Sprintf("Method: %s, URI: %s, RemoteAddr: %s, StatusCode: %d, Duration: %s", r.Method, redactedURI, r.RemoteAddr, lrw.statusCode, duration)
        }

        log.Println(logMessage)
    }
}

type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
    lrw.statusCode = code
    lrw.ResponseWriter.WriteHeader(code)
}

func main() {
    if err := loadAPIKeys("api_tokens.txt"); err != nil {
        log.Fatalf("Failed to load API keys: %v", err)
    }

    http.HandleFunc("/ping", logRequest(pingHandler))
    log.Println("Server starting on port 8080...")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

