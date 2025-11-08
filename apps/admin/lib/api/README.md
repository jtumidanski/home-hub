# API Client

This directory contains API client utilities for communicating with the backend gateway.

## Files

- **client.ts** - Base API client with auth headers and error handling
- **[domain].ts** - Domain-specific API functions (e.g., users.ts, tasks.ts)

## Usage

```tsx
import { apiClient } from "@/lib/api/client";

const data = await apiClient<ResponseType>("/api/endpoint", {
  method: "POST",
  body: JSON.stringify(payload),
});
```

## Architecture Notes

- All API calls go through the gateway at `/api/*`
- Gateway injects tenant/household context via headers
- Never include tenant_id or household_id in request bodies
