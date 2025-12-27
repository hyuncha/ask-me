package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"cleaners-ai/pkg/textextractor"
)

type TextExtractionHandler struct {
	extractor *textextractor.TextExtractor
}

func NewTextExtractionHandler() *TextExtractionHandler {
	return &TextExtractionHandler{
		extractor: textextractor.NewTextExtractor(),
	}
}

type ExtractTextResponse struct {
	Text     string `json:"text"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
}

// ExtractText handles POST /api/extract-text
func (h *TextExtractionHandler) ExtractText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(50 << 20) // 50MB max
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract text based on file type
	var extractedText string
	var fileType string

	switch ext {
	case ".txt", ".md":
		extractedText = string(fileBytes)
		fileType = "text"
	case ".pdf":
		extractedText, err = h.extractor.ExtractFromPDF(fileBytes)
		if err != nil {
			http.Error(w, "Failed to extract text from PDF: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fileType = "pdf"
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		// For images, we could use OCR in the future
		// For now, just return a placeholder
		extractedText = fmt.Sprintf("[이미지 파일: %s]\n", header.Filename)
		fileType = "image"
	default:
		http.Error(w, "Unsupported file type: "+ext, http.StatusBadRequest)
		return
	}

	response := ExtractTextResponse{
		Text:     extractedText,
		FileName: header.Filename,
		FileType: fileType,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
