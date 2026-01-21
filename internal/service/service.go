package service

import (
	"gophermart/internal/repository"
	"todo-app1"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

// Сервис аутентификации
type Authorization interface {
	// Функцонал:
	// Регистрация пользователей
	CreateUser(user todo.User) (int, error)
	// Генерация jwt токенов
	GenerateToken(username, password string) (string, error)
	// Валидация jwt токенов
	ParseToken(token string) (int, error)
}

// Сервис работы со списками задач
type TodoList interface {
	// Функцонал:
	Create(userId int, list todo.TodoList) (int, error)
	GetAll(userId int) ([]todo.TodoList, error)
	GetById(userId, listId int) (todo.TodoList, error)
	Delete(userId, listId int) error
	Update(userId, listId int, input todo.UpdateListInput) error
}

// Сервис работы с задачами
type TodoItem interface {
	// Функцонал:
	Create(userId, listId int, item todo.TodoItem) (int, error)
	GetAll(userId, listId int) ([]todo.TodoItem, error)
	GetById(userId, itemId int) (todo.TodoItem, error)
	Delete(userId, itemId int) error
	Update(userId, itemId int, input todo.UpdateItemInput) error
}

// Здесь определены предметные области (доменные зоны).
// ❗Предметная область это круг задач (сферы реального мира) решаемых приложением.
// Получается тут три предметные области: аутентификация, работа со списками, работа с задачами.
type Service struct {
	// Сервис аутентификации, со своим функционалом.
	Authorization
	// Сервис работы со списками задач, со своим функционалом.
	TodoList
	// Сервис работы с задачами, со своим функционалом.
	TodoItem
}

// Вызываается из main
func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
		TodoList:      NewTodoListService(repos.TodoList),
		TodoItem:      NewTodoItemService(repos.TodoItem, repos.TodoList),
	}
}
