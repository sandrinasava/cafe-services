package handlers

import (
	"net/http"

	"github.com/sandrinasava/cafe-services/order-service/models"
)

// Хендлер для регистрации
func RegistHandler(authClient *models.AuthClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Неправильный метод запроса", http.StatusMethodNotAllowed)
			return
		}

		var credentials struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Email    string `json:"email"`
		}

		if credentials.Username == "" || credentials.Password == "" || credentials.Email == "" {
			http.Error(w, "Имя пользователя, пароль и email обязательны", http.StatusBadRequest)
			return
		}

		err := authClient.Register(r.Context(), credentials.Username, credentials.Password, credentials.Email)
		if err != nil {
			http.Error(w, "Ошибка при регистрации", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
