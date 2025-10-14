package user

import (
	"context"
	"log"

	"github.com/MercerMorning/go_example/auth/internal/converter"
	desc "github.com/MercerMorning/go_example/auth/pkg/user_v1"
)

func (i *Implementation) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	userInfo := converter.ToUserInfoFromDesc(req)

	// Сначала создаем пользователя в локальном сервисе
	id, err := i.userService.Create(ctx, userInfo)
	if err != nil {
		return nil, err
	}

	// Затем вызываем other_service через gRPC
	otherServiceReq := &desc.CreateRequest{
		Name:            req.GetName(),
		Email:           req.GetEmail(),
		Password:        req.GetPassword(),
		PasswordConfirm: req.GetPasswordConfirm(),
		Role:            req.GetRole(),
	}

	otherServiceResp, err := i.userClient.Create(ctx, otherServiceReq)
	if err != nil {
		log.Printf("Failed to create user in other_service: %v", err)
		// Возвращаем успешный ответ от локального сервиса, даже если other_service не сработал
		// В реальном приложении здесь может быть другая логика обработки ошибок
	} else {
		log.Printf("Successfully created user in other_service with ID: %d", otherServiceResp.GetId())
	}

	return converter.ToCreateResponseFromID(id), nil
}
