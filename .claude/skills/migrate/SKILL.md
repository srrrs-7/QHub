---
name: migrate
description: Database migration operations using Atlas
args: status|apply|diff|new [name]
---

# Migrate Skill

Database migration operations using Atlas.

## Usage

```bash
# Check migration status
/migrate status

# Apply pending migrations
/migrate apply

# Generate new migration from schema changes
/migrate diff [name]

# Create empty migration file
/migrate new [name]
```

## Implementation

When invoked:

1. Change to database directory: `cd apps/pkgs/db`
2. Execute Atlas command based on action:
   - `status`: `make atlas-status`
   - `apply`: `make atlas-apply`
   - `diff [name]`: `make atlas-diff NAME=[name]`
   - `new [name]`: `make atlas-new NAME=[name]`
3. Display migration status and any pending migrations
4. If schema changes detected, remind to run `make sqlc-gen`

## Migration Workflow

1. Update schema: Edit `apps/pkgs/db/schema.sql`
2. Generate migration: `/migrate diff add_new_table`
3. Review migration file in `apps/pkgs/db/migrations/`
4. Apply migration: `/migrate apply`
5. Regenerate queries: `make sqlc-gen` (if queries changed)

## Examples

```bash
# Check what migrations are pending
/migrate status

# Apply all pending migrations
/migrate apply

# Create migration from schema changes
/migrate diff add_user_email

# Create empty migration file
/migrate new custom_data_migration
```
