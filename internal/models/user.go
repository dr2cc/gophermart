package models

type User struct {
	ID int `json:"-" db:"id"`
	// Name     string `json:"name" binding:"required"`
	// Username string `json:"username" binding:"required"`
	Login    string  `json:"login" binding:"required"`
	Password string  `json:"password" binding:"required"` // Хэш или сырой из запроса
	Balance  float64 `json:"-"`
}
