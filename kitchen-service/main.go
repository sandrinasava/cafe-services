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
	topicIn := cfg.Kafka.TopicIn
	topicOut := cfg.Kafka.TopicOut

	//создаю консьюмера
	kReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(kafkaBroker, ","),
		GroupID: "kitchen-group",
		Topic:   topicIn,
	})
	//создаю продюсера
	kWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaBroker},
		Topic:    topicOut,
		Balancer: &kafka.LeastBytes{},
	})
	defer kWriter.Close()
	//создаю канал, читающий сигналы ос
	stop := make(chan os.Signal, 1)
	//подписка на оповещение от ос о сигналах SIGINT(нажатие ctrl+c) и SIGTERM(завершение процесса)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	//создаю контекст для корректного завершения цикла при сигналах остановки
	readctx, readCancel := context.WithCancel(context.Background())
	go func() {
		<-stop
		readCancel()
	}()

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
			err = kWriter.WriteMessages(context.Background(), kafka.Message{
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

	// закрытие консьюмера
	if err := kWriter.Close(); err != nil {
		log.Printf("Не удалось закрыть продюсера Kafka: %v", err)
	}
	// Ожидание завершения всех операций
	<-shutdownctx.Done()

	log.Println("Kitchen-service остановлен")
}
