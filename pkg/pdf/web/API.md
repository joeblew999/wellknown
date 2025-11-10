# PDF Form API Documentation

The web server provides a REST API for all PDF form operations.

## Base URL
```
http://localhost:8080
```

## Web Pages (HTML)

### Home
```
GET /
```
Returns: 5-step workflow home page

### Browse Forms
```
GET /1-browse
```
Returns: Browse forms page with state selector

### Download Form
```
GET /2-download
```
Returns: Download form page

### Inspect Fields
```
GET /3-inspect
```
Returns: Inspect PDF fields page

### Fill Form
```
GET /4-fill
```
Returns: Fill form page with case management

### Test
```
GET /5-test
```
Returns: Test runner page

---

## API Endpoints (JSON)

### Browse Forms

**List All States**
```
GET /api/browse
```
Response:
```json
{
  "States": ["VIC", "NSW", "QLD", "SA", "WA", "TAS", "ACT", "NT"],
  "Forms": null
}
```

**List Forms by State**
```
GET /api/browse?state=QLD
```
Response:
```json
{
  "States": null,
  "Forms": [
    {
      "State": "QLD",
      "FormName": "Vehicle Registration Transfer Application",
      "FormCode": "F3520",
      "DirectPDFURL": "https://...",
      "OnlineAvailable": false
    }
  ]
}
```

---

### Download Form

```
POST /api/download
Content-Type: application/x-www-form-urlencoded

form_code=F3520
```

Response:
```json
{
  "PDFPath": "../../data/downloads/F3520.pdf",
  "Form": {
    "FormCode": "F3520",
    "FormName": "Vehicle Registration Transfer Application",
    "State": "QLD"
  },
  "Metadata": "../../data/downloads/F3520.pdf.meta.json"
}
```

---

### Inspect Form

```
POST /api/inspect
Content-Type: application/x-www-form-urlencoded

pdf_path=downloads/f3520.pdf
```

Response:
```json
{
  "TemplatePath": "../../data/templates/f3520_template.json",
  "FieldCount": 52,
  "Fields": ["Text1", "Text2", "Text3", ...]
}
```

---

### Case Management

**Create New Case**
```
POST /api/cases/create
Content-Type: application/x-www-form-urlencoded

form_code=F3520
case_name=Vehicle Sale 2025
entity_name=john_smith
```

Response:
```json
{
  "case": {
    "case_metadata": {
      "case_id": "john_smith_F3520_20251110_123456.789012",
      "case_name": "Vehicle Sale 2025",
      "created_at": "2025-11-10T12:34:56Z",
      "updated_at": "2025-11-10T12:34:56Z"
    },
    "form_reference": {
      "form_code": "F3520"
    },
    "fields": {}
  },
  "case_path": "../../data/cases/john_smith/john_smith_F3520_20251110_123456.789012.json"
}
```

**List All Cases**
```
GET /api/cases/list
```

Response:
```json
[
  {
    "path": "../../data/cases/john_smith/john_smith_F3520_*.json",
    "metadata": {
      "case_id": "john_smith_F3520_20251110_123456.789012",
      "case_name": "Vehicle Sale 2025",
      "created_at": "2025-11-10T12:34:56Z"
    },
    "form_code": "F3520"
  }
]
```

**List Cases by Entity**
```
GET /api/cases/list?entity=john_smith
```

**Load Case Data**
```
GET /api/cases/load?case_id=john_smith_F3520_20251110_123456.789012
```

Response:
```json
{
  "case_metadata": {
    "case_id": "john_smith_F3520_20251110_123456.789012",
    "case_name": "Vehicle Sale 2025",
    "created_at": "2025-11-10T12:34:56Z"
  },
  "form_reference": {
    "form_code": "F3520"
  },
  "fields": {
    "Text1": "John",
    "Text2": "Smith"
  }
}
```

---

### Fill Form from Case

```
POST /api/fill
Content-Type: application/x-www-form-urlencoded

case_id=john_smith_F3520_20251110_123456.789012
flatten=true
```

Response:
```json
{
  "OutputPath": "../../data/outputs/temp_data_filled.pdf",
  "InputPDF": "../../data/downloads/F3520.pdf",
  "Flattened": true
}
```

---

## Error Responses

All endpoints return appropriate HTTP status codes:
- `200 OK` - Success
- `400 Bad Request` - Missing or invalid parameters
- `404 Not Found` - Resource not found
- `405 Method Not Allowed` - Wrong HTTP method
- `500 Internal Server Error` - Server error

Error response format:
```
HTTP/1.1 400 Bad Request
Content-Type: text/plain

form_code is required
```

---

## Example Usage

### cURL Examples

**Browse forms for QLD:**
```bash
curl http://localhost:8080/api/browse?state=QLD
```

**Download form:**
```bash
curl -X POST http://localhost:8080/api/download \
  -d 'form_code=F3520'
```

**Create case:**
```bash
curl -X POST http://localhost:8080/api/cases/create \
  -d 'form_code=F3520' \
  -d 'case_name=My Vehicle Sale' \
  -d 'entity_name=john_smith'
```

**List all cases:**
```bash
curl http://localhost:8080/api/cases/list
```

**Fill form from case:**
```bash
curl -X POST http://localhost:8080/api/fill \
  -d 'case_id=john_smith_F3520_20251110_123456.789012' \
  -d 'flatten=true'
```

### JavaScript Examples

**Browse forms:**
```javascript
const response = await fetch('/api/browse?state=QLD');
const data = await response.json();
console.log(data.Forms);
```

**Create case:**
```javascript
const formData = new FormData();
formData.append('form_code', 'F3520');
formData.append('case_name', 'My Sale');
formData.append('entity_name', 'john_smith');

const response = await fetch('/api/cases/create', {
  method: 'POST',
  body: formData
});
const data = await response.json();
console.log(data.case.case_metadata.case_id);
```

---

## Server Configuration

**Start server:**
```bash
pdfform serve
pdfform serve --port 3000
```

**Default settings:**
- Port: 8080
- Data directory: `../../data`
- Templates: Embedded in binary
