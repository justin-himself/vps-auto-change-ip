package main

import (
    "bufio"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "os/exec"
    "sync"
    "time"
    "fmt"
    "strings"
)

var validAPIKeys map[string]struct{}

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

func pingAddressesInParallel(addresses []string) map[string]bool {
    var wg sync.WaitGroup
    results := make(map[string]bool)
    mutex := sync.Mutex{}

    for _, addr := range addresses {
        wg.Add(1)
        go func(addr string) {
            defer wg.Done()
            result := pingAddress(addr)
            mutex.Lock()
            results[addr] = result
            mutex.Unlock()
        }(addr)
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

    results := pingAddressesInParallel(addresses)
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

        duration := time.Since(start)
        logMessage := fmt.Sprintf("Method: %s, URI: %s, RemoteAddr: %s, StatusCode: %d, Duration: %s", r.Method, r.RequestURI, r.RemoteAddr, lrw.statusCode, duration)

        // Pattern match and redact token from the log message
        if token := r.URL.Query().Get("token"); token != "" && isAPIKeyValid(token) {
            redactedURI := strings.Replace(r.RequestURI, "token="+token, "token=[redacted]", 1)
            logMessage = fmt.Sprintf("Method: %s, URI: %s, RemoteAddr: %s, StatusCode: %d, Duration: %s", r.Method, redactedURI, r.RemoteAddr, lrw.statusCode, duration)
        }

        log.Println(logMessage)
    }
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

