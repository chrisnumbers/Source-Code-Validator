package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"source-code-validator/server/internal/util"
)

func ValidateSourceCode(url string, requirements *multipart.FileHeader) (*string, error) {
	fileType, isValid, err := util.IsValidFile(requirements)
	if err != nil || !isValid {
		return nil, fmt.Errorf("invalid file type: %w", err)
	}

	var requirementText string
	switch fileType {
	case "text/plain":
		// Handle text file
		file, err := requirements.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open text file: %w", err)
		}
		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				return
			}
		}(file)

		content, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read text file: %w", err)
		}
		requirementText = string(content)

	case "application/pdf":
		var err error
		requirementText, err = util.ExtractTextFromPDF(requirements)
		if err != nil {
			return nil, fmt.Errorf("failed to extract text from PDF: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
	fmt.Println("Requirements text:", requirementText)

	user, repo, branch, err := util.ParseGitHubURLRegex(url)
	if err != nil {
		return nil, fmt.Errorf("error parsing GitHub URL: %w", err)
	}
	ownerRepo := fmt.Sprintf("%s/%s", user, repo)
	fileContents, err := util.ListFiles(ownerRepo, branch, "", &[]string{})
	if err != nil {
		return nil, fmt.Errorf("error listing files: %w", err)
	}

	fileBang, err := util.GetFileContent(*fileContents)
	if err != nil {
		return nil, fmt.Errorf("error getting file contents: %w", err)
	}
	// c.IndentedJSON(http.StatusOK, fileBang)

	consultation, err := util.ConsultChatGPT(fileBang, requirementText)
	if err != nil {
		return nil, fmt.Errorf("error consulting ChatGPT: %w", err)
	}
	return &consultation, nil
}
