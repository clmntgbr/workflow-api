package di

import (
	"go-api/internal/infrastructure/config"

	"gorm.io/gorm"
)

type Container struct{}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	return &Container{}
}
