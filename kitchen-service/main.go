package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	topicIn  = "new_orders"
	topicOut = "ready_orders"
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
	//создаю консьюмера
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(kafkaBroker, ","),
		GroupID: "kitchen-group",
		Topic:   topicIn,
	})
	//создаю продюсера
	w := kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    topicOut,
		Balancer: &kafka.LeastBytes{},
	}

	//создаю канал, читающий сигналы ос
	stop := make(chan os.Signal, 1)
	//подписка на оповещение от ос о сигналах SIGINT(нажатие ctrl+c) и SIGTERM(завершение процесса)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	//создаю контекст с отменой по сигналу
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-stop
		cancel()
	}()

	go func() {
		for {
			// читаю из брокера если ctx не отменен
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					break
				}
				log.Printf("неудачное чтение из брокера: %v", err)
				time.Sleep(1 * time.Second) //задержка при повторной попытке чтения
				continue
			}
			// достаю данные из сообщения и десериализую
			var order Order
			if err := json.Unmarshal(m.Value, &order); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			// Имитация приготовления заказа
			time.Sleep(2 * time.Second)

			order.Status = "ready"

			//сериализация
			message, err := json.Marshal(order)
			if err != nil {
				log.Printf("неудачная сериализация: %v", err)
				continue
			}
			//отправка сообщения в брокер
			err = w.WriteMessages(context.Background(), kafka.Message{
				Key:   []byte(order.ID),
				Value: message,
			})
			if err != nil {
				log.Printf("ошибка отправки сообщения в брокер: %v", err)
				continue
			}

			log.Printf("Заказ %s готов", order.ID)
		}
	}()

	//далее закрываю консьюмер, если сервер остановился, чтобы избежать утечки данных
	//ожидание сигнала
	<-stop
	log.Println("Остановка Kitchen-service")
	//закрытие консьюмера
	if err := r.Close(); err != nil {
		log.Printf("Не удалось закрыть консьюмер: %v", err)
	}

	log.Println("Kitchen-service остановлен")
}
