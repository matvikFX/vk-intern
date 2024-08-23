# Тестовое задание на стажировку VK

Разработка API для KV-хранилища на базе Tarantool

## Запуск

### В Docker-контейнере
```bash
docker-compose up --build
```

### Локально
```bash
docker-compose up tarantool
# Обязательно прописать путь к конфигу, или добавить переменную окружения CONFIG_PATH
go run cmd/main.go --config=config/local.yaml
```

## Описание API

### Коды ответов
- 200 OK - Запрос успешно обработан
- 201 Created - Запрос успешно обработан и данные записаны
- 400 Bad Request - Неверный запрос
- 401 Unauthorized - Пользователь не авторизован
- 404 Not Found - Неверно указан путь
- 405 Method Not Allowed - Неправильный метод
- 500 Internal Server Error - Ошибка на стороне сервера

### Примеры правильных запросов

`/api/login` \
Запрос:
```bash
curl --location 'http://localhost:8080/api/login' \
--header 'Content-Type: application/json' \
--data '{
    "username": "admin",
    "password": "presale"
}'
```
Ответ:
```json
{"token": "user_token"}
```

`/api/write` \
Запрос:
```bash
curl --location 'http://localhost:8080/api/write' \
--header 'Authorization: Bearer user_token' \
--header 'Content-Type: application/json' \
--data '{
    "data": {
        "key1": "value1",
        "key2": 2
    }
}'
```
Ответ:
```json
{"status": "success"}
```

`/api/read` \
Запрос:
```bash
curl --location 'http://localhost:8080/api/read' \
--header 'Authorization: Bearer user_token' \
--header 'Content-Type: application/json' \
--data '{
	"keys": ["key1", "key2"]
}'
```
Ответ:
```json
{
	"data": {
		"key1": "value1", 
		"key2": 2
	}
}
```

## Дополнительные сведения

При выполнении запроса `api/read` пользователь может указать несущесвтующие ключи, в таком случае сервер вернет эти ключи со значем null. Подумал, что данное решение будет намного лучше, чем выдавать пользователю ошибку, так как не все запрошенные ключи могут быть пустыми.
