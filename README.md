# Home Hub

A modular, microservice-based home information platform designed to power kiosk-style displays and administrative interfaces for shared households.

## Overview

Home Hub delivers household dashboards with tasks, meals, weather, calendar, and reminders through a multi-tenant, multi-household architecture with Google/GitHub-based authentication.

## Architecture

- **Microservices by domain**: Each domain (users, calendar, weather, tasks, meals, reminders) owns its own database schema, API, and workers
- **Gateway pattern**: Single public entry point via nginx proxy
- **Stateless APIs**: All state persisted in PostgreSQL; workers handle time-based logic
- **Poll-based UI**: No WebSockets for predictable, resilient kiosk design

## Applications

### Kiosk (`apps/kiosk/`)
The primary household dashboard interface for always-on displays (tablets, Raspberry Pi touchscreens, wall-mounted monitors).

**Features:**
- User profile display with OAuth authentication
- Household information and member list
- Responsive design for tablet/desktop
- Touch-optimized interface (≥44x44px targets)
- Dark mode support

**Port:** 5173
**Access:** http://localhost:3000/kiosk/

### Admin Portal (`apps/admin/`)
Administrative interface for managing users, households, and system configuration.

**Features:**
- User management (view, edit, roles)
- Household management (create, edit, associate users)
- Dashboard with system statistics
- Administrative controls

**Port:** 5174
**Access:** http://localhost:3000/admin/

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Node.js 20+ (for local development)
- Go 1.24+ (for backend services)

### Environment Setup

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Configure OAuth credentials in `.env`:
   ```bash
   # Google OAuth
   GOOGLE_CLIENT_ID=your-google-client-id
   GOOGLE_CLIENT_SECRET=your-google-client-secret
   GOOGLE_COOKIE_SECRET=generate-random-32-chars

   # GitHub OAuth
   GITHUB_CLIENT_ID=your-github-client-id
   GITHUB_CLIENT_SECRET=your-github-client-secret
   GITHUB_COOKIE_SECRET=generate-random-32-chars
   ```

### Running with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

**Services started:**
- PostgreSQL (port 5432)
- svc-users (backend service)
- Admin portal (port 5174)
- Kiosk app (port 5173)
- OAuth2 proxies (Google: 4180, GitHub: 4181)
- nginx proxy (port 80)
- pgAdmin (port 4000)

**Access points:**
- Main proxy: http://localhost
- Admin: http://localhost/admin
- Kiosk: http://localhost/kiosk
- API: http://localhost/api
- pgAdmin: http://localhost:4000

### Local Development

#### Kiosk App

```bash
cd apps/kiosk
npm install
npm run dev
# → http://localhost:5173
```

#### Admin Portal

```bash
cd apps/admin
npm install
npm run dev
# → http://localhost:5174
```

#### Backend Service (svc-users)

```bash
cd apps/svc-users
go build
./svc-users
# → http://localhost:8080
```

## Project Structure

```
home-hub/
├── apps/
│   ├── kiosk/              # Kiosk dashboard app (Next.js)
│   ├── admin/              # Admin portal (Next.js)
│   ├── svc-users/          # Users/households service (Go)
│   ├── gateway/            # API gateway (planned)
│   └── workers/            # Background jobs (planned)
├── docs/
│   └── PROJECT_KNOWLEDGE.md  # Architecture documentation
├── dev/
│   └── active/             # Dev docs for complex tasks
├── nginx.conf              # nginx proxy configuration
├── docker-compose.yml      # Docker services definition
└── README.md               # This file
```

## Authentication Flow

1. User visits `/kiosk` or `/admin`
2. nginx checks authentication via OAuth2 proxy
3. If unauthenticated, redirects to `/oauth2/{provider}/start`
4. After OAuth flow completes, redirects back to original URL
5. Session cookie persists authentication state
6. Frontend calls `/api/me` to get user data

## Development Workflow

### Adding New Features

1. **Plan**: Use `/dev-docs [description]` to create structured docs
2. **Implement**: Follow architecture patterns in `docs/PROJECT_KNOWLEDGE.md`
3. **Test**: Run locally with `npm run dev` or `docker-compose up`
4. **Document**: Update dev docs and README

### Code Guidelines

- **TypeScript** for all frontend code
- **Go** for all backend services
- **Tailwind CSS** for styling
- **shadcn/ui** for UI components
- **JSON:API** format for all API responses

### Branching Strategy

- `main`: Production-ready code
- Feature branches: `feature/kiosk-dashboard`, `feature/task-rollover`, etc.
- Follow conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`

## Key Technologies

- **Frontend**: Next.js 16, React 19, TypeScript, Tailwind CSS 4
- **Backend**: Go 1.24, GORM, PostgreSQL 16
- **Infrastructure**: Docker, nginx, OAuth2 Proxy
- **Authentication**: Google OIDC, GitHub OAuth

## Documentation

- **Architecture**: `/docs/PROJECT_KNOWLEDGE.md` - Complete system design
- **Dev Methodology**: `/dev/README.md` - Dev docs pattern guide
- **Kiosk App**: `/apps/kiosk/README.md` - Kiosk-specific documentation
- **Admin App**: `/apps/admin/README.md` - Admin-specific documentation

## Contributing

1. Create a feature branch from `main`
2. Make your changes following code guidelines
3. Write tests for new functionality
4. Update documentation as needed
5. Submit a pull request with clear description

## Troubleshooting

### OAuth Not Working

- Verify OAuth credentials in `.env`
- Check redirect URIs match configuration
- Ensure cookie domain is set correctly (`COOKIE_DOMAIN`)
- Check OAuth proxy logs: `docker-compose logs oauth2-proxy-google`

### Cannot Access Apps

- Verify nginx is running: `docker-compose ps nginx`
- Check service health: `curl http://localhost/health`
- View nginx logs: `docker-compose logs nginx`
- Ensure ports are not already in use

### Database Connection Issues

- Verify PostgreSQL is running: `docker-compose ps db`
- Check database credentials match `.env`
- View database logs: `docker-compose logs db`

## License

Private - Home Hub Project