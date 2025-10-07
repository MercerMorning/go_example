package service

import (
	"context"

	"github.com/MercerMorning/go_example/auth/internal/model"
)

type UserService interface {
	Create(ctx context.Context, info *model.UserInfo) (int64, error)
	Get(ctx context.Context, id int64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, id int64, info *model.UserUpdate) error
	Delete(ctx context.Context, id int64) error
}
