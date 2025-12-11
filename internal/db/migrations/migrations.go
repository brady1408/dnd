package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/nav-inc/pomegranate"
)

//go:embed 00001_initial_schema 00002_expanded_character_sheet 00003_add_size_gender
var embedded embed.FS

// Run executes all pending migrations against the database
func Run(db *sql.DB) error {
	migrationFS := pomegranate.FromEmbed(embedded, ".")

	migrations, err := pomegranate.ReadMigrationFS(migrationFS)
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	log.Printf("Found %d migrations", len(migrations))

	// Run all migrations forward to latest (empty string = latest)
	// false = don't prompt for confirmation
	err = pomegranate.MigrateForwardTo("", db, migrations, false)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}
