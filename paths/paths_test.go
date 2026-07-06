package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMediaBaseDirDefault checks that the default value (no env override) is
// an absolute path that ends with the expected "data/media" suffix.
func TestMediaBaseDirDefault(t *testing.T) {
	// Ensure the variable is unset so we exercise the default branch.
	t.Setenv("MEDIA_DOWNLOAD_DIR", "")

	dir := MediaBaseDir()

	if !filepath.IsAbs(dir) {
		t.Errorf("MediaBaseDir() = %q; want an absolute path", dir)
	}

	// Use forward slashes for the suffix check so the test passes on both
	// Windows (backslash) and Unix (forward slash).
	normalised := filepath.ToSlash(dir)
	if !strings.HasSuffix(normalised, "data/media") {
		t.Errorf("MediaBaseDir() = %q; expected suffix \"data/media\"", dir)
	}
}

// TestMediaBaseDirEnvOverride checks that MEDIA_DOWNLOAD_DIR is honoured.
func TestMediaBaseDirEnvOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MEDIA_DOWNLOAD_DIR", tmp)

	dir := MediaBaseDir()

	if !filepath.IsAbs(dir) {
		t.Errorf("MediaBaseDir() = %q; want an absolute path", dir)
	}

	// The returned value should resolve to the same location as tmp.
	absTmp, _ := filepath.Abs(tmp)
	if dir != absTmp {
		t.Errorf("MediaBaseDir() = %q; want %q", dir, absTmp)
	}
}

// TestMediaBaseDirRelativeEnv checks that a relative path in MEDIA_DOWNLOAD_DIR
// is still converted to absolute.
func TestMediaBaseDirRelativeEnv(t *testing.T) {
	t.Setenv("MEDIA_DOWNLOAD_DIR", "relative/path/here")

	dir := MediaBaseDir()

	if !filepath.IsAbs(dir) {
		t.Errorf("MediaBaseDir() with relative env = %q; want an absolute path", dir)
	}
}

// TestGetMediaPath checks that GetMediaPath returns an absolute path and that
// it starts with MediaBaseDir().
func TestGetMediaPath(t *testing.T) {
	t.Setenv("MEDIA_DOWNLOAD_DIR", "")

	got := GetMediaPath("images/foo.jpg")

	if !filepath.IsAbs(got) {
		t.Errorf("GetMediaPath() = %q; want an absolute path", got)
	}

	base := MediaBaseDir()
	if !strings.HasPrefix(got, base) {
		t.Errorf("GetMediaPath() = %q; want it to start with MediaBaseDir() = %q", got, base)
	}
}

// TestGetMediaPathWithEnvOverride checks that GetMediaPath honours the env var.
func TestGetMediaPathWithEnvOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("MEDIA_DOWNLOAD_DIR", tmp)

	got := GetMediaPath("audio/voice.ogg")

	absTmp, _ := filepath.Abs(tmp)
	expected := filepath.Join(absTmp, "audio/voice.ogg")
	if got != expected {
		t.Errorf("GetMediaPath() = %q; want %q", got, expected)
	}
}

// TestEnsureDataDirectoriesCreatesMediaBaseDir checks that EnsureDataDirectories
// creates the directory returned by MediaBaseDir().
func TestEnsureDataDirectoriesCreatesMediaBaseDir(t *testing.T) {
	tmp := t.TempDir()
	customMedia := filepath.Join(tmp, "custom-media")

	t.Setenv("MEDIA_DOWNLOAD_DIR", customMedia)

	// Temporarily point DataDir at a subdir of tmp so we don't pollute the
	// real ./data tree. We can't change the const, but we verify the custom
	// media dir is created (which is the key behaviour we're testing).
	if err := EnsureDataDirectories(); err != nil {
		t.Fatalf("EnsureDataDirectories() error: %v", err)
	}

	if _, err := os.Stat(customMedia); os.IsNotExist(err) {
		t.Errorf("EnsureDataDirectories() did not create %s", customMedia)
	}
}
