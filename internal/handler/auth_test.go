package handler

import (
	"bytes"
	"gophermart/internal/models"
	"gophermart/internal/service"
	mock_service "gophermart/internal/service/mocks"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"go.uber.org/mock/gomock"
)

func TestHandler_signUp(t *testing.T) {
	// mockBehavior - поведение мока
	// принимает объект (структуру) мока сервиса и модель (структуру) пользователя
	type mockBehavior func(s *mock_service.MockAuthorization, user models.User)

	testTable := []struct {
		name                 string
		inputBody            string
		inputUser            models.User
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "OK",
			inputBody: `{"login": "username", "password": "qwerty"}`,
			inputUser: models.User{
				Login:    "username",
				Password: "qwerty",
			},
			mockBehavior: func(s *mock_service.MockAuthorization, user models.User) {
				// У объекта мока s вызываем метод EXPECT (s.EXPECT()) = создаем поведение для s
				// Через . указываем, что ожидаем получить вызов CreateUser с параметром (user models.User)
				// Через следующую . пишем ожидание того, что вызов CreateUser вернет ID 1 и nil (ошибка)
				s.EXPECT().CreateUser(user).Return(1, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"id":1}`,
		},
	}

	for _, test := range testTable {
		// По сути повторяем здесь упрощенную только под задачи теста "точку сборки" Run.
		t.Run(test.name, func(t *testing.T) {

			// 1. Сборка зависимостей: ctrl--> services--> handler
			// Инициализация моков.
			// Создаем "ложный" контроллер (требование библиотеки- создавать контроллер
			// и "финишировать" его по выполнению теста)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// Создаем "ложный" сервис
			auth := mock_service.NewMockAuthorization(ctrl)
			// Передавая параметры (auth, test.inputUser) мы указываем, что
			// ожидаем получить вызов метода сервиса, а в качестве аргумента inputUser models.User
			test.mockBehavior(auth, test.inputUser)
			//Создаем объект сервисов, но передадим аргументом для интерфейса авторизации наш "ложный" auth.
			services := &service.Service{Authorization: auth}
			// Инициализируем хендлер
			handler := controller{services}

			// 2. Init Endpoint
			r := gin.New()
			r.POST("/register", handler.signUp)

			// Create Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/register",
				//bytes.NewBufferString(test.inputBody) создает тело нашего запроса
				//и реализует интерфейс io.Reader
				bytes.NewBufferString(test.inputBody))

			// Make Request
			r.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, w.Code, test.expectedStatusCode)
			assert.Equal(t, w.Body.String(), test.expectedResponseBody)
		})
	}
}
