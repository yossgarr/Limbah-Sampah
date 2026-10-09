package database

import (
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Embed seluruh file .sql yang ada di folder migrations
//go:embed migrations/*.sql
var migrationFS embed.FS

func RunMigrations(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("gagal membuat driver postgres migration: %w", err)
	}

	sourceDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("gagal membaca embedded migrations: %w", err)
	}

	m, err := migrate.NewWithInstance(
		"iofs", sourceDriver,
		"postgres", driver,
	)
	if err != nil {
		return fmt.Errorf("gagal inisialisasi migration instance: %w", err)
	}

	// Jalankan migrasi sampai versi paling baru
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("gagal menjalankan migration up: %w", err)
	}

	log.Println("✅ Database migration berhasil dijalankan / sudah up-to-date")
	return nil
}
