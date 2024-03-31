package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
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
			kv.store[parts[0]] = parts[1]
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
		fmt.Fprintf(file, "%s,%s\n", key, value)
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
			kv.store[parts[0]] = parts[1]
		}
	}

	kv.saveToFile()
}

func loadAPIKeys() {
	apiKeys = make(map[string]bool)
	file, err := os.Open("api_keys.txt")
	if err != nil {
		panic(fmt.Sprintf("Failed to load API keys: %v", err))
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
		panic("No API keys loaded; ensure api_keys.txt is not empty.")
	}
}

func authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow GET requests without API key except for the root path
		if r.Method != "GET" || r.URL.Path == "/db.txt" {
			apiKeyHeader := r.Header.Get("X-API-Key")
			if _, exists := apiKeys[apiKeyHeader]; !exists {
				http.Error(w, "Unauthorized: API key is invalid or missing", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}


func serveReadme(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("readme.txt")
	if err != nil {
		http.Error(w, "Failed to read readme.txt", http.StatusInternalServerError)
		return
	}
	w.Write(content)
}

func serveDb(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("db.txt")
	if err != nil {
		http.Error(w, "Failed to read db.txt", http.StatusInternalServerError)
		return
	}
	w.Write(content)
}


func main() {
	loadAPIKeys()
	kv := NewKVStore()
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.Method == "GET" {
			serveReadme(w, r)
			return
		}

		if r.URL.Path == "/db.txt" && r.Method == "GET" {
			serveDb(w, r)
			return
		}

		key := strings.TrimPrefix(r.URL.Path, "/")
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
	}

	authenticatedHandler := authenticate(handler)
	http.HandleFunc("/", authenticatedHandler)

	fmt.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}
}

