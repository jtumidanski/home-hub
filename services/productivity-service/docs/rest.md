# REST API

All endpoints are prefixed with `/api/v1` and require JWT authentication. Request and response bodies use JSON:API format.

## Endpoints

### GET /api/v1/tasks

Lists all non-deleted tasks.

**Parameters:** None

**Response:** JSON:API array of `tasks` resources.

| Attribute       | Type    |
|-----------------|---------|
| title           | string  |
| notes           | string  |
| status          | string  |
| dueOn           | string  |
| rolloverEnabled | boolean |
| ownerUserId     | string  |
| completedAt     | string  |
| deletedAt       | string  |
| createdAt       | string  |
| updatedAt       | string  |

`dueOn` is formatted as `YYYY-MM-DD`.

**Error Conditions:**

| Status | Condition    |
|--------|--------------|
| 500    | Query failed |

---

### POST /api/v1/tasks

Creates a new task with status "pending". If `ownerUserId` is not provided, defaults to the authenticated user.

**Parameters:** None

**Request Model:** JSON:API `tasks` resource.

| Attribute       | Type    | Required |
|-----------------|---------|----------|
| title           | string  | yes      |
| notes           | string  | no       |
| dueOn           | string  | no       |
| rolloverEnabled | boolean | no       |
| ownerUserId     | string  | no       |

`dueOn` is parsed as `YYYY-MM-DD`.

**Response:** JSON:API `tasks` resource. Status 201.

**Error Conditions:**

| Status | Condition             |
|--------|-----------------------|
| 400    | Title is required     |
| 500    | Creation failed       |

---

### GET /api/v1/tasks/{id}

Returns a task by ID.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Response:** JSON:API `tasks` resource.

**Error Conditions:**

| Status | Condition  |
|--------|------------|
| 400    | Invalid ID |
| 404    | Not found  |

---

### PATCH /api/v1/tasks/{id}

Updates a task. Status transitions between "pending" and "completed" automatically manage completedAt and completedByUID.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Request Model:** JSON:API `tasks` resource.

| Attribute       | Type    | Required |
|-----------------|---------|----------|
| title           | string  | yes      |
| notes           | string  | no       |
| status          | string  | yes      |
| dueOn           | string  | no       |
| rolloverEnabled | boolean | no       |
| ownerUserId     | string  | no       |

**Response:** JSON:API `tasks` resource.

**Error Conditions:**

| Status | Condition      |
|--------|----------------|
| 400    | Invalid ID     |
| 404    | Not found      |
| 500    | Update failed  |

---

### DELETE /api/v1/tasks/{id}

Soft-deletes a task.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Response:** 204 No Content.

**Error Conditions:**

| Status | Condition      |
|--------|----------------|
| 400    | Invalid ID     |
| 404    | Not found      |
| 500    | Delete failed  |

---

### POST /api/v1/tasks/restorations

Restores a soft-deleted task within the 3-day restoration window. Creates an audit record.

**Parameters:** None

**Request Model:** JSON:API `task-restorations` resource.

| Attribute | Type   | Required |
|-----------|--------|----------|
| taskId    | string | yes      |

**Response:** JSON:API `task-restorations` resource. Status 201.

| Attribute | Type   |
|-----------|--------|
| createdAt | string |

**Error Conditions:**

| Status | Condition              |
|--------|------------------------|
| 400    | Invalid taskId         |
| 400    | Task is not deleted    |
| 400    | Restore window expired |
| 404    | Task not found         |
| 500    | Restore failed         |

---

### GET /api/v1/reminders

Lists all reminders.

**Parameters:** None

**Response:** JSON:API array of `reminders` resources.

| Attribute        | Type    |
|------------------|---------|
| title            | string  |
| notes            | string  |
| scheduledFor     | string  |
| ownerUserId      | string  |
| active           | boolean |
| lastDismissedAt  | string  |
| lastSnoozedUntil | string  |
| createdAt        | string  |
| updatedAt        | string  |

**Error Conditions:**

| Status | Condition    |
|--------|--------------|
| 500    | Query failed |

---

### POST /api/v1/reminders

Creates a new reminder. If `ownerUserId` is not provided, defaults to the authenticated user.

**Parameters:** None

**Request Model:** JSON:API `reminders` resource.

| Attribute    | Type   | Required |
|--------------|--------|----------|
| title        | string | yes      |
| notes        | string | no       |
| scheduledFor | string | yes      |
| ownerUserId  | string | no       |

`scheduledFor` is parsed as RFC3339.

**Response:** JSON:API `reminders` resource. Status 201.

**Error Conditions:**

| Status | Condition        |
|--------|------------------|
| 400    | Invalid date     |
| 400    | Validation failed |
| 500    | Creation failed  |

---

### GET /api/v1/reminders/{id}

Returns a reminder by ID.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Response:** JSON:API `reminders` resource.

**Error Conditions:**

| Status | Condition  |
|--------|------------|
| 400    | Invalid ID |
| 404    | Not found  |

---

### PATCH /api/v1/reminders/{id}

Updates a reminder.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Request Model:** JSON:API `reminders` resource.

| Attribute    | Type   | Required |
|--------------|--------|----------|
| title        | string | yes      |
| notes        | string | no       |
| scheduledFor | string | yes      |
| ownerUserId  | string | no       |

**Response:** JSON:API `reminders` resource.

**Error Conditions:**

| Status | Condition      |
|--------|----------------|
| 400    | Invalid ID     |
| 400    | Invalid date   |
| 404    | Not found      |
| 500    | Update failed  |

---

### DELETE /api/v1/reminders/{id}

Hard-deletes a reminder.

**Parameters:**

| Name | In   | Type | Required |
|------|------|------|----------|
| id   | path | UUID | yes      |

**Response:** 204 No Content.

**Error Conditions:**

| Status | Condition      |
|--------|----------------|
| 400    | Invalid ID     |
| 404    | Not found      |
| 500    | Delete failed  |

---

### POST /api/v1/reminders/dismissals

Dismisses a reminder and creates an audit record.

**Parameters:** None

**Request Model:** JSON:API `reminder-dismissals` resource.

| Attribute  | Type   | Required |
|------------|--------|----------|
| reminderId | string | yes      |

**Response:** JSON:API `reminder-dismissals` resource. Status 201.

| Attribute | Type   |
|-----------|--------|
| createdAt | string |

**Error Conditions:**

| Status | Condition            |
|--------|----------------------|
| 400    | Invalid reminderId   |
| 400    | Validation failed    |
| 404    | Reminder not found   |
| 500    | Dismiss failed       |

---

### POST /api/v1/reminders/snoozes

Snoozes a reminder for a specified duration and creates an audit record.

**Parameters:** None

**Request Model:** JSON:API `reminder-snoozes` resource.

| Attribute       | Type   | Required |
|-----------------|--------|----------|
| reminderId      | string | yes      |
| durationMinutes | int    | yes      |

Allowed values for durationMinutes: 10, 30, 60.

**Response:** JSON:API `reminder-snoozes` resource. Status 201.

| Attribute       | Type   |
|-----------------|--------|
| durationMinutes | int    |
| snoozedUntil    | string |
| createdAt       | string |

**Error Conditions:**

| Status | Condition               |
|--------|-------------------------|
| 400    | Invalid reminderId      |
| 400    | Invalid snooze parameters |
| 500    | Snooze failed           |

---

### GET /api/v1/summary/tasks?date=YYYY-MM-DD

Returns aggregated task counts. The client computes "today" in its local timezone and supplies it as `date`; the server uses it to compute the `completedTodayCount` and `overdueCount` boundaries.

**Parameters:**

| Name | In    | Type   | Required | Format     |
|------|-------|--------|----------|------------|
| date | query | string | yes      | YYYY-MM-DD |

**Response:** JSON:API `task-summaries` resource with ID "current".

| Attribute           | Type |
|---------------------|------|
| pendingCount        | int  |
| completedTodayCount | int  |
| overdueCount        | int  |

**Error Conditions:**

| Status | Condition                                     |
|--------|-----------------------------------------------|
| 400    | Missing, empty, or malformed `date` parameter |
| 500    | Query failed                                  |

---

### GET /api/v1/summary/reminders

Returns aggregated reminder counts.

**Parameters:** None

**Response:** JSON:API `reminder-summaries` resource with ID "current".

| Attribute     | Type |
|---------------|------|
| dueNowCount   | int  |
| upcomingCount | int  |
| snoozedCount  | int  |

**Error Conditions:**

| Status | Condition    |
|--------|--------------|
| 500    | Query failed |

---

### GET /api/v1/summary/dashboard?date=YYYY-MM-DD

Returns a combined dashboard summary. Same `date` semantics as `/summary/tasks`.

**Parameters:**

| Name | In    | Type   | Required | Format     |
|------|-------|--------|----------|------------|
| date | query | string | yes      | YYYY-MM-DD |

**Response:** JSON:API `dashboard-summaries` resource with ID "current".

| Attribute        | Type   |
|------------------|--------|
| pendingTaskCount | int    |
| dueReminderCount | int    |
| generatedAt      | string |

**Error Conditions:**

| Status | Condition                                     |
|--------|-----------------------------------------------|
| 400    | Missing, empty, or malformed `date` parameter |
| 500    | Query failed                                  |

**Error Conditions:**

| Status | Condition    |
|--------|--------------|
| 500    | Query failed |
