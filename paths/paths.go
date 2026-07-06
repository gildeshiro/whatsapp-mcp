package paths

import (
	"os"
	"path/filepath"
)

// MediaBaseDir returns the absolute path to the media download directory.
// If the MEDIA_DOWNLOAD_DIR environment variable is set and non-empty, that
// value is resolved to an absolute path; otherwise the default ./data/media
// is used. The env var is read on every call so that tests can override it
// with t.Setenv without needing a sync.Once reset mechanism.
func MediaBaseDir() string {
	if dir := os.Getenv("MEDIA_DOWNLOAD_DIR"); dir != "" {
		if abs, err := filepath.Abs(dir); err == nil {
			return abs
		}
	}
	if abs, err := filepath.Abs(DataMediaDir); err == nil {
		return abs
	}
	// unreachable in practice, but keeps the function pure
	return DataMediaDir
}

// DataDir is the base data directory for the application.
const DataDir = "./data"

// Data subdirectories for organizing different types of data.
const (
	DataDBDir    = DataDir + "/db"
	DataMediaDir = DataDir + "/media"
)

// Storage paths for migrations and other persistent data.
const (
	MigrationsDir = "storage/migrations"
)

// File paths for databases, logs, and other files.
const (
	MessagesDBPath     = DataDBDir + "/messages.db"
	WhatsAppAuthDBPath = DataDBDir + "/whatsapp_auth.db"
	WhatsAppLogPath    = DataDir + "/whatsapp.log"
	QRCodePath         = "./qr.png"
)

// EnsureDataDirectories ensures that all required data directories exist.
// This includes the configurable media base dir (honouring MEDIA_DOWNLOAD_DIR).
func EnsureDataDirectories() error {
	dirs := []string{
		DataDir,
		DataDBDir,
		MediaBaseDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// GetMediaPath returns the absolute path for a media file given its relative path.
// The base directory is determined by MediaBaseDir (honours MEDIA_DOWNLOAD_DIR).
func GetMediaPath(relativePath string) string {
	return filepath.Join(MediaBaseDir(), relativePath)
}
