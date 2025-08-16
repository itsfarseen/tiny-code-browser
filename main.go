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
	LISTEN_ADDR = "0.0.0.0:8080"
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

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>File Browser</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: #f8f9fa;
            --container-bg: white;
            --text-color: #333;
            --header-bg: #343a40;
            --header-text: white;
            --border-color: #ddd;
            --hover-bg: #f8f9fa;
            --code-bg: #f8f9fa;
            --code-border: #e9ecef;
            --secondary-text: #6c757d;
            --button-bg: #6c757d;
            --button-hover: #5a6268;
            --font-size: 13px;
        }

        [data-theme="dark"] {
            --bg-color: #1a1a1a;
            --container-bg: #2d2d2d;
            --text-color: #e0e0e0;
            --header-bg: #1f1f1f;
            --header-text: #e0e0e0;
            --border-color: #404040;
            --hover-bg: #3a3a3a;
            --code-bg: #1e1e1e;
            --code-border: #404040;
            --secondary-text: #a0a0a0;
            --button-bg: #555;
            --button-hover: #666;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, sans-serif;
            margin: 0;
            padding: 0;
            background: var(--bg-color);
            color: var(--text-color);
            transition: background 0.3s, color 0.3s;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: var(--container-bg);
            overflow: hidden;
            min-height: 100vh;
            transition: background 0.3s;
        }
        .header {
            background: var(--header-bg);
            color: var(--header-text);
            padding: 15px 20px;
            border-bottom: 1px solid var(--border-color);
            display: flex;
            justify-content: space-between;
            align-items: center;
            transition: background 0.3s, border-color 0.3s;
        }
        .header-left {
            display: flex;
            align-items: center;
            gap: 15px;
        }
        .header-left h2 {
            margin: 0;
        }
        .header-right {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .path {
            font-family: 'JetBrains Mono', monospace;
            font-size: 14px;
            color: var(--secondary-text);
            word-break: break-all;
            transition: color 0.3s;
        }
        .header-button {
            background: var(--button-bg);
            color: var(--header-text);
            border: none;
            padding: 8px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            text-decoration: none;
            display: inline-flex;
            align-items: center;
            transition: background 0.3s;
        }
        .header-button:hover {
            background: var(--button-hover);
        }
        .font-controls {
            display: flex;
            gap: 5px;
        }
        .font-btn {
            background: var(--button-bg);
            color: var(--header-text);
            border: none;
            padding: 6px 10px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 12px;
            transition: background 0.3s;
        }
        .font-btn:hover {
            background: var(--button-hover);
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
            padding: 12px 15px;
            border-bottom: 1px solid var(--border-color);
            text-decoration: none;
            color: var(--text-color);
            transition: background 0.15s, border-color 0.3s, color 0.3s;
            min-height: 44px;
        }
        .file-item:hover {
            background: var(--hover-bg);
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
            color: var(--secondary-text);
            font-size: 12px;
            font-family: 'JetBrains Mono', monospace;
            transition: color 0.3s;
        }
        .code-content {
            background: var(--code-bg);
            border-top: 1px solid var(--code-border);
            border-bottom: 1px solid var(--code-border);
            padding: 20px;
            font-family: 'JetBrains Mono', monospace;
            font-size: var(--font-size);
            line-height: 1.5;
            white-space: pre;
            overflow-x: auto;
            overflow-y: auto;
            color: var(--text-color);
            margin: 0 -20px;
            transition: background 0.3s, border-color 0.3s, color 0.3s, font-size 0.2s;
        }
        .back-button {
            display: inline-block;
            padding: 4px 16px;
            background: var(--button-bg);
            color: var(--header-text);
            text-decoration: none;
            border-radius: 4px;
            font-size: 14px;
            transition: background 0.3s;
            display: flex;
            align-items: center;
        }
        .back-button:hover {
            background: var(--button-hover);
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
            <div class="header-left">
                <h2>File Browser</h2>
                <div class="path">{{.CurrentPath}}</div>
                {{if ne .CurrentPath ""}}
                    <a href="/browse/{{.CurrentPath | dirname}}" class="back-button"><< Back</a>
                {{end}}
            </div>
            <div class="header-right">
                {{if .IsFile}}
                <div class="font-controls">
                    <button class="font-btn" onclick="changeFontSize(-1)">A-</button>
                    <button class="font-btn" onclick="changeFontSize(1)">A+</button>
                </div>
                {{end}}
                <button class="header-button" onclick="toggleTheme()">🌙</button>
            </div>
        </div>
        <div class="content">
            {{if .IsFile}}
                <h3>{{.FileName}}</h3>
                <div class="code-content">{{.Content}}</div>
            {{else}}
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

    <script>
        function toggleTheme() {
            const body = document.body;
            const button = document.querySelector('.header-button');
            
            if (body.getAttribute('data-theme') === 'dark') {
                body.removeAttribute('data-theme');
                button.textContent = '🌙';
                localStorage.setItem('theme', 'light');
            } else {
                body.setAttribute('data-theme', 'dark');
                button.textContent = '☀️';
                localStorage.setItem('theme', 'dark');
            }
        }

        function changeFontSize(delta) {
            const root = document.documentElement;
            const currentSize = parseInt(getComputedStyle(root).getPropertyValue('--font-size'));
            const newSize = Math.max(5, Math.min(32, currentSize + delta));
            root.style.setProperty('--font-size', newSize + 'px');
            localStorage.setItem('fontSize', newSize);
        }

        // Apply saved theme and font size on page load
        document.addEventListener('DOMContentLoaded', function() {
            const savedTheme = localStorage.getItem('theme');
            const button = document.querySelector('.header-button');
            
            if (savedTheme === 'dark') {
                document.body.setAttribute('data-theme', 'dark');
                button.textContent = '☀️';
            }

            const savedFontSize = localStorage.getItem('fontSize');
            if (savedFontSize) {
                document.documentElement.style.setProperty('--font-size', savedFontSize + 'px');
            }
        });
    </script>
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
