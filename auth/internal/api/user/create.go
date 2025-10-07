package user

import (
	"context"

	desc "github.com/MercerMorning/go_example/auth/pkg/user_v1"
)

func (i *Implementation) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	// id, err := i.userService.Create(ctx, converter.ToNoteInfoFromDesc(req.GetInfo()))
	// if err != nil {
	// 	return nil, err
	// }

	// log.Printf("inserted note with id: %d", id)

	return &desc.CreateResponse{
		Id: 11,
	}, nil
	// return &desc.CreateResponse{
	// 	Id: id,
	// }, nil
}
