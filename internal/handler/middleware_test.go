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
		t.Run(test.name, func(t *testing.T) {
			// Controller (ctrl) — это верхний управляющий объект для всей системы ♊имитаторов в тесте.
			// Контроллер — это не мок сервиса и не HTTP‑контроллер из MVC, а управляющая тестовая сущность,
			// контролирующая моки (имитаторы) и гарантирует корректность их поведения.
			// Он выступает координатором для мок‑объектов. Его задачи:
			// 1. Задавать область и срок жизни ♊имитаторов (например, .NewMockAuthorization(ctrl)).
			// 2. Хранить и управлять ожиданиями (expectations) (что и когда должны были вызвать методы ♊имитаторов).
			// 3. Определять, когда тест должен завершиться с ошибкой, если что-то пошло не так.
			ctrl := gomock.NewController(t) // Эта строка создаёт контроллер ctrl, который «владеет» всеми ♊имитаторами, созданными в этом тесте.
			// Создаёт мок‑объект (сгенерированный NewMockAuthorization) для интерфейса Authorization
			// Мок‑объект привязывается к ctrl и будет использоваться вместо реального сервиса.
			auth := mock_service.NewMockAuthorization(ctrl)
			// Вызываем функцию‑поле (структуры testTable) mockBehavior в текущем тест‑кейсе,
			// чтобы настроить поведение мока auth (например, какие методы вызываются и что возвращают)
			// для входных данных test.token
			test.mockBehavior(auth, test.token)
			// Создаёт экземпляр слоя сервисов service.Service, в котором вместо реального Authorization подставляется мок‑объект auth
			services := &service.Service{Authorization: auth}
			// Создаём экземпляр контроллера/хендлера controller, в который передаётся структура services;
			// теперь хендлер использует мок‑сервис при своей работе (авторизации).
			handler := controller{services}

			// Инициализируем новый роутер Gin r (для теста, без реального запуска HTTP‑сервера)
			r := gin.New()
			// Заглушка маршрута: регистрируем GET‑роут /identity, где:
			// * сначала выполняется middleware‑метод handler.userIdentity (проверка токена, извлечение ID пользователя в контексте);
			// * затем анонимный хендлер (func(c *gin.Context)) получает id из контекста по ключу userCtx
			// и возвращает его как строку (тело) ответа с кодом 200
			r.GET("/identity", handler.userIdentity, func(c *gin.Context) {
				id, _ := c.Get(userCtx)
				c.String(200, "%d", id)
			})
			// Создаём объект httptest.ResponseRecorder,
			// который ловит HTTP‑ответ (код, заголовки, тело) имитируя HTTP‑сервер.
			w := httptest.NewRecorder()
			// Создаём тестовый HTTP‑запрос (пустой в данном случае): метод GET, путь /identity, без тела.
			// Так как это тест авторизации, передача nil в качестве тела для GET-запроса — это норма.
			req := httptest.NewRequest("GET", "/identity", nil)
			// Имитируем передачу токена.
			// Устанавливаем в запрос заголовок с именем test.headerName ("Authorization")
			// и значением test.headerValue (строка вида "Bearer <token_value>"). Получается (Authorization: Bearer ...)
			//
			// 📌"Тестирование граничных условий."
			// Сила этой строки проявляется в Table-Driven Tests. Мы можем передавать разные комбинации в test.
			// К примеру:
			// ("OK") Валидный кейс: headerName: "Authorization", headerValue: "Bearer token".
			// Пустой заголовок: headerName: "Authorization", headerValue: "".
			// Опечатка в ключе: headerName: "Auth", headerValue: "Bearer ...".
			// Неверный формат: headerName: "Authorization", headerValue: "Basic 12345".
			// Пустое значение: headerName: "Authorization", headerValue: "Bearer " (токен забыли).
			// // Почему это важно для Middleware?
			// Middleware‑метод handler.userIdentity выполняет три шага, которые зависят от этой строки:
			// 1️⃣ Extract: Достает строку из хедера (то, что здесь Set).
			// 2️⃣ Parse & Validate: Проверяет подпись токена.
			// 3️⃣ Set Context: Если всё ок, кладет userId в контекст Gin
			req.Header.Set(test.headerName, test.headerValue)

			// Make Request.
			// Передаёт запрос req в роутер Gin r (вызывает метод ServeHTTP у объекта роутера),
			// который обрабатывает его и записывает результат в w (имитирует работу HTTP‑сервера в памяти).
			r.ServeHTTP(w, req)

			// Проверка (Assert, точнее "проверка утверждения").
			// Проверяет, что HTTP‑статус‑код ответа (w.Code) совпадает с ожидаемым test.expectedStatusCode (200, 400, ...).
			assert.Equal(t, test.expectedStatusCode, w.Code)
			// Проверяет, что тело ответа (w.Body.String()) совпадает с ожидаемым test.expectedResponseBody (часто строковое представление JSON‑ответа).
			assert.Equal(t, test.expectedResponseBody, w.Body.String())
		})
	}
}
