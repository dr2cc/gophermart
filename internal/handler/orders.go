package handler

import (
	"bytes"
	"errors"
	"fmt"
	"gophermart/internal/service"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Логика работы слоя http обработчиков (соответственно и каждого обработчика):
// 1️⃣ Принимаем данные от клиента (обычно в формате json).
// 2️⃣ Мапим (преобразуем в конкретную объектную модель, структуру) принятые данные по нашей внутренней структуре.
// 3️⃣ Передаем данные в службу нашего приложения.
// 4️⃣ Возвращаем клиенту response.

func (h *Handler) createOrder(c *gin.Context) {
	// Content-Type: text/plain
	// ```
	// 0
	// ```

	// 1️⃣ Принимаем данные из сети
	// Получаем "сырые" данные
	body, err := c.GetRawData()
	if err != nil {
		// // Это стандартный вызов ошибки в gin
		//c.String(http.StatusBadRequest, "Error reading request body")
		// Но у нас есть кастомный
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	// "Возвращаем" данные в тело запроса.
	// Это важно, так как GetRawData опустошил c.Request.Body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// Преобразуем в строку
	text := string(body)
	if text == "" {
		newErrorResponse(c, http.StatusUnprocessableEntity, "order number required")
		return
	}

	// 3️⃣ Передаем данные в службу нашего приложения.
	err = h.services.Order.RecordOrder(text)
	// 4️⃣ Возвращаем клиенту response.
	// - `200` — номер заказа уже был загружен этим пользователем;
	// - `202` StatusAccepted — новый номер заказа принят в обработку;
	// - `400` — неверный формат запроса;
	// - `401` — пользователь не аутентифицирован;
	// - `409` — номер заказа уже был загружен другим пользователем;
	// - `422` StatusUnprocessableEntity  — неверный формат номера заказа;
	// - `500` — внутренняя ошибка сервера.

	if err != nil {

		// Мапим ошибку на HTTP код
		switch {
		case errors.Is(err, service.ErrOrderAlreadyExists):
			//c.String(http.StatusConflict, "Order already registered")
			newErrorResponse(c, http.StatusConflict, "Order already registered")
		case errors.Is(err, service.ErrInvalidOrderFormat):
			errMessage := fmt.Sprintf("Invalid format: %s", err.Error())
			newErrorResponse(c, http.StatusUnprocessableEntity, errMessage)
		default:
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	c.String(http.StatusAccepted, "")
}
