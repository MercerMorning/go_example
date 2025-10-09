package user

import (
	"context"

	"github.com/MercerMorning/go_example/auth/internal/converter"
	desc "github.com/MercerMorning/go_example/auth/pkg/user_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i *Implementation) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	user, err := i.userService.Get(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return converter.ToDescFromUser(user), nil
}




