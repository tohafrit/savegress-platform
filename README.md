# Savegress Platform

Production-ready landing page and API for Savegress CDC platform.

## Project Structure

```
savegress-platform/
├── frontend/                    # Next.js 14 application
├── backend/                     # Fastify API
├── docker/                      # Docker configuration
│   ├── docker-compose.yml       # Production
│   ├── docker-compose.local.yml # Local development
│   ├── Caddyfile                # Production Caddy config
│   ├── Caddyfile.local          # Local Caddy config
│   ├── frontend.Dockerfile
│   └── backend.Dockerfile
├── Makefile                     # Easy commands
├── .env.example
└── README.md
```

## Tech Stack

- **Frontend**: Next.js 14 (App Router), TypeScript, Tailwind CSS, shadcn/ui
- **Backend**: Fastify, TypeScript, Prisma ORM
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Reverse Proxy**: Caddy 2

## Quick Start (Local Development)

### Option 1: Docker (Recommended)

The easiest way to run everything locally:

```bash
# Build and start all services
make dev-build

# Or without rebuild
make dev
```

Open http://localhost in your browser.

**Available commands:**
```bash
make dev          # Start all services
make dev-build    # Build and start all services
make dev-down     # Stop all services
make dev-logs     # View logs
make dev-clean    # Stop and remove all containers + volumes
```

### Option 2: Native Development

Run frontend and backend separately (requires local PostgreSQL):

```bash
# Install dependencies
make install

# Terminal 1: Run frontend
make frontend
# Opens at http://localhost:3000

# Terminal 2: Run backend
make backend
# API at http://localhost:3001
```

## Local URLs

When running with Docker (`make dev`):
- **Frontend**: http://localhost
- **API**: http://localhost/api/early-access
- **Health check**: http://localhost/health
- **PostgreSQL**: localhost:5432 (user: savegress, password: localdev123)
- **Redis**: localhost:6379 (password: localdev123)

## Testing the Setup

After running `make dev-build`, test that everything works:

```bash
# Check health endpoint
curl http://localhost/health

# Test early access form submission
curl -X POST http://localhost/api/early-access \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "company": "Test Company"}'
```

## Development Workflow

### Database Migrations

```bash
# Create a new migration
make db-migrate

# Open Prisma Studio (visual database browser)
make db-studio

# Reset database (deletes all data!)
make db-reset
```

### Viewing Logs

```bash
# All services
make dev-logs

# Specific service
cd docker && docker-compose -f docker-compose.local.yml logs -f backend
cd docker && docker-compose -f docker-compose.local.yml logs -f frontend
cd docker && docker-compose -f docker-compose.local.yml logs -f postgres
```

### Rebuilding After Code Changes

```bash
# Rebuild specific service
cd docker && docker-compose -f docker-compose.local.yml up --build frontend

# Rebuild everything
make dev-build
```

## Production Deployment

### Prerequisites

- Hetzner server (or any VPS) with Docker installed
- Domain (savegress.com) pointed to your server
- Ports 80 and 443 open

### Deploy

```bash
# On your server
git clone <repo-url> savegress-platform
cd savegress-platform

# Create environment file
cp .env.example .env
nano .env  # Set secure passwords for DB_PASSWORD and REDIS_PASSWORD

# Start services
make prod-build

# Check logs
make prod-logs
```

### Production Commands

```bash
make prod         # Start production services
make prod-build   # Build and start
make prod-down    # Stop services
make prod-logs    # View logs
```

## Environment Variables

### Production (.env)

```bash
DB_PASSWORD=your_secure_password_here
REDIS_PASSWORD=your_secure_redis_password
```

### Local Development

Local development uses hardcoded values (no .env needed):
- Database: `postgresql://savegress:localdev123@postgres:5432/savegress`
- Redis: `redis://:localdev123@redis:6379`

## Troubleshooting

### Port 80 already in use

```bash
# Find what's using port 80
sudo lsof -i :80

# Stop the process or change the port in docker-compose.local.yml
```

### Database connection issues

```bash
# Check if postgres is running
cd docker && docker-compose -f docker-compose.local.yml ps postgres

# View postgres logs
cd docker && docker-compose -f docker-compose.local.yml logs postgres

# Connect to database directly
docker exec -it savegress-postgres psql -U savegress -d savegress
```

### Frontend not loading

```bash
# Check frontend logs
cd docker && docker-compose -f docker-compose.local.yml logs frontend

# Rebuild frontend
cd docker && docker-compose -f docker-compose.local.yml up --build frontend
```

### Clean restart

```bash
# Remove everything and start fresh
make dev-clean
make dev-build
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/api/early-access` | Submit early access request |

### POST /api/early-access

```json
{
  "email": "user@company.com",
  "company": "Company Name",
  "currentSolution": "kafka|dms|fivetran|none|other",
  "dataVolume": "<10GB|10-100GB|100GB-1TB|>1TB",
  "message": "Optional message"
}
```

## License

Proprietary
