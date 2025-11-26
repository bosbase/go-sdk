# API Rules Documentation - Go SDK Documentation

API Rules are collection access controls and data filters that determine who can perform actions on your collections and what data they can access.

## Overview

Each collection has 5 standard API rules, corresponding to specific API actions:

- **`listRule`** - Controls read/list access
- **`viewRule`** - Controls read/view access  
- **`createRule`** - Controls create access
- **`updateRule`** - Controls update access
- **`deleteRule`** - Controls delete access

Auth collections have two additional rules:

- **`manageRule`** - Admin-like permissions for managing auth records
- **`authRule`** - Additional constraints applied during authentication

## Rule Values

Each rule can be set to one of three values:

### 1. `nil` (Locked)
Only authorized superusers can perform the action.

```go
_, err := client.Collections.Update("products", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "listRule": nil,
    },
})
```

### 2. `""` (Empty String - Public)
Anyone (superusers, authorized users, and guests) can perform the action.

```go
_, err := client.Collections.Update("products", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "listRule": "",
    },
})
```

### 3. Non-empty String (Filter Expression)
Only users satisfying the filter expression can perform the action.

```go
_, err := client.Collections.Update("products", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "listRule": `@request.auth.id != ""`,
    },
})
```

## Default Permissions

When you create a base collection without specifying rules, BosBase applies opinionated defaults:

- `listRule` and `viewRule` default to an empty string (`""`), so guests and authenticated users can query records.
- `createRule` defaults to `@request.auth.id != ""`, restricting writes to authenticated users or superusers.
- `updateRule` and `deleteRule` default to `@request.auth.id != "" && createdBy = @request.auth.id`, which limits mutations to the record creator (superusers still bypass rules).

Every base collection now includes hidden system fields named `createdBy` and `updatedBy`. BosBase adds those fields automatically when a collection is created and manages their values server-side: `createdBy` always captures the authenticated actor that inserted the record (or stays empty for anonymous writes) and cannot be overridden later, while `updatedBy` is overwritten on each write (or cleared for anonymous writes).

## Setting Rules

### Bulk Rule Updates

Set multiple rules at once:

```go
_, err := client.Collections.Update("products", &bosbase.CrudMutateOptions{
    Body: map[string]interface{}{
        "listRule":  `@request.auth.id != ""`,
        "viewRule":  `@request.auth.id != ""`,
        "createRule": `@request.auth.id != ""`,
        "updateRule": `@request.auth.id != "" && createdBy = @request.auth.id`,
        "deleteRule": nil, // Only superusers
    },
})
```

## Common Rule Patterns

### Allow Only Authenticated Users

```go
rules := map[string]interface{}{
    "listRule":  `@request.auth.id != ""`,
    "viewRule":  `@request.auth.id != ""`,
    "createRule": `@request.auth.id != ""`,
    "updateRule": `@request.auth.id != ""`,
    "deleteRule": `@request.auth.id != ""`,
}
```

### Owner-Based Access

```go
rules := map[string]interface{}{
    "viewRule":  `@request.auth.id != "" && author = @request.auth.id`,
    "updateRule": `@request.auth.id != "" && author = @request.auth.id`,
    "deleteRule": `@request.auth.id != "" && author = @request.auth.id`,
}
```

### Public Read, Authenticated Write

```go
rules := map[string]interface{}{
    "listRule":  "", // Public
    "viewRule":  "", // Public
    "createRule": `@request.auth.id != ""`,
    "updateRule": `@request.auth.id != "" && createdBy = @request.auth.id`,
    "deleteRule": `@request.auth.id != "" && createdBy = @request.auth.id`,
}
```

## Related Documentation

- [API Rules and Filters](./API_RULES_AND_FILTERS.md) - Complete API rules documentation
- [Collections](./COLLECTIONS.md) - Collection configuration

