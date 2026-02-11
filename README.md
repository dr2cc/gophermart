# Накопительная система лояльности «Гофермарт» 

Дипломный проект первой половины курса «Продвинутый Go-разработчик».

### Описание API:
- [Specification](SPECIFICATION.md)

## Работа с докером (11.02.26 не реализована)

### Запуск/выключение с docker compose
```bash
   cd /path/to/project/dir
   cp deploy/local/compose.yaml ./
   docker compose up -d   #запуск
   docker compose stop    #выключение
   docker compose down -v #выключение с очисткой БД
```
### Запуск контейнера приложения
```bash
   docker compose run --rm loyalty bash
```

Если хотите сохранять историю команд и избежать проблем с правами в генерируемых файлах,
раскоментируйте строки в сервисе loyalty в **/path/to/project/dir/compose.yaml**. 
Нормально работает только в Линуксе. **Избегайте подобного в продакшине!**

#### Примечание
Запуск процесса от определенного пользователя в docker compose
```yaml
services:
  loyalty:
   #...
   user: 1000:1000 #${CURRENT_UID}:${CURRENT_GID}
```
В большинсве случаев 1000:1000 работает нормально, узнать их можно с помощью команд:
```bash
   id -u
   id -g
```

### Работа с миграциями
```bash
   docker compose run --rm loyalty bash
   migrate --help
```
