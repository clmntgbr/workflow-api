package migrate

import (
	"database/sql"
	"fmt"

	"go-api/migrations"

	"github.com/pressly/goose/v3"
)

const dir = "."

func Up(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	goose.SetBaseFS(migrations.FS)
	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func Down(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	goose.SetBaseFS(migrations.FS)
	if err := goose.Down(db, dir); err != nil {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}

func Status(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	goose.SetBaseFS(migrations.FS)
	if err := goose.Status(db, dir); err != nil {
		return fmt.Errorf("migrate status: %w", err)
	}
	return nil
}
