package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed template.html
var htmlTemplate string

var (
	ROOT_DIR    = getEnvOrDefault("ROOT_DIR", "./srv")
	LISTEN_ADDR = getEnvOrDefault("LISTEN_ADDR", "0.0.0.0:8080")
	APP_TITLE   = getEnvOrDefault("APP_TITLE", "Tiny Code Browser")
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

type FileInfo struct {
	Name  string
	IsDir bool
	Size  int64
	Path  string
}

type PageData struct {
	CurrentPath string
	Files       []FileInfo
	Content     string
	IsFile      bool
	FileName    string
	AppTitle    string
}

func main() {
	tmpl := template.Must(template.New("page").Funcs(template.FuncMap{
		"dirname": func(path string) string {
			dir := filepath.Dir(path)
			if dir == "." || dir == "/" {
				return ""
			}
			return dir
		},
	}).Parse(htmlTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/browse", http.StatusFound)
	})

	http.HandleFunc("/browse/", func(w http.ResponseWriter, r *http.Request) {
		// Extract relative path from URL
		urlPath := strings.TrimPrefix(r.URL.Path, "/browse")
		if urlPath == "" {
			urlPath = "/"
		}

		// Build absolute path within ROOT_DIR
		fullPath := filepath.Join(ROOT_DIR, urlPath)

		// Security check: ensure the resolved path is within ROOT_DIR
		cleanPath, err := filepath.Abs(fullPath)
		if err != nil {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		rootAbs, err := filepath.Abs(ROOT_DIR)
		if err != nil {
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		if !strings.HasPrefix(cleanPath, rootAbs) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		stat, err := os.Stat(cleanPath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		data := PageData{
			CurrentPath: urlPath,
			AppTitle:    APP_TITLE,
		}

		if stat.IsDir() {
			files, err := os.ReadDir(cleanPath)
			if err != nil {
				http.Error(w, "Cannot read directory", http.StatusInternalServerError)
				return
			}

			for _, file := range files {
				relativePath := filepath.Join(urlPath, file.Name())
				var size int64

				if !file.IsDir() {
					if info, err := file.Info(); err == nil {
						size = info.Size()
					}
				}

				data.Files = append(data.Files, FileInfo{
					Name:  file.Name(),
					IsDir: file.IsDir(),
					Size:  size,
					Path:  relativePath,
				})
			}
		} else {
			// Display file content
			content, err := os.ReadFile(cleanPath)
			if err != nil {
				http.Error(w, "Cannot read file", http.StatusInternalServerError)
				return
			}

			data.IsFile = true
			data.FileName = filepath.Base(cleanPath)
			data.Content = string(content)
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}
	})

	fmt.Printf("Server starting on http://localhost%s\n", LISTEN_ADDR)
	fmt.Printf("Serving files from: %s\n", ROOT_DIR)
	log.Fatal(http.ListenAndServe(LISTEN_ADDR, nil))
}
