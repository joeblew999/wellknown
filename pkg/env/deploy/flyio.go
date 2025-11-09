package deploy

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joeblew999/wellknown/pkg/env"
)

// ================================================================
// Fly.io CLI Wrapper
// ================================================================
// Self-contained Fly.io automation that models all flyctl commands
// Leverages registry knowledge for forward engineering (secrets, config)

// ================================================================
// Installation
// ================================================================

// Install installs flyctl via go install
func Install() error {
	fmt.Println("📦 Installing flyctl...")
	cmd := exec.Command("go", "install", "github.com/superfly/flyctl@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install flyctl: %w", err)
	}
	fmt.Println("✅ flyctl installed successfully")
	return nil
}

// ================================================================
// Auth
// ================================================================

// AuthWhoami checks who is currently logged in
func AuthWhoami() error {
	cmd := exec.Command("flyctl", "auth", "whoami")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// AuthLogin performs interactive browser login
func AuthLogin() error {
	fmt.Println("🔐 Opening browser for Fly.io login...")
	cmd := exec.Command("flyctl", "auth", "login")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// AuthLogout logs out the current user
func AuthLogout() error {
	cmd := exec.Command("flyctl", "auth", "logout")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ================================================================
// Apps
// ================================================================

// AppsCreate creates a new Fly.io app
func AppsCreate(name, org string) error {
	args := []string{"apps", "create"}
	if name != "" {
		args = append(args, name)
	}
	if org != "" {
		args = append(args, "--org", org)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// AppsList lists all apps
func AppsList() error {
	cmd := exec.Command("flyctl", "apps", "list")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// AppsDestroy destroys an app (WARNING: destructive)
func AppsDestroy(name string) error {
	cmd := exec.Command("flyctl", "apps", "destroy", name, "--yes")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// AppExists checks if an app exists
func AppExists(name string) (bool, error) {
	cmd := exec.Command("flyctl", "apps", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(output), name), nil
}

// ================================================================
// Launch
// ================================================================

// Launch launches a new app (creates fly.toml if missing)
func Launch(noDeploy bool) error {
	args := []string{"launch"}
	if noDeploy {
		args = append(args, "--no-deploy")
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// ================================================================
// Volumes
// ================================================================

// VolumesCreate creates a persistent volume
func VolumesCreate(volumeName, app, region string, sizeGB int) error {
	args := []string{"volumes", "create", volumeName}
	if app != "" {
		args = append(args, "--app", app)
	}
	if region != "" {
		args = append(args, "--region", region)
	}
	if sizeGB > 0 {
		args = append(args, "--size", fmt.Sprintf("%d", sizeGB))
	}
	args = append(args, "--yes")

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// VolumesList lists volumes for an app
func VolumesList(app string) error {
	args := []string{"volumes", "list"}
	if app != "" {
		args = append(args, "--app", app)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// VolumesDestroy destroys a volume
func VolumesDestroy(volumeID string) error {
	cmd := exec.Command("flyctl", "volumes", "destroy", volumeID, "--yes")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ================================================================
// Secrets (FORWARD ENGINEERING - uses registry knowledge)
// ================================================================

// SecretsImport imports secrets from .env file using registry knowledge
// This is the FORWARD ENGINEERING approach:
//  1. Registry defines which vars are secrets
//  2. Load values from envFilePath
//  3. Generate NAME=VALUE format for ONLY secret vars
//  4. Pipe to flyctl secrets import
func SecretsImport(registry *env.Registry, envFilePath, app string) error {
	// 1. Load environment file
	secrets, err := env.LoadSecrets(env.SecretsSource{
		FilePath:     envFilePath,
		PreferEncrypted: true, // Try .age file if plaintext missing
	})
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", envFilePath, err)
	}

	// 2. Get secret vars from registry
	secretVars := registry.GetSecrets()
	if len(secretVars) == 0 {
		fmt.Println("💡 No secret variables defined in registry")
		return nil
	}

	// 3. Build NAME=VALUE format for secrets only
	var lines []string
	for _, v := range secretVars {
		value, exists := secrets[v.Name]
		if !exists || value == "" {
			fmt.Printf("⚠️  Warning: Secret %s not found in %s\n", v.Name, envFilePath)
			continue
		}
		lines = append(lines, fmt.Sprintf("%s=%s", v.Name, value))
	}

	if len(lines) == 0 {
		return fmt.Errorf("no secret values found in %s", envFilePath)
	}

	// 4. Pipe to flyctl secrets import
	secretsInput := strings.Join(lines, "\n")

	args := []string{"secrets", "import"}
	if app != "" {
		args = append(args, "--app", app)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdin = strings.NewReader(secretsInput)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("🔐 Importing %d secrets to Fly.io...\n", len(lines))
	return cmd.Run()
}

// SecretsList lists all secrets for an app
func SecretsList(app string) error {
	args := []string{"secrets", "list"}
	if app != "" {
		args = append(args, "--app", app)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// SecretsUnset removes a secret
func SecretsUnset(app, name string) error {
	args := []string{"secrets", "unset", name}
	if app != "" {
		args = append(args, "--app", app)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ================================================================
// Deploy
// ================================================================

// Deploy deploys the app
func Deploy(app string) error {
	args := []string{"deploy"}
	if app != "" {
		args = append(args, "--app", app)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Status shows app status
func Status(app string) error {
	args := []string{"status"}
	if app != "" {
		args = append(args, "--app", app)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Logs tails app logs
func Logs(app string, follow bool) error {
	args := []string{"logs"}
	if app != "" {
		args = append(args, "--app", app)
	}
	if follow {
		args = append(args, "-f")
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ================================================================
// SSH
// ================================================================

// SSH opens SSH console to app machine
func SSH(app string) error {
	args := []string{"ssh", "console"}
	if app != "" {
		args = append(args, "--app", app)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// Open opens the app in a browser
func Open(app string) error {
	args := []string{"open"}
	if app != "" {
		args = append(args, "--app", app)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ================================================================
// Helpers (FORWARD ENGINEERING)
// ================================================================

// ReadFlyTomlConfig reads fly.toml and extracts app name and region
func ReadFlyTomlConfig() (appName, region string, err error) {
	file, err := os.Open("fly.toml")
	if err != nil {
		return "", "", fmt.Errorf("failed to open fly.toml: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Parse: app = 'env-demo'
		if strings.HasPrefix(line, "app") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				appName = strings.Trim(strings.TrimSpace(parts[1]), "'\"")
			}
		}

		// Parse: primary_region = 'sjc'
		if strings.HasPrefix(line, "primary_region") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				region = strings.Trim(strings.TrimSpace(parts[1]), "'\"")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("error reading fly.toml: %w", err)
	}

	return appName, region, nil
}

// ExportSecretsForFly exports secrets in NAME=VALUE format for Fly.io
// Uses registry knowledge to know which vars are secrets
func ExportSecretsForFly(registry *env.Registry, envFilePath string) (string, error) {
	// Load environment file
	secrets, err := env.LoadSecrets(env.SecretsSource{
		FilePath:     envFilePath,
		PreferEncrypted: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to load %s: %w", envFilePath, err)
	}

	// Get secret vars from registry
	secretVars := registry.GetSecrets()
	if len(secretVars) == 0 {
		return "", fmt.Errorf("no secret variables defined in registry")
	}

	// Build NAME=VALUE format
	var buf bytes.Buffer
	for _, v := range secretVars {
		value, exists := secrets[v.Name]
		if !exists || value == "" {
			continue
		}
		buf.WriteString(fmt.Sprintf("%s=%s\n", v.Name, value))
	}

	return buf.String(), nil
}
