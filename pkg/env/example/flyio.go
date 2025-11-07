package main

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

// FlyInstall installs flyctl via go install
func FlyInstall() error {
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

// FlyAuthWhoami checks who is currently logged in
func FlyAuthWhoami() error {
	cmd := exec.Command("flyctl", "auth", "whoami")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// FlyAuthLogin performs interactive browser login
func FlyAuthLogin() error {
	fmt.Println("🔐 Opening browser for Fly.io login...")
	cmd := exec.Command("flyctl", "auth", "login")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// FlyAuthLogout logs out the current user
func FlyAuthLogout() error {
	cmd := exec.Command("flyctl", "auth", "logout")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ================================================================
// Apps
// ================================================================

// FlyAppsCreate creates a new Fly.io app
func FlyAppsCreate(name, org string) error {
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

// FlyAppsList lists all apps
func FlyAppsList() error {
	cmd := exec.Command("flyctl", "apps", "list")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// FlyAppsDestroy destroys an app (WARNING: destructive)
func FlyAppsDestroy(name string) error {
	cmd := exec.Command("flyctl", "apps", "destroy", name, "--yes")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// FlyAppExists checks if an app exists
func FlyAppExists(name string) (bool, error) {
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

// FlyLaunch launches a new app (creates fly.toml if missing)
func FlyLaunch(noDeploy bool) error {
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

// FlyVolumesCreate creates a persistent volume
func FlyVolumesCreate(volumeName, app, region string, sizeGB int) error {
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

// FlyVolumesList lists volumes for an app
func FlyVolumesList(app string) error {
	args := []string{"volumes", "list"}
	if app != "" {
		args = append(args, "--app", app)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// FlyVolumesDestroy destroys a volume
func FlyVolumesDestroy(volumeID string) error {
	cmd := exec.Command("flyctl", "volumes", "destroy", volumeID, "--yes")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ================================================================
// Secrets (FORWARD ENGINEERING - uses registry knowledge)
// ================================================================

// FlySecretsImport imports secrets from .env.local using registry knowledge
// This is the FORWARD ENGINEERING approach:
//  1. Registry defines which vars are secrets
//  2. Load values from .env.local
//  3. Generate NAME=VALUE format for ONLY secret vars
//  4. Pipe to flyctl secrets import
func FlySecretsImport(app string) error {
	// 1. Load .env.local
	secrets, err := env.LoadSecrets(env.SecretsSource{
		FilePath:     ".env.local",
		TryEncrypted: true, // Try .env.local.age if .env.local missing
	})
	if err != nil {
		return fmt.Errorf("failed to load .env.local: %w", err)
	}

	// 2. Get secret vars from registry
	secretVars := AppRegistry.GetSecrets()
	if len(secretVars) == 0 {
		fmt.Println("💡 No secret variables defined in registry")
		return nil
	}

	// 3. Build NAME=VALUE format for secrets only
	var lines []string
	for _, v := range secretVars {
		value, exists := secrets[v.Name]
		if !exists || value == "" {
			fmt.Printf("⚠️  Warning: Secret %s not found in .env.local\n", v.Name)
			continue
		}
		lines = append(lines, fmt.Sprintf("%s=%s", v.Name, value))
	}

	if len(lines) == 0 {
		return fmt.Errorf("no secret values found in .env.local")
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

// FlySecretsList lists all secrets for an app
func FlySecretsList(app string) error {
	args := []string{"secrets", "list"}
	if app != "" {
		args = append(args, "--app", app)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// FlySecretsUnset removes a secret
func FlySecretsUnset(app, name string) error {
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

// FlyDeploy deploys the app
func FlyDeploy(app string) error {
	args := []string{"deploy"}
	if app != "" {
		args = append(args, "--app", app)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// FlyStatus shows app status
func FlyStatus(app string) error {
	args := []string{"status"}
	if app != "" {
		args = append(args, "--app", app)
	}

	cmd := exec.Command("flyctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// FlyLogs tails app logs
func FlyLogs(app string, follow bool) error {
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

// FlySSH opens SSH console to app machine
func FlySSH(app string) error {
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

// ================================================================
// Helpers (FORWARD ENGINEERING)
// ================================================================

// ParseFlyToml reads fly.toml and extracts app name and region
func ParseFlyToml() (appName, region string, err error) {
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
func ExportSecretsForFly() (string, error) {
	// Load .env.local
	secrets, err := env.LoadSecrets(env.SecretsSource{
		FilePath:     ".env.local",
		TryEncrypted: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to load .env.local: %w", err)
	}

	// Get secret vars from registry
	secretVars := AppRegistry.GetSecrets()
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
