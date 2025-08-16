package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	ROOT_DIR    = "./srv"
	LISTEN_ADDR = ":8080"
)

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
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>File Browser</title>
    <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, sans-serif;
            margin: 0;
            padding: 20px;
            background: #f8f9fa;
            color: #333;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: #343a40;
            color: white;
            padding: 15px 20px;
            border-bottom: 1px solid #ddd;
        }
        .path {
            font-family: 'JetBrains Mono', monospace;
            font-size: 14px;
            color: #6c757d;
        }
        .content {
            padding: 20px;
        }
        .file-list {
            list-style: none;
            padding: 0;
            margin: 0;
        }
        .file-item {
            display: flex;
            align-items: center;
            padding: 8px 12px;
            border-bottom: 1px solid #eee;
            text-decoration: none;
            color: #333;
            transition: background 0.15s;
        }
        .file-item:hover {
            background: #f8f9fa;
        }
        .file-icon {
            margin-right: 8px;
            width: 16px;
            text-align: center;
        }
        .file-name {
            flex: 1;
            font-family: 'JetBrains Mono', monospace;
            font-size: 14px;
        }
        .file-size {
            color: #6c757d;
            font-size: 12px;
            font-family: 'JetBrains Mono', monospace;
        }
        .code-content {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 4px;
            padding: 20px;
            font-family: 'JetBrains Mono', monospace;
            font-size: 13px;
            line-height: 1.5;
            white-space: pre-wrap;
            overflow-x: auto;
            color: #333;
        }
        .back-button {
            display: inline-block;
            padding: 8px 16px;
            background: #6c757d;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            font-size: 14px;
            margin-bottom: 15px;
        }
        .back-button:hover {
            background: #5a6268;
        }
        .error {
            color: #dc3545;
            background: #f8d7da;
            padding: 10px;
            border-radius: 4px;
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>File Browser</h2>
            <div class="path">{{.CurrentPath}}</div>
        </div>
        <div class="content">
            {{if .IsFile}}
                {{if ne .CurrentPath ""}}
                    <a href="/browse{{.CurrentPath | dirname}}" class="back-button">← Back</a>
                {{end}}
                <h3>{{.FileName}}</h3>
                <div class="code-content">{{.Content}}</div>
            {{else}}
                {{if ne .CurrentPath ""}}
                    <a href="/browse{{.CurrentPath | dirname}}" class="back-button">← Back</a>
                {{end}}
                <ul class="file-list">
                    {{range .Files}}
                    <li>
                        <a href="/browse{{.Path}}" class="file-item">
                            <span class="file-icon">{{if .IsDir}}📁{{else}}📄{{end}}</span>
                            <span class="file-name">{{.Name}}</span>
                            {{if not .IsDir}}
                                <span class="file-size">{{.Size}} bytes</span>
                            {{end}}
                        </a>
                    </li>
                    {{end}}
                </ul>
            {{end}}
        </div>
    </div>
</body>
</html>
`

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
		}

		if stat.IsDir() {
			// List directory contents
			files, err := os.ReadDir(cleanPath)
			if err != nil {
				http.Error(w, "Cannot read directory", http.StatusInternalServerError)
				return
			}

			for _, file := range files {
				relativePath := filepath.Join(urlPath, file.Name())
				info, err := file.Info()
				if err != nil {
					http.Error(w, "Cannot stat file", http.StatusInternalServerError)
					return
				}
				data.Files = append(data.Files, FileInfo{
					Name:  file.Name(),
					IsDir: file.IsDir(),
					Size:  info.Size(),
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
