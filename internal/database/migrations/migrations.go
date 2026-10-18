// Package migrations holds versioned schema migrations.
//
// Each migration lives in its own file named <timestamp>_<description>.go
// and registers itself in init(). Bun derives the migration name from the
// file name, so keep the naming scheme.
//
// Migrations deliberately declare their own snapshot structs instead of
// importing internal/models: a migration must keep producing the same schema
// even after the models change. Write follow-up migrations with ALTER TABLE
// (raw SQL via db.ExecContext) or new snapshot structs.
package migrations

import "github.com/uptrace/bun/migrate"

// Migrations is the registry used by the migrator.
var Migrations = migrate.NewMigrations()
