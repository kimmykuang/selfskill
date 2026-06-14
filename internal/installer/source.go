package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimmykuang/selfskill/internal/frontmatter"
)

// FetchResult represents a fetched skill ready for installation.
type FetchResult struct {
	Name    string // skill name
	DirPath string // local directory path (for dir-based skills)
	IsDir   bool   // true if DirPath is set
	Content []byte // raw content (for single-file skills)
	Source  string // original source URL (set only for remote sources like URLSource)
}

// Source represents a skill source that can be fetched.
type Source interface {
	Fetch() ([]FetchResult, error)
}

// ResolveSource determines the source type from the input string.
func ResolveSource(input string) Source {
	return resolveSourceWithClient(input, &http.Client{})
}

func resolveSourceWithClient(input string, client *http.Client) Source {
	if isURL(input) {
		return &URLSource{url: resolveGitHubURL(input), origURL: input, httpClient: client}
	}

	info, err := os.Stat(input)
	if err != nil {
		// Treat as URL if can't stat
		return &URLSource{url: input, origURL: input, httpClient: client}
	}

	if !info.IsDir() {
		// Single .md file
		return &LocalFileSource{path: input}
	}

	// Check if it's a skill directory (contains SKILL.md)
	if _, err := os.Stat(filepath.Join(input, "SKILL.md")); err == nil {
		return &LocalDirSource{dir: input}
	}

	// Parent directory containing multiple skill subdirectories
	return &LocalParentDirSource{dir: input}
}

// URLSource downloads a single file from a URL.
type URLSource struct {
	url        string
	origURL    string // original user-supplied URL (before resolveGitHubURL)
	httpClient *http.Client
}

func (s *URLSource) Fetch() ([]FetchResult, error) {
	content, err := httpGet(s.httpClient, s.url)
	if err != nil {
		return nil, err
	}

	name := filenameFromURL(s.url)
	// Try to get name from frontmatter
	meta, _, err := frontmatter.Parse(content)
	if err == nil && meta != nil {
		if v, ok := meta["name"].(string); ok && v != "" {
			name = v
		}
	}

	src := s.origURL
	if src == "" {
		src = s.url
	}
	return []FetchResult{{Name: name, Content: content, IsDir: false, Source: src}}, nil
}

// LocalFileSource copies a single .md file.
type LocalFileSource struct {
	path string
}

func (s *LocalFileSource) Fetch() ([]FetchResult, error) {
	content, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSuffix(filepath.Base(s.path), ".md")
	meta, _, err := frontmatter.Parse(content)
	if err == nil && meta != nil {
		if v, ok := meta["name"].(string); ok && v != "" {
			name = v
		}
	}

	return []FetchResult{{Name: name, Content: content, IsDir: false}}, nil
}

// LocalDirSource copies an entire skill directory.
type LocalDirSource struct {
	dir string
}

func (s *LocalDirSource) Fetch() ([]FetchResult, error) {
	name := filepath.Base(s.dir)

	// Try to get name from SKILL.md frontmatter
	content, err := os.ReadFile(filepath.Join(s.dir, "SKILL.md"))
	if err == nil {
		meta, _, err := frontmatter.Parse(content)
		if err == nil && meta != nil {
			if v, ok := meta["name"].(string); ok && v != "" {
				name = v
			}
		}
	}

	return []FetchResult{{Name: name, DirPath: s.dir, IsDir: true}}, nil
}

// LocalParentDirSource scans a directory for skill subdirectories.
type LocalParentDirSource struct {
	dir string
}

func (s *LocalParentDirSource) Fetch() ([]FetchResult, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	var results []FetchResult
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subDir := filepath.Join(s.dir, e.Name())
		if _, err := os.Stat(filepath.Join(subDir, "SKILL.md")); err != nil {
			continue // not a skill directory
		}

		src := &LocalDirSource{dir: subDir}
		fetched, err := src.Fetch()
		if err != nil {
			continue
		}
		results = append(results, fetched...)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no skill directories found in %s", s.dir)
	}
	return results, nil
}

// Helper functions

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func resolveGitHubURL(url string) string {
	// Convert GitHub blob URLs to raw URLs
	// https://github.com/user/repo/blob/main/path -> https://raw.githubusercontent.com/user/repo/main/path
	if strings.Contains(url, "github.com") && strings.Contains(url, "/blob/") {
		url = strings.Replace(url, "github.com", "raw.githubusercontent.com", 1)
		url = strings.Replace(url, "/blob/", "/", 1)
	}
	return url
}

func filenameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return "unknown"
	}
	name := parts[len(parts)-1]
	return strings.TrimSuffix(name, ".md")
}

func httpGet(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP GET %s: status %d", url, resp.StatusCode)
	}

	// Limit response to 10MB to prevent OOM
	const maxSize = 10 << 20
	limited := io.LimitReader(resp.Body, maxSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %w", url, err)
	}
	if len(data) > maxSize {
		return nil, fmt.Errorf("response from %s exceeds 10MB limit", url)
	}
	return data, nil
}
