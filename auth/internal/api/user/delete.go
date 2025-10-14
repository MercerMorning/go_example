package user

import (
	"context"

	desc "github.com/MercerMorning/go_example/auth/pkg/user_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	err := i.userService.Delete(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &emptypb.Empty{}, nil
}








