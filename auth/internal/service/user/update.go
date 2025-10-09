package user

import (
	"context"

	"github.com/MercerMorning/go_example/auth/internal/model"
)

func (s *serv) Update(ctx context.Context, id int64, info *model.UserUpdate) error {
	return s.txManager.ReadCommitted(ctx, func(ctx context.Context) error {
		return s.userRepository.Update(ctx, id, info)
	})
}




