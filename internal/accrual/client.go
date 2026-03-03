package accrual

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	HTTPClient *resty.Client
}

// Call from app
func NewClient(address string) *Client {
	address = strings.Trim(address, "/")
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "http://" + address
	}

	return &Client{
		HTTPClient: resty.New().
			SetTimeout(time.Duration(5) * time.Second).
			SetBaseURL(address),
	}
}

func (c *Client) GetAccrual(ctx context.Context, orderNum string) (*OrderResponse, time.Duration, error) {
	const op = "accrual.GetAccrual"

	// Мапим респонс на наш dto
	var result OrderResponse

	// Выполнение запроса (метод resty R() создает новый экземпляр запроса) к серверу accrual
	// Как работает: в этот момент Go открывает TCP-соединение, отправляет HTTP-заголовки и ждет ответа.
	// Используется клиент HTTPClient (*resty.Client), созданный в NewClient.
	// У него есть свой тайм-аут (здесь 5 секунд), который подстрахует, если контекст ctx вдруг окажется бесконечным.
	resp, err := c.HTTPClient.R().
		SetContext(ctx).
		SetResult(&result). // Resty сам сделает json.Unmarshal в result, если статус 2xx
		Get("/api/orders/" + orderNum)

	if err != nil {
		return nil, 0, fmt.Errorf("%s: request failed: %w", op, err)
	}

	// Обрабатываем ошибки
	switch resp.StatusCode() {
	case http.StatusOK: // 200
		return &result, 0, nil

	case http.StatusTooManyRequests: // 429
		// У Resty удобный метод Header().Get()
		seconds, _ := strconv.Atoi(resp.Header().Get("Retry-After"))
		if seconds == 0 {
			seconds = 1 // Дефолтная пауза, если заголовок пустой
		}
		return nil, time.Duration(seconds) * time.Second, ErrTooManyRequests

	case http.StatusNoContent: //204
		return nil, 0, ErrOrderNotRegistered

	default:
		return nil, 0, fmt.Errorf("%s: unexpected status code: %d", op, resp.StatusCode())
	}
}
