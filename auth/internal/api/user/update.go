package user

import (
	"context"

	"github.com/MercerMorning/go_example/auth/internal/converter"
	desc "github.com/MercerMorning/go_example/auth/pkg/user_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) Update(ctx context.Context, req *desc.UpdateRequest) (*emptypb.Empty, error) {
	userUpdate := converter.ToUserUpdateFromDesc(req)
	
	err := i.userService.Update(ctx, req.GetId(), userUpdate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &emptypb.Empty{}, nil
}
