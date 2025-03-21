package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-redis/redis/v8"

	"github.com/sandrinasava/cafe-services/order-service/models"
)

// Хендлер для получения статуса заказа
func StatusHandler(rdb *redis.Client, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Query().Get("id")
		if orderID == "" {
			http.Error(w, "ID заказа не предоставлен", http.StatusBadRequest)
			return
		}
		var order models.Order
		// Поиск заказа в кэше
		cachedOrder, err := rdb.Get(r.Context(), orderID).Result()
		if err == nil {
			err = json.Unmarshal([]byte(cachedOrder), &order)
			if err != nil {
				http.Error(w, "Ошибка при обработке заказа", http.StatusInternalServerError)
				return
			}
		} else if err == redis.Nil {
			// Поиск заказа в базе данных
			err = db.QueryRowContext(r.Context(), "SELECT id, customer, items, status FROM orders WHERE id=$1", orderID).Scan(
				&order.ID, &order.Customer, &order.Items, &order.Status)
			if err != nil {
				http.Error(w, "Заказ не найден", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "Ошибка при обработке заказа", http.StatusInternalServerError)
			return
		}

		err = json.Unmarshal([]byte(cachedOrder), &order)
		if err != nil {
			http.Error(w, "Ошибка при обработке заказа", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	}
}
