package user

import (
	"context"

	"github.com/MercerMorning/go_example/auth/internal/model"
)

func (s *serv) Get(ctx context.Context, id int64) (*model.User, error) {
	return s.userRepository.Get(ctx, id)
}

func (s *serv) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.userRepository.GetByEmail(ctx, email)
}




