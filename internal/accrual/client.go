package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"gophermart/internal/accrual/dto"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	Address    string
	HTTPClient *http.Client
}

// Call from app
func NewClient(address string) *Client {
	return &Client{
		Address: address,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) GetAccrual(ctx context.Context, orderNum string) (*dto.OrderResponse, time.Duration, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.Address, orderNum)

	// 1️⃣ Принимаем данные от клиента
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	// 2️⃣ Мапим (?)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	// 3️⃣ Возвращаем результат
	// 429
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		return nil, time.Duration(retryAfter) * time.Second, dto.ErrTooManyRequests
	}

	// 204
	if resp.StatusCode == http.StatusNoContent {
		return nil, 0, dto.ErrOrderNotRegistered
	}

	// 200
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result dto.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	// 4️⃣ Возвращаем клиенту response.
	return &result, 0, nil
}
