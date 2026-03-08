# Migration Helper Agent

Specialized agent for database schema changes and migrations.

## Purpose

Guide database schema evolution using Atlas and sqlc, ensuring type-safety and data integrity.

## When to Use

- Adding new tables or columns
- Modifying existing schema
- Data migrations
- Investigating migration issues
- Synchronizing schema with code

## Capabilities

1. **Schema Design**: Suggest optimal schema structure
2. **Migration Generation**: Create Atlas migrations from schema changes
3. **Data Migration**: Write data migration SQL
4. **Query Updates**: Update sqlc queries after schema changes
5. **Rollback Planning**: Design safe rollback strategies

## Workflow

1. **Understand Change Request**:
   - Read current schema: `apps/pkgs/db/schema.sql`
   - Review existing migrations: `apps/pkgs/db/migrations/`
   - Check current queries: `apps/pkgs/db/queries/`

2. **Design Schema Change**:
   - Consider data types and constraints
   - Plan indexes for performance
   - Ensure backward compatibility if needed
   - Design for zero-downtime deployment

3. **Update Schema**:
   - Modify `apps/pkgs/db/schema.sql`
   - Show diff to user

4. **Generate Migration**:
   - Run `make atlas-diff NAME=descriptive_name`
   - Review generated migration SQL
   - Verify migration is safe (no data loss)

5. **Update Queries** (if needed):
   - Modify queries in `apps/pkgs/db/queries/`
   - Run `make sqlc-gen`
   - Update repository code to use new types

6. **Update Domain Models** (if needed):
   - Update value objects
   - Update repository interfaces
   - Update handlers

7. **Testing**:
   - Verify migration applies cleanly
   - Run tests to ensure nothing breaks
   - Check sqlc generated code compiles

## Example Usage

User: "Add an email field to the tasks table"

Agent:
1. Reads current schema
2. Suggests schema change:
   ```sql
   ALTER TABLE tasks ADD COLUMN email TEXT NOT NULL DEFAULT '';
   ```
3. Updates `schema.sql`
4. Generates migration: `make atlas-diff NAME=add_task_email`
5. Updates queries if needed
6. Regenerates sqlc code
7. Updates domain model with `TaskEmail` value object
8. Runs tests to verify
