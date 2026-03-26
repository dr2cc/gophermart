package handler

import (
	"errors"
	"gophermart/internal/service"
	mock_service "gophermart/internal/service/mocks"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestController_userIdentity(t *testing.T) {
	// Объявляем ♊имитацию поведения
	type mockBehavior func(s *mock_service.MockAuthorization, token string)

	testTable := []struct {
		name                 string
		headerName           string
		headerValue          string
		token                string
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "OK",
			headerName:  "Authorization",
			headerValue: "Bearer token",
			token:       "token",
			// ♊ имитация поведения:
			mockBehavior: func(s *mock_service.MockAuthorization, token string) {
				// 1️⃣ При обращении к объекту s мы будем ОЖИДАТЬ()
				s.EXPECT().
					// 2️⃣ что вызов метода ParseToken(token)
					ParseToken(token). // это метод (структуры MockOrderMockRecorder) "притворяющийся" методом ParseToken(accessToken string) структуры authService, реализующей интерфейс service.Authorization
					// 3️⃣ Вернет() тестируемому коду 1 (ID пользователя) и nil (отсутствие ошибки).
					Return(1, nil)
				// В данном случае, как только программа вызовет метод ParseToken, ♊имитатор мгновенно отдаст ей 1 (ID пользователя) и nil (отсутствие ошибки).
				// Так логика продолжит тестирование, не обращаясь к реальной базе данных.
			},
			expectedStatusCode:   200,
			expectedResponseBody: "1",
		},
		{
			name:                 "Invalid Header Name",
			headerName:           "",
			headerValue:          "Bearer token",
			token:                "token",
			mockBehavior:         func(s *mock_service.MockAuthorization, token string) {},
			expectedStatusCode:   401,
			expectedResponseBody: `{"message":"empty auth header"}`,
		},
		{
			name:                 "Invalid Header Value",
			headerName:           "Authorization",
			headerValue:          "Bearr token",
			token:                "token",
			mockBehavior:         func(s *mock_service.MockAuthorization, token string) {},
			expectedStatusCode:   401,
			expectedResponseBody: `{"message":"invalid auth header"}`,
		},
		{
			name:                 "Empty Token",
			headerName:           "Authorization",
			headerValue:          "Bearer ",
			token:                "token",
			mockBehavior:         func(s *mock_service.MockAuthorization, token string) {},
			expectedStatusCode:   401,
			expectedResponseBody: `{"message":"token is empty"}`,
		},
		{
			name:        "Parse Error",
			headerName:  "Authorization",
			headerValue: "Bearer token",
			token:       "token",
			mockBehavior: func(s *mock_service.MockAuthorization, token string) {
				s.EXPECT().ParseToken(token).Return(0, errors.New("invalid token"))
			},
			expectedStatusCode:   401,
			expectedResponseBody: `{"message":"invalid token"}`,
		},
	}

	for _, test := range testTable {
		// .Run() - по сути повторяем здесь упрощенную (только под задачи теста) "точку сборки"
		t.Run(test.name, func(t *testing.T) {
			// Controller (ctrl) — это верхний управляющий объект для всей системы ♊имитаторов в тесте.
			// Контроллер — это не мок сервиса и не HTTP‑контроллер из MVC, а управляющая тестовая сущность,
			// контролирующая моки (имитаторы) и гарантирует корректность их поведения.
			// Он выступает координатором для мок‑объектов. Его задачи:
			// 1. Задавать область и срок жизни ♊имитаторов (например, .NewMockAuthorization(ctrl)).
			// 2. Хранить и управлять ожиданиями (expectations) (что и когда должны были вызвать методы ♊имитаторов).
			// 3. Определять, когда тест должен завершиться с ошибкой, если что-то пошло не так.
			ctrl := gomock.NewController(t) // Эта строка создаёт контроллер ctrl, который «владеет» всеми ♊имитаторами, созданными в этом тесте.
			// Теперь создаем ♊имитацию сервиса, который "притворяется" реальной бизнес-логикой (интерфейсом Authorization).
			// Здесь мок‑объект auth явно привязывается к контроллеру ctrl: все его ожидания (EXPECT()....) будут записаны и затем проверены именно через ctrl
			auth := mock_service.NewMockAuthorization(ctrl)
			// 📍Передавая параметры (auth, test.inputUser) мы указываем, что
			// ожидаем получить вызов метода сервиса, а в качестве аргумента token
			test.mockBehavior(auth, test.token)
			// Создаем объект сервисов, но передадим аргументом для интерфейса авторизации наш "ложный" auth.
			// service.Service думает, что работает с настоящей базой или API, хотя на самом деле он работает с нашим «контролируемым» моком.
			services := &service.Service{Authorization: auth}
			// Инициализируем хендлер.
			// Структура controller получает объект services, внутри которого уже есть наш мок.
			// Хендлер не знает, как проверять пользователя, он лишь делегирует это сервису.
			// Теперь при вызове handler.signIn внутри сработает цепочка, ведущая к моку.
			handler := controller{services}

			// 2. Init Endpoint
			r := gin.New()
			// 1. Заглушка маршрута: регистрируем временный роут /identity, чтобы проверить,
			// пропустит ли его middleware и корректно ли передаст данные дальше.
			r.GET("/identity",
				// 2. Цепочка вызовов: Сначала срабатывает handler.userIdentity (наше middleware).
				// Если авторизация успешна, управление переходит в анонимную функцию func(c *gin.Context).
				handler.userIdentity, func(c *gin.Context) {
					// const userCtx = "userId"
					// 3. Передача контекста: Внутри middleware делается c.Set(userCtx, userId).
					// Этот тест проверяет, что ID пользователя действительно "доехал" до обработчика через контекст Gin.
					id, _ := c.Get(userCtx)
					// 4. Проверка результата: c.String(200, "%d", id) выводит ID в тело ответа.
					c.String(200, "%d", id)
				})
			// Создадим запрос к тестируемуму методу
			w := httptest.NewRecorder()
			// Создаем «пустой» запрос. Так как это тест авторизации, передача nil в качестве тела для GET-запроса — это норма.
			req := httptest.NewRequest("GET", "/identity", nil)
			// Ключевой момент. Имитируем передачу токена.
			// Поскольку мы не запускаем реальный веб-сервер, мы вручную наполняем структуру http.Request.
			// test.headerName: У нас "Authorization".
			// test.headerValue: У нас строка вида "Bearer <token_value>".
			// Middleware внутри себя будет вызывать c.GetHeader("Authorization").
			// Благодаря этой строчке в тесте, Gin «подумает», что заголовок пришел от реального клиента.
			// 2. Тестирование граничных условий
			// Сила этой строки проявляется в Table-Driven Tests. Мы можем передавать разные комбинации в test.
			// К примеру:
			// ("OK") Валидный кейс: headerName: "Authorization", headerValue: "Bearer token".
			// Пустой заголовок: headerName: "Authorization", headerValue: "".
			// Опечатка в ключе: headerName: "Auth", headerValue: "Bearer ...".
			// Неверный формат: headerName: "Authorization", headerValue: "Basic 12345".
			// Пустое значение: headerName: "Authorization", headerValue: "Bearer " (токен забыли).
			// 3. Почему это важно для Middleware.
			// Наш userIdentity() выполняет три шага, которые зависят от этой строки:
			// 1️⃣ Extract: Достает строку из хедера (то, что вы Set в тесте).
			// 2️⃣ Parse & Validate: Проверяет подпись токена.
			// 3️⃣ Set Context: Если всё ок, кладет userId в контекст Gin
			req.Header.Set(test.headerName, test.headerValue)

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
