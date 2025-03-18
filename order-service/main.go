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
)

// определяю топик, в который будут отпраляться сообщения
const (
	topic = "new_orders"
)

type Order struct {
	ID       string   `json:"id"`
	Customer string   `json:"customer"`
	Items    []string `json:"items"`
	Status   string   `json:"status"`
}

func main() {
	//получаю конфиги
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Неудачная загрузка конфигураций: %v", err)
	}

	kafkaBroker := cfg.Kafka.Broker
	redisHost := cfg.Redis.Host

	// создание нового клиента Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "",
		DB:       0,
	})
	//создаю продюсера
	kWriter := kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	//хендлер для обработки заказа
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

		order.Status = "received"
		// сериализация
		message, err := json.Marshal(order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// отправляю заказ в Kafka
		err = kWriter.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(order.ID),
			Value: message,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// cохраняю заказ в кэше Redis на 1 час
		ctx := context.Background()
		err = rdb.Set(ctx, order.ID, string(message), 1*time.Hour).Err()
		if err != nil {
			log.Printf("Ошибка кеширования сообщения: %v", err)
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Сообщение отправлено: %s\n", order.ID)
	})
	// Добавляю таймауты для сервера для предотвращения долгих блокировок
	srv := &http.Server{
		Addr:         ":8081",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	// Запуск сервера в отдельной горутине
	go func() {
		log.Printf("Сервис заказов слушает на порту 8081")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	//Graceful Shutdown
	//создаю канал, читающий сигналы ос
	stop := make(chan os.Signal, 1)
	//подписка на оповещение от ос о сигналах SIGINT(нажатие ctrl+c) и SIGTERM(завершение процесса)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	//ожидание сигнала
	<-stop
	log.Println("Остановка Kitchen-service")
	// создаю контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// завершение работы http-сервера
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Не удалось корректно остановить сервер: %v", err)
	}
	//закрытие консьюмера
	if err := kWriter.Close(); err != nil {
		log.Fatalf("Не удалось закрыть консьюмер: %v", err)
	}

	log.Println("Kitchen-service остановлен")
}
