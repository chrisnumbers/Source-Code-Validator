package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

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

func ParseGitHubURLRegex(rawURL string) (user, repo, branch string, err error) {
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

func GetFileContent(urls []string) ([]string, error) {
	var results []string

	for _, url := range urls {
		response, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				fmt.Println("Error closing response body:", err)
				return
			}
		}(response.Body)

		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		results = append(results, string(bodyBytes))
	}
	return results, nil

}

func ListFiles(ownerRepo, branch, path string, fileContents *[]string) (*[]string, error) {
	contents, err := fetchContents(ownerRepo, path)
	if err != nil {
		return fileContents, err
	}

	for _, item := range contents {
		switch item.Type {
		case "file":
			rawURL := fmt.Sprintf(rawBaseURL, ownerRepo, branch, item.Path)
			// Check if the file is allowed based on its extension
			if isAllowedFile(rawURL) {
				*fileContents = append(*fileContents, rawURL)
			}

		case "dir":
			if _, err := ListFiles(ownerRepo, branch, item.Path, fileContents); err != nil {
				return fileContents, err
			}

		}
	}

	return fileContents, nil
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
			return
		}
	}(resp.Body)

	var contents []GitHubContent
	err = json.NewDecoder(resp.Body).Decode(&contents)
	return contents, err
}

var allowedExtensions = []string{
	// Code
	".go", ".py", ".js", ".ts", ".java", ".c", ".cpp", ".cs", ".rb", ".rs", ".php",

	// Web
	".html", ".htm", ".css", ".scss", ".sass", ".json", ".xml",

	// Markup & Config
	".md", ".txt", ".sh", ".yml", ".yaml", ".ini", ".cfg", ".toml", ".env", ".dockerfile",

	// Build/Project
	".gradle", ".makefile", ".mk", ".bat", ".ps1",

	// Templates
	".ejs", ".hbs", ".pug", ".jinja", ".njk",
}

var allowedFileNames = []string{"dockerfile", "makefile", "readme.md", "license", "changelog.md"}

func isAllowedFile(path string) bool {
	lower := strings.ToLower(filepath.Base(path))
	for _, name := range allowedFileNames {
		if lower == name {
			return true
		}
	}
	for _, ext := range allowedExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}
