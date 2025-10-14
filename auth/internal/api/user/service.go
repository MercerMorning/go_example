package user

import (
	"github.com/MercerMorning/go_example/auth/internal/service"
	desc "github.com/MercerMorning/go_example/auth/pkg/user_v1"
)

type Implementation struct {
	desc.UnimplementedUserV1Server
	userService service.UserService
	userClient  desc.UserV1Client
}

func NewImplementation(userService service.UserService, userClient desc.UserV1Client) *Implementation {
	return &Implementation{
		userService: userService,
		userClient:  userClient,
	}
}
