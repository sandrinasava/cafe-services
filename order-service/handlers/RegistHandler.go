package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/sandrinasava/cafe-services/order-service/models"
)

// RegistHandler godoc
// @Summary Регистрация нового пользователя
// @Description Обработчик для регистрации нового пользователя
// @ID regist-handler
// @Accept json
// @Produce json
// @Param credentials body models.Credentials true "Учетные данные пользователя"
// @Success 303 "Перенаправление на страницу логина"
// @Failure 400 {object} map[string]interface{} "Неправильное тело запроса"
// @Failure 405 {object} map[string]interface{} "Неправильный метод запроса"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /register [post]
func RegistHandler(authClient *models.AuthClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Неправильный метод запроса", http.StatusMethodNotAllowed)
			return
		}
		credentials := models.Credentials{}

		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			http.Error(w, "Неправильное тело запроса", http.StatusBadRequest)
			return
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
