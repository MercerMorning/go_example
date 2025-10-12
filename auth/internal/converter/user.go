package converter

import (
	"github.com/MercerMorning/go_example/auth/internal/model"
	desc "github.com/MercerMorning/go_example/auth/pkg/user_v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToUserInfoFromDesc конвертирует CreateRequest в UserInfo
func ToUserInfoFromDesc(req *desc.CreateRequest) *model.UserInfo {
	return &model.UserInfo{
		Name:     req.GetName(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Role:     req.GetRole().String(),
	}
}

// ToUserUpdateFromDesc конвертирует UpdateRequest в UserUpdate
func ToUserUpdateFromDesc(req *desc.UpdateRequest) *model.UserUpdate {
	update := &model.UserUpdate{}

	if req.GetName() != nil {
		update.Name = &req.GetName().Value
	}

	if req.GetEmail() != nil {
		update.Email = &req.GetEmail().Value
	}

	return update
}

// ToDescFromUser конвертирует User в GetResponse
func ToDescFromUser(user *model.User) *desc.GetResponse {
	var updatedAt *timestamppb.Timestamp
	if user.UpdatedAt.Valid {
		updatedAt = timestamppb.New(user.UpdatedAt.Time)
	}

	role := desc.Role_USER
	if user.Info.Role == "ADMIN" {
		role = desc.Role_ADMIN
	}

	return &desc.GetResponse{
		Id:        user.ID,
		Name:      user.Info.Name,
		Email:     user.Info.Email,
		Role:      role,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: updatedAt,
	}
}

// ToCreateResponseFromID конвертирует ID в CreateResponse
func ToCreateResponseFromID(id int64) *desc.CreateResponse {
	return &desc.CreateResponse{
		Id: id,
	}
}

// ToUpdateRequestFromDesc создает UpdateRequest из GetResponse (для внутреннего использования)
func ToUpdateRequestFromDesc(req *desc.UpdateRequest) *model.UserUpdate {
	update := &model.UserUpdate{}

	if req.GetName() != nil {
		update.Name = &req.GetName().Value
	}

	if req.GetEmail() != nil {
		update.Email = &req.GetEmail().Value
	}

	return update
}






