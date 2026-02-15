package handler

import (
	"gophermart/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Логика работы слоя http обработчиков (соответственно и каждого обработчика):
// 1️⃣ Принимаем данные от клиента (обычно в формате json).
// 2️⃣ Мапим (преобразуем в конкретную объектную модель, структуру) принятые данные по нашей внутренней структуре.
// 3️⃣ Передаем данные в службу нашего приложения.
// 4️⃣ Возвращаем клиенту response.

// Handler
// - В качестве методов будет иметь
// все эндпойнты и инициализатор роутера.
// - В качестве зависимости Handler имеет
// указатель на структуру сервисов.
type Handler struct {
	services *service.Service
}

// Called from app
func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

// Called from app
func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api")
	{
		user := api.Group("/user")
		{
			user.POST("/register", h.signUp)
			user.POST("/login", h.signIn)

			// 		user.POST("/orders", h.createOrder)
			// 		user.GET("/orders", h.readOrders)

			// 		user.GET("/withdrawals", h.createWithdrawals)

			// 		balance := user.Group("/balance")
			// 		{
			// 			balance.GET("/", h.readBalance)
			// 			balance.POST("/withdraw", h.createWithdraw)
			// 		}
		}
	}

	return router
}
