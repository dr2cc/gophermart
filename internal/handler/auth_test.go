package handler

import (
	"bytes"
	"errors"
	"gophermart/internal/models"
	"gophermart/internal/service"
	mock_service "gophermart/internal/service/mocks"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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
			// ожидания от мока
			mockBehavior: func(s *mock_service.MockAuthorization, user models.User) {
				// Тело функции.
				// У объекта мока s вызываем метод EXPECT (s.EXPECT()) = создаем поведение для s
				// Через . указываем, что ожидаем получить вызов CreateUser с параметром (user models.User)
				// Через следующую . пишем ожидание того, что вызов CreateUser вернет ID 1 и nil (ошибка)
				s.EXPECT().CreateUser(user).Return(1, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"id":1}`,
		},
		{
			name:      "Empty fields",
			inputBody: `{"password": "qwerty"}`,
			mockBehavior: func(s *mock_service.MockAuthorization, user models.User) {
				// тело функции не указываем, т.к вызова метода сервиса не будет- нет данных
			},
			expectedStatusCode:   400,
			expectedResponseBody: `{"message":"invalid input body"}`,
		},
		{
			name:      "Service failure",
			inputBody: `{"login": "username", "password": "qwerty"}`,
			inputUser: models.User{
				Login:    "username",
				Password: "qwerty",
			},
			// ожидания от мока
			mockBehavior: func(s *mock_service.MockAuthorization, user models.User) {
				// Тело функции.
				// ожидание того, что вызов CreateUser вернет ID 1 и errors.New("service failure")
				s.EXPECT().CreateUser(user).Return(1, errors.New("service failure"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: `{"message":"service failure"}`,
		},
	}

	for _, test := range testTable {
		// По сути повторяем здесь упрощенную (только под задачи теста) "точку сборки" app.Run()
		t.Run(test.name, func(t *testing.T) {
			// 1. Сборка зависимостей: ctrl--> services--> handler
			// Инициализация моков.
			// Создаем "ложный" контроллер (требование библиотеки- "финишировать" контроллер в 2026 уже не обязательно)
			ctrl := gomock.NewController(t)
			// defer ctrl.Finish()
			// Создаем "ложный" сервис, который притворяется реальной бизнес-логикой (интерфейсом Authorization).
			auth := mock_service.NewMockAuthorization(ctrl)
			// Передавая параметры (auth, test.inputUser) мы указываем, что
			// ожидаем получить вызов метода сервиса, а в качестве аргумента inputUser (models.User)
			test.mockBehavior(auth, test.inputUser)
			// Создаем объект сервисов, но передадим аргументом для интерфейса авторизации наш "ложный" auth.
			// service.Service думает, что работает с настоящей базой или API, хотя на самом деле он работает с нашим «контролируемым» моком.
			services := &service.Service{Authorization: auth}
			// Инициализируем хендлер.
			// Структура controller получает объект services, внутри которого уже есть наш мок.
			// Хендлер не знает, как регистрировать пользователя, он лишь делегирует это сервису.
			// Теперь при вызове handler.signUp внутри сработает цепочка, ведущая к моку.
			handler := controller{services}

			// 2. Init Endpoint
			r := gin.New()
			// Регистрируем конкретную функцию контроллера (тестируемый метод хендлера signUp) на маршрут /register
			// Gin теперь знает: «Если придет POST-запрос на этот адрес, нужно запустить именно этот код».
			r.POST("/register", handler.signUp)

			// Создадим запрос к тестируемуму методу
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/register",
				//bytes.NewBufferString(test.inputBody) создает тело нашего запроса
				//и реализует интерфейс io.Reader
				bytes.NewBufferString(test.inputBody))

			// Make Request
			// Вызываем метод ServeHTTP у объекта роутера
			// Отдаем роутеру «виртуальный» запрос (req) и «записывающее устройство» (w).
			// Роутер прогоняет запрос через свои механизмы, вызывает хендлер, тот вызывает сервис (мок!),
			// получает ответ и записывает результат в w.
			r.ServeHTTP(w, req)

			// Проверка (Assert, точнее "проверка утверждения").
			// 1. Проверяем Status Code
			// Мы ожидаем определенный код если тест проходит по нашему плану.
			// test.expectedStatusCode уже заранее определен в структуре кейса.
			assert.Equal(t, test.expectedStatusCode, w.Code)
			// 2. Проверяем тело ответа.
			// w.Body.String() возвращает ответ хендлера.
			// Мы сравниваем это значение с ожидаемой строкой из нашего тест-кейса.
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}
