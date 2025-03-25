package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/sandrinasava/cafe-services/order-service/models"
)

// OrderHandler godoc
// @Summary Создание нового заказа
// @Description Обработчик для создания нового заказа
// @ID order-handler
// @Accept json
// @Produce json
// @Param customer body string true "Customer Name"
// @Param items body string true "Items"
// @Success 201 {string} string "Заказ успешно создан"
// @Failure 400 {object} map[string]interface{} "Неправильное тело запроса"
// @Failure 401 {object} map[string]interface{} "Недействительный токен"
// @Failure 405 {object} map[string]interface{} "Метод не доступен"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /order [post]
func OrderHandler(rdb *redis.Client, db *sql.DB, authClient *models.AuthClient, kWriter *kafka.Writer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не доступен", http.StatusMethodNotAllowed)
			return
		}

		// достаю данные из сообщения и десериализую
		var order models.Order
		if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// достаю токен из заголовков
		token := r.Header.Get("Authorization")
		if token == "" {
			// Перенаправление на страницу авторизации
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// валидация токена
		valid, err := authClient.ValidateToken(r.Context(), token)
		if err != nil {
			log.Printf("Ошибка при валидации токена: %v", err)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if !valid {
			log.Println("Недействительный токен")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		order.ID = uuid.New()
		order.Status = "received"
		// сериализация
		message, err := json.Marshal(order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// отправляю заказ в Kafka
		err = kWriter.WriteMessages(r.Context(), kafka.Message{
			Key:   []byte(order.ID.String()),
			Value: message,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// сохраняю заказ в кэше Redis на 1 час
		err = rdb.Set(r.Context(), order.ID.String(), string(message), 1*time.Hour).Err()
		if err != nil {
			log.Printf("Ошибка кеширования сообщения: %v", err)
		}

		// Сохранение заказа в PostgreSQL
		_, err = db.ExecContext(r.Context(), "INSERT INTO orders (id, customer, items, status) VALUES ($1, $2, $3, $4)",
			order.ID, order.Customer, order.Items, order.Status)
		if err != nil {
			http.Error(w, "Ошибка при сохранении заказа в базу данных", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Сообщение отправлено: %s\n", order.ID)
	}
}
