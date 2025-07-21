package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"

	"github.com/dotenv-org/godotenvvault"
	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
	"github.com/openai/openai-go"
)

func main() {
	// Load environment variables
	err := godotenvvault.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := gin.Default()
	router.GET("/validate", getValidateSource)

	router.Run("localhost:8080")
}

func isValidFile(fileHeader *multipart.FileHeader) (string, bool, error) {
	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		return "", false, err
	}
	defer file.Close()

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


func getValidateSource(c *gin.Context) {
	url := c.PostForm("url")
	fmt.Println(url)

	requirements, err := c.FormFile("requirements")
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "failed to validate"})
		return
	}

	fileType, isValid, err := isValidFile(requirements)
	if err != nil || !isValid {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Uploaded file is not a valid text or PDF file"})
		return
	}

	var requirementText string
	switch fileType {
	case "text/plain":
		// Handle text file
		file, err := requirements.Open()
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Uploaded file is not a valid text file"})
			return
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "failed to output text file content"})
			return
		}
		requirementText = string(content)
	
	case "application/pdf":
		var err error
		requirementText, err = extractTextFromPDF(requirements)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "failed to extract text from PDF"})
			return
		}
	
	default:
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Unsupported file type"})
		return
	}
	fmt.Println("Requirements text:", requirementText)


	user, repo, branch, err := parseGitHubURLRegex(url)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Error parsing GitHub URL"})
		return
	}
	ownerRepo := fmt.Sprintf("%s/%s", user, repo)
	fileContents, err := listFiles(ownerRepo, branch, "", &[]string{})
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error listing files"})
		return
	}

	fileBang, err := getFileContent(*fileContents)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "error getting file contents"})
		return
	}
	// c.IndentedJSON(http.StatusOK, fileBang)


	consultation, err := consultChatGPT(fileBang, requirementText)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "error consulting ChatGPT"})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": consultation})
	
}

const (
	apiURL     = "https://api.github.com/repos/%s/contents/%s"
	rawBaseURL = "https://raw.githubusercontent.com/%s/refs/heads/%s/%s" // Note: 'refs/heads/' is used to get the raw URL for the branch
)

type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "dir"
	DownloadURL string `json:"download_url"`
}

func fetchContents(ownerRepo, path string) ([]GitHubContent, error) {
	url := fmt.Sprintf(apiURL, ownerRepo, path)
	req, _ := http.NewRequest("GET", url, nil)

	// Optional: Add your GitHub token here to avoid rate-limiting
	// token := os.Getenv("GITHUB_TOKEN")
	// if token != "" {
	//     req.Header.Set("Authorization", "Bearer "+token)
	// }

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var contents []GitHubContent
	err = json.NewDecoder(resp.Body).Decode(&contents)
	return contents, err
}

func listFiles(ownerRepo, branch, path string, fileContents *[]string) (*[]string, error) {
	contents, err := fetchContents(ownerRepo, path)
	if err != nil {
		return fileContents, err
	}

	for _, item := range contents {
		switch item.Type {
		case "file":
			rawURL := fmt.Sprintf(rawBaseURL, ownerRepo, branch, item.Path)
			*fileContents = append(*fileContents, rawURL)
		case "dir":
			if _, err := listFiles(ownerRepo, branch, item.Path, fileContents); err != nil {
				return fileContents, err
			}

		}
	}

	return fileContents, nil
}

func parseGitHubURLRegex(rawURL string) (user, repo, branch string, err error) {
	// Regex pattern breakdown:
	// ^https://github.com/            → match GitHub base URL
	// ([^/]+)                         → capture group 1: username
	// /([^/]+)                        → capture group 2: repository name
	// (/tree/([^/]+))?                → optional group: '/tree/branchname', capture group 4: branch
	pattern := `^https://github\.com/([^/]+)/([^/]+)(/tree/([^/]+))?`

	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(rawURL)

	if len(matches) == 0 {
		return "", "", "", fmt.Errorf("invalid GitHub URL")
	}

	user = matches[1]
	repo = matches[2]

	if len(matches) >= 5 && matches[4] != "" {
		branch = matches[4]
	} else {
		branch = "main" // default if not explicitly included
	}

	return user, repo, branch, nil
}

func getFileContent(urls []string) ([]string, error) {
	results := []string{}

	for _, url := range urls {
		response, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		results = append(results, string(bodyBytes))
	}
	return results, nil

}

func consultChatGPT(content []string, requirements string) (string, error) {
	client := openai.NewClient()

	prompt := fmt.Sprintf(`
You are a software design analyst. Use the CLEAR framework (Context, Lens, Expectations, Analysis, Recommendations) to evaluate how well the following code meets the specified system design requirements.

--- SYSTEM DESIGN REQUIREMENTS ---
%s

--- CODE CONTENT ---
%s

Use the following format in your response:

**Context**: What is the context of this system and the intent behind the code?
**Lens**: What criteria or perspective are you using to evaluate the code against the requirements?
**Expectations**: What does the system design require in terms of architecture, logic, or design?
**Analysis**: How does the code satisfy (or fail to satisfy) each requirement? Provide reasoning and examples.
**Recommendations**: What can be improved in the code to better align with the design requirements?
`, requirements, content[1])

	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		return "Error getting ChatGPT response", nil
	}
	return chatCompletion.Choices[0].Message.Content, nil
}

func extractTextFromPDF(requirementsFile *multipart.FileHeader) (string, error) {
	// Open the PDF file
	file, err := requirementsFile.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()
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
