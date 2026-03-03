package handler

import (
	"gophermart/internal/service"
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_handler_createOrder(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		services *service.Service
		// Named input parameters for target function.
		c *gin.Context
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.services)
			h.createOrder(tt.c)
		})
	}
}
