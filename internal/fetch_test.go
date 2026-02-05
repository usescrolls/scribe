package scribe

import (
	"archive/zip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFetchResult_Cleanup_NonCached(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-fetch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	result := &FetchResult{
		ContentDir: tmpDir,
		IsCached:   false,
	}

	result.Cleanup()

	if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
		t.Error("expected non-cached ContentDir to be removed after Cleanup()")
	}
}

func TestFetchResult_Cleanup_Cached(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-fetch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	result := &FetchResult{
		ContentDir: tmpDir,
		IsCached:   true,
	}

	result.Cleanup()

	if _, err := os.Stat(tmpDir); err != nil {
		t.Error("expected cached ContentDir to remain after Cleanup()")
	}
}

func TestFetchResult_Cleanup_Nil(t *testing.T) {
	var result *FetchResult
	result.Cleanup()
}

func TestFetchResult_Cleanup_EmptyContentDir(t *testing.T) {
	result := &FetchResult{
		ContentDir: "",
		IsCached:   false,
	}
	result.Cleanup()
}

func TestFetchAndDiscoverSkills_Local(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-fetch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	InitLoggerCLI(false)

	skillContent := `---
name: test-skill
description: A test skill
---
# Test Skill

This is a test skill.
`
	if err := os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte(skillContent), 0o644); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}

	source := &SourceInfo{
		Type:      "local",
		LocalPath: tmpDir,
	}

	skills, result, err := FetchAndDiscoverSkills(source)
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills() error: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "test-skill" {
		t.Errorf("expected skill name 'test-skill', got %q", skills[0].Name)
	}
	if result.ContentDir != "" {
		t.Errorf("expected empty ContentDir for local source, got %q", result.ContentDir)
	}
}

func TestFetchAndDiscoverSkills_LocalWithSubpath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-fetch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	InitLoggerCLI(false)

	subDir := filepath.Join(tmpDir, "skills")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	skillContent := `---
name: sub-skill
description: A skill in a subdirectory
---
# Sub Skill
`
	if err := os.WriteFile(filepath.Join(subDir, "SKILL.md"), []byte(skillContent), 0o644); err != nil {
		t.Fatalf("failed to write SKILL.md: %v", err)
	}

	source := &SourceInfo{
		Type:      "local",
		LocalPath: tmpDir,
		Subpath:   "skills",
	}

	skills, _, err := FetchAndDiscoverSkills(source)
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills() error: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "sub-skill" {
		t.Errorf("expected skill name 'sub-skill', got %q", skills[0].Name)
	}
}

func TestFetchAndDiscoverSkills_UnsupportedType(t *testing.T) {
	InitLoggerCLI(false)

	source := &SourceInfo{Type: "unsupported"}
	_, _, err := FetchAndDiscoverSkills(source)
	if err == nil {
		t.Error("expected error for unsupported source type")
	}
}

func TestFetchAndDiscoverSkills_WellKnown(t *testing.T) {
	InitLoggerCLI(false)

	source := &SourceInfo{Type: "well-known"}
	_, _, err := FetchAndDiscoverSkills(source)
	if err == nil {
		t.Error("expected error for well-known source type")
	}
}

func TestFindZipCommonRoot(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{
			name:     "empty file list",
			files:    nil,
			expected: "",
		},
		{
			name:     "all share common root",
			files:    []string{"project/", "project/file1.txt", "project/sub/file2.txt"},
			expected: "project/",
		},
		{
			name:     "no common root - multiple top dirs",
			files:    []string{"dir1/file1.txt", "dir2/file2.txt"},
			expected: "",
		},
		{
			name:     "no common root - file at top level",
			files:    []string{"file.txt"},
			expected: "",
		},
		{
			name:     "single file in directory",
			files:    []string{"root/file.txt"},
			expected: "root/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var zipFiles []*zip.File
			for _, name := range tt.files {
				fh := &zip.FileHeader{Name: name}
				zipFiles = append(zipFiles, &zip.File{FileHeader: *fh})
			}

			result := findZipCommonRoot(zipFiles)
			if result != tt.expected {
				t.Errorf("findZipCommonRoot() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDownloadAndExtractZip(t *testing.T) {
	// Create a zip file in memory
	tmpDir, err := os.MkdirTemp("", "scribe-zip-src-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	zipPath := filepath.Join(tmpDir, "test.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	w := zip.NewWriter(zipFile)

	files := []struct {
		name    string
		content string
	}{
		{"myproject/SKILL.md", "---\nname: zip-skill\ndescription: A zipped skill\n---\n# Zip Skill\n"},
		{"myproject/README.md", "# README\n"},
	}
	for _, f := range files {
		fw, err := w.Create(f.name)
		if err != nil {
			t.Fatalf("failed to create file in zip: %v", err)
		}
		if _, err := fw.Write([]byte(f.content)); err != nil {
			t.Fatalf("failed to write file content: %v", err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatalf("failed to close zip file: %v", err)
	}

	// Serve the zip via httptest
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, zipPath)
	}))
	defer srv.Close()

	extractDir, err := DownloadAndExtractZip(srv.URL + "/test.zip")
	if err != nil {
		t.Fatalf("DownloadAndExtractZip() error: %v", err)
	}
	defer func() { _ = os.RemoveAll(extractDir) }()

	// Common root "myproject/" should be stripped
	skillPath := filepath.Join(extractDir, "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("expected SKILL.md at root of extracted dir, got error: %v", err)
	}

	readmePath := filepath.Join(extractDir, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		t.Errorf("expected README.md at root of extracted dir, got error: %v", err)
	}
}

func TestDownloadAndExtractZip_NoCommonRoot(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-zip-src-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	zipPath := filepath.Join(tmpDir, "flat.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	w := zip.NewWriter(zipFile)
	fw, err := w.Create("SKILL.md")
	if err != nil {
		t.Fatalf("failed to create file in zip: %v", err)
	}
	if _, err := fw.Write([]byte("---\nname: flat-skill\ndescription: Flat\n---\n# Flat\n")); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	if err := zipFile.Close(); err != nil {
		t.Fatalf("failed to close zip file: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, zipPath)
	}))
	defer srv.Close()

	extractDir, err := DownloadAndExtractZip(srv.URL + "/flat.zip")
	if err != nil {
		t.Fatalf("DownloadAndExtractZip() error: %v", err)
	}
	defer func() { _ = os.RemoveAll(extractDir) }()

	skillPath := filepath.Join(extractDir, "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("expected SKILL.md at root of extracted dir, got error: %v", err)
	}
}

func TestDownloadAndExtractZip_BadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := DownloadAndExtractZip(srv.URL + "/missing.zip")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}
