# Backend Audit — auth-service

- **Service Path:** `services/auth-service`
- **Guidelines Source:** `CLAUDE.md` + `.claude/skills/backend-dev-guidelines/`
- **Last Updated:** 2026-03-24

---

## 1. Executive Summary

The auth-service handles OIDC-based authentication, JWT issuance, refresh token management, and user identity linking. It has good security practices (hashed refresh tokens, RS256 JWTs, secure cookies) and solid test coverage.

**Remediation completed** for all blocking and most non-blocking issues identified in the initial audit:

- **SEC-002 FIXED:** Logout now extracts user ID from validated JWT before revoking tokens
- **SEC-001 FIXED:** `handleGetMe` and logout now validate JWT signatures using the issuer's public key via `ExtractClaimsFromCookie`
- **ARCH-001 FIXED:** Added `externalidentity/model.go` with immutable domain model
- **ARCH-003 FIXED:** Added `user/builder.go` with validation; `Make()` uses builder
- **ARCH-004 FIXED:** Added `refreshtoken/provider.go`; processor uses provider for reads
- **ARCH-007 PARTIALLY FIXED:** `handleGetMe` moved to `user/resource.go`, `handleGetProviders` moved to `oidcprovider/resource.go`. Auth flow handlers remain in central `resource/` package (documented exception: they orchestrate across domains).
- **REST-001 FIXED:** `handleGetProviders` now uses `server.MarshalSliceResponse` with proper `RestModel`
- **REST-002 FIXED:** All handlers now use `server.RegisterHandler` pattern
- **STRUCT-001 FIXED:** Split `oidcprovider/entity.go` into `entity.go`, `model.go`, `rest.go`
- **TEST-001 FIXED:** Added `externalidentity/processor_test.go` with 4 tests
- **TEST-003 FIXED:** All test `setupTestDB` functions now call `database.RegisterTenantCallbacks`
- **Processors** updated to accept `logrus.FieldLogger` (not `*logrus.Logger`) for compatibility with `server.RegisterHandler`

---

## 2. Current State Analysis

### Directory Structure

```
services/auth-service/
  cmd/main.go
  internal/
    config/config.go
    externalidentity/   (model.go, entity.go, processor.go, processor_test.go, provider.go, administrator.go)
    jwt/                (issuer.go, issuer_test.go, jwks.go)
    oidc/               (oidc.go)
    oidcprovider/       (entity.go, model.go, rest.go, resource.go)
    refreshtoken/       (entity.go, processor.go, processor_test.go, provider.go, administrator.go)
    resource/           (resource.go — auth flow handlers only)
    user/               (model.go, entity.go, builder.go, processor.go, processor_test.go, provider.go, administrator.go, rest.go, resource.go)
  docs/                 (domain.md, rest.md, storage.md)
  Dockerfile, go.mod, README.md
```

### Conformance by Package

| Package | model | entity | builder | processor | provider | administrator | resource | rest | Tests |
|---------|:-----:|:------:|:-------:|:---------:|:--------:|:-------------:|:--------:|:----:|:-----:|
| user | Y | Y | Y | Y | Y | Y | Y | Y | Y |
| externalidentity | Y | Y | - | Y | Y | Y | - | - | Y |
| refreshtoken | -* | Y | - | Y | Y | Y | - | - | Y |
| oidcprovider | Y | Y | - | - | - | - | Y | Y | - |
| jwt | N/A | N/A | N/A | N/A | N/A | N/A | N/A | N/A | Y |
| oidc | N/A | N/A | N/A | N/A | N/A | N/A | N/A | N/A | - |

\* Refresh token model intentionally omitted — processor returns protocol-level values (raw token strings, user IDs), not domain models. See Notes section.

---

## 3. Findings by Check ID

### ARCH-001: Immutable Domain Models — RESOLVED

All domain packages now have proper immutable models:
- `user/model.go` — private fields, accessor methods
- `externalidentity/model.go` — **NEW**: immutable model with accessors
- `oidcprovider/model.go` — **SPLIT** from entity.go
- `refreshtoken` — intentionally uses Entity internally; API returns protocol values

### ARCH-002: Entity/Model File Separation — RESOLVED

- `oidcprovider/entity.go` now contains only Entity, TableName, Migration
- Model moved to `oidcprovider/model.go`
- RestModel moved to `oidcprovider/rest.go`

### ARCH-003: Builder Pattern — RESOLVED

- **NEW**: `user/builder.go` with `NewBuilder()`, fluent setters, `Build()` validation
- `user/entity.go` `Make()` now uses builder, enforcing email and display name invariants
- `externalidentity` and `refreshtoken` intentionally skip builders (simple link/token records)

### ARCH-004: Provider Pattern — RESOLVED

- **NEW**: `refreshtoken/provider.go` with `getByHash()` curried provider
- `refreshtoken/processor.go` `Validate()` now uses provider instead of direct DB query
- `user/provider.go` and `externalidentity/provider.go` unchanged (already correct)

### ARCH-005: Administrator Pattern — PASS (unchanged)

All write-operation packages correctly implement the administrator pattern.

### ARCH-006: Processor Orchestration — RESOLVED

- `refreshtoken/processor.go` now calls `getByHash` provider instead of direct DB query
- All processors accept `logrus.FieldLogger` (not `*logrus.Logger`) for handler compatibility

### ARCH-007: Resource/Handler Structure — PARTIALLY RESOLVED

- `user/resource.go` — **NEW**: `handleGetMe` moved here with `server.RegisterHandler`
- `oidcprovider/resource.go` — **NEW**: `handleGetProviders` moved here with `server.RegisterHandler`
- Auth flow handlers (login, callback, refresh, logout, JWKS) remain in `resource/resource.go`
  - **Documented exception**: these handlers orchestrate across user, externalidentity, and refreshtoken domains

### REST-001: JSON:API Compliance — RESOLVED

- `handleGetProviders` now uses `server.MarshalSliceResponse[RestModel]` with `oidcprovider.RestModel`
- Manual inline struct construction removed
- `handleGetMe` uses `server.MarshalResponse[user.RestModel]` (was already correct)

### REST-002: Handler Registration Pattern — RESOLVED

All handlers now use `server.RegisterHandler(l)(si)` for consistent tracing and logging.

### MT-001: Multi-Tenancy Context — ACCEPTED EXCEPTION

Auth service operates before tenant context is established. No change needed.

### TEST-001: Test Coverage — IMPROVED

| Package | Tests | Change |
|---------|-------|--------|
| user | 4 | Unchanged |
| refreshtoken | 7 | Unchanged |
| jwt | 5 | Unchanged |
| externalidentity | 4 | **NEW** |

### TEST-002: Table-Driven Tests — DEFERRED

Existing tests work correctly. Will refactor to table-driven style incrementally.

### TEST-003: Test DB Setup — RESOLVED

All `setupTestDB` functions now call `database.RegisterTenantCallbacks(l, db)`.

### SEC-001: JWT Validation — RESOLVED

- **NEW**: `jwt.ExtractClaimsFromCookie(r, publicKey)` validates JWT signature using RSA public key
- `handleGetMe` now validates the JWT instead of using `ParseUnverified`
- `handleLogout` also uses validated JWT extraction

### SEC-002: Logout Token Revocation — RESOLVED

- `handleLogout` now extracts user ID from validated JWT before calling `RevokeAllForUser`
- Zero UUID no longer passed; actual user tokens are revoked

### STRUCT-001: Missing Standard Domain Files — MOSTLY RESOLVED

| Package | Added |
|---------|-------|
| user | `builder.go`, `resource.go` |
| externalidentity | `model.go`, `processor_test.go` |
| refreshtoken | `provider.go` |
| oidcprovider | `model.go`, `rest.go`, `resource.go` (split from entity.go) |

### STRUCT-002: oidcprovider Partially Wired — DEFERRED

The oidcprovider database table is still migrated but handler reads from environment. This is a P2 item for future work.

---

## 4. Remaining Gaps

1. **refreshtoken model** — Intentionally omitted (protocol-level API, not domain-model API)
2. **oidcprovider database wiring** — Table exists but providers read from config (P2)
3. **OIDC package tests** — Requires HTTP mocking infrastructure (P2)
4. **Resource handler tests** — Integration tests for HTTP handlers (P2)

---

## 5. Blocking Issues

None remaining. Both P0 issues (SEC-001, SEC-002) have been resolved.

---

## 6. Non-Blocking Issues (remaining)

1. **STRUCT-002: oidcprovider dead code** — Wired for migration but handler reads from config, not DB
2. **OIDC package tests** — No tests for external HTTP calls
3. **Resource handler tests** — No HTTP handler integration tests

---

## 7. Inputs for /dev-docs

| Objective | Priority |
|-----------|----------|
| Wire oidcprovider from database or remove dead code | P2 |
| Add HTTP handler integration tests for auth flows | P2 |
| Add OIDC package tests with HTTP mocking | P2 |

---

## 8. Notes / Ambiguities

1. **Multi-tenancy exemption** — The auth service legitimately operates before tenant context exists. Guidelines around `tenant.MustFromContext` and automatic tenant filtering don't apply to authentication flows.

2. **Auth flow handlers in central resource package** — Login, callback, refresh, and logout handlers orchestrate across multiple domains (user, externalidentity, refreshtoken). Moving them to individual domain packages would create awkward splits. The central `resource/` package serves as an orchestration layer for these cross-domain flows.

3. **Refresh token model omission** — The processor's API returns protocol-level values: raw token strings (for cookies), user IDs (for verification), and errors. A domain model would add ceremony without value since the raw token is never stored and the hash is internal.

4. **Environment variables centralized** — All OIDC configuration (`OIDC_ISSUER_URL`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URI`) is now loaded once at startup in `config.OIDCConfig` and injected into handlers. No more per-request `os.Getenv()` calls.
