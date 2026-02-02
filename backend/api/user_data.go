package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// UserDataHandler handles user data folder operations
type UserDataHandler struct {
	dataFolderPath string
}

// NewUserDataHandler creates a new user data handler
func NewUserDataHandler() *UserDataHandler {
	// Default data folder path (can be configured via environment variable)
	dataPath := os.Getenv("USER_DATA_PATH")
	if dataPath == "" {
		dataPath = "./user-data"
	}

	// Ensure data folder exists
	ensureDataFolderStructure(dataPath)

	return &UserDataHandler{
		dataFolderPath: dataPath,
	}
}

// ensureDataFolderStructure creates the required folder structure
func ensureDataFolderStructure(basePath string) {
	folders := []string{
		"profiles",
		"templates",
		"logs",
		"cache",
		"indicators",
		"drawings",
		"workspaces",
	}

	for _, folder := range folders {
		fullPath := filepath.Join(basePath, folder)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			fmt.Printf("[UserData] Failed to create folder %s: %v\n", fullPath, err)
		}
	}
}

// FileEntry represents a file or directory
type FileEntry struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Type     string    `json:"type"` // "file" or "directory"
	Size     int64     `json:"size,omitempty"`
	Modified time.Time `json:"modified,omitempty"`
}

// HandleGetDataFolderPath returns the data folder path
func (h *UserDataHandler) HandleGetDataFolderPath(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	absPath, err := filepath.Abs(h.dataFolderPath)
	if err != nil {
		http.Error(w, "Failed to get absolute path", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"path":    absPath,
		"success": true,
	}

	json.NewEncoder(w).Encode(response)
}

// HandleListFiles lists files in data folder or subfolder
func (h *UserDataHandler) HandleListFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get path from query parameter
	targetPath := r.URL.Query().Get("path")
	if targetPath == "" {
		targetPath = h.dataFolderPath
	}

	// Security: Ensure path is within data folder
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	absBase, _ := filepath.Abs(h.dataFolderPath)
	if !filepath.HasPrefix(absTarget, absBase) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// List files
	entries, err := os.ReadDir(absTarget)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	files := make([]FileEntry, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileType := "file"
		if entry.IsDir() {
			fileType = "directory"
		}

		files = append(files, FileEntry{
			Name:     entry.Name(),
			Path:     filepath.Join(absTarget, entry.Name()),
			Type:     fileType,
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
	}

	response := map[string]interface{}{
		"files":   files,
		"success": true,
	}

	json.NewEncoder(w).Encode(response)
}

// HandleDownloadFile downloads a file from data folder
func (h *UserDataHandler) HandleDownloadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get file path from query parameter
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "Missing file path", http.StatusBadRequest)
		return
	}

	// Security: Ensure path is within data folder
	absFile, err := filepath.Abs(filePath)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	absBase, _ := filepath.Abs(h.dataFolderPath)
	if !filepath.HasPrefix(absFile, absBase) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Check if file exists
	info, err := os.Stat(absFile)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if info.IsDir() {
		http.Error(w, "Cannot download directory", http.StatusBadRequest)
		return
	}

	// Open file
	file, err := os.Open(absFile)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(absFile)))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

	// Stream file to response
	io.Copy(w, file)
}

// HandleUploadFile uploads a file to data folder
func (h *UserDataHandler) HandleUploadFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 100MB)
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get target path
	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = h.dataFolderPath
	}

	// Security: Ensure path is within data folder
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	absBase, _ := filepath.Abs(h.dataFolderPath)
	if !filepath.HasPrefix(absTarget, absBase) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Ensure target directory exists
	if err := os.MkdirAll(absTarget, 0755); err != nil {
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	// Create destination file
	destPath := filepath.Join(absTarget, header.Filename)
	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, file)
	if err != nil {
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"path":    destPath,
		"message": "File uploaded successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// HandleDeleteFile deletes a file from data folder
func (h *UserDataHandler) HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		http.Error(w, "Missing file path", http.StatusBadRequest)
		return
	}

	// Security: Ensure path is within data folder
	absFile, err := filepath.Abs(req.Path)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	absBase, _ := filepath.Abs(h.dataFolderPath)
	if !filepath.HasPrefix(absFile, absBase) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Delete file or directory
	err = os.RemoveAll(absFile)
	if err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "File deleted successfully",
	}

	json.NewEncoder(w).Encode(response)
}

// HandleCreateDirectory creates a directory in data folder
func (h *UserDataHandler) HandleCreateDirectory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		http.Error(w, "Missing directory path", http.StatusBadRequest)
		return
	}

	// Security: Ensure path is within data folder
	absDir, err := filepath.Abs(req.Path)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	absBase, _ := filepath.Abs(h.dataFolderPath)
	if !filepath.HasPrefix(absDir, absBase) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Create directory
	err = os.MkdirAll(absDir, 0755)
	if err != nil {
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"path":    absDir,
		"message": "Directory created successfully",
	}

	json.NewEncoder(w).Encode(response)
}
