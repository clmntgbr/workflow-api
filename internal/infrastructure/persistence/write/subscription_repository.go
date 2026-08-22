package write

import (
	"context"
	"errors"

	domainsubscription "go-api/internal/domain/subscription"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type subscriptionWriteRepository struct {
	db *gorm.DB
}

func NewSubscriptionWriteRepository(db *gorm.DB) domainsubscription.SubscriptionWriteRepository {
	return &subscriptionWriteRepository{db: db}
}

func (r *subscriptionWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *subscriptionWriteRepository) Save(ctx context.Context, subscription *domainsubscription.Subscription) error {
	return DBWithContext(ctx, r.db).Create(subscriptionModelFromDomain(subscription)).Error
}

func (r *subscriptionWriteRepository) Update(ctx context.Context, subscription *domainsubscription.Subscription) error {
	return DBWithContext(ctx, r.db).Save(subscriptionModelFromDomain(subscription)).Error
}

func (r *subscriptionWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainsubscription.Subscription, error) {
	var model SubscriptionModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return subscriptionDomainFromModel(&model), nil
}

func (r *subscriptionWriteRepository) GetByStripeSubscriptionID(
	ctx context.Context,
	stripeSubscriptionID string,
) (*domainsubscription.Subscription, error) {
	var model SubscriptionModel
	err := DBWithContext(ctx, r.db).
		Where("stripe_subscription_id = ?", stripeSubscriptionID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return subscriptionDomainFromModel(&model), nil
}

func (r *subscriptionWriteRepository) GetByStripeCustomerID(
	ctx context.Context,
	stripeCustomerID string,
) (*domainsubscription.Subscription, error) {
	var model SubscriptionModel
	err := DBWithContext(ctx, r.db).
		Where("stripe_customer_id = ?", stripeCustomerID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return subscriptionDomainFromModel(&model), nil
}
