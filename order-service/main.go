package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"

	pb "github.com/sandrinasava/cafe-services/auth-service/proto"
)

// определяю топик, в который будут отправляться сообщения
const (
	topic = "new_orders"
)

type Order struct {
	ID       string   `json:"id"`
	Customer string   `json:"customer"`
	Items    []string `json:"items"`
	Status   string   `json:"status"`
}

type AuthClient struct {
	client pb.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(address string) (*AuthClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к auth-service: %w", err)
	}
	return &AuthClient{
		client: pb.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}

func (c *AuthClient) ValidateToken(ctx context.Context, token string) (bool, error) {
	resp, err := c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		return false, fmt.Errorf("не удалось валидировать токен: %w", err)
	}
	return resp.Valid, nil
}

func main() {
	// получаю конфиги
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Неудачная загрузка конфигураций: %v", err)
	}

	kafkaBroker := cfg.Kafka.Broker
	redisHost := cfg.Redis.Host
	authServiceAddress := cfg.AuthService.Address

	// создание нового клиента Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	// создание клиента gRPC для auth-service
	authClient, err := NewAuthClient(authServiceAddress)
	if err != nil {
		log.Fatalf("Не удалось создать клиента для auth-service: %v", err)
	}
	defer authClient.Close()

	//создаю продюсера
	kWriter := kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer kWriter.Close()

	// хендлер для обработки заказа
	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не доступен", http.StatusMethodNotAllowed)
			return
		}

		// достаю данные из сообщения и десериализую
		var order Order
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// достаю токен из заголовков
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Токен авторизации не предоставлен", http.StatusUnauthorized)
			return
		}

		// валидация токена
		valid, err := authClient.ValidateToken(r.Context(), token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if !valid {
			http.Error(w, "Недействительный токен", http.StatusUnauthorized)
			return
		}

		// TODO(Aleksandrina): Добавить валидацию заказа перед отправкой в Kafka [2024-01-01]
		order.Status = "received"
		// сериализация
		message, err := json.Marshal(order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// отправляю заказ в Kafka
		err = kWriter.WriteMessages(r.Context(), kafka.Message{
			Key:   []byte(order.ID),
			Value: message,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// сохраняю заказ в кэше Redis на 1 час
		err = rdb.Set(r.Context(), order.ID, string(message), 1*time.Hour).Err()
		if err != nil {
			log.Printf("Ошибка кеширования сообщения: %v", err)
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Сообщение отправлено: %s\n", order.ID)
	})

	// Добавляю таймауты для сервера для предотвращения долгих блокировок
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Запуск сервера в отдельной горутине
	go func() {
		log.Printf("Сервис заказов слушает на порту %d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Graceful Shutdown
	// создаю канал, читающий сигналы ос
	stop := make(chan os.Signal, 1)
	// подписка на оповещение от ос о сигналах SIGINT(нажатие ctrl+c) и SIGTERM(завершение процесса)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// ожидание сигнала
	<-stop
	log.Println("Остановка Order-service")

	// создаю контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// завершение работы http-сервера
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Не удалось корректно остановить сервер: %v", err)
	}

	// закрытие консьюмера
	if err := kWriter.Close(); err != nil {
		log.Printf("Не удалось закрыть продюсера Kafka: %v", err)
	}

	// закрытие клиента Redis
	if err := rdb.Close(); err != nil {
		log.Printf("Не удалось закрыть клиент Redis: %v", err)
	}

	log.Println("Order-service остановлен")
}
