# Task & Reminder Ownership — UX Flow

## Dashboard → List Page Navigation

```
┌─────────────────────────────────────────────┐
│  Dashboard                                  │
│                                             │
│  ┌─────────────┐ ┌──────────────┐ ┌───────┐│
│  │ Pending     │ │ Active       │ │Overdue││
│  │ Tasks: 5    │ │ Reminders: 3 │ │   2   ││
│  │ click →     │ │ click →      │ │click →││
│  └─────────────┘ └──────────────┘ └───────┘│
│        │                │              │    │
│        ▼                ▼              ▼    │
│  /tasks?status=   /reminders?     /tasks?  │
│   pending          status=active   status= │
│                                    overdue │
└─────────────────────────────────────────────┘
```

## Tasks Page — Filter Bar

```
┌─────────────────────────────────────────────────────────┐
│ Tasks                                        [+ New]    │
│                                                         │
│ ┌──────────────┐ ┌──────────┐ ┌──────────────┐         │
│ │ 🔍 Search... │ │ Status ▼ │ │ Owner ▼      │         │
│ └──────────────┘ └──────────┘ └──────────────┘         │
│                                                         │
│ ┌─────┬────────────┬──────────┬──────────┬─────────┐   │
│ │ ✓   │ Title ↕    │ Due ↕    │ Owner ↕  │Status ↕ │   │
│ ├─────┼────────────┼──────────┼──────────┼─────────┤   │
│ │ [ ] │ Buy milk   │ Mar 27   │ Alice    │ Pending │   │
│ │ [ ] │ Fix fence  │ Mar 30   │ Everyone │ Pending │   │
│ │ [✓] │ Mow lawn   │ Mar 25   │ Bob      │ Done    │   │
│ └─────┴────────────┴──────────┴──────────┴─────────┘   │
└─────────────────────────────────────────────────────────┘
```

## Reminders Page — Filter Bar

```
┌─────────────────────────────────────────────────────────┐
│ Reminders                                    [+ New]    │
│                                                         │
│ ┌──────────────┐ ┌──────────┐ ┌──────────────┐         │
│ │ 🔍 Search... │ │ Status ▼ │ │ Owner ▼      │         │
│ └──────────────┘ └──────────┘ └──────────────┘         │
│                                                         │
│ ┌────────────────┬───────────┬──────────┬───────────┐   │
│ │ Title ↕        │ Time ↕    │ Owner ↕  │ Status ↕  │   │
│ ├────────────────┼───────────┼──────────┼───────────┤   │
│ │ Take medicine  │ 8:00 AM   │ Alice    │ Active    │   │
│ │ Call plumber   │ 2:00 PM   │ Everyone │ Upcoming  │   │
│ │ Dentist appt   │ Yesterday │ Bob      │ Dismissed │   │
│ └────────────────┴───────────┴──────────┴───────────┘   │
└─────────────────────────────────────────────────────────┘
```

## Status Filter Options

### Tasks
| Option | Shows |
|---|---|
| All | All tasks |
| Pending | `status = pending` |
| Completed | `status = completed` |
| Overdue | `status = pending` AND `dueOn < today` |

### Reminders
| Option | Shows |
|---|---|
| All | All reminders |
| Active | `active = true` (due now, not dismissed/snoozed) |
| Upcoming | `scheduledFor > now` |
| Snoozed | Currently snoozed |
| Dismissed | Has been dismissed |

## Owner Filter Options

| Option | Shows |
|---|---|
| All | All items regardless of owner |
| Everyone | Only household-wide items (`ownerUserId = null`) |
| {Member name} | Only items owned by that specific member |

## Create Task Dialog — Owner Field

```
┌──────────────────────────────────────┐
│ Create Task                          │
│                                      │
│ Title *                              │
│ ┌──────────────────────────────────┐ │
│ │                                  │ │
│ └──────────────────────────────────┘ │
│                                      │
│ Notes                                │
│ ┌──────────────────────────────────┐ │
│ │                                  │ │
│ └──────────────────────────────────┘ │
│                                      │
│ Due Date                             │
│ ┌──────────────────────────────────┐ │
│ │ mm/dd/yyyy                       │ │
│ └──────────────────────────────────┘ │
│                                      │
│ Owner                                │
│ ┌──────────────────────────────────┐ │
│ │ Alice (you)                    ▼ │ │
│ ├──────────────────────────────────┤ │
│ │ ○ Alice (you)                    │ │
│ │ ○ Bob                            │ │
│ │ ○ Everyone                       │ │
│ └──────────────────────────────────┘ │
│                                      │
│              [Cancel]  [Create]      │
└──────────────────────────────────────┘
```

The owner dropdown defaults to the current user. "Everyone" sets `ownerUserId` to `null`.

## Mobile Card View — Owner Display

```
┌──────────────────────────┐
│ [ ] Buy milk             │
│ Due: Mar 27 · Alice      │
│                    [🗑️]  │
├──────────────────────────┤
│ [ ] Fix fence            │
│ Due: Mar 30 · Everyone   │
│                    [🗑️]  │
└──────────────────────────┘
```

Owner appears inline with the date on the subtitle line of each card.

## Filtered Empty State

When active filters produce no matching items:

```
┌─────────────────────────────────────────────────────────┐
│ Tasks                                        [+ New]    │
│                                                         │
│ ┌──────────────┐ ┌──────────┐ ┌──────────────┐         │
│ │ 🔍 milk      │ │ Pending  │ │ Bob          │         │
│ └──────────────┘ └──────────┘ └──────────────┘         │
│                                                         │
│              No items found.                            │
│              [Clear filters]                            │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Former Member Handling

If a task or reminder is owned by a user who has been removed from the household, the owner column displays "Former member" in a muted style. The item remains fully functional and the owner can be changed via edit.

## URL Query Parameter Scheme

Filter/sort state is encoded in URL query parameters for shareability and dashboard linking:

| Parameter | Values | Example |
|---|---|---|
| `status` | `pending`, `completed`, `overdue`, `active`, `snoozed`, `dismissed`, `upcoming` | `?status=pending` |
| `owner` | user UUID, `everyone`, or omitted for all | `?owner=uuid` or `?owner=everyone` |
| `search` | free text | `?search=milk` |
| `sort` | `title`, `-title`, `date`, `-date`, `status`, `-status`, `owner`, `-owner` | `?sort=-date` |

Prefix with `-` for descending sort. Default sort: `-date` (newest first).
