package note

import (
	"github.com/MercerMorning/go_example/auth/internal/client/db"
	"github.com/MercerMorning/go_example/auth/internal/repository"
	"github.com/MercerMorning/go_example/auth/internal/service"
)

type serv struct {
	noteRepository repository.UserRepository
	txManager      db.TxManager
}

func NewService(
	noteRepository repository.UserRepository,
	txManager db.TxManager,
) service.UserService {
	return &serv{
		noteRepository: noteRepository,
		txManager:      txManager,
	}
}
