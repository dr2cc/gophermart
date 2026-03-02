package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Address    string
	HTTPClient *http.Client
}

// Call from app
func NewClient(address string) *Client {
	// ❌ Если адрес не начинается с http, добавляем его сами
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "http://" + address
	}
	//
	return &Client{
		Address: address,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) GetAccrual(ctx context.Context, orderNum string) (*OrderResponse, time.Duration, error) {
	const op = "accrual.GetAccrual"

	// Без защиты "от дурака" HTTPClient (*http.Client) в Go не понимает,
	// как обращаться к адресу, если в начале не указан протокол.
	// Желательный формат ACCRUAL_SYSTEM_ADDRESS
	// export ACCRUAL_SYSTEM_ADDRESS=http://localhost:8090
	url := fmt.Sprintf("%s/api/orders/%s", c.Address, orderNum)

	// Создаем http-запрос к серверу accrual
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	// Выполнение запроса (Do)
	// Как работает: в этот момент Go открывает TCP-соединение, отправляет HTTP-заголовки и ждет ответа.
	// Используется клиент HTTPClient (*http.Client), созданный в NewClient.
	// У него есть свой тайм-аут (здесь 5 секунд), который подстрахует, если контекст ctx вдруг окажется бесконечным.
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("called %s from %s: %w", "HTTPClient", op, err)
	}
	defer resp.Body.Close()

	// Обрабатываем ошибки
	// 429
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		return nil, time.Duration(retryAfter) * time.Second, ErrTooManyRequests
	}

	// 204
	if resp.StatusCode == http.StatusNoContent {
		return nil, 0, ErrOrderNotRegistered
	}

	// 200
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Мапим респонс на наш dto
	var result OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("called %s from %s: %w", "NewDecoder", op, err)
	}

	return &result, 0, nil
}
