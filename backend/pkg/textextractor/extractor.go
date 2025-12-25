package textextractor

import (
	"bytes"
	"fmt"
	

	"github.com/ledongthuc/pdf"
)

type TextExtractor struct{}

func NewTextExtractor() *TextExtractor {
	return &TextExtractor{}
}

// ExtractFromPDF extracts text from PDF bytes
func (e *TextExtractor) ExtractFromPDF(pdfBytes []byte) (string, error) {
	reader := bytes.NewReader(pdfBytes)
	pdfReader, err := pdf.NewReader(reader, int64(len(pdfBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create PDF reader: %w", err)
	}

	var text bytes.Buffer
	numPages := pdfReader.NumPage()

	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page := pdfReader.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		// Extract text content from page
		content, err := page.GetPlainText(nil)
		if err != nil {
			// Continue to next page if this one fails
			continue
		}

		text.WriteString(content)
		text.WriteString("\n\n")
	}

	return text.String(), nil
}

// ExtractFromText reads plain text
func (e *TextExtractor) ExtractFromText(textBytes []byte) string {
	return string(textBytes)
}

// Future: Add OCR for images
// func (e *TextExtractor) ExtractFromImage(imageBytes []byte) (string, error) {
//     // Use Tesseract OCR or Google Vision API
//     return "", nil
// }
