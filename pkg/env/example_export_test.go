package env_test

import (
	"fmt"
	"os"

	"github.com/joeblew999/wellknown/pkg/env"
)

// ExampleRegistry_Export demonstrates basic export
func ExampleRegistry_Export() {
	vars := []env.EnvVar{
		{Name: "SERVER_PORT"},
		{Name: "LOG_LEVEL"},
	}

	registry := env.NewRegistry(vars)

	// Set environment variables
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("LOG_LEVEL", "info")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("LOG_LEVEL")
	}()

	// Export in simple format
	output := registry.Export(env.ExportOptions{
		Format:       env.FormatSimple,
		IncludeEmpty: false,
	})

	fmt.Println(output)

	// Output:
	// SERVER_PORT=8080
	// LOG_LEVEL=info
}

// ExampleRegistry_ExportSimple shows the convenience method
func ExampleRegistry_ExportSimple() {
	vars := []env.EnvVar{
		{Name: "API_URL"},
		{Name: "TIMEOUT"},
	}

	registry := env.NewRegistry(vars)

	os.Setenv("API_URL", "https://api.example.com")
	os.Setenv("TIMEOUT", "30")
	defer func() {
		os.Unsetenv("API_URL")
		os.Unsetenv("TIMEOUT")
	}()

	// ExportSimple is a shorthand for simple format, non-empty values
	output := registry.ExportSimple()
	fmt.Println(output)

	// Output:
	// API_URL=https://api.example.com
	// TIMEOUT=30
}

// ExampleRegistry_ExportSecrets demonstrates exporting only secrets
func ExampleRegistry_ExportSecrets() {
	vars := []env.EnvVar{
		{Name: "PUBLIC_KEY", Secret: false},
		{Name: "PRIVATE_KEY", Secret: true},
		{Name: "API_SECRET", Secret: true},
	}

	registry := env.NewRegistry(vars)

	os.Setenv("PUBLIC_KEY", "public_value")
	os.Setenv("PRIVATE_KEY", "private_value")
	os.Setenv("API_SECRET", "secret_value")
	defer func() {
		os.Unsetenv("PUBLIC_KEY")
		os.Unsetenv("PRIVATE_KEY")
		os.Unsetenv("API_SECRET")
	}()

	// Export only secrets
	output := registry.ExportSecrets()
	fmt.Println(output)

	// Output:
	// PRIVATE_KEY=private_value
	// API_SECRET=secret_value
}

// ExampleRegistry_ExportRequired shows exporting required variables
func ExampleRegistry_ExportRequired() {
	vars := []env.EnvVar{
		{Name: "REQUIRED_VAR", Required: true},
		{Name: "OPTIONAL_VAR", Required: false},
	}

	registry := env.NewRegistry(vars)

	os.Setenv("REQUIRED_VAR", "value")
	os.Setenv("OPTIONAL_VAR", "optional")
	defer func() {
		os.Unsetenv("REQUIRED_VAR")
		os.Unsetenv("OPTIONAL_VAR")
	}()

	// Export only required variables (includes empty to show missing)
	output := registry.ExportRequired()
	fmt.Println(output)

	// Output:
	// REQUIRED_VAR=value
}

// ExampleRegistry_Export_maskSecrets shows masking secret values
func ExampleRegistry_Export_maskSecrets() {
	vars := []env.EnvVar{
		{Name: "PUBLIC_VAR", Secret: false},
		{Name: "SECRET_VAR", Secret: true},
	}

	registry := env.NewRegistry(vars)

	os.Setenv("PUBLIC_VAR", "visible")
	os.Setenv("SECRET_VAR", "hidden")
	defer func() {
		os.Unsetenv("PUBLIC_VAR")
		os.Unsetenv("SECRET_VAR")
	}()

	// Export with masked secrets
	output := registry.Export(env.ExportOptions{
		Format:       env.FormatSimple,
		MaskSecrets:  true,
		IncludeEmpty: false,
	})

	fmt.Println(output)

	// Output:
	// PUBLIC_VAR=visible
	// SECRET_VAR=***
}

// ExampleRegistry_ExportSystemd shows systemd format
func ExampleRegistry_ExportSystemd() {
	vars := []env.EnvVar{
		{Name: "SERVICE_PORT"},
		{Name: "SERVICE_NAME"},
	}

	registry := env.NewRegistry(vars)

	os.Setenv("SERVICE_PORT", "8080")
	os.Setenv("SERVICE_NAME", "myapp")
	defer func() {
		os.Unsetenv("SERVICE_PORT")
		os.Unsetenv("SERVICE_NAME")
	}()

	// Export in systemd format
	output := registry.ExportSystemd()
	fmt.Println(output)

	// Output:
	// Environment="SERVICE_PORT=8080"
	// Environment="SERVICE_NAME=myapp"
}

// ExampleRegistry_ExportK8s shows Kubernetes YAML format
func ExampleRegistry_ExportK8s() {
	vars := []env.EnvVar{
		{Name: "APP_NAME"},
		{Name: "APP_VERSION"},
	}

	registry := env.NewRegistry(vars)

	os.Setenv("APP_NAME", "myapp")
	os.Setenv("APP_VERSION", "1.0.0")
	defer func() {
		os.Unsetenv("APP_NAME")
		os.Unsetenv("APP_VERSION")
	}()

	// Export in Kubernetes format
	output := registry.ExportK8s()
	fmt.Println(output)

	// Output:
	// - name: APP_NAME
	//   value: "myapp"
	// - name: APP_VERSION
	//   value: "1.0.0"
}

// ExampleRegistry_Export_combinedFilters shows combining filters
func ExampleRegistry_Export_combinedFilters() {
	vars := []env.EnvVar{
		{Name: "PUBLIC_OPTIONAL", Secret: false, Required: false},
		{Name: "PUBLIC_REQUIRED", Secret: false, Required: true},
		{Name: "SECRET_OPTIONAL", Secret: true, Required: false},
		{Name: "SECRET_REQUIRED", Secret: true, Required: true},
	}

	registry := env.NewRegistry(vars)

	// Set all variables
	os.Setenv("PUBLIC_OPTIONAL", "value1")
	os.Setenv("PUBLIC_REQUIRED", "value2")
	os.Setenv("SECRET_OPTIONAL", "value3")
	os.Setenv("SECRET_REQUIRED", "value4")
	defer func() {
		os.Unsetenv("PUBLIC_OPTIONAL")
		os.Unsetenv("PUBLIC_REQUIRED")
		os.Unsetenv("SECRET_OPTIONAL")
		os.Unsetenv("SECRET_REQUIRED")
	}()

	// Export only required secrets
	output := registry.Export(env.ExportOptions{
		Format:       env.FormatSimple,
		SecretsOnly:  true,
		RequiredOnly: true,
		IncludeEmpty: false,
	})

	fmt.Println(output)

	// Output:
	// SECRET_REQUIRED=value4
}

// ExampleRegistry_Export_emptyValues shows including empty values
func ExampleRegistry_Export_emptyValues() {
	vars := []env.EnvVar{
		{Name: "SET_VAR"},
		{Name: "UNSET_VAR"},
	}

	registry := env.NewRegistry(vars)

	os.Setenv("SET_VAR", "value")
	os.Unsetenv("UNSET_VAR")
	defer os.Unsetenv("SET_VAR")

	// Without IncludeEmpty
	output1 := registry.Export(env.ExportOptions{
		Format:       env.FormatSimple,
		IncludeEmpty: false,
	})
	fmt.Println("Without empty:")
	fmt.Println(output1)

	// With IncludeEmpty
	output2 := registry.Export(env.ExportOptions{
		Format:       env.FormatSimple,
		IncludeEmpty: true,
	})
	fmt.Println("With empty:")
	fmt.Println(output2)

	// Output:
	// Without empty:
	// SET_VAR=value
	// With empty:
	// SET_VAR=value
	// UNSET_VAR=
}

// ExampleExportOptions shows configuring export options
func ExampleExportOptions() {
	// Configure export with multiple options
	opts := env.ExportOptions{
		Format:       env.FormatSystemd,
		SecretsOnly:  true,
		RequiredOnly: false,
		IncludeEmpty: false,
		MaskSecrets:  true,
	}

	fmt.Printf("Format: %s\n", opts.Format)
	fmt.Printf("Secrets only: %v\n", opts.SecretsOnly)
	fmt.Printf("Mask secrets: %v\n", opts.MaskSecrets)

	// Output:
	// Format: systemd
	// Secrets only: true
	// Mask secrets: true
}

// ExampleRegistry_Export_docker shows Docker format (same as simple)
func ExampleRegistry_Export_docker() {
	vars := []env.EnvVar{
		{Name: "CONTAINER_PORT"},
	}

	registry := env.NewRegistry(vars)

	os.Setenv("CONTAINER_PORT", "3000")
	defer os.Unsetenv("CONTAINER_PORT")

	// Docker format is same as simple
	output := registry.Export(env.ExportOptions{
		Format:       env.FormatDocker,
		IncludeEmpty: false,
	})

	fmt.Println(output)

	// Output:
	// CONTAINER_PORT=3000
}
