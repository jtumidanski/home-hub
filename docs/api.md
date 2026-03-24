# Home Hub — API Specification

This document defines the versioned API surface for Home Hub.

The API is split across services, but all external routes are exposed through a single ingress hostname using path-prefix routing.

All API endpoints are versioned from the start.

Base path:

    /api/v1

The API follows JSON:API conventions where applicable.

Core services:

- auth-service
- account-service
- productivity-service

---

## 1. API Conventions

### 1.1 Versioning

All endpoints are rooted at:

    /api/v1

### 1.2 Media Type

Requests and responses should use:

    application/vnd.api+json

### 1.3 Resource Style

The API uses resource-oriented paths.

Examples:

    /api/v1/tasks
    /api/v1/tasks/{id}
    /api/v1/tasks/restorations

### 1.4 Resource Types

Resource `type` values are plural.

Examples:

- users
- tenants
- households
- memberships
- preferences
- contexts
- tasks
- task-restorations
- reminders
- reminder-snoozes
- reminder-dismissals
- task-summaries
- reminder-summaries
- dashboard-summaries

### 1.5 Error Shape

Errors should be returned as JSON:API error objects.

Each error should include, where possible:

- status
- code
- title
- detail

### 1.6 Includes

Where supported, JSON:API include syntax may be used.

Example:

    /api/v1/summary/tasks?include=tasks

### 1.7 Context

Requests are authenticated through auth-service-issued JWTs transported in secure HTTP-only cookies.

Services resolve access using:

- authenticated user identity
- tenant context
- household context

---

## 2. Routing by Service

### 2.1 auth-service

    /api/v1/auth/*
    /api/v1/users/*

### 2.2 account-service

    /api/v1/tenants/*
    /api/v1/households/*
    /api/v1/memberships/*
    /api/v1/preferences/*
    /api/v1/contexts/*

### 2.3 productivity-service

    /api/v1/tasks/*
    /api/v1/reminders/*
    /api/v1/summary/*

---

## 3. Auth Service API

### 3.1 Get Providers

Endpoint:

    GET /api/v1/auth/providers

Purpose:

- return enabled OIDC providers

Example response:

    {
      "data": [
        {
          "type": "auth-providers",
          "id": "google",
          "attributes": {
            "displayName": "Google"
          }
        }
      ]
    }

### 3.2 Begin Login

Endpoint:

    GET /api/v1/auth/login/{provider}

Example:

    /api/v1/auth/login/google?redirect=/app

Purpose:

- initiate OIDC login flow

Notes:

- redirect must be validated
- auth-service performs provider redirect

### 3.3 Callback

Endpoint:

    GET /api/v1/auth/callback/{provider}

Purpose:

- receive provider callback
- exchange code
- resolve or create user
- issue cookies
- redirect to frontend

### 3.4 Refresh Token

Endpoint:

    POST /api/v1/auth/token/refresh

Purpose:

- rotate refresh token
- issue new access token

### 3.5 Logout

Endpoint:

    POST /api/v1/auth/logout

Purpose:

- revoke current refresh session
- clear cookies

### 3.6 JWKS

Endpoint:

    GET /api/v1/auth/.well-known/jwks.json

Purpose:

- expose public keys for downstream JWT validation

### 3.7 Current User

Endpoint:

    GET /api/v1/users/me

Resource type:

    users

Purpose:

- return identity-only current user resource

Example response:

    {
      "data": {
        "type": "users",
        "id": "usr_123",
        "attributes": {
          "email": "user@example.com",
          "displayName": "Jane Doe",
          "givenName": "Jane",
          "familyName": "Doe",
          "avatarUrl": "https://example.com/avatar.png",
          "createdAt": "2026-03-24T12:00:00Z",
          "updatedAt": "2026-03-24T12:00:00Z"
        }
      }
    }

Notes:

- no tenant data here
- no household data here
- no membership data here

---

## 4. Account Service API

### 4.1 Tenants

#### List Tenants

    GET /api/v1/tenants

#### Create Tenant

    POST /api/v1/tenants

Resource type:

    tenants

Example request:

    {
      "data": {
        "type": "tenants",
        "attributes": {
          "name": "Tumidanski Household"
        }
      }
    }

#### Get Tenant

    GET /api/v1/tenants/{id}

Attributes:

- name
- createdAt
- updatedAt

Relationships:

- households
- memberships

Notes:

- tenant deletion is out of scope for v1

---

### 4.2 Households

#### List Households

    GET /api/v1/households

Supported filters:

- filter[tenant]=<tenant-id>

#### Create Household

    POST /api/v1/households

Resource type:

    households

Example request:

    {
      "data": {
        "type": "households",
        "attributes": {
          "name": "Main Home",
          "timezone": "America/Detroit",
          "units": "imperial"
        },
        "relationships": {
          "tenant": {
            "data": {
              "type": "tenants",
              "id": "ten_123"
            }
          }
        }
      }
    }

#### Get Household

    GET /api/v1/households/{id}

#### Update Household

    PATCH /api/v1/households/{id}

Attributes:

- name
- timezone
- units
- createdAt
- updatedAt

Relationships:

- tenant
- memberships

Rules:

- only owner may create additional households
- household deletion is out of scope for v1

---

### 4.3 Memberships

#### List Memberships

    GET /api/v1/memberships

Supported filters:

- filter[tenant]=<tenant-id>
- filter[household]=<household-id>
- filter[user]=<user-id>
- filter[role]=owner|admin|editor|viewer

#### Create Membership

    POST /api/v1/memberships

Resource type:

    memberships

Example request:

    {
      "data": {
        "type": "memberships",
        "attributes": {
          "role": "editor"
        },
        "relationships": {
          "tenant": {
            "data": {
              "type": "tenants",
              "id": "ten_123"
            }
          },
          "household": {
            "data": {
              "type": "households",
              "id": "hh_123"
            }
          },
          "user": {
            "data": {
              "type": "users",
              "id": "usr_456"
            }
          }
        }
      }
    }

#### Update Membership

    PATCH /api/v1/memberships/{id}

#### Delete Membership

    DELETE /api/v1/memberships/{id}

Attributes:

- role
- createdAt
- updatedAt

Relationships:

- tenant
- household
- user

Rules:

- roles may differ by household
- removing the only owner is not allowed in v1 unless replacement logic exists

---

### 4.4 Preferences

#### List Preferences

    GET /api/v1/preferences

Purpose:

- return the current user preference resource for the active tenant context in v1

#### Get Preference

    GET /api/v1/preferences/{id}

#### Update Preference

    PATCH /api/v1/preferences/{id}

Resource type:

    preferences

Attributes:

- theme
- createdAt
- updatedAt

Relationships:

- tenant
- activeHousehold

Theme allowed values:

- light
- dark

Example request to update theme:

    {
      "data": {
        "type": "preferences",
        "id": "pref_123",
        "attributes": {
          "theme": "dark"
        }
      }
    }

Example request to switch active household:

    {
      "data": {
        "type": "preferences",
        "id": "pref_123",
        "relationships": {
          "activeHousehold": {
            "data": {
              "type": "households",
              "id": "hh_456"
            }
          }
        }
      }
    }

Rules:

- one preference resource per user per tenant
- backend is canonical
- frontend may cache locally
- activeHousehold must belong to same tenant and be authorized

---

### 4.5 Current Context

#### Get Current Context

    GET /api/v1/contexts/current

Resource type:

    contexts

Purpose:

- return resolved current application context for authenticated user

Response is a single read-only resource with id:

    current

Example response:

    {
      "data": {
        "type": "contexts",
        "id": "current",
        "attributes": {
          "resolvedTheme": "dark",
          "resolvedRole": "owner",
          "canCreateHousehold": true
        },
        "relationships": {
          "tenant": {
            "data": {
              "type": "tenants",
              "id": "ten_123"
            }
          },
          "activeHousehold": {
            "data": {
              "type": "households",
              "id": "hh_456"
            }
          },
          "preference": {
            "data": {
              "type": "preferences",
              "id": "pref_123"
            }
          },
          "memberships": {
            "data": [
              {
                "type": "memberships",
                "id": "mem_1"
              }
            ]
          }
        }
      }
    }

Supported include example:

    /api/v1/contexts/current?include=tenant,activeHousehold,preference,memberships

Rules:

- read-only
- identity remains in auth-service
- if stored active household is invalid, service must resolve safe fallback or return selection-required state

---

## 5. Productivity Service API

### 5.1 Tasks

#### List Tasks

    GET /api/v1/tasks

Supported filters:

- filter[status]=pending|completed
- filter[deleted]=true|false
- filter[dueOn]=YYYY-MM-DD
- filter[dueOn][lte]=YYYY-MM-DD
- filter[dueOn][gte]=YYYY-MM-DD

Supported sorts:

- sort=dueOn
- sort=-dueOn
- sort=createdAt
- sort=-createdAt
- sort=updatedAt
- sort=-updatedAt

#### Create Task

    POST /api/v1/tasks

Resource type:

    tasks

Example request:

    {
      "data": {
        "type": "tasks",
        "attributes": {
          "title": "Take out trash",
          "notes": "Bins go out tonight",
          "status": "pending",
          "dueOn": "2026-03-24",
          "rolloverEnabled": true
        }
      }
    }

#### Get Task

    GET /api/v1/tasks/{id}

#### Update Task

    PATCH /api/v1/tasks/{id}

Example request to complete:

    {
      "data": {
        "type": "tasks",
        "id": "task_123",
        "attributes": {
          "status": "completed"
        }
      }
    }

#### Delete Task

    DELETE /api/v1/tasks/{id}

Delete behavior:

- soft delete only
- sets deletedAt
- hidden from normal active lists

Attributes:

- title
- notes
- status
- dueOn
- rolloverEnabled
- completedAt
- deletedAt
- createdAt
- updatedAt

Relationships:

- completedByUser

Rules:

- status values: pending, completed
- setting status to completed completes task
- setting status back to pending reopens task
- completedAt is service-managed

---

### 5.2 Task Restorations

#### Create Task Restoration

    POST /api/v1/tasks/restorations

Resource type:

    task-restorations

Example request:

    {
      "data": {
        "type": "task-restorations",
        "relationships": {
          "task": {
            "data": {
              "type": "tasks",
              "id": "task_123"
            }
          }
        }
      }
    }

Relationships:

- task
- createdByUser

Rules:

- task must exist
- task must be soft-deleted
- task must be within 3-day restore window
- successful creation clears deletedAt

---

### 5.3 Reminders

#### List Reminders

    GET /api/v1/reminders

Supported filters:

- filter[active]=true|false
- filter[scheduledFor][gte]=ISO-8601 datetime
- filter[scheduledFor][lte]=ISO-8601 datetime

Supported sorts:

- sort=scheduledFor
- sort=-scheduledFor
- sort=createdAt
- sort=-createdAt

#### Create Reminder

    POST /api/v1/reminders

Resource type:

    reminders

Example request:

    {
      "data": {
        "type": "reminders",
        "attributes": {
          "title": "Leave for appointment",
          "notes": "Bring insurance card",
          "scheduledFor": "2026-03-24T14:30:00Z"
        }
      }
    }

#### Get Reminder

    GET /api/v1/reminders/{id}

#### Update Reminder

    PATCH /api/v1/reminders/{id}

#### Delete Reminder

    DELETE /api/v1/reminders/{id}

Attributes:

- title
- notes
- scheduledFor
- active
- lastDismissedAt
- lastSnoozedUntil
- createdAt
- updatedAt

Rules:

- reminders are one-time only in v1
- active is service-derived, not client-controlled

---

### 5.4 Reminder Snoozes

#### Create Reminder Snooze

    POST /api/v1/reminders/snoozes

Resource type:

    reminder-snoozes

Example request:

    {
      "data": {
        "type": "reminder-snoozes",
        "attributes": {
          "durationMinutes": 30
        },
        "relationships": {
          "reminder": {
            "data": {
              "type": "reminders",
              "id": "rem_123"
            }
          }
        }
      }
    }

Attributes:

- durationMinutes
- snoozedUntil
- createdAt

Relationships:

- reminder
- createdByUser

Rules:

- allowed durationMinutes values: 10, 30, 60
- service computes effective snoozedUntil

---

### 5.5 Reminder Dismissals

#### Create Reminder Dismissal

    POST /api/v1/reminders/dismissals

Resource type:

    reminder-dismissals

Example request:

    {
      "data": {
        "type": "reminder-dismissals",
        "relationships": {
          "reminder": {
            "data": {
              "type": "reminders",
              "id": "rem_123"
            }
          }
        }
      }
    }

Relationships:

- reminder
- createdByUser

Rules:

- successful creation updates dismissal state

---

### 5.6 Task Summary

#### Get Task Summary

    GET /api/v1/summary/tasks

Resource type:

    task-summaries

Response is a single read-only resource with id:

    current

Attributes:

- pendingCount
- completedTodayCount
- overdueCount

Optional relationships:

- tasks

Supported include:

    /api/v1/summary/tasks?include=tasks

Example response:

    {
      "data": {
        "type": "task-summaries",
        "id": "current",
        "attributes": {
          "pendingCount": 6,
          "completedTodayCount": 2,
          "overdueCount": 1
        }
      }
    }

---

### 5.7 Reminder Summary

#### Get Reminder Summary

    GET /api/v1/summary/reminders

Resource type:

    reminder-summaries

Response is a single read-only resource with id:

    current

Attributes:

- dueNowCount
- upcomingCount
- snoozedCount

Optional relationships:

- reminders

Supported include:

    /api/v1/summary/reminders?include=reminders

Example response:

    {
      "data": {
        "type": "reminder-summaries",
        "id": "current",
        "attributes": {
          "dueNowCount": 1,
          "upcomingCount": 3,
          "snoozedCount": 1
        }
      }
    }

---

### 5.8 Dashboard Summary

#### Get Dashboard Summary

    GET /api/v1/summary/dashboard

Resource type:

    dashboard-summaries

Response is a single read-only resource with id:

    current

Attributes:

- householdName
- timezone
- pendingTaskCount
- dueReminderCount
- generatedAt

Relationships:

- taskSummary
- reminderSummary

Example response:

    {
      "data": {
        "type": "dashboard-summaries",
        "id": "current",
        "attributes": {
          "householdName": "Main Home",
          "timezone": "America/Detroit",
          "pendingTaskCount": 6,
          "dueReminderCount": 1,
          "generatedAt": "2026-03-24T13:15:00Z"
        },
        "relationships": {
          "taskSummary": {
            "data": {
              "type": "task-summaries",
              "id": "current"
            }
          },
          "reminderSummary": {
            "data": {
              "type": "reminder-summaries",
              "id": "current"
            }
          }
        }
      }
    }

---

## 6. Authorization Rules

### 6.1 General

Every service must validate:

- authenticated user
- tenant context
- household context where applicable

### 6.2 Account Rules

- only owner may create additional households
- role mutations must be authorized
- invalid active household assignment must be rejected

### 6.3 Productivity Rules

- all rows are scoped by tenant_id and household_id
- operations outside authorized context must be rejected
- summary data must respect same scope

---

## 7. Onboarding Flow

Onboarding is frontend-orchestrated.

Recommended sequence:

1. authenticate through auth-service
2. call:

       GET /api/v1/users/me

3. call:

       GET /api/v1/contexts/current

4. if no valid context exists:
   - create tenant
   - create household
   - initialize or fetch preference
5. refetch:

       GET /api/v1/contexts/current

No dedicated onboarding endpoint is required in v1.

---

## 8. Health Endpoints

All services expose:

    /healthz
    /readyz

Used by:

- local checks
- Docker validation
- Kubernetes readiness and liveness

---

## 9. Notes

- all APIs are versioned from the start
- all services use JSON:API conventions where applicable
- auth-service owns identity
- account-service owns app context
- productivity-service owns household productivity data
- frontend bootstraps using users/me and contexts/current
