package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

const (
	topic = "ready_orders"
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
	dbHost := cfg.Postgres.Host
	dbPort := cfg.Postgres.Port
	dbUser := cfg.Postgres.User
	dbPass := cfg.Postgres.Password
	dbName := cfg.Postgres.DBName

	//запускаю бд
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", connStr)
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

			//Добавляю заказ в бд
			_, err = db.Exec(`INSERT INTO orders (id, customer, items, status) VALUES ($1, $2, $3, $4)`,
				order.ID, order.Customer, strings.Join(order.Items, ","), order.Status)
			if err != nil {
				log.Printf("Не удалось добавить заказ в бд: %v", err)
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
	// Ожидание завершения всех операций
	<-shutdownctx.Done()

	log.Println("Kitchen-service остановлен")
}
