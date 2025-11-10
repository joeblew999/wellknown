package pdfform

import (
	"os"
	"path/filepath"
	"sync"
)

// Default configuration constants
const (
	DefaultDataDirName      = ".data"
	DefaultCatalogDirName   = "catalog"
	DefaultDownloadsDirName = "downloads"
	DefaultTemplatesDirName = "templates"
	DefaultOutputsDirName   = "outputs"
	DefaultCasesDirName     = "cases"
	DefaultTempDirName      = "temp"
	DefaultCertsDirName     = "certs"
	DefaultCatalogFileName  = "australian_transfer_forms.csv"
	DefaultCertFileName     = "cert.pem"
	DefaultKeyFileName      = "key.pem"

	// Docker paths
	DockerAppDir  = "/app"
	DockerDataDir = "/app/.data"
)

// Config holds all configuration paths and settings for the PDF form system
type Config struct {
	// Base directory - all paths are relative to this
	DataDir string

	// Subdirectories
	CatalogDir   string // Catalog files (CSV)
	DownloadsDir string // Downloaded PDFs
	TemplatesDir string // Field templates (JSON)
	OutputsDir   string // Filled PDFs
	CasesDir     string // Case files
	TempDir      string // Temporary files
	CertsDir     string // HTTPS certificates

	// File names
	CatalogFile string // australian_transfer_forms.csv
	CertFile    string // cert.pem
	KeyFile     string // key.pem

	// System temp directory (for OS-level temp files)
	SystemTempDir string
}

var (
	// defaultConfig is the package-level default configuration
	defaultConfig *Config
	configMutex   sync.RWMutex
)

// DefaultConfig returns the default configuration
// This uses ".data" as the base directory with standard subdirectories
// The .data prefix is a convention for hidden data directories
// In Docker, this will be mounted as a volume (e.g., /app/.data)
func DefaultConfig() *Config {
	return &Config{
		DataDir:       DefaultDataDirName,
		CatalogDir:    DefaultCatalogDirName,
		DownloadsDir:  DefaultDownloadsDirName,
		TemplatesDir:  DefaultTemplatesDirName,
		OutputsDir:    DefaultOutputsDirName,
		CasesDir:      DefaultCasesDirName,
		TempDir:       DefaultTempDirName,
		CertsDir:      DefaultCertsDirName,
		CatalogFile:   DefaultCatalogFileName,
		CertFile:      DefaultCertFileName,
		KeyFile:       DefaultKeyFileName,
		SystemTempDir: os.TempDir(),
	}
}

// NewConfig creates a new configuration with the given base data directory
func NewConfig(dataDir string) *Config {
	cfg := DefaultConfig()
	cfg.DataDir = dataDir
	return cfg
}

// SetDefaultConfig sets the package-level default configuration
// This should be called early in main() to configure the entire system
func SetDefaultConfig(cfg *Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	defaultConfig = cfg
}

// GetDefaultConfig returns the current default configuration
// If not set, returns a new default config
func GetDefaultConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if defaultConfig == nil {
		return DefaultConfig()
	}
	return defaultConfig
}

// Path building methods

// CatalogPath returns the full path to the catalog directory
func (c *Config) CatalogPath() string {
	return filepath.Join(c.DataDir, c.CatalogDir)
}

// CatalogFilePath returns the full path to the catalog CSV file
func (c *Config) CatalogFilePath() string {
	return filepath.Join(c.DataDir, c.CatalogDir, c.CatalogFile)
}

// DownloadsPath returns the full path to the downloads directory
func (c *Config) DownloadsPath() string {
	return filepath.Join(c.DataDir, c.DownloadsDir)
}

// TemplatesPath returns the full path to the templates directory
func (c *Config) TemplatesPath() string {
	return filepath.Join(c.DataDir, c.TemplatesDir)
}

// OutputsPath returns the full path to the outputs directory
func (c *Config) OutputsPath() string {
	return filepath.Join(c.DataDir, c.OutputsDir)
}

// CasesPath returns the full path to the cases directory
func (c *Config) CasesPath() string {
	return filepath.Join(c.DataDir, c.CasesDir)
}

// TestScenariosPath returns the full path to the test scenarios directory
func (c *Config) TestScenariosPath() string {
	return filepath.Join(c.DataDir, c.CasesDir, "test_scenarios")
}

// TempPath returns the full path to the temp directory
func (c *Config) TempPath() string {
	return filepath.Join(c.DataDir, c.TempDir)
}

// CertsPath returns the full path to the certs directory
func (c *Config) CertsPath() string {
	return filepath.Join(c.DataDir, c.CertsDir)
}

// CertFilePath returns the full path to the certificate file
func (c *Config) CertFilePath() string {
	return filepath.Join(c.DataDir, c.CertsDir, c.CertFile)
}

// KeyFilePath returns the full path to the key file
func (c *Config) KeyFilePath() string {
	return filepath.Join(c.DataDir, c.CertsDir, c.KeyFile)
}

// EntityCasesPath returns the full path to a specific entity's cases directory
func (c *Config) EntityCasesPath(entityName string) string {
	return filepath.Join(c.DataDir, c.CasesDir, entityName)
}

// EnsureDirectories creates all necessary directories
func (c *Config) EnsureDirectories() error {
	dirs := []string{
		c.CatalogPath(),
		c.DownloadsPath(),
		c.TemplatesPath(),
		c.OutputsPath(),
		c.CasesPath(),
		c.TempPath(),
		c.CertsPath(),
		c.TestScenariosPath(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// Helper functions for backward compatibility and convenience

// GetCatalogPath returns the catalog CSV path using default config
func GetCatalogPath() string {
	return GetDefaultConfig().CatalogFilePath()
}

// GetDownloadsPath returns the downloads directory using default config
func GetDownloadsPath() string {
	return GetDefaultConfig().DownloadsPath()
}

// GetTemplatesPath returns the templates directory using default config
func GetTemplatesPath() string {
	return GetDefaultConfig().TemplatesPath()
}

// GetOutputsPath returns the outputs directory using default config
func GetOutputsPath() string {
	return GetDefaultConfig().OutputsPath()
}

// GetCasesPath returns the cases directory using default config
func GetCasesPath() string {
	return GetDefaultConfig().CasesPath()
}

// GetTestScenariosPath returns the test scenarios directory using default config
func GetTestScenariosPath() string {
	return GetDefaultConfig().TestScenariosPath()
}

// GetTempPath returns the temp directory using default config
func GetTempPath() string {
	return GetDefaultConfig().TempPath()
}

// GetEntityCasesPath returns the entity cases directory using default config
func GetEntityCasesPath(entityName string) string {
	return GetDefaultConfig().EntityCasesPath(entityName)
}

// FindDataDir searches for the data directory in common locations
// Priority: ENV variable > Docker path > Current dir > Parent dirs > Default
func FindDataDir() string {
	// 1. Check environment variable
	if dataDir := os.Getenv("PDFFORM_DATA_DIR"); dataDir != "" {
		return dataDir
	}

	// 2. Check if running in Docker
	if _, err := os.Stat(DockerAppDir); err == nil {
		return DockerDataDir
	}

	// 3. Search in current directory and parents
	searchPaths := []string{
		DefaultDataDirName,                                    // .data
		filepath.Join("..", DefaultDataDirName),              // ../.data
		filepath.Join("..", "..", DefaultDataDirName),        // ../../.data
		filepath.Join("..", "..", "..", DefaultDataDirName),  // ../../../.data
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// 4. Default: assume current directory
	return DefaultDataDirName
}
