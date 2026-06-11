package repository

import (
    "context"
    "smieci-sms/internal/model"
)

type UserRepository interface {
    SaveUser(ctx context.Context, user model.User, address model.Address) error
    ListUsers(ctx context.Context) ([]model.User, error)
}

type userRepository struct {
    db *DB
}

func NewUserRepository(db *DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) SaveUser(ctx context.Context, user model.User, address model.Address) error {
    // TODO: implement location lookup by address and persist user with location_id
    return nil
}

func (r *userRepository) ListUsers(ctx context.Context) ([]model.User, error) {
    // TODO: implement query logic
    return []model.User{}, nil
}
