package schema

import (
	"fmt"

	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/processed"
	"go-api/internal/infrastructure/persistence/write"

	"gorm.io/gorm"
)

func Models() []any {
	return []any{
		&write.UserModel{},
		&write.OrganizationModel{},
		&write.UserOrganizationModel{},
		&write.WorkflowModel{},
		&write.EndpointModel{},
		&write.StepModel{},
		&write.ConnectionModel{},
		&write.WorkflowRunModel{},
		&write.StepRunModel{},
		&write.InsightModel{},
		&write.VariableModel{},
		&outbox.OutboxEvent{},
		&processed.ProcessedEvent{},
	}
}

func AssertModelsMatchDB(db *gorm.DB) error {
	for _, model := range Models() {
		if err := assertModelMatchesDB(db, model); err != nil {
			return err
		}
	}
	return nil
}

func assertModelMatchesDB(db *gorm.DB, model any) error {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return fmt.Errorf("parse model: %w", err)
	}

	table := stmt.Schema.Table
	if !db.Migrator().HasTable(table) {
		return fmt.Errorf("schema drift: table %q missing in database (run migrations)", table)
	}

	columnTypes, err := db.Migrator().ColumnTypes(table)
	if err != nil {
		return fmt.Errorf("schema drift: inspect table %q: %w", table, err)
	}

	dbColumns := make(map[string]struct{}, len(columnTypes))
	for _, col := range columnTypes {
		dbColumns[col.Name()] = struct{}{}
	}

	modelColumns := make(map[string]struct{})
	for _, field := range stmt.Schema.Fields {
		if field.DBName == "" || field.IgnoreMigration {
			continue
		}
		modelColumns[field.DBName] = struct{}{}
		if _, ok := dbColumns[field.DBName]; !ok {
			return fmt.Errorf(
				"schema drift: table %q column %q exists on model but not in database",
				table,
				field.DBName,
			)
		}
	}

	for name := range dbColumns {
		if _, ok := modelColumns[name]; !ok {
			return fmt.Errorf(
				"schema drift: table %q column %q exists in database but not on model",
				table,
				name,
			)
		}
	}

	return nil
}
