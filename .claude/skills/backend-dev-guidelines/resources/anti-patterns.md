
---
title: Anti-Patterns
description: Common pitfalls to avoid when implementing Golang microservices.
---

# Anti-Patterns

| Anti-Pattern | Why It’s Wrong |
|---------------|----------------|
| Business logic in handlers | Breaks separation of concerns |
| Mutable public fields | Violates immutability |
| Database logic in processors | Violates functional purity |
| Hardcoded topics | Breaks environment portability |
| Missing validation | Allows invalid domain states |
| Passing TenantId as param | Must use context extraction |
| Skipping header decorators | Breaks tracing and tenancy propagation |
| Global context usage | Breaks request isolation |

**Always** prefer pure, context-aware, curried, and testable functions.
