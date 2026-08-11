package write

import (
	"context"

	"gorm.io/gorm"
)

type txCtxKey struct{}

func ContextWithTx(ctx context.Context, db *gorm.DB) context.Context {
	if db == nil {
		return ctx
	}
	return context.WithValue(ctx, txCtxKey{}, db)
}

func DBWithContext(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if v, ok := ctx.Value(txCtxKey{}).(*gorm.DB); ok && v != nil {
		return v.WithContext(ctx)
	}
	return defaultDB.WithContext(ctx)
}
