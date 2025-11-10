# Docker Quick Start

## Build

```bash
docker build -t pdfform:latest .
```

## Run

### Simple (no persistence)
```bash
docker run -p 8080:8080 pdfform:latest
```

### With persistence
```bash
docker run -d \
  --name pdfform \
  -p 8080:8080 \
  -v pdfform-data:/app/.data \
  pdfform:latest
```

### Development (mount local .data)
```bash
docker run -d \
  --name pdfform \
  -p 8080:8080 \
  -v $(pwd)/.data:/app/.data \
  -e PDFFORM_DATA_DIR=/app/.data \
  pdfform:latest
```

## Docker Compose

```bash
# Start
docker-compose up -d

# Logs
docker-compose logs -f

# Stop
docker-compose down
```

## Access

- Web GUI: http://localhost:8080
- API: http://localhost:8080/api/*

## CLI Commands in Docker

```bash
# Browse forms
docker exec pdfform /app/pdfform 1-browse

# List cases
docker exec pdfform /app/pdfform 5-test
```

## Data

Volume mount: `/app/.data`

```bash
# Backup
docker run --rm \
  -v pdfform-data:/data \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/pdfform-$(date +%Y%m%d).tar.gz -C /data .

# Restore
docker run --rm \
  -v pdfform-data:/data \
  -v $(pwd)/backups:/backup \
  alpine tar xzf /backup/pdfform-20250110.tar.gz -C /data
```

See [DEPLOYMENT.md](DEPLOYMENT.md) for full documentation.
