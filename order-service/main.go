package main

import (
	"context"
	"database/sql"
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

	"github.com/sandrinasava/cafe-services/order-service/handlers"
	"github.com/sandrinasava/cafe-services/order-service/models"
)

func main() {
	// получаю конфиги
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Неудачная загрузка конфигураций: %v", err)
	}

	kafkaBroker := cfg.Kafka.Broker
	redisHost := cfg.Redis.Host
	authServiceAddress := cfg.AuthService.Address
	dbDSN := cfg.DB.DSN

	// создание нового клиента Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	// создание клиента gRPC для auth-service
	authClient, err := models.NewAuthClient(authServiceAddress)
	if err != nil {
		log.Fatalf("Не удалось создать клиента для auth-service: %v", err)
	}
	defer authClient.Close()

	// Подключение к PostgreSQL
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	// Создание консюмера Kafka для получения сообщений о доставке
	kReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   models.TopicOrderDelivered,
		GroupID: "order-service",
	})
	defer kReader.Close()

	//создаю продюсера
	kWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    models.TopicNewOrders,
		Balancer: &kafka.LeastBytes{},
	})
	defer kWriter.Close()

	// регистрация маршрутов

	http.HandleFunc("/order", handlers.OrderHandler(rdb, db, authClient, kWriter))

	http.HandleFunc("/order/status", handlers.StatusHandler(rdb, db))

	http.HandleFunc("/login", handlers.AuthZHandler(rdb, db, authClient))

	http.HandleFunc("/register", handlers.RegistHandler(authClient))

	// Добавляю таймауты для сервера для предотвращения долгих блокировок
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Запуск сервера в отдельной горутине
	go func() {
		log.Printf("Сервис заказов слушает на порту %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Обработка сообщений из Kafka
	go func() {
		for {
			m, err := kReader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Ошибка при чтении сообщения из Kafka: %v", err)
				continue
			}

			var order models.Order
			err = json.Unmarshal(m.Value, &order)
			if err != nil {
				log.Printf("Ошибка при десериализации сообщения: %v", err)
				continue
			}

			// Обновление статуса заказа в Redis и PostgreSQL
			err = rdb.Set(context.Background(), order.ID.String(), string(m.Value), 1*time.Hour).Err()
			if err != nil {
				log.Printf("Ошибка обновления статуса заказа в Redis: %v", err)
			}

			_, err = db.ExecContext(context.Background(), "UPDATE orders SET status=$1 WHERE id=$2", order.Status, order.ID)
			if err != nil {
				log.Printf("Ошибка обновления статуса заказа в PostgreSQL: %v", err)
			}
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

	// закрытие консьюмера
	if err := kReader.Close(); err != nil {
		log.Printf("Не удалось закрыть консьюмера Kafka: %v", err)
	}

	// закрытие клиента Redis
	if err := rdb.Close(); err != nil {
		log.Printf("Не удалось закрыть клиент Redis: %v", err)
	}
	// закрытие бд
	if err := db.Close(); err != nil {
		log.Printf("Не удалось закрыть соединение с базой данных: %v", err)
	}

	log.Println("Order-service остановлен")
}
