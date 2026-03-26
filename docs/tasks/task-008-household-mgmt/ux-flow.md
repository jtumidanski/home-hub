# Household Management — UX Flow

## 1. First-Login Join Flow (Modified Onboarding)

```
User authenticates via OIDC
         │
         ▼
  Frontend fetches /contexts/current
         │
         ▼
  Has tenant + household? ──Yes──► Dashboard
         │
         No
         │
         ▼
  Fetch GET /invitations/mine?filter[status]=pending
         │
         ▼
  Has pending invitations? ──No──► Standard Onboarding
         │                          (Create Tenant → Create Household)
         Yes
         │
         ▼
  ┌─────────────────────────────┐
  │   Invitation Selection      │
  │                             │
  │   "You've been invited!"    │
  │                             │
  │   ┌───────────────────────┐ │
  │   │ The Smith Home        │ │
  │   │ Role: Editor          │ │
  │   │ Invited by: Jane S.   │ │
  │   │ Expires: Apr 2, 2026  │ │
  │   │ [Accept]  [Decline]   │ │
  │   └───────────────────────┘ │
  │                             │
  │   ┌───────────────────────┐ │
  │   │ Mom & Dad's Place     │ │
  │   │ Role: Viewer          │ │
  │   │ Invited by: Mom       │ │
  │   │ Expires: Apr 1, 2026  │ │
  │   │ [Accept]  [Decline]   │ │
  │   └───────────────────────┘ │
  │                             │
  │   ─── or ───                │
  │                             │
  │   [Create my own household] │
  └─────────────────────────────┘
         │
    Accept one ──────────────────► POST /invitations/{id}/accept
         │                                    │
         ▼                                    ▼
  Membership created              Active household auto-switched
  Tenant assigned                 Redirect to Dashboard
  Preference created
```

**Key behaviors**:
- Accepting one invitation immediately redirects to dashboard.
- Remaining pending invitations stay pending for later management.
- "Create my own household" links to the standard onboarding flow.
- If all invitations are declined, the user proceeds to standard onboarding.

---

## 2. Household Members Page

Accessible from the existing "Households" navigation item or as a sub-section.

### 2.1 Privileged View (Owner / Admin)

```
┌────────────────────────────────────────────────────┐
│  Household Members            [Invite Member]      │
│                                                    │
│  Members (3)                                       │
│  ┌──────────────────────────────────────────────┐  │
│  │ 👤 Jane Smith          Owner     Jan 2026    │  │
│  │    jane@example.com    ⚠ Sole owner          │  │
│  │                                              │  │
│  │ 👤 John Smith          Admin ▾   Feb 2026    │  │
│  │    john@example.com        [Remove]          │  │
│  │                                              │  │
│  │ 👤 Alex Smith          Editor ▾  Mar 2026    │  │
│  │    alex@example.com        [Remove]          │  │
│  └──────────────────────────────────────────────┘  │
│                                                    │
│  Pending Invitations (1)                           │
│  ┌──────────────────────────────────────────────┐  │
│  │ ✉ pat@example.com      Viewer                │  │
│  │   Invited by Jane · Expires Apr 2            │  │
│  │                            [Revoke]          │  │
│  └──────────────────────────────────────────────┘  │
│                                                    │
│  ──────────────────────────────────────────────    │
│  [Leave Household]                                 │
└────────────────────────────────────────────────────┘
```

**Interactions**:
- Role dropdown (▾) allows changing a member's role. Admin cannot modify owners.
- [Remove] button shows confirmation dialog before removing.
- [Invite Member] opens a dialog with email input and role selector.
- [Revoke] cancels a pending invitation with confirmation.
- [Leave Household] shows confirmation dialog. Blocked if last owner.

### 2.2 Non-Privileged View (Editor / Viewer)

```
┌────────────────────────────────────────────────────┐
│  Household Members                                 │
│                                                    │
│  Members (3)                                       │
│  ┌──────────────────────────────────────────────┐  │
│  │ 👤 Jane Smith          Owner     Jan 2026    │  │
│  │    jane@example.com    ⚠ Sole owner          │  │
│  │                                              │  │
│  │ 👤 John Smith          Admin     Feb 2026    │  │
│  │    john@example.com                          │  │
│  │                                              │  │
│  │ 👤 Alex Smith          Editor    Mar 2026    │  │
│  │    alex@example.com                          │  │
│  └──────────────────────────────────────────────┘  │
│                                                    │
│  Pending Invitations (1)                           │
│  ┌──────────────────────────────────────────────┐  │
│  │ ✉ pat@example.com      Viewer                │  │
│  │   Invited by Jane · Expires Apr 2            │  │
│  └──────────────────────────────────────────────┘  │
│                                                    │
│  ──────────────────────────────────────────────    │
│  [Leave Household]                                 │
└────────────────────────────────────────────────────┘
```

**Differences from privileged view**:
- No [Invite Member] button.
- Roles displayed as plain text (no dropdown).
- No [Remove] or [Revoke] buttons.
- [Leave Household] is still available.

---

## 3. Invite Member Dialog

```
┌─────────────────────────────────┐
│  Invite Member                  │
│                                 │
│  Email                          │
│  ┌───────────────────────────┐  │
│  │ user@example.com          │  │
│  └───────────────────────────┘  │
│                                 │
│  Role                           │
│  ┌───────────────────────────┐  │
│  │ Viewer               ▾   │  │
│  └───────────────────────────┘  │
│  Options: Viewer, Editor, Admin │
│                                 │
│  [Cancel]        [Send Invite]  │
└─────────────────────────────────┘
```

**Validation**:
- Email is required and must be valid format.
- Shows inline error if invitation already exists (409) or user is already a member (422).

---

## 4. Post-Onboarding Invitation Notification

For users who already have a tenant/household and receive new invitations:

```
┌──────────────────────────────────────┐
│  Navigation Sidebar                  │
│                                      │
│  Home                                │
│    Dashboard                         │
│  Productivity                        │
│    Tasks                             │
│    Reminders                         │
│  Management                          │
│    Households                        │
│    Members  (2) ◄── badge            │
└──────────────────────────────────────┘
```

- A badge on the "Households" nav item indicates pending invitations for the current user.
- The `pendingInvitationCount` from app context drives this badge.
- The households page includes a section showing the user's own pending invitations with accept/decline actions.
- Household members is a sub-page at `/app/households/:id/members`, accessed from the households list.

---

## 5. Confirmation Dialogs

### Remove Member
```
┌─────────────────────────────────────┐
│  Remove Member                      │
│                                     │
│  Are you sure you want to remove    │
│  Alex Smith from this household?    │
│  They will lose access to all       │
│  household data.                    │
│                                     │
│  [Cancel]              [Remove]     │
└─────────────────────────────────────┘
```

### Leave Household
```
┌─────────────────────────────────────┐
│  Leave Household                    │
│                                     │
│  Are you sure you want to leave     │
│  "The Smith Home"? You will lose    │
│  access to all household data.      │
│                                     │
│  [Cancel]               [Leave]     │
└─────────────────────────────────────┘
```

### Last Owner Block
```
┌─────────────────────────────────────┐
│  Cannot Leave                       │
│                                     │
│  You are the only owner of this     │
│  household. Assign another owner    │
│  before leaving.                    │
│                                     │
│                          [OK]       │
└─────────────────────────────────────┘
```
