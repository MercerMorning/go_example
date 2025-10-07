package repository

import (
	"context"

	"github.com/MercerMorning/go_example/auth/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, info *model.NoteInfo) (int64, error)
}
