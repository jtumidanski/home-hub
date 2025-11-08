# TypeScript Types

This directory contains shared TypeScript type definitions.

## Organization

```
types/
  api.ts         - API request/response types
  models.ts      - Domain model types (User, Household, Task, etc.)
  components.ts  - Component prop types
```

## Guidelines

- Use interfaces for object shapes
- Use types for unions and complex types
- Export all types for use across the application
- Keep types aligned with backend API contracts (dto-js package)

## Note

In the future, types will be generated from OpenAPI specs in `packages/dto-js/`.
