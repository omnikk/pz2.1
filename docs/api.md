# API Documentation

## Auth Service (порт 8081)

### POST /v1/auth/login
Получение токена.

**Request:**
```json
{"username": "student", "password": "student"}
```

**Response 200:**
```json
{"access_token": "demo-token", "token_type": "Bearer"}
```

**Ошибки:** 400 — неверный формат, 401 — неверные данные

---

### GET /v1/auth/verify
Проверка токена.

**Headers:** `Authorization: Bearer demo-token`

**Response 200:**
```json
{"valid": true, "subject": "student"}
```

**Response 401:**
```json
{"valid": false, "error": "unauthorized"}
```

---

## Tasks Service (порт 8082)

Все запросы требуют заголовок `Authorization: Bearer <token>`

| Метод | Путь | Описание | Коды |
|-------|------|----------|------|
| POST | /v1/tasks | Создать задачу | 201, 400, 401 |
| GET | /v1/tasks | Список задач | 200, 401 |
| GET | /v1/tasks/{id} | Получить задачу | 200, 401, 404 |
| PATCH | /v1/tasks/{id} | Обновить задачу | 200, 401, 404 |
| DELETE | /v1/tasks/{id} | Удалить задачу | 204, 401, 404 |

### Переменные окружения

| Переменная | Сервис | По умолчанию |
|------------|--------|--------------|
| AUTH_PORT | Auth | 8081 |
| TASKS_PORT | Tasks | 8082 |
| AUTH_BASE_URL | Tasks | http://localhost:8081 |