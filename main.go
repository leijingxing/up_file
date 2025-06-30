// main.go

package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// **修改点 1**: 将 UPLOAD_ROOT_DIR 声明为包级变量，以便在所有函数中共享。
var UPLOAD_ROOT_DIR string

// sanitizeFilename ... (此函数无需修改)
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	return name
}

// aliveHandler ... (此函数无需修改)
func aliveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	io.WriteString(w, "file_upload.rs")
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	headerKey := r.Header.Get("HYZH-KEY")
	if headerKey != "" && headerKey != "HYzh221015" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Printf("Unauthorized access attempt with key: %s", headerKey)
		return
	}

	query := r.URL.Query()
	filename := query.Get("filename")
	proj := query.Get("proj")
	dir := query.Get("dir")

	if filename == "" || proj == "" || dir == "" {
		http.Error(w, "Missing query parameters: filename, proj, dir", http.StatusBadRequest)
		return
	}

	safeProj := sanitizeFilename(proj)
	safeDir := sanitizeFilename(dir)

	// **修改点 2**: 使用全局的 UPLOAD_ROOT_DIR 构建绝对路径。
	// 原来的代码: targetDir := filepath.Join(safeProj, safeDir)
	targetDir := filepath.Join(UPLOAD_ROOT_DIR, safeProj, safeDir)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		// 在日志中提供更详细的错误信息，这对于调试 launchd 服务至关重要
		log.Printf("Fatal: Error creating directory %s: %v. Check permissions and path.", targetDir, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("开始上传文件 '%s' 到 %s", filename, targetDir)

	reader, err := r.MultipartReader()
	if err != nil {
		log.Printf("Error getting multipart reader: %v", err)
		http.Error(w, "Error processing multipart form", http.StatusBadRequest)
		return
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading next multipart part: %v", err)
			continue
		}
		defer part.Close() // 确保每个 part 都被关闭

		safeFilename := sanitizeFilename(filename)
		filePath := filepath.Join(targetDir, safeFilename)

		dst, err := os.Create(filePath)
		if err != nil {
			log.Printf("Fatal: Error creating file %s: %v", filePath, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(dst, part); err != nil {
			log.Printf("Error writing to file %s: %v", filePath, err)
			dst.Close() // 立即关闭文件以避免句柄泄露
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		dst.Close() // 写入成功后关闭文件
		log.Printf("文件 '%s' 保存成功", filePath)
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "上传成功")
}

func latestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Missing query parameter: path", http.StatusBadRequest)
		return
	}

	// **修改点 3**: 同样，使用 UPLOAD_ROOT_DIR 来构建要读取的目录的绝对路径。
	// 原来的代码: safePath := sanitizeFilename(path)
	safePath := filepath.Join(UPLOAD_ROOT_DIR, sanitizeFilename(path))

	entries, err := os.ReadDir(safePath)
	if err != nil {
		log.Printf("Error reading directory %s: %v", safePath, err)
		if os.IsNotExist(err) {
			http.Error(w, "Directory not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	var latestTime uint64 = 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			log.Printf("Could not get file info for %s: %v", entry.Name(), err)
			continue
		}
		modifiedTime := info.ModTime()
		elapsed := time.Since(modifiedTime)
		timeInSecs := uint64(elapsed.Seconds())

		if timeInSecs > latestTime {
			latestTime = timeInSecs
		}
	}

	responseBody := strconv.FormatUint(latestTime, 10)
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, responseBody)
}

func main() {
	// **修改点 4**: 初始化我们之前声明的包级变量 UPLOAD_ROOT_DIR。
	// 这段逻辑本身是正确的，只是之前没有被正确使用。
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Fatal: Failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	UPLOAD_ROOT_DIR = os.Getenv("UPLOAD_DIR")
	if UPLOAD_ROOT_DIR == "" {
		log.Println("UPLOAD_DIR environment variable not set, using default 'uploads' directory relative to the executable.")
		UPLOAD_ROOT_DIR = filepath.Join(exeDir, "uploads")
	} else {
		log.Printf("Using custom upload directory from UPLOAD_DIR env: %s", UPLOAD_ROOT_DIR)
	}
    // 确保根上传目录存在
    if err := os.MkdirAll(UPLOAD_ROOT_DIR, 0755); err != nil {
        log.Fatalf("Fatal: Failed to create root upload directory %s: %v", UPLOAD_ROOT_DIR, err)
    }

    log.Printf("Root upload directory is set to: %s", UPLOAD_ROOT_DIR)

	mux := http.NewServeMux()
	mux.HandleFunc("/", aliveHandler)
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/latest", latestHandler)

	addr := "0.0.0.0:8080"
	log.Printf("Starting server on %s", addr)

	// 注意：这里的 err 变量与上面的 err 是不同的作用域，没问题。
	serverErr := http.ListenAndServe(addr, mux)
	if serverErr != nil {
		// 使用 Fatalf 可以确保在服务器启动失败时，程序会带着错误信息退出。
		log.Fatalf("Fatal: Server failed to start: %v", serverErr)
	}
}