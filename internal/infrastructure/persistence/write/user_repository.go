package write

import (
	"context"
	"errors"

	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userWriteRepository struct {
	db *gorm.DB
}

func NewUserWriteRepository(db *gorm.DB) domainuser.UserWriteRepository {
	return &userWriteRepository{db: db}
}

func (r *userWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *userWriteRepository) Save(ctx context.Context, user *domainuser.User) error {
	return DBWithContext(ctx, r.db).Create(userModelFromDomain(user)).Error
}

func (r *userWriteRepository) Update(ctx context.Context, user *domainuser.User) error {
	return DBWithContext(ctx, r.db).Save(userModelFromDomain(user)).Error
}

func (r *userWriteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return DBWithContext(ctx, r.db).Delete(&UserModel{}, id).Error
}

func (r *userWriteRepository) DeleteByClerkID(ctx context.Context, clerkID string) error {
	return DBWithContext(ctx, r.db).Where("clerk_id = ?", clerkID).Delete(&UserModel{}).Error
}

func (r *userWriteRepository) GetByClerkID(ctx context.Context, clerkID string) (*domainuser.User, error) {
	var model UserModel
	err := DBWithContext(ctx, r.db).Where("clerk_id = ?", clerkID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return userDomainFromModel(&model), nil
}

func (r *userWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	var model UserModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return userDomainFromModel(&model), nil
}

func (r *userWriteRepository) GetBySubscriptionID(
	ctx context.Context,
	subscriptionID uuid.UUID,
) (*domainuser.User, error) {
	var model UserModel
	err := DBWithContext(ctx, r.db).
		Where("subscription_id = ?", subscriptionID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return userDomainFromModel(&model), nil
}
