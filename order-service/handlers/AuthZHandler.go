package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/sandrinasava/cafe-services/order-service/models"
)

// Хендлер для получения статуса заказа
func AuthZHandler(rdb *redis.Client, db *sql.DB, authClient *models.AuthClient, kafkaBroker string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Неправильный метод запроса", http.StatusMethodNotAllowed)
			return
		}

		var credentials struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			http.Error(w, "Неправильное тело запроса", http.StatusBadRequest)
			return
		}

		if credentials.Username == "" || credentials.Password == "" {
			http.Error(w, "Имя пользователя и пароль обязательны", http.StatusBadRequest)
			return
		}

		token, err := authClient.Login(r.Context(), credentials.Username, credentials.Password)
		if err != nil {
			http.Error(w, "Неверное имя пользователя или пароль", http.StatusUnauthorized)
			return
		}

		// Устанавливаем токен в куках
		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})

		http.Redirect(w, r, "/order", http.StatusSeeOther)
	}
}
