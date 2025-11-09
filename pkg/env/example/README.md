# env-demo - Environment Management Demo

A working demonstration application showing the [github.com/joeblew999/wellknown/pkg/env](../) package in action. This app runs an HTTP server and demonstrates environment variable management across local and production environments with built-in secrets encryption.

## Quick Start

```bash
# Build the application
go build -o env-demo .

# View available commands
./env-demo help

# Start the HTTP server
./env-demo serve
```

Visit [http://localhost:8080](http://localhost:8080) to see the demo.

## What's This?

This is a **complete, working application** that demonstrates:

- ✅ Registry-driven environment management
- ✅ Separate local and production configurations
- ✅ Built-in secrets encryption (Age)
- ✅ Template-based env file generation
- ✅ Workflow automation for deployment
- ✅ HTTP server showcasing environment variable usage
- ✅ Ready for Fly.io deployment

## Available Commands

### HTTP Server

```bash
./env-demo serve    # Start HTTP server on $SERVER_PORT (default: 8080)
./env-demo health   # CLI health check
```

### Workflow Automation

```bash
./env-demo sync-registry      # Sync deployment configs and templates
./env-demo sync-environments  # Merge secrets into environments
./env-demo finalize           # Encrypt files for deployment
./env-demo ko-build           # Build with ko (fast 12MB Docker image)
```

## HTTP Endpoints

When running with `serve`, these endpoints are available:

- **GET /** - Homepage showing app status and environment info
- **GET /health** - Health check endpoint (JSON)
- **GET /env** - Environment variables showcase (hides secret values)
- **GET /feature-demo** - Feature flag demonstration (FEATURE_BETA)
- **GET /database** - Database connection status

## Workflow

The typical workflow for managing environments:

### 1. Define Your Variables

Edit `registry.go` to define your environment variables:

```go
var AppRegistry = env.NewRegistry([]env.Variable{
    {Key: "SERVER_PORT", DefaultValue: "8080", Group: "Server"},
    {Key: "DATABASE_URL", Secret: true, Required: true, Group: "Database"},
    // ... more variables
})
```

### 2. Sync Registry

```bash
./env-demo sync-registry
```

This generates:
- `.env.local` - Local environment template
- `.env.production` - Production environment template
- `.env.secrets.local` - Local secrets template
- `.env.secrets.production` - Production secrets template

### 3. Fill in Secrets

Edit the secrets files with actual values:

```bash
# .env.secrets.local
DATABASE_URL=postgresql://localhost/myapp_dev
STRIPE_API_KEY=sk_test_...

# .env.secrets.production
DATABASE_URL=postgresql://prod-host/myapp
STRIPE_API_KEY=sk_live_...
```

### 4. Sync Environments

```bash
./env-demo sync-environments
```

This merges secrets into the environment templates and validates required variables.

### 5. Finalize for Deployment

```bash
./env-demo finalize
```

This:
- Generates Age encryption key (if needed)
- Encrypts all environment files to `.age` format
- Adds encrypted files to git (safe to commit)

### 6. Deploy

```bash
# Deploy to Fly.io
flyctl deploy

# Or run locally with Docker
docker-compose up
```

## Deployment to Fly.io

This app is production-ready and can be deployed to Fly.io:

```bash
# 1. Install flyctl
go install github.com/superfly/flyctl@latest

# 2. Login
flyctl auth login

# 3. Launch app (reads fly.toml for name/region)
flyctl launch --no-deploy

# 4. Create volume
flyctl volumes create pb_data --region [your-region] --size 1

# 5. Import secrets (after running sync-environments)
flyctl secrets import < .env.production

# 6. Deploy
flyctl deploy

# 7. Check status
flyctl status

# 8. View logs
flyctl logs
```

## Configuration

The demo app uses these environment variables (defined in `registry.go`):

**Server:**
- `SERVER_PORT` - HTTP server port (default: 8080)
- `LOG_LEVEL` - Logging level (default: info)

**Database:**
- `DATABASE_URL` - Database connection URL (secret, required)

**APIs:**
- `STRIPE_API_KEY` - Stripe API key (secret, required)
- `SENDGRID_API_KEY` - SendGrid API key (secret, optional)

**Features:**
- `FEATURE_BETA` - Enable beta features (default: false)

## Files

- **main.go** - Simplified CLI (5 commands)
- **server.go** - HTTP server implementation
- **workflow.go** - Workflow command implementations
- **registry.go** - Environment variable registry (edit this!)
- **commands.go** - Low-level commands (preserved as reference, not used)
- **Dockerfile** - Multi-stage Docker build
- **fly.toml** - Fly.io deployment configuration
- **docker-compose.yml** - Local Docker development

## Development

### Local Development (No Docker)

```bash
# Run directly with Go (fastest iteration)
go run . serve
```

### Docker Development

```bash
# Option 1: Ko build (recommended - fast, 12MB image)
go run . ko-build                            # Build with ko
IMAGE=ko.local/example-...:latest docker-compose up

# Option 2: Dockerfile build (traditional)
docker-compose up --build                    # Build and run

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Building Binary

```bash
# Local binary
go build -o env-demo .

# Docker image with ko (12MB)
go run . ko-build

# Docker image with Dockerfile
docker build -t env-demo:latest .
```

### Running Tests

```bash
# Run tests (in parent pkg/env directory)
cd .. && go test ./...
```

## Documentation

- [WORKFLOW.md](./WORKFLOW.md) - Detailed workflow guide
- [GIT_SAFETY.md](./GIT_SAFETY.md) - Git safety guidelines for secrets
- [pkg/env docs](../) - Library documentation

## Architecture

This demo app shows the **correct way** to use the `pkg/env` library:

1. **Registry-Driven** - Single source of truth in `registry.go`
2. **Workflow-Based** - Use the 3 workflow commands for all operations
3. **Secrets-Separated** - Secrets live in separate `.env.secrets.*` files
4. **Encryption-Ready** - All secrets can be encrypted with Age
5. **Deployment-Ready** - Works with Docker, Fly.io, or any platform

## Why This Structure?

The app previously had 44 commands. Now it has just **5 commands**:

- **2 server commands** (serve, health) - Run the HTTP application
- **3 workflow commands** (sync-registry, sync-environments, finalize) - Manage environments

This makes it:
- ✅ Easier to understand
- ✅ Easier to deploy
- ✅ Better demonstration of the library
- ✅ Production-ready example

For low-level operations, use the `pkg/env` library directly in your own code. The old 41 commands are preserved in `commands.go` as reference.

## License

See parent directory for license information.
