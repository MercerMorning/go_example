package user

import (
	"github.com/MercerMorning/go_example/auth/internal/client/db"
	"github.com/MercerMorning/go_example/auth/internal/repository"
	"github.com/MercerMorning/go_example/auth/internal/service"
)

type serv struct {
	userRepository repository.UserRepository
	txManager      db.TxManager
}

func NewService(
	userRepository repository.UserRepository,
	txManager db.TxManager,
) service.UserService {
	return &serv{
		userRepository: userRepository,
		txManager:      txManager,
	}
}
