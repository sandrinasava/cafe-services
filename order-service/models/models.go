package models

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	pb "github.com/sandrinasava/go-proto-module"
)

// Credentials представляет структуру данных для регистрации или авторизации
// @Description Структура данных, содержащая учетные данные пользователя
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// Order представляет структуру заказа
// @Description Заказ, содержащий информацию о клиенте, товарах и статусе
type Order struct {
	ID       uuid.UUID `json:"id"`
	Customer uuid.UUID `json:"customer"`
	Items    []string  `json:"items"`
	Status   string    `json:"status"`
}

// AuthClient представляет клиент для взаимодействия с auth-service через gRPC
type AuthClient struct {
	Client pb.AuthServiceClient
	Conn   *grpc.ClientConn
}

// Close закрывает соединение с auth-service
func (c *AuthClient) Close() error {
	return c.Conn.Close()
}

// ValidateToken проверяет валидность токена
func (c *AuthClient) ValidateToken(ctx context.Context, token string) (bool, error) {
	resp, err := c.Client.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		return false, fmt.Errorf("не удалось валидировать токен: %w", err)
	}
	return resp.Valid, nil
}

// Login выполняет вход пользователя и возвращает токен
func (c *AuthClient) Login(ctx context.Context, username, password string) (string, error) {
	resp, err := c.Client.Login(ctx, &pb.LoginRequest{Username: username, Password: password})
	if err != nil {
		return "", fmt.Errorf("не удалось выполнить вход: %w", err)
	}
	return resp.Token, nil
}

func (c *AuthClient) Register(ctx context.Context, username, password, email string) error {
	_, err := c.Client.Register(ctx, &pb.RegisterRequest{Username: username, Password: password, Email: email})
	if err != nil {
		return fmt.Errorf("не удалось зарегистрироваться: %w", err)
	}
	return nil
}

// NewAuthClient создает новый клиент для взаимодействия с auth-service по gRPC
func NewAuthClient(address string) (*AuthClient, error) {
	//установка соединения с сервером gRPC
	conn, err := grpc.NewClient(address)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к auth-service: %w", err)
	}
	//создание клиента
	return &AuthClient{
		Client: pb.NewAuthServiceClient(conn),
		Conn:   conn,
	}, nil
}
