# Using pkg/env as a Library

This guide shows how to use `pkg/env` and `pkg/env/workflow` as a library in your own Go projects.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Basic Usage](#basic-usage)
- [Workflow Functions](#workflow-functions)
- [Advanced Usage](#advanced-usage)
- [Testing Your Code](#testing-your-code)
- [Complete Example](#complete-example)

## Installation

```bash
go get github.com/joeblew999/wellknown/pkg/env
```

## Quick Start

The simplest way to use the library:

```go
package main

import (
    "fmt"
    "github.com/joeblew999/wellknown/pkg/env"
)

func main() {
    // 1. Define your environment variables
    registry := env.NewRegistry([]env.EnvVar{
        {
            Name:        "API_KEY",
            Description: "Your API key",
            Secret:      true,
            Required:    true,
            Group:       "API",
        },
        {
            Name:        "PORT",
            Description: "Server port",
            Default:     "8080",
            Group:       "Server",
        },
    })

    // 2. Access values
    apiKey := registry.ByName("API_KEY").GetString()
    port := registry.ByName("PORT").GetInt()

    fmt.Printf("Starting server on port %d\n", port)
}
```

## Core Concepts

### Registry-Driven Architecture

The library uses a **single source of truth** pattern:

```
Registry → Templates → Environment Files → Deployment
```

1. **Registry**: Define ALL environment variables once
2. **Templates**: Generate .env files from registry
3. **Secrets Management**: Separate secrets from public config
4. **Encryption**: Use Age encryption for secure storage

### Key Components

- **Registry**: Central definition of all environment variables
- **EnvVar**: Individual variable with metadata (name, description, default, secret, required)
- **Environment**: Represents a .env file (local, production, secrets)
- **Workflow**: High-level orchestration functions

## Basic Usage

### 1. Creating a Registry

```go
package myapp

import "github.com/joeblew999/wellknown/pkg/env"

var AppRegistry = env.NewRegistry([]env.EnvVar{
    // Server Configuration
    {
        Name:        "SERVER_HOST",
        Description: "Server hostname",
        Default:     "localhost",
        Group:       "Server",
    },
    {
        Name:        "SERVER_PORT",
        Description: "Server port",
        Default:     "8080",
        Group:       "Server",
    },

    // Database Configuration
    {
        Name:        "DATABASE_URL",
        Description: "PostgreSQL connection string",
        Secret:      true,
        Required:    true,
        Group:       "Database",
    },
    {
        Name:        "DATABASE_MAX_CONNECTIONS",
        Description: "Maximum database connections",
        Default:     "25",
        Group:       "Database",
    },

    // API Keys
    {
        Name:        "STRIPE_SECRET_KEY",
        Description: "Stripe API secret key",
        Secret:      true,
        Required:    true,
        Group:       "Payment",
    },
})
```

### 2. Accessing Values

```go
// String values
host := AppRegistry.ByName("SERVER_HOST").GetString()

// Integer values
port := AppRegistry.ByName("SERVER_PORT").GetInt()

// Boolean values (supports: true, false, 1, 0, yes, no)
debug := AppRegistry.ByName("DEBUG_MODE").GetBool()

// Check if set
if dbURL := AppRegistry.ByName("DATABASE_URL").GetString(); dbURL != "" {
    // Connect to database
}
```

### 3. Generating Environment Files

```go
// Generate .env.local template
content := env.Local.Generate(AppRegistry, "My Application")
os.WriteFile(env.Local.FileName, []byte(content), 0600)

// Generate .env.production template
content = env.Production.Generate(AppRegistry, "My Application")
os.WriteFile(env.Production.FileName, []byte(content), 0600)
```

### 4. Loading Secrets

```go
// Load secrets from .env.secrets.local (prefers encrypted .age version)
secrets, err := env.LoadSecrets(env.SecretsSource{
    FilePath:        env.SecretsLocal.FileName,
    PreferEncrypted: true,
})
if err != nil {
    log.Fatal(err)
}

// Access loaded secrets
stripeKey := secrets["STRIPE_SECRET_KEY"]
```

### 5. Validation

```go
// Validate that all required variables are set
if err := AppRegistry.ValidateRequired(); err != nil {
    log.Fatalf("Missing required environment variables: %v", err)
}
```

## Workflow Functions

The `pkg/env/workflow` package provides high-level orchestration functions.

### 1. Registry Sync Workflow

Syncs registry changes to all configuration files:

```go
import "github.com/joeblew999/wellknown/pkg/env/workflow"

result, err := workflow.SyncRegistryWorkflow(workflow.RegistrySyncOptions{
    Registry:           AppRegistry,
    AppName:            "My Application",
    CreateSecretsFiles: true,
    DeploymentConfigs: []workflow.DeploymentConfig{
        {
            FilePath:    "Dockerfile",
            StartMarker: "# === AUTO-GENERATED ===",
            EndMarker:   "# === END ===",
            Generator: func(r *env.Registry) (string, error) {
                return r.GenerateDockerfileDocs(env.DockerfileDocsOptions{}), nil
            },
        },
    },
})

if err != nil {
    log.Fatal(err)
}

// Check results
fmt.Printf("Updated: %v\n", result.UpdatedFiles)
fmt.Printf("Generated: %v\n", result.GeneratedFiles)
fmt.Printf("Warnings: %v\n", result.Warnings)
```

### 2. Environments Sync Workflow

Merges secrets into environment templates:

```go
result, err := workflow.SyncEnvironmentsWorkflow(workflow.EnvironmentsSyncOptions{
    Registry:          AppRegistry,
    AppName:           "My Application",
    LocalEnv:          env.Local,
    ProductionEnv:     env.Production,
    ValidateRequired:  true,
})

if err != nil {
    log.Fatal(err)
}

// Check if validation passed
if result.HasWarnings() {
    for _, warn := range result.Warnings {
        log.Printf("Warning: %s", warn)
    }
}
```

### 3. Finalize Workflow

Encrypts environment files and optionally stages them in git:

```go
result, err := workflow.FinalizeWorkflow(workflow.FinalizeOptions{
    Environments:      env.AllEnvironmentFiles(),
    EncryptionKeyPath: ".age/key.txt",
    GitAdd:            true,
})

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Encrypted: %v\n", result.GeneratedFiles)
```

## Advanced Usage

### Custom Environment Files

```go
// Define custom environment
customEnv := &env.Environment{
    Name:     "staging",
    FileName: ".env.staging",
    BaseDir:  ".",
}

// Generate template
content := customEnv.Generate(AppRegistry, "My App - Staging")
os.WriteFile(customEnv.FullPath(), []byte(content), 0600)
```

### Custom Base Directory

```go
// Use different directory for environment files
projectEnv := env.Local.WithBaseDir("./config")

// Now operates in ./config directory
content := projectEnv.Generate(AppRegistry, "My App")
os.WriteFile(projectEnv.FullPath(), []byte(content), 0600)
```

### Filtering Variables

```go
// Get only secrets
secrets := AppRegistry.GetSecrets()
for _, secret := range secrets {
    fmt.Printf("Secret: %s\n", secret.Name)
}

// Get only required variables
required := AppRegistry.GetRequired()
for _, req := range required {
    fmt.Printf("Required: %s\n", req.Name)
}

// Group by category
groups := AppRegistry.GetByGroup()
for groupName, vars := range groups {
    fmt.Printf("Group: %s (%d variables)\n", groupName, len(vars))
    for _, v := range vars {
        fmt.Printf("  - %s\n", v.Name)
    }
}
```

### Syncing Deployment Configs

```go
// Sync Dockerfile
err := env.SyncFileSection(env.SyncOptions{
    FilePath:    "Dockerfile",
    StartMarker: "# === AUTO-GENERATED ===",
    EndMarker:   "# === END ===",
    Content:     AppRegistry.GenerateDockerfileDocs(env.DockerfileDocsOptions{}),
})

// Sync docker-compose.yml
err = env.SyncFileSection(env.SyncOptions{
    FilePath:    "docker-compose.yml",
    StartMarker: "# === AUTO-GENERATED ===",
    EndMarker:   "# === END ===",
    Content:     AppRegistry.GenerateDockerComposeEnv([]string{}),
})

// Sync fly.toml
tomlEnv := AppRegistry.GenerateTOMLEnv("env", []string{})
tomlSecrets := AppRegistry.GenerateTOMLSecretsList("secrets")
err = env.SyncFileSection(env.SyncOptions{
    FilePath:    "fly.toml",
    StartMarker: "# === AUTO-GENERATED ===",
    EndMarker:   "# === END ===",
    Content:     tomlEnv + "\n" + tomlSecrets,
})
```

### Custom Output Writers

All workflow functions support custom output writers for logging:

```go
var logBuf bytes.Buffer

result, err := workflow.SyncRegistryWorkflow(workflow.RegistrySyncOptions{
    Registry:     AppRegistry,
    AppName:      "My App",
    OutputWriter: &logBuf, // Capture workflow output
})

// Log output
fmt.Println(logBuf.String())
```

## Testing Your Code

### Mocking Environment Variables

```go
func TestMyFunction(t *testing.T) {
    // Set test environment
    os.Setenv("API_KEY", "test_key")
    os.Setenv("PORT", "9999")
    defer func() {
        os.Unsetenv("API_KEY")
        os.Unsetenv("PORT")
    }()

    // Test your code
    result := MyFunction()
    // assertions...
}
```

### Testing Workflows

```go
func TestWorkflow(t *testing.T) {
    // Create temp directory
    tmpDir, err := os.MkdirTemp("", "test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    // Change to temp dir
    origDir, _ := os.Getwd()
    defer os.Chdir(origDir)
    os.Chdir(tmpDir)

    // Create test registry
    registry := env.NewRegistry([]env.EnvVar{
        {Name: "TEST_VAR", Description: "Test", Default: "value"},
    })

    // Run workflow
    result, err := workflow.SyncRegistryWorkflow(workflow.RegistrySyncOptions{
        Registry: registry,
        AppName:  "Test App",
    })

    // Verify results
    if err != nil {
        t.Fatalf("Workflow failed: %v", err)
    }

    if len(result.UpdatedFiles) != 2 {
        t.Errorf("Expected 2 files, got %d", len(result.UpdatedFiles))
    }
}
```

## Complete Example

Here's a complete example showing all major features:

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/joeblew999/wellknown/pkg/env"
    "github.com/joeblew999/wellknown/pkg/env/workflow"
)

// Define your application's environment registry
var AppRegistry = env.NewRegistry([]env.EnvVar{
    // Server
    {
        Name:        "SERVER_HOST",
        Description: "Server hostname",
        Default:     "localhost",
        Group:       "Server",
    },
    {
        Name:        "SERVER_PORT",
        Description: "Server port",
        Default:     "8080",
        Group:       "Server",
    },

    // Database
    {
        Name:        "DATABASE_URL",
        Description: "PostgreSQL connection URL",
        Secret:      true,
        Required:    true,
        Group:       "Database",
    },

    // API
    {
        Name:        "API_KEY",
        Description: "External API key",
        Secret:      true,
        Required:    true,
        Group:       "API",
    },
})

func main() {
    // Phase 1: Initialize project (run once)
    if len(os.Args) > 1 && os.Args[1] == "init" {
        initProject()
        return
    }

    // Phase 2: Normal application startup
    startApplication()
}

func initProject() {
    fmt.Println("Initializing project environment...")

    // Sync registry to all configs
    result, err := workflow.SyncRegistryWorkflow(workflow.RegistrySyncOptions{
        Registry:           AppRegistry,
        AppName:            "My Application",
        CreateSecretsFiles: true,
        DeploymentConfigs: []workflow.DeploymentConfig{
            {
                FilePath:    "Dockerfile",
                StartMarker: "# === AUTO-GENERATED ===",
                EndMarker:   "# === END ===",
                Generator: func(r *env.Registry) (string, error) {
                    return r.GenerateDockerfileDocs(env.DockerfileDocsOptions{}), nil
                },
            },
        },
    })

    if err != nil {
        log.Fatalf("Failed to initialize: %v", err)
    }

    fmt.Printf("Created/updated: %v\n", result.UpdatedFiles)
    fmt.Printf("Generated: %v\n", result.GeneratedFiles)
    fmt.Println("\nNext steps:")
    fmt.Println("1. Edit .env.secrets.local and .env.secrets.production")
    fmt.Println("2. Run: go run . sync")
}

func startApplication() {
    // Validate required variables
    if err := AppRegistry.ValidateRequired(); err != nil {
        log.Fatalf("Environment validation failed: %v", err)
    }

    // Get configuration
    host := AppRegistry.ByName("SERVER_HOST").GetString()
    port := AppRegistry.ByName("SERVER_PORT").GetInt()

    // Start your application
    fmt.Printf("Starting server on %s:%d\n", host, port)
    // ... rest of your application code
}
```

## Best Practices

### 1. Define Registry in Dedicated File

Create `registry.go` in your project:

```go
package myapp

import "github.com/joeblew999/wellknown/pkg/env"

var AppRegistry = env.NewRegistry([]env.EnvVar{
    // Define all variables here
})
```

### 2. Group Related Variables

```go
env.EnvVar{
    Name:  "DATABASE_URL",
    Group: "Database", // Makes documentation clearer
}
```

### 3. Mark Secrets Appropriately

```go
env.EnvVar{
    Name:   "API_KEY",
    Secret: true, // Will be separated into secrets files
}
```

### 4. Set Sensible Defaults

```go
env.EnvVar{
    Name:    "LOG_LEVEL",
    Default: "info", // Development-friendly default
}
```

### 5. Validate on Startup

```go
func main() {
    if err := AppRegistry.ValidateRequired(); err != nil {
        log.Fatalf("Missing required config: %v", err)
    }
    // ... start application
}
```

### 6. Use Workflows for Automation

Don't write custom logic - use the provided workflow functions:

```go
// ✅ Good: Use workflow functions
result, err := workflow.SyncRegistryWorkflow(opts)

// ❌ Bad: Reimplementing workflow logic
// Custom code to sync files...
```

### 7. Keep CLI Separate from Library

- **Library code**: Pure functions, no `os.Exit()`, return errors
- **CLI code**: User-friendly output, handle errors, call `os.Exit()`

## Differences from CLI Usage

When using as a library vs. the CLI tool:

| Aspect | CLI Usage | Library Usage |
|--------|-----------|---------------|
| Error handling | Prints and exits | Returns errors |
| Output | Formatted with emojis | Plain or custom writer |
| Workflow | Step-by-step prompts | Single function calls |
| Configuration | Command-line args | Options structs |
| State management | File-based | In-memory + files |

## Integration Patterns

### Pattern 1: Initialization Script

```go
// tools/init-env.go
package main

import (
    "github.com/joeblew999/wellknown/pkg/env/workflow"
    "myapp/internal/config"
)

func main() {
    result, err := workflow.SyncRegistryWorkflow(workflow.RegistrySyncOptions{
        Registry: config.AppRegistry,
        AppName:  "MyApp",
        CreateSecretsFiles: true,
    })
    // Handle result...
}
```

### Pattern 2: Runtime Validation

```go
// internal/config/config.go
package config

import "github.com/joeblew999/wellknown/pkg/env"

func MustLoad() *Config {
    if err := AppRegistry.ValidateRequired(); err != nil {
        panic(err)
    }

    return &Config{
        Host: AppRegistry.ByName("SERVER_HOST").GetString(),
        Port: AppRegistry.ByName("SERVER_PORT").GetInt(),
    }
}
```

### Pattern 3: Build-Time Generation

```go
//go:generate go run tools/generate-env.go

// tools/generate-env.go
package main

import "myapp/internal/config"

func main() {
    // Generate docs
    content := config.AppRegistry.GenerateDockerfileDocs(...)
    // Write to files...
}
```

## Further Reading

- See `example/` directory for a complete working CLI example
- Check `workflow/` tests for advanced usage patterns
- Read `WORKFLOW.md` for understanding the 3-phase workflow pattern

## Support

For issues, questions, or contributions:
- GitHub: https://github.com/joeblew999/wellknown
- Issues: https://github.com/joeblew999/wellknown/issues
