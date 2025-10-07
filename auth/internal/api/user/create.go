package user

import (
	"context"

	"github.com/MercerMorning/go_example/auth/internal/converter"
	desc "github.com/MercerMorning/go_example/auth/pkg/user_v1"
)

func (i *Implementation) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	userInfo := converter.ToUserInfoFromDesc(req)
	
	id, err := i.userService.Create(ctx, userInfo)
	if err != nil {
		return nil, err
	}

	return converter.ToCreateResponseFromID(id), nil
}
