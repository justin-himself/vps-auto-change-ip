package main

import (
    "bufio"
    "encoding/base64"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "strings"
    "sync"
    "regexp"
)

var apiKeys map[string]bool

type KVStore struct {
    sync.RWMutex
    store map[string]string
}

func NewKVStore() *KVStore {
    kv := &KVStore{
        store: make(map[string]string),
    }
    kv.loadFromFile()
    return kv
}

func (kv *KVStore) loadFromFile() {
    file, err := os.Open("db.txt")
    if err != nil {
        if os.IsNotExist(err) {
            return
        }
        panic(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        parts := strings.SplitN(scanner.Text(), ",", 2)
        if len(parts) == 2 {
            decodedValue, err := base64.StdEncoding.DecodeString(parts[1])
            if err != nil {
                log.Printf("Error decoding Base64 data: %v", err)
                continue
            }
            kv.store[parts[0]] = string(decodedValue)
        }
    }
}

func (kv *KVStore) saveToFile() {
    file, err := os.Create("db.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    for key, value := range kv.store {
        encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
        fmt.Fprintf(file, "%s,%s\n", key, encodedValue)
    }
}

func (kv *KVStore) Set(key, value string) {
    kv.Lock()
    defer kv.Unlock()
    kv.store[key] = value
    kv.saveToFile()
}

func (kv *KVStore) Delete(key string) {
    kv.Lock()
    defer kv.Unlock()
    delete(kv.store, key)
    kv.saveToFile()
}

func (kv *KVStore) Get(key string) (string, bool) {
    kv.RLock()
    defer kv.RUnlock()
    value, exists := kv.store[key]
    return value, exists
}

func (kv *KVStore) ReplaceStore(newContents string) {
    kv.Lock()
    defer kv.Unlock()
    kv.store = make(map[string]string)

    scanner := bufio.NewScanner(strings.NewReader(newContents))
    for scanner.Scan() {
        parts := strings.SplitN(scanner.Text(), ",", 2)
        if len(parts) == 2 {
            decodedValue, err := base64.StdEncoding.DecodeString(parts[1])
            if err != nil {
                log.Printf("Error decoding Base64 data: %v", err)
                continue
            }
            kv.store[parts[0]] = string(decodedValue)
        }
    }

    kv.saveToFile()
}

func loadAPIKeys() {
    apiKeys = make(map[string]bool)
    file, err := os.Open("api_keys.txt")
    if err != nil {
        log.Fatalf("Failed to load API keys: %v", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        key := scanner.Text()
        if key != "" {
            apiKeys[key] = true
        }
    }

    if len(apiKeys) == 0 {
        log.Fatal("No API keys loaded; ensure api_keys.txt is not empty.")
    }
}

func authenticate(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "GET" {
            apiKeyHeader := r.Header.Get("X-API-Key")
            if _, exists := apiKeys[apiKeyHeader]; !exists {
                http.Error(w, "Unauthorized: API key is invalid or missing", http.StatusUnauthorized)
                return
            }
        }
        next(w, r)
    }
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Use the responseRecorder to capture the status code
        rec := newResponseRecorder(w)

        next(rec, r)

        // Log the method, URI, and status code in one line
        log.Printf("%s %s %d", r.Method, r.URL.RequestURI(), rec.statusCode)
    }
}

type responseRecorder struct {
    http.ResponseWriter
    statusCode int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
    // Default the status code to 200 for cases where WriteHeader is not called
    return &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
}

func (r *responseRecorder) WriteHeader(code int) {
    r.statusCode = code
    r.ResponseWriter.WriteHeader(code)
}

func serveReadme(w http.ResponseWriter, r *http.Request) {
    content, err := ioutil.ReadFile("readme.txt")
    if err != nil {
        http.Error(w, "Failed to read readme.txt", http.StatusInternalServerError)
        return
    }
    w.Write(content)
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if err != nil {
        return false
    }
    return !info.IsDir()
}

func main() {
    loadAPIKeys()
    kv := NewKVStore()

    authenticatedHandler := authenticate(loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
        allowedCharsRegex := regexp.MustCompile(`^[0-9a-zA-Z\$\-\_\.\+\!\*'\(\);\ / \?\:\@\=\&\ \ \"\<\>\#\%\{\}\|\ \\ \^\~\[\]]+$`)

        // Verify if the URL contains only allowed characters
        if !allowedCharsRegex.MatchString(r.URL.RequestURI()) {
            http.Error(w, "URL contains invalid characters", http.StatusBadRequest)
            return
        }
        if r.URL.RequestURI() == "/" && r.Method == "GET" {
            serveReadme(w, r)
            return
        }

        key := r.URL.RequestURI()
        switch r.Method {
        case "POST", "PUT":
            body, err := ioutil.ReadAll(r.Body)
            if err != nil {
                http.Error(w, "Failed to read request body", http.StatusBadRequest)
                return
            }
            if r.Method == "POST" {
                kv.Set(key, string(body))
                fmt.Fprintf(w, "Key %s set successfully", key)
            } else if r.Method == "PUT" {
                kv.ReplaceStore(string(body))
                fmt.Fprint(w, "Store replaced successfully")
            }
        case "DELETE":
            kv.Delete(key)
            fmt.Fprintf(w, "Key %s deleted successfully", key)
        case "GET":
            value, exists := kv.Get(key)
            if !exists {
                http.Error(w, "Key not found", http.StatusNotFound)
                return
            }
            fmt.Fprint(w, value)
        default:
            http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
        }
    }))
    http.HandleFunc("/", authenticatedHandler)

    // Check for SSL certificate and key
    certFile := "ssl/fullchain.pem"
    keyFile := "ssl/privkey.pem"

    if fileExists(certFile) && fileExists(keyFile) {
        // Both certificate and key files exist, start HTTPS server in a new goroutine
        go func() {
            log.Println("HTTPS server is running on https://localhost:8443")
            err := http.ListenAndServeTLS(":8443", certFile, keyFile, nil)
            if err != nil {
                log.Fatalf("Failed to start HTTPS server: %v", err)
            }
        }()
    } else {
        // Certificate or key file does not exist, log a warning
        log.Println("Warning: SSL certificate or key not found. HTTPS server will not start.")
    }

    // Always start HTTP server
    log.Println("HTTP server is running on http://localhost:8081")
    err := http.ListenAndServe(":8081", nil)
    if err != nil {
        log.Fatalf("Failed to start HTTP server: %v", err)
    }

    select {} // Keep the main goroutine running
}
