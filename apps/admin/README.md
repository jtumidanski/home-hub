# Home Hub Admin Portal

Administrative interface for the Home Hub multi-tenant household management platform.

## Technology Stack

- **Next.js 16** - React framework with App Router
- **React 19** - UI library
- **TypeScript 5** - Type-safe JavaScript
- **Tailwind CSS 4** - Utility-first CSS framework
- **shadcn/ui** - High-quality React components (Radix UI + Tailwind)

## Getting Started

### Prerequisites

- Node.js 20.x LTS
- npm 10.x or pnpm 8.x

### Installation

```bash
# Install dependencies
cd apps/admin
npm install

# Copy environment variables
cp .env.local.example .env.local
```

### Development

```bash
# Start development server on port 5174
npm run dev
```

Open [http://localhost:5174](http://localhost:5174) in your browser.

### Building

```bash
# Create production build
npm run build

# Start production server
npm run start
```

### Linting

```bash
# Run ESLint
npm run lint
```

## Environment Variables

Create a `.env.local` file based on `.env.local.example`:

```bash
# API Gateway URL
NEXT_PUBLIC_API_URL=http://localhost:3000/api

# Application Environment
NEXT_PUBLIC_APP_ENV=development
```

**Note:** Variables prefixed with `NEXT_PUBLIC_` are exposed to the browser.

## Project Structure

```
apps/admin/
├── app/                    # Next.js App Router
│   ├── layout.tsx         # Root layout
│   ├── page.tsx           # Home page
│   └── globals.css        # Global styles
├── components/
│   ├── ui/                # shadcn/ui components
│   ├── layout/            # Header, Sidebar, Footer
│   └── features/          # Feature-specific components
├── lib/
│   ├── api/               # API client utilities
│   ├── hooks/             # Custom React hooks
│   └── utils.ts           # Utility functions
├── types/                 # TypeScript type definitions
├── public/                # Static assets
├── Dockerfile             # Production container
├── .dockerignore          # Docker ignore patterns
├── .env.local.example     # Environment variable template
├── components.json        # shadcn/ui configuration
├── next.config.ts         # Next.js configuration
├── tailwind.config.ts     # Tailwind CSS configuration
└── tsconfig.json          # TypeScript configuration
```

## Docker Usage

### Build Docker Image

```bash
# From project root
docker-compose build admin
```

### Run Container

```bash
# Start admin service
docker-compose up admin

# Start all services
docker-compose up -d
```

The admin portal will be available at [http://localhost:5174](http://localhost:5174).

## Adding shadcn/ui Components

```bash
# Add a new component
npx shadcn@latest add [component-name]

# Example: Add a dialog component
npx shadcn@latest add dialog
```

Components are automatically added to `components/ui/`.

## Development Guidelines

### Import Aliases

Use the `@/*` import alias to reference files from the root:

```typescript
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/api/client";
import type { User } from "@/types/models";
```

### API Client

All API calls should use the client in `lib/api/client.ts`:

```typescript
import { get, post } from "@/lib/api/client";

// GET request
const users = await get<User[]>("/users");

// POST request
const newUser = await post<User>("/users", {
  name: "John Doe",
  email: "john@example.com",
});
```

**Important:** Never include `tenant_id` or `household_id` in request bodies. The gateway injects these via headers.

### Layout Components

The admin portal uses a three-component layout:

- **Header** - Top navigation with user menu
- **Sidebar** - Side navigation (hidden on mobile)
- **Footer** - Footer with links and copyright

Update `app/layout.tsx` to integrate these components.

### Styling

This project uses Tailwind CSS with shadcn/ui components:

```tsx
// Using Tailwind classes
<div className="flex items-center gap-4 p-6" />

// Using shadcn/ui components
<Button variant="outline" size="lg">
  Click me
</Button>

// Combining both
<Card className="hover:shadow-lg transition-shadow">
  <CardHeader>...</CardHeader>
</Card>
```

### Dark Mode

shadcn/ui includes dark mode support out of the box. Use Tailwind's `dark:` prefix:

```tsx
<div className="bg-white dark:bg-neutral-950 text-black dark:text-white">
  Content
</div>
```

## Architecture Notes

### Gateway Pattern

All API requests go through the Home Hub gateway at `/api/*`:

```
Admin Portal → nginx proxy → Gateway → Domain Services
```

### Authentication

Authentication is handled by the gateway using Google OIDC. The gateway:
- Verifies user identity
- Injects context headers (`X-Tenant-ID`, `X-Household-ID`, `X-User-ID`)
- Returns a service JWT for inter-service communication

### Security

- **Never expose** `household_id` in API responses
- **Always trust** the gateway-injected headers, not client data

### Polling vs WebSockets

The admin portal uses **30-second polling** for dashboard updates, not WebSockets. This approach:
- Works well with CDNs and caching
- Simplifies deployment (stateless)
- Suitable for kiosk displays on Raspberry Pi 4

## Deployment

The admin portal is designed for:
- **Docker** - Containerized deployment
- **Kubernetes** - Orchestration with Helm charts (planned)
- **Raspberry Pi 4** - Kiosk mode on local devices

### Port Assignments

- Development: `5174`
- Production (container): `5174`
- Nginx proxy (planned): `3000` (routes `/admin/` → admin service)

## Troubleshooting

### Port already in use

```bash
# Find process using port 5174
lsof -i :5174

# Kill the process
kill -9 <PID>
```

### Docker build fails

```bash
# Clean Docker cache
docker system prune -a

# Rebuild without cache
docker-compose build --no-cache admin
```

### Hot reload not working

```bash
# Ensure you're in the admin directory
cd apps/admin

# Restart the dev server
npm run dev
```

## Related Documentation

- [Project Architecture](/docs/PROJECT_KNOWLEDGE.md)
- [Dev Docs Pattern](/dev/README.md)
- [CLAUDE.md](/CLAUDE.md)
- [Next.js Documentation](https://nextjs.org/docs)
- [shadcn/ui Documentation](https://ui.shadcn.com)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)

## Contributing

This project follows the Home Hub development methodology:

1. Create dev docs for complex tasks (`/dev-docs [task]`)
2. Update context frequently during implementation
3. Mark tasks complete in the checklist
4. Document decisions and blockers

See `/dev/README.md` for details.

## License

Proprietary - Home Hub Platform
