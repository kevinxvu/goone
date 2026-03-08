# Database Migrations

This directory contains database migration files managed by [Goose](https://github.com/pressly/goose).

## Migration File Format

Each migration file follows the naming convention:
```
YYYYMMDDHHMMSS_description.sql
```

Example: `20240101000001_create_users_table.sql`

## Migration File Structure

Each SQL migration file must have both `Up` and `Down` sections:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
```

## Commands

### Run all pending migrations
```bash
make migrate
# or
go run cmd/migration/main.go up
```

### Check migration status
```bash
go run cmd/migration/main.go status
```

### Rollback last migration
```bash
make migrate.undo
# or
go run cmd/migration/main.go down
```

### Create new migration
```bash
go run cmd/migration/main.go create add_users_table sql
```

### Other commands
```bash
# Show current version
go run cmd/migration/main.go version

# Rollback all migrations
go run cmd/migration/main.go reset

# Redo last migration
go run cmd/migration/main.go redo

# Migrate to specific version
go run cmd/migration/main.go up-to 20240101000001

# Rollback to specific version
go run cmd/migration/main.go down-to 20240101000001
```

## Best Practices

1. **Always test both Up and Down**: Ensure both directions work correctly
2. **Keep migrations small**: One logical change per migration
3. **Never modify existing migrations**: Create new ones instead
4. **Use transactions**: Wrap DDL statements in `StatementBegin/End`
5. **Be idempotent**: Use `IF NOT EXISTS` and `IF EXISTS` where possible
6. **Copy model structs**: Don't reference models directly to avoid schema drift
7. **Sequential versioning**: Use timestamps (YYYYMMDDHHMMSS) for version numbers

## Database-Specific Notes

### MySQL/MariaDB
- Always specify `ENGINE=InnoDB ROW_FORMAT=DYNAMIC`
- Use `DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
- Soft deletes: Add `deleted_at TIMESTAMP NULL DEFAULT NULL` with index

### PostgreSQL
- Use `SERIAL` or `BIGSERIAL` for auto-increment
- Timestamps: `created_at TIMESTAMPTZ DEFAULT NOW()`
- Soft deletes: Add `deleted_at TIMESTAMPTZ` with index

### SQLite
- Auto-increment: `INTEGER PRIMARY KEY AUTOINCREMENT`
- Timestamps: `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`
- Limited ALTER TABLE support - plan migrations carefully

## Migration History

Goose tracks migration history in the `goose_db_version` table:
- `id`: Migration version number
- `version_id`: Unique identifier
- `is_applied`: Whether migration has been applied
- `tstamp`: When migration was applied

## Troubleshooting

### Failed migration
If a migration fails, Goose will rollback the transaction automatically (if using `StatementBegin/End`). Fix the issue and re-run.

### Out-of-order migrations
Use the `--allow-missing` flag to apply missing migrations:
```bash
go run cmd/migration/main.go -allow-missing up
```

### Manual fix
To manually fix migration state, directly modify the `goose_db_version` table (use with caution).
