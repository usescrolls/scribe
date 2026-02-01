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

// ResolveSource resolves the plugin source based on type
// For github/git/url sources, it passes through the source definition
// For zip sources, it downloads and extracts the plugin files locally
func (s *Server) ResolveSource(name string, source PluginSource) (interface{}, error) {
	switch source.Source {
	case "github":
		// Pass through - Claude Code handles GitHub sources directly
		if source.Repo == "" {
			return nil, fmt.Errorf("github source requires repo")
		}
		resolved := map[string]interface{}{
			"source": "github",
			"repo":   source.Repo,
		}
		if source.Ref != "" {
			resolved["ref"] = source.Ref
		}
		return resolved, nil

	case "git", "url":
		// Pass through - Claude Code handles git URLs directly
		if source.URL == "" {
			return nil, fmt.Errorf("git/url source requires url")
		}
		resolved := map[string]interface{}{
			"source": "url",
			"url":    source.URL,
		}
		if source.Ref != "" {
			resolved["ref"] = source.Ref
		}
		return resolved, nil

	case "zip":
		// Download zip file and extract to plugins directory
		if source.URL == "" {
			return nil, fmt.Errorf("zip source requires url")
		}
		if err := s.downloadZip(name, source.URL); err != nil {
			return nil, err
		}
		return "./plugins/" + name, nil

	default:
		return nil, fmt.Errorf("unsupported source type: %s (supported: github, git, url, zip)", source.Source)
	}
}

// downloadZip downloads a zip file from a URL and extracts it to the plugins directory
func (s *Server) downloadZip(name string, zipURL string) error {
	targetDir := filepath.Join(s.hubDir, PluginsDirName, name)
	Logger.Info("downloading zip plugin", "name", name, "url", zipURL, "target", targetDir)

	// Remove existing if present
	if err := os.RemoveAll(targetDir); err != nil {
		Logger.Warn("failed to remove existing plugin directory", "path", targetDir, "error", err)
	}

	// Download the zip file
	resp, err := http.Get(zipURL)
	if err != nil {
		return fmt.Errorf("failed to download zip: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download zip: status %d", resp.StatusCode)
	}

	// Create a temporary file for the zip
	tmpFile, err := os.CreateTemp("", "plugin-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Copy the response body to the temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to save zip: %w", err)
	}
	tmpFile.Close()

	// Open the zip file
	zipReader, err := zip.OpenReader(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer zipReader.Close()

	// Create the target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract files
	for _, file := range zipReader.File {
		// Handle the case where zip contains a root folder with the plugin name
		// e.g., test-runner/plugin.json should extract to targetDir/plugin.json
		filePath := file.Name

		// Strip the first directory component if it matches the plugin name
		parts := strings.SplitN(filePath, "/", 2)
		if len(parts) > 1 && parts[0] == name {
			filePath = parts[1]
		}

		if filePath == "" {
			continue
		}

		destPath := filepath.Join(targetDir, filePath)

		// Check for zip slip vulnerability
		if !strings.HasPrefix(destPath, filepath.Clean(targetDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path in zip: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, file.Mode())
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Extract file
		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %w", err)
		}

		dstFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("failed to create file: %w", err)
		}

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			srcFile.Close()
			dstFile.Close()
			return fmt.Errorf("failed to extract file: %w", err)
		}

		srcFile.Close()
		dstFile.Close()
	}

	Logger.Info("plugin extracted successfully", "name", name)
	return nil
}
