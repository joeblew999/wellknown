# PDF Form Filler

```sh

make build
make serve
```

http://localhost:8080

Nice to see the SSE events:
http://127.0.0.1:8080/gui/events

A Go package and CLI tool for working with fillable PDF forms. This tool helps you:
- Extract form field names from any fillable PDF
- Fill PDFs with JSON data (supports both local PDFs and URLs)
- Flatten/lock PDFs to make them read-only
- Create reusable JSON templates for common forms (like VIC/NSW GTO forms, government forms, etc.)

## Features

- **Form Field Extraction**: Automatically discover all form fields in a PDF
- **JSON-based filling**: Use structured JSON to fill out forms programmatically
- **PDF URL Support**: Download and fill PDFs directly from URLs
- **Field Locking**: Flatten/lock filled PDFs to prevent further editing
- **Template Generation**: Create reusable JSON templates from any fillable PDF

## Quick Start

### 🎯 5-Step Numbered Workflow

The CLI provides a simple numbered workflow that guides you through the entire process:

```bash
cd examples/pdfform

# 1️⃣ Browse available government forms
./pdfform 1-browse                    # See all states
./pdfform 1-browse --state VIC        # See Victoria forms

# 2️⃣ Download a form
./pdfform 2-download F3520            # Download Queensland form

# 3️⃣ Inspect the form fields
./pdfform 3-inspect f3520.pdf         # Creates f3520_template.json

# 4️⃣ Edit the template (add your data), then fill
./pdfform 4-fill f3520_template.json           # Fill the form
./pdfform 4-fill f3520_template.json --flatten # Fill and lock fields
./pdfform 4-fill --test vba_basic              # Use a test case

# 5️⃣ Run automated tests (optional)
./pdfform 5-test                      # List all test cases
./pdfform 5-test vba_basic            # Run specific test
./pdfform 5-test --all                # Run all tests
```

Each step shows you what to do next, making it easy to follow the workflow!

### Alternative: Working with Your Own PDFs

If you have a PDF from another source (not in the catalog):

```bash
cd examples/pdfform

# Skip steps 1-2, start with inspect
./pdfform 3-inspect myform.pdf -o template.json

# Edit template.json and add your data

# Fill the PDF
./pdfform 4-fill template.json -o filled.pdf
```

## Usage

### 🎯 Numbered Workflow Commands (Recommended)

**Step 1: Browse Forms**
```bash
pdfform 1-browse                    # List all available forms
pdfform 1-browse --state VIC        # List forms for a specific state
```

**Step 2: Download Form**
```bash
pdfform 2-download F3520            # Download form by code
pdfform 2-download F3520 -o pdfs/   # Download to specific directory
```

**Step 3: Inspect Fields**
```bash
pdfform 3-inspect form.pdf          # Extract fields to <formname>_template.json
pdfform 3-inspect form.pdf -o out.json
```

**Step 4: Fill Form**
```bash
pdfform 4-fill data.json            # Fill the form
pdfform 4-fill data.json --flatten  # Fill and lock fields
pdfform 4-fill --test vba_basic     # Use a test case
```

**Step 5: Test Forms**
```bash
pdfform 5-test                      # List all test cases
pdfform 5-test vba_basic            # Run specific test
pdfform 5-test --all                # Run all tests
```

### JSON Format

The `pdf_url` field can be either:
- **HTTP/HTTPS URL**: Downloads the PDF from the internet
- **Local file path**: Uses a PDF file from your filesystem (absolute or relative path)

```json
{
  "pdf_url": "https://example.com/form.pdf",  // URL example
  "fields": {
    "field_name_1": "value1",
    "field_name_2": "value2"
  }
}
```

```json
{
  "pdf_url": "/path/to/local/form.pdf",  // Local file example
  "fields": {
    "field_name_1": "value1",
    "field_name_2": "value2"
  }
}
```

## Testing Framework

The tool includes a built-in testing framework for managing multiple test scenarios.

### Directory Structure
```
examples/pdfform/testdata/
├── pdfs/          # Test PDF files
├── cases/         # Test case JSON files
├── outputs/       # Generated output PDFs
└── README.md      # Testing documentation
```

### Test Case Format
```json
{
  "name": "mytest",
  "description": "Description of test",
  "pdf_url": "testdata/pdfs/form.pdf",
  "fields": {
    "FieldName": "value"
  },
  "expect_error": false
}
```

### CLI Testing Commands

**List all test cases:**
```bash
./pdfform 5-test
```

**Run specific test:**
```bash
./pdfform 5-test mytest
```

**Run all tests:**
```bash
./pdfform 5-test --all
```

**Use test case with fill command:**
```bash
./pdfform 4-fill --test mytest
```

### Writing Go Tests

```go
import "github.com/joeblew999/wellknown/pkg/pdf"

func TestMyForm(t *testing.T) {
    testCase, err := pdfform.LoadTestCase("testdata/cases/mytest.json")
    if err != nil {
        t.Fatal(err)
    }

    output, err := pdfform.RunTestCase(testCase, "testdata/outputs")
    if err != nil && !testCase.ExpectError {
        t.Errorf("Test failed: %v", err)
    }
}
```

See [examples/pdfform/testdata/README.md](examples/pdfform/testdata/README.md) for detailed testing documentation.

## Examples

See [examples/pdfform/](examples/pdfform/) for a complete working example.

## Australian Government Forms Catalog

The tool includes a built-in catalog of Australian government transfer forms (vehicle registration transfers for all states and territories).

### Browse Forms Catalog

**List all states:**
```bash
./pdfform 1-browse
```

**List forms for a specific state:**
```bash
./pdfform 1-browse --state VIC
./pdfform 1-browse --state NSW
```

**Download a form by code:**
```bash
./pdfform 2-download F3520        # Queensland form
./pdfform 2-download VRPIN00613   # Victoria form
```

### Forms Data

The catalog is stored in [australian_transfer_forms.csv](australian_transfer_forms.csv) and includes:
- All 8 Australian states and territories (VIC, NSW, QLD, SA, WA, TAS, ACT, NT)
- Direct PDF download URLs where available
- Information pages for each form
- Notes about online availability and deadlines

### Example: Download and Fill Government Form

```bash
# 1. Browse available forms
./pdfform 1-browse --state QLD

# 2. Download the form
./pdfform 2-download F3520 -o pdfs/

# 3. Inspect form fields
./pdfform 3-inspect pdfs/f3520.pdf

# 4. Edit f3520_template.json with your data

# 5. Fill the form
./pdfform 4-fill f3520_template.json --flatten
```

## Dual Library Support

This tool uses **two PDF libraries** with automatic fallback:

1. **pdfcpu** (primary) - Fast, feature-rich Go library
2. **benoitkugler/pdf** (fallback) - Better compatibility with signed/complex PDFs

When filling a PDF:
- Tries `pdfcpu` first (faster, more features)
- If `pdfcpu` fails, automatically falls back to `benoitkugler/pdf`
- Works with **signed PDFs** that pdfcpu can't handle!

### Example Output
```
⚠️  pdfcpu failed (pdfcpu: missing form data), trying benoitkugler library...
✅ Test passed: testdata/outputs/vba_basic_filled.pdf
```

## Limitations

- **Complex Forms**: Some advanced PDF features (JavaScript, calculations) may not work after filling
- **Signature Fields**: Digital signature fields are preserved but not signed



