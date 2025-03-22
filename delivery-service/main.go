package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
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
	DB_DSN := cfg.DB.DSN
	topic := cfg.Kafka.Topic
	redisHost := cfg.Redis.Host
	redisPassword := cfg.Redis.Password

	// создание нового клиента Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPassword,
		DB:       0,
	})
	defer rdb.Close()

	db, err := sql.Open("postgres", DB_DSN)
	if err != nil {
		log.Fatalf("неудачное соединение с бд: %v", err)
	}
	defer db.Close()

	//создаю консьюмера
	kReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(kafkaBroker, ","),
		GroupID: "delivery-group",
		Topic:   topic,
	})

	//создаю канал, читающий сигналы ос
	stop := make(chan os.Signal, 1)
	//подписка на оповещение от ос о сигналах SIGINT(нажатие ctrl+c) и SIGTERM(завершение процесса)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	//создаю контекст с отменой по сигналу
	readctx, readCancel := context.WithCancel(context.Background())
	go func() {
		<-stop
		readCancel()
	}()

	// запускаю горутину
	go func() {
		for {
			// читаю из брокера если ctx не отменен
			m, err := kReader.ReadMessage(readctx)
			if err != nil {
				if readctx.Err() != nil {
					break
				}
				log.Printf("неудачное чтение из брокера: %v", err)
				time.Sleep(1 * time.Second) //задержка при повторной попытке чтения
				continue
			}
			// достаю данные из сообщения и десериализую
			var order Order
			if err := json.Unmarshal(m.Value, &order); err != nil {
				log.Printf("неудачная сериализация сообщения: %v", err)
				continue
			}

			// Имитация доставки заказа
			time.Sleep(3 * time.Second)

			order.Status = "delivered"

			// Обновление статуса заказа в Redis и PostgreSQL
			err = rdb.Set(context.Background(), order.ID, string(m.Value), 1*time.Hour).Err()
			if err != nil {
				log.Printf("Ошибка обновления статуса заказа в Redis: %v", err)
			}

			_, err = db.ExecContext(context.Background(), "UPDATE orders SET status=$1 WHERE id=$2", order.Status, order.ID)
			if err != nil {
				log.Printf("Ошибка обновления статуса заказа в PostgreSQL: %v", err)
				continue
			}

			log.Printf("Заказ %s доставлен", order.ID)
		}
	}()

	// Graceful Shutdown

	// Контекст с таймаутом для корректного завершения всех операций
	shutdownctx, shutdowCancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer shutdowCancel()
	// ожидание сигнала
	<-stop

	log.Println("Остановка Kitchen-service")

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
	// Ожидание завершения всех операций
	<-shutdownctx.Done()

	log.Println("Kitchen-service остановлен")
}
