package pdfform

import (
	"fmt"
	"os"
	"os/exec"
)

// CertManager handles HTTPS certificate generation and management
type CertManager struct {
	config *Config
}

// NewCertManager creates a new certificate manager
func NewCertManager(config *Config) *CertManager {
	return &CertManager{config: config}
}

// IsMkcertInstalled checks if mkcert is installed
func (cm *CertManager) IsMkcertInstalled() bool {
	_, err := exec.LookPath("mkcert")
	return err == nil
}

// InstallMkcert installs mkcert using go install
func (cm *CertManager) InstallMkcert() error {
	fmt.Println("📦 Installing mkcert...")
	cmd := exec.Command("go", "install", "filippo.io/mkcert@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// EnsureMkcert ensures mkcert is installed
func (cm *CertManager) EnsureMkcert() error {
	if cm.IsMkcertInstalled() {
		return nil
	}

	fmt.Println("🔧 mkcert not found, installing...")
	if err := cm.InstallMkcert(); err != nil {
		return fmt.Errorf("failed to install mkcert: %w", err)
	}

	fmt.Println("✅ mkcert installed successfully")
	return nil
}

// CertsExist checks if certificates already exist
func (cm *CertManager) CertsExist() bool {
	certPath := cm.config.CertFilePath()
	keyPath := cm.config.KeyFilePath()

	_, certErr := os.Stat(certPath)
	_, keyErr := os.Stat(keyPath)

	return certErr == nil && keyErr == nil
}

// GenerateCerts generates new certificates using mkcert
func (cm *CertManager) GenerateCerts() error {
	// Ensure mkcert is installed
	if err := cm.EnsureMkcert(); err != nil {
		return err
	}

	// Ensure certs directory exists
	certsDir := cm.config.CertsPath()
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		return fmt.Errorf("failed to create certs directory: %w", err)
	}

	// Install CA if not already installed
	fmt.Println("🔐 Setting up local CA...")
	installCmd := exec.Command("mkcert", "-install")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		fmt.Printf("⚠️  Warning: Could not install CA: %v\n", err)
		fmt.Println("   Certificates will still work, but may show warnings")
	}

	// Get local IPs for certificate
	ips, err := GetLocalIPs()
	if err != nil {
		return fmt.Errorf("failed to get local IPs: %w", err)
	}

	// Build mkcert command with localhost and all local IPs
	args := []string{
		"-cert-file", cm.config.CertFilePath(),
		"-key-file", cm.config.KeyFilePath(),
		"localhost",
		"127.0.0.1",
		"::1",
	}
	args = append(args, ips...)

	fmt.Printf("🔐 Generating certificates for localhost and %d LAN IP(s)...\n", len(ips))
	cmd := exec.Command("mkcert", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate certificates: %w", err)
	}

	fmt.Printf("✅ Certificates generated:\n")
	fmt.Printf("   📄 %s\n", cm.config.CertFilePath())
	fmt.Printf("   🔑 %s\n", cm.config.KeyFilePath())

	return nil
}

// EnsureCerts ensures certificates exist, generating them if needed
func (cm *CertManager) EnsureCerts() error {
	if cm.CertsExist() {
		return nil
	}

	fmt.Println("📜 No certificates found, generating new ones...")
	return cm.GenerateCerts()
}

// GetCertPaths returns the paths to the certificate files
func (cm *CertManager) GetCertPaths() (certPath, keyPath string, err error) {
	if !cm.CertsExist() {
		return "", "", fmt.Errorf("certificates do not exist")
	}

	return cm.config.CertFilePath(), cm.config.KeyFilePath(), nil
}

// RegenerateCerts removes existing certs and generates new ones
func (cm *CertManager) RegenerateCerts() error {
	// Remove existing certs
	certPath := cm.config.CertFilePath()
	keyPath := cm.config.KeyFilePath()

	os.Remove(certPath)
	os.Remove(keyPath)

	// Generate new ones
	return cm.GenerateCerts()
}

// ShowCertInfo displays information about the certificates
func (cm *CertManager) ShowCertInfo() error {
	if !cm.CertsExist() {
		fmt.Println("❌ No certificates found")
		fmt.Printf("   Run 'pdfform certs generate' to create them\n")
		return nil
	}

	certPath := cm.config.CertFilePath()
	keyPath := cm.config.KeyFilePath()

	fmt.Println("📜 Certificate Information:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📄 Certificate: %s\n", certPath)
	fmt.Printf("🔑 Key:         %s\n", keyPath)

	// Show file sizes and modification times
	certInfo, err := os.Stat(certPath)
	if err == nil {
		fmt.Printf("\n   Created: %s\n", certInfo.ModTime().Format("2006-01-02 15:04:05"))
		fmt.Printf("   Size:    %d bytes\n", certInfo.Size())
	}

	// Get domains from certificate
	ips, _ := GetLocalIPs()
	fmt.Println("\n📡 Valid for:")
	fmt.Println("   • localhost")
	fmt.Println("   • 127.0.0.1")
	for _, ip := range ips {
		fmt.Printf("   • %s\n", ip)
	}

	fmt.Println("\n🔒 Trust:")
	fmt.Println("   Desktop browsers: Trusted automatically (via mkcert CA)")
	fmt.Println("   iOS devices: Just visit the root URL in Safari, accept the cert prompt, and you're done!")

	return nil
}

// Helper function to ensure certs for the default config
func EnsureCerts() error {
	config := GetDefaultConfig()
	cm := NewCertManager(config)
	return cm.EnsureCerts()
}

// Helper to get cert paths from default config
func GetCertPaths() (certPath, keyPath string, err error) {
	config := GetDefaultConfig()
	cm := NewCertManager(config)
	return cm.GetCertPaths()
}
