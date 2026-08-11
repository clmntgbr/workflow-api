package command

import (
	"database/sql"
	"fmt"

	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/migrate"
	"go-api/internal/infrastructure/persistence/schema"

	"github.com/spf13/cobra"
)

func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "up",
			Short: "Apply all pending SQL migrations",
			RunE: func(cmd *cobra.Command, args []string) error {
				sqlDB, err := openSQL()
				if err != nil {
					return err
				}
				defer sqlDB.Close()

				if err := migrate.Up(sqlDB); err != nil {
					return err
				}
				fmt.Println("Migrations applied successfully")
				return nil
			},
		},
		&cobra.Command{
			Use:   "down",
			Short: "Roll back the last SQL migration",
			RunE: func(cmd *cobra.Command, args []string) error {
				sqlDB, err := openSQL()
				if err != nil {
					return err
				}
				defer sqlDB.Close()

				if err := migrate.Down(sqlDB); err != nil {
					return err
				}
				fmt.Println("Last migration rolled back successfully")
				return nil
			},
		},
		&cobra.Command{
			Use:   "status",
			Short: "Show migration status",
			RunE: func(cmd *cobra.Command, args []string) error {
				sqlDB, err := openSQL()
				if err != nil {
					return err
				}
				defer sqlDB.Close()
				return migrate.Status(sqlDB)
			},
		},
		&cobra.Command{
			Use:   "check",
			Short: "Fail if persistence models and database schema diverge",
			RunE: func(cmd *cobra.Command, args []string) error {
				env := config.Load()
				db := config.ConnectDatabase(env)
				if err := schema.AssertModelsMatchDB(db); err != nil {
					return err
				}
				fmt.Println("Schema check passed: models match database")
				return nil
			},
		},
	)

	return cmd
}

func openSQL() (*sql.DB, error) {
	env := config.Load()
	db := config.ConnectDatabase(env)
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	return sqlDB, nil
}
