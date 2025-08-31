# Migration

A Go database migration tool that supports PostgreSQL databases. This tool allows you to manage database schema changes through versioned migration files.

## Features

- **Generate migrations**: Create timestamped up/down migration files
- **Run migrations**: Apply pending migrations to your database
- **Rollback migrations**: Revert applied migrations
- **Migration status**: Check which migrations have been applied
- **PostgreSQL support**: Currently supports PostgreSQL databases

## Installation

```bash
git clone <repository-url>
cd migration
go mod download
```

## Setup

1. Create the migrations directory:
```bash
mkdir -p db/migrations
```

2. Ensure your PostgreSQL database is running and accessible.

## Configuration

The tool uses hardcoded configuration in `cmd/migrate/main.go`:
- **Database**: PostgreSQL at localhost:5432
- **Database name**: `gooolib_migration_development`
- **Username/Password**: `postgres/postgres`
- **Migration directory**: `./db/migrations`

## Commands

### Generate a Migration

Create a new migration with up and down SQL files:

```bash
go run cmd/migrate/main.go generate [migration_name]
```

This creates two files:
- `{timestamp}_{migration_name}.up.sql` - SQL to apply the migration
- `{timestamp}_{migration_name}.down.sql` - SQL to revert the migration

Example:
```bash
go run cmd/migrate/main.go generate create_users_table
# Creates:
# 20250830133803_create_users_table.up.sql
# 20250830133803_create_users_table.down.sql
```

### Apply Migrations (Up)

Apply all pending migrations:

```bash
go run cmd/migrate/main.go up
```

### Rollback Migrations (Down)

Rollback the last applied migration:

```bash
go run cmd/migrate/main.go down
```

Or use the alias:
```bash
go run cmd/migrate/main.go rollback
```

### Reset All Migrations

Rollback all applied migrations:

```bash
go run cmd/migrate/main.go reset
```

### Check Migration Status

View the current migration status:

```bash
go run cmd/migrate/main.go status
```

## Migration Files

Migration files follow this naming convention:
```
{timestamp}_{description}.{direction}.sql
```

Where:
- `{timestamp}`: Format `YYYYMMDDHHMMSS` (e.g., `20250830133803`)
- `{description}`: Descriptive name using underscores or hyphens
- `{direction}`: Either `up` or `down`

### Example Migration Files

**20250830133803_create_users.up.sql**:
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**20250830133803_create_users.down.sql**:
```sql
DROP TABLE users;
```

## Project Structure

```
.
├── cmd/
│   └── migrate/
│       └── main.go          # CLI entry point
├── command/                 # Command implementations
│   ├── command.go          # Command parser and router
│   ├── generate.go         # Generate command
│   ├── up.go              # Up command
│   ├── down.go            # Down/rollback command
│   ├── reset.go           # Reset command
│   └── status.go          # Status command
├── config/
│   └── config.go          # Configuration structures
├── migrate/               # Core migration logic
│   ├── migrate.go         # Main migration orchestrator
│   ├── file.go           # Migration file handling
│   ├── entity.go         # Database entities
│   └── repository.go     # Database operations
├── db/
│   └── migrations/       # Migration files directory
└── go.mod
```

## Database Schema

The tool automatically creates a `migrations` table to track applied migrations:

```sql
CREATE TABLE migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT NOW()
);
```

## Development

### Running Tests

```bash
go test ./...
```

### Dependencies

- `github.com/lib/pq` - PostgreSQL driver
- `github.com/stretchr/testify` - Testing framework
- `github.com/gooolib/errors` - Error handling

## License

Licensed under the terms specified in the LICENSE file.