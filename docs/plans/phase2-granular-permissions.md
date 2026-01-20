# Phase 2: Granular Permissions (Future)

## Overview

Extend the permission system to support per-entity-type permissions. This allows fine-grained control like "can edit transactions but only view employees".

## Prerequisites

- Phase 1 completed
- `member_permissions` table exists with `entity_type` column

## Entity Types

| Entity Type | Value | Description |
|-------------|-------|-------------|
| Global | `NULL` | Applies to all entities (Phase 1 default) |
| Transactions | `transactions` | Transaction CRUD |
| Employees | `employees` | Employee management |
| Salaries | `salaries` | Salary records |
| Salary Costs | `salary_costs` | Salary cost breakdowns |
| Categories | `categories` | Category management |
| VAT | `vats` | VAT configuration |
| VAT Settings | `vat_settings` | VAT automation |
| Bank Accounts | `bank_accounts` | Bank account management |
| Forecasts | `forecasts` | Forecast viewing |

## Database

The `member_permissions` table from Phase 1 already supports this:

```sql
-- Global permission (Phase 1)
INSERT INTO member_permissions (user_id, organisation_id, entity_type, can_view, can_edit, can_delete)
VALUES (1, 1, NULL, true, true, false);

-- Entity-specific permission (Phase 2)
INSERT INTO member_permissions (user_id, organisation_id, entity_type, can_view, can_edit, can_delete)
VALUES (1, 1, 'transactions', true, true, true);
```

## Permission Resolution Logic

```go
func (a *APIService) hasPermission(userID, orgID int64, entityType string, action string) bool {
    // 1. Check entity-specific permission first
    entityPerm := a.dbService.GetMemberPermission(userID, orgID, entityType)
    if entityPerm != nil {
        return checkAction(entityPerm, action)
    }

    // 2. Fall back to global permission
    globalPerm := a.dbService.GetMemberPermission(userID, orgID, nil)
    if globalPerm != nil {
        return checkAction(globalPerm, action)
    }

    // 3. Default: view only
    return action == "view"
}
```

## API Changes

### Update Member Endpoint

```json
PATCH /organisations/:id/members/:userID
{
  "permissions": [
    { "entityType": null, "canView": true, "canEdit": false, "canDelete": false },
    { "entityType": "transactions", "canView": true, "canEdit": true, "canDelete": true }
  ]
}
```

### Handler Permission Checks

Each handler needs to check permissions before CRUD operations:

```go
func (h *Handlers) CreateTransaction(...) {
    if !h.apiService.HasPermission(userID, orgID, "transactions", "edit") {
        c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
        return
    }
    // ... proceed with creation
}
```

## Frontend Changes

### Permission Matrix UI

New component: `PermissionMatrix.vue`

```
                    View    Edit    Delete
─────────────────────────────────────────
Global (default)    [x]     [ ]     [ ]
Transactions        [x]     [x]     [x]
Employees           [x]     [x]     [ ]
Forecasts           [x]     [ ]     [ ]
...
```

### Member Edit Dialog

Extend to show permission matrix when editing non-owner members.

### Navigation Guards

Hide menu items user can't access:

```typescript
const canViewEmployees = computed(() =>
  hasPermission('employees', 'view')
)
```

## Implementation Steps

1. **Backend: Permission checking middleware**
   - Create `PermissionMiddleware` that extracts entity type from route
   - Add to protected routes

2. **Backend: Update handlers**
   - Add permission checks to all CRUD handlers
   - Return 403 Forbidden for denied actions

3. **Backend: Update member API**
   - Support array of permissions in update payload
   - Validate permission combinations

4. **Frontend: Permission matrix component**
   - Checkbox grid for entity × action
   - Save button updates all permissions at once

5. **Frontend: Navigation guards**
   - Fetch user permissions on login
   - Hide inaccessible menu items
   - Redirect if accessing forbidden route

6. **Frontend: Disabled UI elements**
   - Disable edit/delete buttons when no permission
   - Show tooltip explaining why disabled

## Testing

### Backend Tests

- Permission inheritance (entity-specific overrides global)
- Permission enforcement on each endpoint
- Cannot grant higher permissions than you have
- SuperAdmin bypasses all checks

### E2E Tests

- User with view-only can't see edit buttons
- User with entity-specific edit can modify that entity
- Permission matrix UI saves correctly

## Security Considerations

1. **Server-side enforcement**: Never trust frontend permission checks
2. **Audit logging**: Log permission changes (future consideration)
3. **Permission escalation**: Users can't grant permissions they don't have
4. **SuperAdmin bypass**: Owners always have full access

## Migration Path

1. Deploy Phase 1 with global permissions only
2. All existing members get global permissions based on role:
   - `admin` → can_view: true, can_edit: true, can_delete: true
   - `editor` → can_view: true, can_edit: true, can_delete: false
   - `read-only` → can_view: true, can_edit: false, can_delete: false
3. Deploy Phase 2 with granular UI
4. SuperAdmins can then customize per-entity permissions
