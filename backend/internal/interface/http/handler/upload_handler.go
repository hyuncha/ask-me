package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type UploadHandler struct {
	uploadDir string
}

func NewUploadHandler(uploadDir string) *UploadHandler {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %v", err))
	}
	return &UploadHandler{
		uploadDir: uploadDir,
	}
}

type UploadResponse struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	FileURL  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
}

// UploadFile handles file uploads
func (h *UploadHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.sendError(w, http.StatusBadRequest, "INVALID_FORM", "Invalid form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "NO_FILE", "No file uploaded")
		return
	}
	defer file.Close()

	// Validate file type (images only for now)
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		h.sendError(w, http.StatusBadRequest, "INVALID_FILE_TYPE", "Only image files are allowed")
		return
	}

	// Generate unique file ID
	fileID := uuid.New().String()
	ext := filepath.Ext(header.Filename)
	fileName := fmt.Sprintf("%s%s", fileID, ext)
	filePath := filepath.Join(h.uploadDir, fileName)

	// Create file
	dst, err := os.Create(filePath)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "FILE_CREATION_FAILED", "Failed to create file")
		return
	}
	defer dst.Close()

	// Copy uploaded file to destination
	fileSize, err := io.Copy(dst, file)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "FILE_WRITE_FAILED", "Failed to write file")
		return
	}

	response := UploadResponse{
		FileID:   fileID,
		FileName: header.Filename,
		FileURL:  fmt.Sprintf("/uploads/%s", fileName),
		FileSize: fileSize,
	}

	h.sendJSON(w, http.StatusOK, response)
}

// ServeFile serves uploaded files
func (h *UploadHandler) ServeFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	// Extract filename from path
	fileName := strings.TrimPrefix(r.URL.Path, "/uploads/")
	if fileName == "" {
		h.sendError(w, http.StatusBadRequest, "INVALID_FILE", "Invalid file name")
		return
	}

	filePath := filepath.Join(h.uploadDir, fileName)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.sendError(w, http.StatusNotFound, "FILE_NOT_FOUND", "File not found")
		return
	}

	// Serve file
	http.ServeFile(w, r, filePath)
}

func (h *UploadHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *UploadHandler) sendError(w http.ResponseWriter, status int, code, message string) {
	h.sendJSON(w, status, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}
	for _, t := range validTypes {
		if contentType == t {
			return true
		}
	}
	return false
}
