package deploy

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ================================================================
// Ko CLI Wrapper
// ================================================================
// Ko is a fast container builder for Go applications
// https://ko.build
//
// Ko produces minimal, distroless images without needing a Dockerfile
// Typical image sizes: 10-15MB vs 50-100MB+ with traditional Docker builds

// ================================================================
// Installation
// ================================================================

// InstallKo installs ko via go install
func InstallKo() error {
	fmt.Println("📦 Installing ko...")
	cmd := exec.Command("go", "install", "github.com/google/ko@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install ko: %w", err)
	}
	fmt.Println("✅ ko installed successfully")
	return nil
}

// ================================================================
// Build
// ================================================================

// KoBuildOptions configures a ko build
type KoBuildOptions struct {
	// Path to build (usually "." for current directory)
	Path string

	// Tag for the image (default: "latest")
	Tag string

	// BaseImage overrides the default base image
	// Default: cgr.dev/chainguard/static:latest (from .ko.yaml)
	BaseImage string

	// Platforms to build for (e.g., "linux/amd64,linux/arm64")
	// Default: current platform
	Platforms string

	// Push to registry instead of loading locally
	Push bool

	// Registry to push to (e.g., "registry.fly.io/myapp")
	// Only used if Push is true
	Registry string

	// Local loads image into local Docker daemon
	// Default: true (for local development)
	Local bool

	// Env vars to set during build (e.g., CGO_ENABLED=0, GOWORK=off)
	Env map[string]string
}

// BuildLocal builds a Go application with ko and loads it into local Docker
// Returns the full image name (e.g., "ko.local/example-abc123:latest")
func BuildLocal(path string) (string, error) {
	return BuildWithKo(KoBuildOptions{
		Path:  path,
		Tag:   "latest",
		Local: true,
		Env: map[string]string{
			"CGO_ENABLED": "0",
			"GOWORK":      "off",
		},
	})
}

// BuildWithKo builds a container image using ko
// Returns the full image reference (e.g., "ko.local/example-abc123:latest")
func BuildWithKo(opts KoBuildOptions) (string, error) {
	// Validate path
	if opts.Path == "" {
		opts.Path = "."
	}

	// Build command
	args := []string{"build"}

	// Local vs Push
	if opts.Local {
		args = append(args, "--local")
	} else if opts.Push {
		args = append(args, "--push")
		if opts.Registry != "" {
			// Set KO_DOCKER_REPO via env
			if opts.Env == nil {
				opts.Env = make(map[string]string)
			}
			opts.Env["KO_DOCKER_REPO"] = opts.Registry
		}
	}

	// Tag
	if opts.Tag != "" {
		args = append(args, "-t", opts.Tag)
	}

	// Base image override
	if opts.BaseImage != "" {
		args = append(args, "--base-import-paths", "--base-image", opts.BaseImage)
	}

	// Platforms
	if opts.Platforms != "" {
		args = append(args, "--platform", opts.Platforms)
	}

	// Path to build
	args = append(args, opts.Path)

	// Create command
	cmd := exec.Command("ko", args...)

	// Set environment variables
	cmd.Env = os.Environ()
	if opts.Env != nil {
		for key, value := range opts.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Capture output to extract image name
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run build
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ko build failed: %w\nstderr: %s", err, stderr.String())
	}

	// Extract image name from output
	// Ko prints the image reference on the last line
	imageName := strings.TrimSpace(stdout.String())
	lines := strings.Split(imageName, "\n")
	if len(lines) > 0 {
		imageName = strings.TrimSpace(lines[len(lines)-1])
	}

	if imageName == "" {
		return "", fmt.Errorf("failed to extract image name from ko output")
	}

	return imageName, nil
}

// ================================================================
// Info & Utilities
// ================================================================

// KoVersion returns the installed ko version
func KoVersion() (string, error) {
	cmd := exec.Command("ko", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get ko version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// KoInstalled checks if ko is installed
func KoInstalled() bool {
	_, err := exec.LookPath("ko")
	return err == nil
}

// ParseKoOutput extracts the image name from ko build output
// Ko outputs progress lines followed by the final image reference
func ParseKoOutput(output string) string {
	// Ko prints the image reference on the last line
	scanner := bufio.NewScanner(strings.NewReader(output))
	var lastLine string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lastLine = line
		}
	}
	return lastLine
}

// GetImageShortName extracts a short name from full ko image reference
// Example: "ko.local/example-abc123:latest" → "ko.local/example-abc123:latest"
// Example: "registry.fly.io/myapp@sha256:..." → "registry.fly.io/myapp@sha256:..."
func GetImageShortName(fullRef string) string {
	// Ko references are already reasonably short
	// Just return as-is for docker-compose usage
	return fullRef
}
