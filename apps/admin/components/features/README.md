# Feature Components

This directory contains feature-specific components.

## Organization

Organize components by feature domain:

```
features/
  households/
    HouseholdList.tsx
    HouseholdForm.tsx
  users/
    UserTable.tsx
    UserInvite.tsx
  tasks/
    TaskBoard.tsx
    TaskCard.tsx
```

## Guidelines

- Keep components focused on a single feature
- Export components from index.ts for clean imports
- Use shadcn/ui components for base UI elements
