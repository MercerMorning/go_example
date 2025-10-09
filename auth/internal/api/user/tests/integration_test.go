package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/MercerMorning/go_example/auth/internal/api/user"
	"github.com/MercerMorning/go_example/auth/internal/client/db"
	"github.com/MercerMorning/go_example/auth/internal/client/db/pg"
	"github.com/MercerMorning/go_example/auth/internal/client/db/transaction"
	"github.com/MercerMorning/go_example/auth/internal/model"
	"github.com/MercerMorning/go_example/auth/internal/repository"
	userRepository "github.com/MercerMorning/go_example/auth/internal/repository/user"
	"github.com/MercerMorning/go_example/auth/internal/service"
	userService "github.com/MercerMorning/go_example/auth/internal/service/user"
	desc "github.com/MercerMorning/go_example/auth/pkg/user_v1"
)

// IntegrationTestSuite содержит настройки для integration тестов
type IntegrationTestSuite struct {
	suite.Suite
	dbClient       db.Client
	txManager      db.TxManager
	userRepository repository.UserRepository
	userService    service.UserService
	userAPI        *user.Implementation
	ctx            context.Context
}

// SetupSuite выполняется один раз перед всеми тестами
func (suite *IntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	
	// Настройка тестовой БД
	suite.setupTestDB()
	
	// Инициализация зависимостей
	suite.setupDependencies()
	
	// Создание таблиц
	suite.createTables()
}

// TearDownSuite выполняется один раз после всех тестов
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.dbClient != nil {
		suite.dbClient.Close()
	}
}

// SetupTest выполняется перед каждым тестом
func (suite *IntegrationTestSuite) SetupTest() {
	// Очистка таблиц перед каждым тестом
	suite.cleanupTables()
}

// setupTestDB настраивает подключение к тестовой БД
func (suite *IntegrationTestSuite) setupTestDB() {
	// Используем тестовую БД или создаем подключение к локальной
	testDSN := os.Getenv("TEST_PG_DSN")
	if testDSN == "" {
		// Fallback на локальную БД для тестов
		testDSN = "postgres://postgres:postgres@localhost:54321/auth_test?sslmode=disable"
	}
	
	client, err := pg.New(suite.ctx, testDSN)
	require.NoError(suite.T(), err, "Failed to connect to test database")
	
	// Проверяем соединение
	err = client.DB().Ping(suite.ctx)
	require.NoError(suite.T(), err, "Failed to ping test database")
	
	suite.dbClient = client
}

// setupDependencies инициализирует все зависимости
func (suite *IntegrationTestSuite) setupDependencies() {
	// Создаем менеджер транзакций
	suite.txManager = transaction.NewTransactionManager(suite.dbClient.DB())
	
	// Создаем репозиторий
	suite.userRepository = userRepository.NewRepository(suite.dbClient)
	
	// Создаем сервис
	suite.userService = userService.NewService(suite.userRepository, suite.txManager)
	
	// Создаем API
	suite.userAPI = user.NewImplementation(suite.userService)
}

// createTables создает необходимые таблицы
func (suite *IntegrationTestSuite) createTables() {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'USER',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE
		);
	`
	
	q := db.Query{
		Name:     "create_users_table",
		QueryRaw: createTableQuery,
	}
	
	_, err := suite.dbClient.DB().ExecContext(suite.ctx, q)
	require.NoError(suite.T(), err, "Failed to create users table")
}

// cleanupTables очищает таблицы
func (suite *IntegrationTestSuite) cleanupTables() {
	truncateQuery := `TRUNCATE TABLE users RESTART IDENTITY CASCADE;`
	
	q := db.Query{
		Name:     "truncate_users_table",
		QueryRaw: truncateQuery,
	}
	
	_, err := suite.dbClient.DB().ExecContext(suite.ctx, q)
	require.NoError(suite.T(), err, "Failed to truncate users table")
}

// TestCreateUserIntegration тестирует создание пользователя с реальной БД
func (suite *IntegrationTestSuite) TestCreateUserIntegration() {
	// Подготовка тестовых данных
	name := gofakeit.Name()
	email := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, true, 13)
	passwordConfirm := password // Для успешного теста пароли должны совпадать
	role := desc.Role_USER

	req := &desc.CreateRequest{
		Name:            name,
		Email:           email,
		Password:        password,
		PasswordConfirm: passwordConfirm,
		Role:            role,
	}

	// Выполнение запроса
	response, err := suite.userAPI.Create(suite.ctx, req)

	// Проверки
	require.NoError(suite.T(), err, "Create user should not return error")
	require.NotNil(suite.T(), response, "Response should not be nil")
	require.Greater(suite.T(), response.Id, int64(0), "User ID should be greater than 0")

	// Проверяем, что пользователь действительно создался в БД
	createdUser, err := suite.userRepository.Get(suite.ctx, response.Id)
	require.NoError(suite.T(), err, "Should be able to get created user")
	require.NotNil(suite.T(), createdUser, "Created user should not be nil")
	require.Equal(suite.T(), name, createdUser.Info.Name, "User name should match")
	require.Equal(suite.T(), email, createdUser.Info.Email, "User email should match")
	require.Equal(suite.T(), password, createdUser.Info.Password, "User password should match")
	require.Equal(suite.T(), "USER", createdUser.Info.Role, "User role should match")
	require.False(suite.T(), createdUser.CreatedAt.IsZero(), "CreatedAt should be set")
}

// TestCreateUserWithDuplicateEmail тестирует создание пользователя с дублирующимся email
func (suite *IntegrationTestSuite) TestCreateUserWithDuplicateEmail() {
	// Создаем первого пользователя
	email := gofakeit.Email()
	firstUser := &model.UserInfo{
		Name:     gofakeit.Name(),
		Email:    email,
		Password: gofakeit.Password(true, true, true, true, true, 13),
		Role:     "USER",
	}

	_, err := suite.userRepository.Create(suite.ctx, firstUser)
	require.NoError(suite.T(), err, "First user creation should succeed")

	// Пытаемся создать второго пользователя с тем же email
	secondUserReq := &desc.CreateRequest{
		Name:            gofakeit.Name(),
		Email:           email, // Дублирующийся email
		Password:        gofakeit.Password(true, true, true, true, true, 13),
		PasswordConfirm: gofakeit.Password(true, true, true, true, true, 13),
		Role:            desc.Role_USER,
	}

	_, err = suite.userAPI.Create(suite.ctx, secondUserReq)
	require.Error(suite.T(), err, "Creating user with duplicate email should fail")
}

// TestCreateUserWithDifferentPasswords тестирует создание пользователя с разными паролями
func (suite *IntegrationTestSuite) TestCreateUserWithDifferentPasswords() {
	req := &desc.CreateRequest{
		Name:            gofakeit.Name(),
		Email:           gofakeit.Email(),
		Password:        "password123",
		PasswordConfirm: "different_password", // Разные пароли
		Role:            desc.Role_USER,
	}

	_, err := suite.userAPI.Create(suite.ctx, req)
	require.Error(suite.T(), err, "Creating user with different passwords should fail")
}

// TestCreateUserWithAdminRole тестирует создание пользователя с ролью ADMIN
func (suite *IntegrationTestSuite) TestCreateUserWithAdminRole() {
	req := &desc.CreateRequest{
		Name:            gofakeit.Name(),
		Email:           gofakeit.Email(),
		Password:        gofakeit.Password(true, true, true, true, true, 13),
		PasswordConfirm: gofakeit.Password(true, true, true, true, true, 13),
		Role:            desc.Role_ADMIN,
	}

	response, err := suite.userAPI.Create(suite.ctx, req)
	require.NoError(suite.T(), err, "Create admin user should not return error")
	require.NotNil(suite.T(), response, "Response should not be nil")

	// Проверяем роль в БД
	createdUser, err := suite.userRepository.Get(suite.ctx, response.Id)
	require.NoError(suite.T(), err, "Should be able to get created admin user")
	require.Equal(suite.T(), "ADMIN", createdUser.Info.Role, "User role should be ADMIN")
}

// TestCreateUserPerformance тестирует производительность создания пользователей
func (suite *IntegrationTestSuite) TestCreateUserPerformance() {
	const userCount = 100
	
	start := time.Now()
	
	for i := 0; i < userCount; i++ {
		req := &desc.CreateRequest{
			Name:            gofakeit.Name(),
			Email:           gofakeit.Email(),
			Password:        gofakeit.Password(true, true, true, true, true, 13),
			PasswordConfirm: gofakeit.Password(true, true, true, true, true, 13),
			Role:            desc.Role_USER,
		}

		_, err := suite.userAPI.Create(suite.ctx, req)
		require.NoError(suite.T(), err, fmt.Sprintf("User %d creation should succeed", i+1))
	}
	
	duration := time.Since(start)
	suite.T().Logf("Created %d users in %v (avg: %v per user)", 
		userCount, duration, duration/userCount)
	
	// Проверяем, что все пользователи создались
	var count int
	countQuery := `SELECT COUNT(*) FROM users;`
	q := db.Query{
		Name:     "count_users",
		QueryRaw: countQuery,
	}
	
	err := suite.dbClient.DB().QueryRowContext(suite.ctx, q).Scan(&count)
	require.NoError(suite.T(), err, "Should be able to count users")
	require.Equal(suite.T(), userCount, count, "All users should be created")
}

// TestIntegrationSuite запускает все integration тесты
func TestIntegrationSuite(t *testing.T) {
	// Пропускаем тесты, если не настроена тестовая БД
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Integration tests are skipped")
	}
	
	suite.Run(t, new(IntegrationTestSuite))
}
