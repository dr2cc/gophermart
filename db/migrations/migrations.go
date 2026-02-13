package migrations

import "embed"

// Эмбедим («захватываем») все .sql файлы из текущей директории
//
//go:embed *.sql
var FS embed.FS
