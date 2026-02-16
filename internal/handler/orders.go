package handler

import (
	"bytes"
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
		c.String(http.StatusBadRequest, "Error reading request body")
		return
	}
	// "Возвращаем" данные в тело запроса.
	// Это важно, так как GetRawData опустошил c.Request.Body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// Преобразуем в строку
	text := string(body)
	if text == "" {
		c.String(http.StatusUnprocessableEntity, "order number required")
		return
	}

	// 3️⃣ Передаем данные в службу нашего приложения.
	err = h.services.Order.RecordOrder(text)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	// 4️⃣ Возвращаем клиенту response.
	// - `200` — номер заказа уже был загружен этим пользователем;
	// - `202` StatusAccepted — новый номер заказа принят в обработку;
	// - `400` — неверный формат запроса;
	// - `401` — пользователь не аутентифицирован;
	// - `409` — номер заказа уже был загружен другим пользователем;
	// - `422` StatusUnprocessableEntity  — неверный формат номера заказа;
	// - `500` — внутренняя ошибка сервера.
	c.String(http.StatusAccepted, "Received: %s", text)
}
