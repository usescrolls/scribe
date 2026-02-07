package scribe

import (
	"archive/zip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestFetchResult_Cleanup_IsCachedWithContent(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "scribe-cleanup-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	result := &FetchResult{
		ContentDir: tmpDir,
		IsCached:   true,
	}
	result.Cleanup()

	// Cached dir should NOT be removed
	if _, err := os.Stat(tmpDir); err != nil {
		t.Error("cached content dir should not be removed by Cleanup()")
	}
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

func TestFetchAndDiscoverSkills_GitHub(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	remoteDir := filepath.Join(tmpDir, "github-repo")
	createTestGitRepo(t, remoteDir, map[string]string{
		"SKILL.md": "---\nname: github-skill\ndescription: GitHub\n---\n# GitHub Skill\n",
	})

	source := &SourceInfo{
		Type:  "github",
		Owner: "testuser",
		Repo:  "github-test",
		URL:   remoteDir,
	}

	skills, result, err := FetchAndDiscoverSkills(source)
	if result != nil {
		defer result.Cleanup()
	}
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills(github) error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "github-skill" {
		t.Errorf("skill name = %q, want 'github-skill'", skills[0].Name)
	}
	if result != nil && !result.IsCached {
		t.Error("expected IsCached=true for github source")
	}
}

func TestFetchAndDiscoverSkills_GitHubWithSubpath(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	remoteDir := filepath.Join(tmpDir, "subpath-repo")
	createTestGitRepo(t, remoteDir, map[string]string{
		"skills/my-skill/SKILL.md": "---\nname: subpath-skill\ndescription: Subpath\n---\n# Subpath\n",
	})

	source := &SourceInfo{
		Type:    "github",
		Owner:   "testuser",
		Repo:    "subpath-test",
		URL:     remoteDir,
		Subpath: "skills",
	}

	skills, result, err := FetchAndDiscoverSkills(source)
	if result != nil {
		defer result.Cleanup()
	}
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills(github+subpath) error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "subpath-skill" {
		t.Errorf("skill name = %q, want 'subpath-skill'", skills[0].Name)
	}
}

func TestFetchAndDiscoverSkills_GitLab(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)
	_ = EnsureScribeDirs()

	remoteDir := filepath.Join(tmpDir, "gitlab-repo")
	createTestGitRepo(t, remoteDir, map[string]string{
		"SKILL.md": "---\nname: gitlab-skill\ndescription: GitLab\n---\n# GitLab Skill\n",
	})

	source := &SourceInfo{
		Type:  "gitlab",
		Owner: "testuser",
		Repo:  "gitlab-test",
		URL:   remoteDir,
	}

	skills, result, err := FetchAndDiscoverSkills(source)
	if result != nil {
		defer result.Cleanup()
	}
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills(gitlab) error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "gitlab-skill" {
		t.Errorf("skill name = %q, want 'gitlab-skill'", skills[0].Name)
	}
}

func TestFetchAndDiscoverSkills_Zip(t *testing.T) {
	tmpDir := setupTempHome(t)
	InitLoggerCLI(false)

	// Create a zip file with a SKILL.md
	zipPath := filepath.Join(tmpDir, "skills.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	w := zip.NewWriter(zipFile)
	fw, _ := w.Create("SKILL.md")
	_, _ = fw.Write([]byte("---\nname: zip-skill\ndescription: Zip\n---\n# Zip\n"))
	_ = w.Close()
	_ = zipFile.Close()

	// Serve the zip
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, zipPath)
	}))
	defer srv.Close()

	source := &SourceInfo{
		Type: "zip",
		URL:  srv.URL + "/skills.zip",
	}

	skills, result, err := FetchAndDiscoverSkills(source)
	if result != nil {
		defer result.Cleanup()
	}
	if err != nil {
		t.Fatalf("FetchAndDiscoverSkills(zip) error: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Name != "zip-skill" {
		t.Errorf("skill name = %q, want 'zip-skill'", skills[0].Name)
	}
	if result != nil && result.IsCached {
		t.Error("expected IsCached=false for zip source")
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

func TestFindZipCommonRoot_SingleDir(t *testing.T) {
	files := []*zip.File{
		{FileHeader: zip.FileHeader{Name: "root/"}},
	}
	result := findZipCommonRoot(files)
	// A single directory entry "root/" counts as a common root prefix
	if result != "root/" {
		t.Errorf("findZipCommonRoot(single dir) = %q, want 'root/'", result)
	}
}

func TestFindZipCommonRoot_MixedRoots(t *testing.T) {
	files := []*zip.File{
		{FileHeader: zip.FileHeader{Name: "a/file1.txt"}},
		{FileHeader: zip.FileHeader{Name: "b/file2.txt"}},
	}
	result := findZipCommonRoot(files)
	if result != "" {
		t.Errorf("findZipCommonRoot(mixed roots) = %q, want ''", result)
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

func TestDownloadAndExtractZip_InvalidZip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("this is not a zip file"))
	}))
	defer srv.Close()

	_, err := DownloadAndExtractZip(srv.URL + "/bad.zip")
	if err == nil {
		t.Error("expected error for invalid zip content")
	}
}

func TestDownloadAndExtractZip_WithSubdirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-zip-test-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a zip with a subdirectory structure (common root)
	zipPath := filepath.Join(tmpDir, "test.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	w := zip.NewWriter(zipFile)

	// Add directory entries with proper permissions
	dirHeader := &zip.FileHeader{Name: "root/"}
	dirHeader.SetMode(0o755 | os.ModeDir)
	_, _ = w.CreateHeader(dirHeader)

	subDirHeader := &zip.FileHeader{Name: "root/sub/"}
	subDirHeader.SetMode(0o755 | os.ModeDir)
	_, _ = w.CreateHeader(subDirHeader)

	fw, _ := w.Create("root/file.txt")
	_, _ = fw.Write([]byte("content"))

	fw2, _ := w.Create("root/sub/nested.txt")
	_, _ = fw2.Write([]byte("nested"))

	_ = w.Close()
	_ = zipFile.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, zipPath)
	}))
	defer srv.Close()

	extractDir, err := DownloadAndExtractZip(srv.URL + "/test.zip")
	if err != nil {
		t.Fatalf("DownloadAndExtractZip error: %v", err)
	}
	defer func() { _ = os.RemoveAll(extractDir) }()

	// Common root "root/" should be stripped
	if _, err := os.Stat(filepath.Join(extractDir, "file.txt")); err != nil {
		t.Error("file.txt not found after extraction (common root should be stripped)")
	}
	if _, err := os.Stat(filepath.Join(extractDir, "sub", "nested.txt")); err != nil {
		t.Error("sub/nested.txt not found after extraction")
	}
}

func TestDownloadAndExtractZip_EmptyZip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-zip-test-*")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	zipPath := filepath.Join(tmpDir, "empty.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	w := zip.NewWriter(zipFile)
	_ = w.Close()
	_ = zipFile.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, zipPath)
	}))
	defer srv.Close()

	extractDir, err := DownloadAndExtractZip(srv.URL + "/empty.zip")
	if err != nil {
		t.Fatalf("DownloadAndExtractZip(empty) error: %v", err)
	}
	defer func() { _ = os.RemoveAll(extractDir) }()

	// Should succeed with empty directory
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		t.Fatalf("read extracted dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries in empty zip extraction, got %d", len(entries))
	}
}

func TestDownloadAndExtractZip_ConnectionError(t *testing.T) {
	_, err := DownloadAndExtractZip("http://localhost:1/nonexistent.zip")
	if err == nil {
		t.Error("expected error for connection failure")
	}
}

func TestExtractZipFile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "scribe-extract-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a zip file
	zipPath := filepath.Join(tmpDir, "test.zip")
	zipFile, _ := os.Create(zipPath)
	w := zip.NewWriter(zipFile)
	fw, _ := w.Create("test.txt")
	_, _ = fw.Write([]byte("extracted content"))
	_ = w.Close()
	_ = zipFile.Close()

	// Open and extract
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("OpenReader error: %v", err)
	}
	defer func() { _ = reader.Close() }()

	destPath := filepath.Join(tmpDir, "output.txt")
	err = extractZipFile(reader.File[0], destPath)
	if err != nil {
		t.Fatalf("extractZipFile error: %v", err)
	}

	data, _ := os.ReadFile(destPath)
	if string(data) != "extracted content" {
		t.Errorf("extracted content = %q, want 'extracted content'", string(data))
	}
}

func TestDownloadAndExtractZip_ZipSlip(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "scribe-zipslip-*")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a zip with a path traversal attempt
	zipPath := filepath.Join(tmpDir, "evil.zip")
	zipFile, _ := os.Create(zipPath)
	w := zip.NewWriter(zipFile)
	fw, _ := w.Create("../../etc/passwd")
	_, _ = fw.Write([]byte("evil"))
	_ = w.Close()
	_ = zipFile.Close()

	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.ServeFile(rw, r, zipPath)
	}))
	defer srv.Close()

	_, err := DownloadAndExtractZip(srv.URL + "/evil.zip")
	if err == nil {
		t.Error("expected error for zip slip attack")
	}
	if !strings.Contains(err.Error(), "invalid file path") {
		t.Errorf("error = %q, want 'invalid file path'", err.Error())
	}
}
