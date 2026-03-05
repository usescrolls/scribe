package scribe

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FetchResult holds the outcome of fetching a source.
// If IsCached is true, the caller must NOT delete ContentDir.
// If IsCached is false and ContentDir != "", the caller must clean up via Cleanup().
type FetchResult struct {
	ContentDir string // root dir containing the fetched content
	SkillsDir  string // ContentDir + subpath (where to discover skills)
	IsCached   bool   // if true, do not delete ContentDir
	IsPrivate  bool   // true if authentication was used to fetch this source
}

// Cleanup removes the content directory only if it is not cached.
func (r *FetchResult) Cleanup() {
	if r == nil {
		return
	}
	if !r.IsCached && r.ContentDir != "" {
		_ = os.RemoveAll(r.ContentDir)
	}
}

// FetchAndDiscoverSkills fetches content from a source and discovers skills.
// Returns the discovered skills and a FetchResult for lifecycle management.
func FetchAndDiscoverSkills(source *SourceInfo, emit ...ProgressEmitter) ([]*Skill, *FetchResult, error) {
	result := &FetchResult{}

	emitProgress(emit, "discover", "parsing", "Parsing source...", "")

	switch source.Type {
	case "local":
		result.SkillsDir = source.LocalPath
		if source.Subpath != "" {
			if !IsSubpathSafe(result.SkillsDir, source.Subpath) {
				return nil, nil, fmt.Errorf("invalid subpath: %q resolves outside the base directory", source.Subpath)
			}
			result.SkillsDir = filepath.Join(result.SkillsDir, source.Subpath)
		}

	case "github", "gitlab", "bitbucket", "git":
		repoDir, isCached, authRequired, err := CloneOrUpdateRepo(source, emit...)
		if err != nil {
			return nil, nil, err
		}
		result.ContentDir = repoDir
		result.IsCached = isCached
		result.IsPrivate = authRequired
		result.SkillsDir = repoDir
		if source.Subpath != "" {
			if !IsSubpathSafe(repoDir, source.Subpath) {
				return nil, nil, fmt.Errorf("invalid subpath: %q resolves outside the repository directory", source.Subpath)
			}
			result.SkillsDir = filepath.Join(repoDir, source.Subpath)
		}

	case "zip":
		emitProgress(emit, "discover", "downloading", "Downloading zip archive...", "")
		tempDir, err := DownloadAndExtractZip(source.URL)
		if err != nil {
			return nil, nil, err
		}
		emitProgress(emit, "discover", "extracting", "Extracting archive...", "")
		result.ContentDir = tempDir
		result.SkillsDir = tempDir

	case "well-known":
		return nil, nil, fmt.Errorf("well-known sources not yet implemented")

	default:
		return nil, nil, fmt.Errorf("unsupported source type: %s", source.Type)
	}

	// Discover skills in the directory
	emitProgress(emit, "discover", "scanning", "Scanning for skills...", "")
	skills, err := DiscoverSkills(result.SkillsDir)
	if err != nil {
		result.Cleanup()
		return nil, nil, err
	}

	emitProgress(emit, "discover", "done", fmt.Sprintf("Found %d skill(s)", len(skills)), "")

	return skills, result, nil
}

// DownloadAndExtractZip downloads and extracts a zip file to a temp directory.
func DownloadAndExtractZip(zipURL string) (string, error) {
	tempDir, err := os.MkdirTemp("", "scribe-zip-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	resp, err := http.Get(zipURL)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to download zip: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to download zip: status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "scribe-download-*.zip")
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		_ = tmpFile.Close()
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to save zip: %w", err)
	}
	_ = tmpFile.Close()

	zipReader, err := zip.OpenReader(tmpPath)
	if err != nil {
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer func() { _ = zipReader.Close() }()

	commonRoot := findZipCommonRoot(zipReader.File)

	for _, file := range zipReader.File {
		filePath := file.Name

		if commonRoot != "" {
			filePath = strings.TrimPrefix(filePath, commonRoot)
			if filePath == "" {
				continue
			}
		}

		destPath := filepath.Join(tempDir, filePath)

		// Check for zip slip vulnerability
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(tempDir)+string(os.PathSeparator)) {
			_ = os.RemoveAll(tempDir)
			return "", fmt.Errorf("invalid file path in zip: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			_ = os.MkdirAll(destPath, file.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			_ = os.RemoveAll(tempDir)
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		if err := extractZipFile(file, destPath); err != nil {
			_ = os.RemoveAll(tempDir)
			return "", err
		}
	}

	return tempDir, nil
}

// findZipCommonRoot checks if all files in a zip share a common root directory.
func findZipCommonRoot(files []*zip.File) string {
	if len(files) == 0 {
		return ""
	}

	var commonRoot string
	for _, file := range files {
		parts := strings.SplitN(file.Name, "/", 2)
		if len(parts) > 1 {
			if commonRoot == "" {
				commonRoot = parts[0] + "/"
			} else if parts[0]+"/" != commonRoot {
				return ""
			}
		} else if !file.FileInfo().IsDir() {
			return ""
		}
	}

	return commonRoot
}

// extractZipFile extracts a single file from a zip archive.
func extractZipFile(file *zip.File, destPath string) error {
	srcFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	return nil
}
