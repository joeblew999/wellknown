# PDF Form Examples

This directory contains example programs showing how to use the PDF form library.

## Examples

### 1. Basic Usage ([basic-usage/](basic-usage/))

Shows the complete workflow:
- Browse forms catalog
- Download a PDF
- Inspect fields
- Fill the PDF

```bash
cd basic-usage
go run main.go
```

### 2. Configuration ([case-management/](case-management/))

Shows how the library auto-discovers the data directory:
- Environment variable detection
- Docker environment detection
- Parent directory search
- Configuration display

```bash
cd case-management
go run main.go
```

## Running the CLI

The main CLI application is in `cmd/pdfform/`:

```bash
# From pkg/pdf directory
cd cmd/pdfform
go run main.go 1-browse

# Or use the Makefile
make build-pdfform
.bin/pdfform --help
```

## Using as a Library

Import the package in your own code:

```go
import pdfform "github.com/joeblew999/wellknown/pkg/pdf"

// Configure - auto-discovers .data directory
dataDir := pdfform.FindDataDir()
config := pdfform.NewConfig(dataDir)
pdfform.SetDefaultConfig(config)

// Use commands
result, err := pdfform.Browse(pdfform.BrowseOptions{
    CatalogPath: config.CatalogFilePath(),
    State: "VIC",
})
```

## Web Server

Start the web server:

```bash
cd cmd/pdfform
go run main.go serve --port 8080
```

Or use the Makefile:

```bash
make run-pdfform-server
```

Then open http://localhost:8080

## Data Directory

All examples use `.data/` in the pkg/pdf directory:

```
pkg/pdf/.data/
├── catalog/              # Form catalogs
├── downloads/            # Downloaded PDFs
├── templates/            # Field templates
├── outputs/              # Filled PDFs
├── cases/                # Case files
└── temp/                 # Temporary files
```

This is automatically created when you run any example.
