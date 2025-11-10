# PDF Form System - Deployment Guide

## Data Directory Convention

The system uses `.data/` as the base directory for all runtime data. This is a standard convention for hidden data directories and works seamlessly across:
- Local development
- Docker containers
- Fly.io deployment
- Any cloud provider

## Directory Structure

```
.data/
├── catalog/              # Form catalogs (tracked in git)
│   └── australian_transfer_forms.csv
├── downloads/            # Downloaded PDFs (gitignored)
├── templates/            # Field templates (gitignored)
├── outputs/              # Filled PDFs (gitignored)
├── cases/                # Case management
│   ├── test_scenarios/   # Test cases (tracked in git)
│   └── {entity}/         # Entity cases (gitignored)
└── temp/                 # Temporary files (gitignored)
```

## Configuration

### Environment Variables

- `PDFFORM_DATA_DIR` - Override the data directory path
  - Local: `./.data`
  - Docker: `/app/.data`
  - Fly.io: `/app/.data` (mounted volume)

### Configuration Priority

1. **ENV variable** `PDFFORM_DATA_DIR` (highest priority)
2. **Docker detection** - `/app/.data` if `/app` exists
3. **Relative path** - `../../.data` for local development

## Local Development

### Setup

```bash
# Clone repository
git clone https://github.com/joeblew999/wellknown.git
cd wellknown/pkg/pdf

# Data directory is created automatically on first run
# Or create manually:
mkdir -p .data/{catalog,downloads,templates,outputs,cases/test_scenarios,temp}

# Build
make build-pdfform

# Run CLI
./.bin/pdfform 1-browse

# Run web server
./.bin/pdfform serve --port 8080
```

### Custom Data Directory

```bash
# Set custom location
export PDFFORM_DATA_DIR=/path/to/custom/data

# Run
./.bin/pdfform serve
```

## Docker Deployment

### Build Image

```bash
cd pkg/pdf
docker build -t pdfform:latest .
```

### Run Container

```bash
# With volume for persistence
docker run -d \
  --name pdfform \
  -p 8080:8080 \
  -v pdfform-data:/app/.data \
  pdfform:latest

# With local .data mount (development)
docker run -d \
  --name pdfform \
  -p 8080:8080 \
  -v $(pwd)/.data:/app/.data \
  pdfform:latest
```

### Docker Compose

```bash
# Start service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop service
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### Docker Configuration

The Dockerfile:
- Uses multi-stage build for minimal image size
- Runs as non-root user (UID 1000)
- Exposes port 8080
- Mounts `/app/.data` as volume
- Copies catalog file into image
- Sets proper permissions

## Fly.io Deployment

### Prerequisites

```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Login
flyctl auth login
```

### Initial Setup

```bash
cd pkg/pdf

# Create app (only once)
flyctl apps create pdfform

# Deploy
flyctl deploy
```

### Volume Management

The app uses a persistent volume for `.data/`:

```bash
# Volume is automatically created from fly.toml
# Initial size: 1GB

# Scale volume if needed
flyctl volumes extend <volume-id> -s 5

# List volumes
flyctl volumes list

# Create snapshot
flyctl volumes snapshots create <volume-id>
```

### Configuration

`fly.toml` settings:
- **Region**: `syd` (Sydney) - change as needed
- **Memory**: 512MB
- **CPU**: 1 shared core
- **Auto-stop**: Enabled (saves costs)
- **Volume**: `/app/.data` mounted from `pdfform_data`
- **Health check**: GET / every 30s

### Deployment Commands

```bash
# Deploy
flyctl deploy

# View logs
flyctl logs

# SSH into container
flyctl ssh console

# Check status
flyctl status

# Scale (add more instances)
flyctl scale count 2

# Open in browser
flyctl open
```

### Environment Variables

```bash
# Set environment variable
flyctl secrets set PDFFORM_DATA_DIR=/app/.data

# List secrets
flyctl secrets list
```

## Data Management

### Backup

#### Docker

```bash
# Create backup
docker run --rm \
  -v pdfform-data:/data \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/pdfform-backup-$(date +%Y%m%d).tar.gz -C /data .

# Restore backup
docker run --rm \
  -v pdfform-data:/data \
  -v $(pwd)/backups:/backup \
  alpine tar xzf /backup/pdfform-backup-20250110.tar.gz -C /data
```

#### Fly.io

```bash
# Create volume snapshot
flyctl volumes snapshots create <volume-id>

# List snapshots
flyctl volumes snapshots list <volume-id>

# Restore from snapshot (create new volume)
flyctl volumes create pdfform_data --snapshot-id <snapshot-id>
```

### Populate Catalog

```bash
# Copy catalog to running container
docker cp .data/catalog/australian_transfer_forms.csv pdfform:/app/.data/catalog/

# Or via volume mount (recommended for initial setup)
cp australian_transfer_forms.csv .data/catalog/
```

## Monitoring

### Health Checks

The application exposes health endpoints:
- `GET /` - Home page (200 OK = healthy)
- Docker: Automatic health check configured
- Fly.io: HTTP check every 30s

### Logs

```bash
# Docker
docker logs -f pdfform

# Docker Compose
docker-compose logs -f

# Fly.io
flyctl logs
flyctl logs -f  # Follow mode
```

### Metrics

```bash
# Fly.io metrics
flyctl metrics

# Docker stats
docker stats pdfform
```

## Scaling

### Vertical Scaling (More Resources)

```bash
# Fly.io - increase memory
flyctl scale memory 1024

# Fly.io - increase CPU
flyctl scale vm shared-cpu-2x

# Docker - set resource limits
docker run -d \
  --memory="1g" \
  --cpus="2" \
  pdfform:latest
```

### Horizontal Scaling (More Instances)

```bash
# Fly.io - add instances
flyctl scale count 3

# Docker - run multiple containers (need load balancer)
docker run -d --name pdfform-1 -p 8081:8080 pdfform:latest
docker run -d --name pdfform-2 -p 8082:8080 pdfform:latest
```

## Security

### HTTPS

- **Fly.io**: Automatic HTTPS with Let's Encrypt
- **Docker**: Use reverse proxy (nginx, caddy)

### Firewall

```bash
# Fly.io - restrict access by IP
flyctl ips allocate-v4  # Get static IP
# Configure firewall rules in Fly.io dashboard
```

### Non-Root User

The Docker image runs as UID 1000 (non-root) for security.

## Troubleshooting

### Data Directory Not Found

```bash
# Check directory exists
docker exec pdfform ls -la /app/.data

# Check permissions
docker exec pdfform ls -ld /app/.data

# Recreate directories
docker exec pdfform mkdir -p /app/.data/{catalog,downloads,templates,outputs,cases,temp}
```

### Volume Mount Issues

```bash
# Check volume
docker volume inspect pdfform-data

# Remove and recreate
docker-compose down -v
docker-compose up -d
```

### Permission Denied

```bash
# Fix ownership (Docker)
docker exec -u root pdfform chown -R pdfform:pdfform /app/.data

# Fix permissions
docker exec -u root pdfform chmod -R 755 /app/.data
```

### Out of Disk Space

```bash
# Fly.io - extend volume
flyctl volumes extend <volume-id> -s 5

# Docker - clean up
docker system prune -a
docker volume prune
```

## Cost Optimization

### Fly.io

- **Auto-stop**: Enabled in fly.toml (free tier friendly)
- **Shared CPU**: Cost-effective for low traffic
- **Minimal instances**: Start with 1, scale as needed
- **Volume size**: Start with 1GB, extend when needed

### Docker

- **Multi-stage build**: Smaller images = faster deploys
- **Alpine base**: Minimal runtime image
- **Resource limits**: Prevent resource exhaustion

## Migration

### From data/ to .data/

Already handled in code. Both paths work:
- Old: `data/`
- New: `.data/` (preferred)

The system automatically uses the new convention.

## Updates

### Rolling Updates (Fly.io)

```bash
# Deploy with zero downtime
flyctl deploy --strategy rolling

# Canary deployment
flyctl deploy --strategy canary
```

### Blue-Green Deployment (Docker)

```bash
# Deploy v2 on different port
docker run -d --name pdfform-v2 -p 8081:8080 pdfform:v2

# Test v2
curl http://localhost:8081

# Swap ports (update load balancer)
# Remove old version
docker stop pdfform-v1
docker rm pdfform-v1
```

## Support

### Logs Location

- Docker: `docker logs pdfform`
- Fly.io: `flyctl logs`
- Local: stdout

### Debug Mode

```bash
# Set log level (if implemented)
export LOG_LEVEL=debug

# Fly.io
flyctl secrets set LOG_LEVEL=debug
```

## Checklist

### Before Deployment

- [ ] Build succeeds locally
- [ ] Tests pass
- [ ] Catalog file exists
- [ ] .data directories created
- [ ] Environment variables set
- [ ] Secrets configured (if any)

### After Deployment

- [ ] Health check passes
- [ ] Can browse forms
- [ ] Can download PDFs
- [ ] Volume persists across restarts
- [ ] Backups configured
- [ ] Monitoring set up
- [ ] HTTPS working (production)
