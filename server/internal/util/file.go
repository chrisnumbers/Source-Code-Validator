package util

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/ledongthuc/pdf"
)

func IsValidFile(fileHeader *multipart.FileHeader) (string, bool, error) {
	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		return "", false, err
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
			return
		}
	}(file)

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", false, err
	}

	// Detect MIME type
	contentType := http.DetectContentType(buffer)

	// Check MIME type and/or extension
	isText := strings.HasPrefix(contentType, "text/plain") ||
		strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".txt")
	isPDF := strings.HasPrefix(contentType, "application/pdf") ||
		strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".pdf")
	if isText || isPDF {
		return contentType, true, nil
	}
	return contentType, false, fmt.Errorf("file is not a valid text or PDF file")
}

func ExtractTextFromPDF(requirementsFile *multipart.FileHeader) (string, error) {
	// Open the PDF file
	file, err := requirementsFile.Open()
	if err != nil {
		return "", err
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
			return
		}
	}(file)
	// Read file into a buffer since pdf.NewReader needs a seeker
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Create a ReadSeeker from the byte slice
	readSeeker := bytes.NewReader(data)

	// Create a PDF reader
	pdfReader, err := pdf.NewReader(readSeeker, int64(len(data)))
	if err != nil {
		return "", err
	}

	// Extract text from all pages
	var textBuilder strings.Builder
	numPages := pdfReader.NumPage()
	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)
		pageText, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		textBuilder.WriteString(pageText)
	}
	rawText := textBuilder.String()
	cleanText := strings.Join(strings.Fields(rawText), " ")
	return cleanText, nil

}
