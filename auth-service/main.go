package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/sandrinasava/cafe-services/auth-service/proto"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

//go:embed init.sql
var initSQL string

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	db        *sql.DB
	jwtSecret string
}

func NewAuthServer(db *sql.DB, jwtSecret string) *AuthServer {
	return &AuthServer{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("не удалось зашифровать пароль: %w", err)
	}

	userID := uuid.New()
	query := `
        INSERT INTO users (id, username, password, email, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err = s.db.ExecContext(ctx, query, userID, req.Username, hashedPassword, req.Email, time.Now())
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_username_key\"" {
			return nil, fmt.Errorf("имя пользователя уже существует")
		}
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			return nil, fmt.Errorf("email уже существует")
		}
		return nil, fmt.Errorf("не удалось зарегистрировать пользователя: %w", err)
	}

	return &pb.RegisterResponse{
		Message: "Пользователь успешно зарегистрирован",
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var user struct {
		ID       uuid.UUID
		Username string
		Password string
	}
	query := `
        SELECT id, username, password FROM users WHERE username = $1
    `
	err := s.db.QueryRowContext(ctx, query, req.Username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("неверное имя пользователя или пароль")
		}
		return nil, fmt.Errorf("не удалось выполнить запрос: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("неверное имя пользователя или пароль")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать токен: %w", err)
	}

	return &pb.LoginResponse{
		Token: tokenString,
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	tokenString := req.Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid: token.Valid,
		Error: "",
	}, nil
}

func runMigrations(db *sql.DB) error {
	_, err := db.Exec(initSQL)
	if err != nil {
		return fmt.Errorf("не удалось выполнить миграции: %w", err)
	}
	log.Println("Миграции успешно выполнены")
	return nil
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Неудачная загрузка конфигураций: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DBDSN)
	if err != nil {
		log.Fatalf("Не удалось открыть соединение с базой данных: %v", err)
	}
	defer db.Close()

	// Проверка подключения к базе данных
	err = db.PingContext(context.Background())
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	log.Printf("Подключение к базе данных успешно установлено")

	// Выполнение миграций
	err = runMigrations(db)
	if err != nil {
		log.Fatalf("Не удалось выполнить миграции: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("Не удалось прослушивать порт: %v", err)
	}
	log.Printf("Аутентификационный сервис слушает на порту %d", cfg.Port)

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, NewAuthServer(db, cfg.JwtSecret))

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Не удалось запустить gRPC сервер: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("Остановка Auth-service")

	// Создаём контекст с таймаутом для корректного завершения сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Корректное завершение gRPC сервера
	s.GracefulStop()

	// Закрытие соединения с базой данных
	if err := db.Close(); err != nil {
		log.Printf("Не удалось закрыть соединение с базой данных: %v", err)
	}

	log.Println("Auth-service остановлен")
}
