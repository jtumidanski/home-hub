# Home Hub Kiosk

The kiosk application is the primary household dashboard interface for Home Hub, designed for always-on displays such as tablets, Raspberry Pi touchscreens, and wall-mounted monitors.

## Features

- **Google/GitHub OAuth Authentication** - Secure sign-in via OAuth providers
- **User Profile Display** - View authenticated user information
- **Household Dashboard** - Display household details and members
- **Responsive Design** - Optimized for tablet and desktop viewports
- **Touch-Friendly** - Large touch targets (≥44x44px) for kiosk displays
- **Dark Mode Support** - Full dark mode theming

## Getting Started

### Prerequisites

- Node.js 20+
- npm or yarn
- Backend services running (gateway, svc-users)

### Installation

```bash
# Install dependencies
npm install
```

### Development

```bash
# Start development server on port 5173
npm run dev

# Access at:
# Direct: http://localhost:5173
# Via proxy: http://localhost:3000/kiosk/
```

### Build

```bash
# Build for production
npm run build

# Start production server
npm start
```

### Linting

```bash
# Run ESLint
npm run lint
```

## Project Structure

```
apps/kiosk/
├── app/                      # Next.js App Router pages
│   ├── layout.tsx           # Root layout with providers
│   ├── page.tsx             # Landing page (user/household display)
│   └── globals.css          # Global styles
├── components/              # React components
│   ├── auth/               # Authentication components
│   │   └── AuthGuard.tsx   # Protected route wrapper
│   ├── layout/             # Layout components (future)
│   ├── household/          # Household components (future)
│   └── ui/                 # shadcn/ui components (future)
├── lib/                    # Utility libraries
│   ├── auth/              # Authentication logic
│   │   ├── api.ts         # fetchMe(), logout()
│   │   ├── AuthContext.tsx # Auth state management
│   │   └── index.ts       # Exports
│   └── api/               # API clients
│       ├── client.ts      # Base API client
│       └── households.ts  # Household API functions
└── public/                # Static assets
```

## Architecture

### Authentication Flow

1. User visits kiosk app
2. `AuthProvider` calls `/api/me` to check authentication
3. If unauthenticated (401), show sign-in buttons
4. User clicks "Sign in with Google/GitHub"
5. Redirects to `/oauth2/{provider}/start`
6. OAuth flow completes, redirects back to kiosk
7. `AuthProvider` refetches user data
8. Landing page displays user profile and household info

### API Integration

- **Base URL**: `/api` (proxied through gateway)
- **Authentication**: Session cookies (HttpOnly)
- **Format**: JSON:API responses
- **Households**: `GET /api/households/{id}`, `GET /api/households/{id}/users`

### State Management

- **Auth State**: React Context (`AuthProvider`)
- **Local State**: React `useState` for household data
- **Data Fetching**: Native `fetch` API with error handling

## Environment Variables

```bash
# API base URL (defaults to /api)
NEXT_PUBLIC_API_URL=/api

# Node environment
NODE_ENV=development
```

## Configuration

### Port Configuration

The kiosk app runs on **port 5173** in development:

```json
// package.json
{
  "scripts": {
    "dev": "next dev -p 5173"
  }
}
```

### Proxy Configuration

In production, nginx proxies `/kiosk/` to the kiosk service:

```nginx
location /kiosk/ {
    proxy_pass http://kiosk:5173/;
}
```

## Development Workflow

### Adding New Components

1. Use shadcn/ui for consistent UI:
   ```bash
   npx shadcn@latest add [component-name]
   ```

2. Create components in appropriate directories:
   - `components/auth/` - Authentication-related
   - `components/household/` - Household-specific
   - `components/layout/` - Layout components
   - `components/ui/` - shadcn/ui components

### API Integration

1. Define interfaces in `lib/api/[domain].ts`
2. Use JSON:API response format
3. Handle errors with `ApiError` class
4. Include `credentials: 'include'` for session cookies

Example:

```typescript
import { get } from '@/lib/api/client';

export async function getResource(id: string) {
  const response = await get<JsonApiResponse<ResourceAttributes>>(
    `/resources/${id}`
  );
  return flattenResource(response.data);
}
```

### Authentication Guards

Protect routes with `AuthGuard`:

```tsx
import { AuthGuard } from '@/components/auth/AuthGuard';

export default function ProtectedPage() {
  return (
    <AuthGuard>
      <YourContent />
    </AuthGuard>
  );
}
```

## Testing

### Manual Testing Checklist

- [ ] Sign in with Google OAuth
- [ ] Sign in with GitHub OAuth
- [ ] User profile displays correctly
- [ ] Household info displays when associated
- [ ] Empty state shows when no household
- [ ] Sign out clears session
- [ ] Refresh preserves authentication
- [ ] Responsive on tablet (768px)
- [ ] Touch targets meet 44x44px minimum
- [ ] Keyboard navigation works

### Performance Testing

- [ ] Lighthouse score >90
- [ ] Initial load <2 seconds
- [ ] Bundle size <200KB gzipped

## Deployment

### Docker

The kiosk app is containerized and deployed via Docker Compose:

```yaml
services:
  kiosk:
    build: ./apps/kiosk
    ports:
      - "5173:5173"
    environment:
      - NEXT_PUBLIC_API_URL=/api
      - NODE_ENV=production
```

### Production Build

```bash
# Build optimized production bundle
npm run build

# Output: .next/ directory
# Serves static pages where possible
```

## Troubleshooting

### Authentication Not Working

1. Verify gateway service is running on port 8080
2. Check session cookies are set with correct domain/path
3. Ensure OAuth redirect URIs are configured correctly
4. Check browser console for CORS errors

### Household Data Not Loading

1. Verify user has `householdId` in profile (`/api/me`)
2. Check `svc-users` service is running
3. Verify household exists in database
4. Check network tab for API errors

### Build Errors

1. Run `npm install` to ensure dependencies are up to date
2. Check TypeScript errors: `npm run build`
3. Verify Tailwind configuration is correct
4. Clear `.next/` directory and rebuild

## Contributing

### Code Style

- Use TypeScript for all new code
- Follow Next.js App Router conventions
- Use Tailwind CSS for styling
- Implement proper error handling
- Add JSDoc comments for public functions

### Commit Messages

Follow conventional commits:
- `feat: add household member list display`
- `fix: resolve auth redirect loop`
- `docs: update README with deployment instructions`

## Resources

- [Next.js Documentation](https://nextjs.org/docs)
- [Tailwind CSS](https://tailwindcss.com)
- [shadcn/ui](https://ui.shadcn.com)
- [Project Architecture](/docs/PROJECT_KNOWLEDGE.md)
- [Dev Docs](/dev/active/kiosk-nextjs-app/)

## License

Private - Home Hub Project
