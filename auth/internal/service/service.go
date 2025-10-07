package service

import (
	"context"

	"github.com/MercerMorning/go_example/auth/internal/model"
)

type UserService interface {
	Create(ctx context.Context, info *model.NoteInfo) (int64, error)
}
