
---
title: File Responsibilities
description: Responsibilities of each standard file in a domain package.
---

# File Responsibilities

## `model.go`
Defines immutable domain objects with private fields and accessor methods.

## `entity.go`

Database entity definitions and migration helpers using GORM. Provides `Make(Entity) (Model, error)` and `Model.ToEntity()`.

## `builder.go`

Fluent API for constructing validated domain models. `Build()` enforces invariants.


## `processor.go`
Pure curried business functions for core business logic. Dependency order: `NewProcessor(log, ctx, db)`.

## `provider.go`

**Read operations** (queries) that fetch data without side effects. This file handles all read-only database access.

**Key Responsibilities:**
- Define query functions returning `database.EntityProvider[T]` or `database.EntityProvider[[]T]`
- Provide `modelFromEntity(e entity) (Model, error)` transformation function
- Use `database.Query[T]` for single-entity lookups
- Use `database.SliceQuery[T]` for multi-entity queries
- Never modify database state

**Typical Signatures:**
```go
// Providers do NOT take tenantId — automatic tenant filtering via GORM callbacks
func getById(id uint32) database.EntityProvider[entity]
func getForAccountInWorld(accountId uint32, worldId world.Id) database.EntityProvider[[]entity]
func getAll() database.EntityProvider[[]entity]
func modelFromEntity(e entity) (Model, error)
```

**Pattern:** Provider functions are curried - they accept query parameters and return a function that takes `*gorm.DB` and returns `model.Provider[T]`. The `*gorm.DB` must have context set via `db.WithContext(ctx)` for automatic tenant filtering. This enables lazy evaluation and composition with `model.Map`, `model.SliceMap`, and `model.ParallelMap`.

**Why This Separation:**
- **Testability** - Read and write operations can be mocked independently
- **Pure composition** - Read path remains side-effect free, enabling functional composition
- **Clear intent** - Code review can quickly identify state-changing operations


## `resource.go`
Route registration and handler definitions for REST endpoints.

**Key Responsibilities:**
- Define `InitializeRoutes(si jsonapi.ServerInformation) func(db *gorm.DB) server.RouteInitializer` for route registration
- Use `server.RegisterHandler(l)(si)` for GET/DELETE handlers (no request body)
- Use `server.RegisterInputHandler[T](l)(si)` for POST/PATCH handlers (with typed request model)
- Implement handler functions matching `server.GetHandler` or `server.InputHandler[T]` signatures
- Map domain errors to HTTP status codes (400, 404, 409, 500)
- **Delegate ALL business logic to processors - NEVER call provider functions directly**
- Use `server.MarshalResponse[T]` for successful responses
- Log errors with context using `d.Logger().WithError(err)`

**Pattern:** Thin handlers that parse input, invoke processors, handle errors, and marshal responses.

**Critical Rules:**
- ✅ `resource.go` → `processor.go` (correct)
- ❌ `resource.go` → `provider.go` (WRONG - bypasses business logic layer)
- ❌ `resource.go` → database/GORM (WRONG - bypasses all layers)


## `rest.go`
Serialization and transformation between domain models and JSON:API.

**Key Responsibilities:**
- Define `RestModel` struct implementing JSON:API interface (`GetName()`, `GetID()`, `SetID()`)
- Define request models (`CreateRequest`, `UpdateRequest`) implementing JSON:API interface
- Implement `Transform(Model) (RestModel, error)` to convert domain models to REST representations
- Implement `TransformSlice([]Model) ([]RestModel, error)` for bulk transformations
- Use flat structure for request models (no nested Data/Type/Attributes)
- Mark ID field with `json:"-"` tag (set via SetID)
- Use pointer fields for optional attributes with `omitempty`


**Pattern:** JSON:API-compliant DTOs with automatic marshaling/unmarshaling via api2go library. No tenant data in payloads.



