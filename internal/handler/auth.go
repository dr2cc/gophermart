package handler

import (
	"errors"
	"gophermart/internal/models"
	"gophermart/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Логика работы слоя http обработчиков (соответственно и каждого обработчика):
// 1️⃣ Принимаем данные от клиента (обычно в формате json).
// 2️⃣ Мапим (преобразуем в конкретную объектную модель, структуру) принятые данные по нашей внутренней структуре.
// 3️⃣ Передаем данные в службу нашего приложения.
// 4️⃣ Возвращаем клиенту response.

// @Summary SignUp
// @Tags auth
// @Description create account
// @ID create-account
// @Accept  json
// @Produce  json
// @Param input body todo.User true "account info"
// @Success 200 {integer} integer 1
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/sign-up [post]
func (h *handler) signUp(c *gin.Context) {
	var input models.User

	// 1️⃣ Принимаем данные из сети, 2️⃣ десериализуем и заполняем (, &input) models
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	// 3️⃣ // Пытаемся создать пользователя, отдаем структуру input сервису
	id, err := h.services.Authorization.CreateUser(input)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			newErrorResponse(c, http.StatusConflict, "The login is already taken") //409
			return
		}
		newErrorResponse(c, http.StatusInternalServerError, err.Error()) // 500
		return
	}

	// TODO: обработать ошибку `400` — неверный формат запроса;

	// 4️⃣ Возвращаем клиенту response
	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

type signInInput struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// @Summary SignIn
// @Tags auth
// @Description login
// @ID login
// @Accept  json
// @Produce  json
// @Param input body signInInput true "credentials"
// @Success 200 {string} string "token"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /auth/sign-in [post]
func (h *handler) signIn(c *gin.Context) {
	var input signInInput

	// 1️⃣ Принимаем данные из сети, 2️⃣ десериализуем и заполняем (, &input) models
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 3️⃣ // Пытаемся создать token, отдаем структуру input сервису
	token, err := h.services.Authorization.GenerateToken(input.Login, input.Password)
	if err != nil {
		// неверная пара логин/пароль
		if errors.Is(err, repository.ErrInvalidCredentials) {
			newErrorResponse(c, http.StatusUnauthorized, "invalid login/password pair") //409
			return
		}
		// остальные ошибки
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 4️⃣ Возвращаем клиенту response
	c.JSON(http.StatusOK, map[string]interface{}{
		"token": token,
	})
}
