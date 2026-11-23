# Database Migrations

The Go backend reuses the existing `redbook` MySQL schema and does not require a destructive migration.

For a fresh database, apply the original baseline schema:

`redbook-backend/src/main/resources/db/schema.sql`

Then apply any required additive migrations from:

`redbook-backend/src/main/resources/db/migration_*.sql`

Alternatively, starting the Go service runs GORM `AutoMigrate` for the mapped models, which creates missing tables and adds missing columns without dropping existing data.
