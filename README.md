# Практика 1

## Выполнил: Студент ЭФМО-02-25 Выборнов Олег Андреевич

---

## Структура проекта

```
tech-ip-sem2/
├── services/
│   ├── auth/
│   │   ├── cmd/auth/
│   │   │   └── main.go
│   │   └── internal/
│   │       ├── http/
│   │       │   └── handler.go
│   │       └── service/
│   │           └── auth.go
│   └── tasks/
│       ├── cmd/tasks/
│       │   └── main.go
│       └── internal/
│           ├── http/
│           │   └── handler.go
│           ├── service/
│           │   └── task.go
│           └── client/authclient/
│               └── http.go
├── shared/
│   ├── middleware/
│   │   ├── requestid.go
│   │   └── logging.go
│   └── httpx/
│       └── client.go
├── docs/
│   └── api.md
├── go.mod
└── README.md
```

---

## 1. Описание границ сервисов

**Auth Service** — отвечает за аутентификацию и выдачу токенов. В учебной реализации хранит фиксированную пару логин/пароль (`student`/`student`) и возвращает предопределённый токен `demo-token`. Предоставляет эндпоинт для проверки токена, который возвращает статус валидности и имя субъекта. Не знает ничего о задачах и бизнес-логике Tasks.

**Tasks Service** — управляет задачами (CRUD) в оперативной памяти. Не хранит информацию о пользователях. Перед выполнением каждой операции обращается к Auth Service для проверки токена с таймаутом 3 секунды. Полностью делегирует авторизацию внешнему сервису.

Границы чётко разделены: Auth занимается только вопросами безопасности, Tasks — только бизнес-логикой работы с задачами.

---

## 2. Схема взаимодействия
```mermaid
sequenceDiagram
    participant C as Client
    participant T as Tasks Service
    participant A as Auth Service

    C->>T: Request with Authorization: Bearer token
    T->>A: GET /v1/auth/verify (таймаут 3с)
    A-->>T: 200 OK (valid: true) / 401 Unauthorized
    T-->>C: 200/201/204 или 401/503
```



---

## 3. Список эндпоинтов

### Auth Service (порт 8081)

| Метод | Путь | Описание | Коды ответов |
|-------|------|----------|--------------|
| POST | `/v1/auth/login` | Получение токена | 200, 400, 401 |
| GET | `/v1/auth/verify` | Проверка валидности токена | 200, 401 |

**POST /v1/auth/login** — запрос:
```json
{
  "username": "student",
  "password": "student"
}
```
Ответ 200:
```json
{
  "access_token": "demo-token",
  "token_type": "Bearer"
}
```

**GET /v1/auth/verify** — заголовки: `Authorization: Bearer demo-token`

Ответ 200:
```json
{
  "valid": true,
  "subject": "student"
}
```

---

### Tasks Service (порт 8082)

Все запросы требуют заголовок `Authorization: Bearer <token>`

| Метод | Путь | Описание | Коды ответов |
|-------|------|----------|--------------|
| POST | `/v1/tasks` | Создание задачи | 201, 400, 401 |
| GET | `/v1/tasks` | Список всех задач | 200, 401 |
| GET | `/v1/tasks/{id}` | Получение задачи по ID | 200, 401, 404 |
| PATCH | `/v1/tasks/{id}` | Обновление задачи | 200, 400, 401, 404 |
| DELETE | `/v1/tasks/{id}` | Удаление задачи | 204, 401, 404 |

---

## 4. Скриншоты Postman с подтверждением работы

### 4.1 POST /v1/auth/login — получение токена (200 OK)
![login](image/p1.png)

### 4.2 GET /v1/auth/verify — проверка токена (200 OK)
![verify](image/p2.png)

### 4.3 POST /v1/tasks — создание задачи (201 Created)
![create task](image/p3.png)

### 4.4 GET /v1/tasks — список задач (200 OK)
![list tasks](image/p4.png)

### 4.5 GET /v1/tasks без токена — отказ в доступе (401 Unauthorized)
![unauthorized](image/p5.png)

---

## 5. Логи с подтверждением прокидывания X-Request-ID

### Логи Auth Service
![auth logs](image/log1.png)

### Логи Tasks Service
![tasks logs](image/log2.png)

---

## 6. Инструкция запуска

### Требования
- Go 1.22+
- Git

### Установка
```bash
git clone https://github.com/omnikk/pz2.1.git
cd pz2.1
go mod download
```

### Запуск Auth Service (Терминал 1)
```powershell
$env:AUTH_PORT="8081"
go run ./services/auth/cmd/auth
```

### Запуск Tasks Service (Терминал 2)
```powershell
$env:TASKS_PORT="8082"
$env:AUTH_BASE_URL="http://localhost:8081"
go run ./services/tasks/cmd/tasks
```

### Переменные окружения

| Переменная | Сервис | По умолчанию |
|------------|--------|--------------|
| AUTH_PORT | Auth | 8081 |
| TASKS_PORT | Tasks | 8082 |
| AUTH_BASE_URL | Tasks | http://localhost:8081 |

---

## 7. Тестирование через curl

```powershell
# Получить токен
Set-Content -Path body.json -Value '{"username":"student","password":"student"}'
curl.exe -s -X POST http://localhost:8081/v1/auth/login -H "Content-Type: application/json" -H "X-Request-ID: req-001" -d "@body.json"

# Проверить токен напрямую
curl.exe -i http://localhost:8081/v1/auth/verify -H "Authorization: Bearer demo-token" -H "X-Request-ID: req-002"

# Создать задачу
Set-Content -Path task.json -Value '{"title":"Do PZ1","description":"split services","due_date":"2026-01-10"}'
curl.exe -i -X POST http://localhost:8082/v1/tasks -H "Content-Type: application/json" -H "Authorization: Bearer demo-token" -H "X-Request-ID: req-003" -d "@task.json"

# Список задач
curl.exe -i http://localhost:8082/v1/tasks -H "Authorization: Bearer demo-token" -H "X-Request-ID: req-004"

# Без токена — должен вернуть 401
curl.exe -i http://localhost:8082/v1/tasks -H "X-Request-ID: req-005"
```

---

## 8. Ответы на контрольные вопросы

**1. Почему межсервисный вызов должен иметь таймаут?**

Таймаут предотвращает каскадные отказы: если Auth Service зависнет, Tasks не будет бесконечно ждать ответа, а вернёт клиенту ошибку через 3 секунды. Без таймаута накапливаются заблокированные горутины, что приводит к исчерпанию ресурсов и отказу всего сервиса.

**2. Чем request-id помогает при диагностике ошибок?**

`X-Request-ID` — уникальный идентификатор, который прокидывается через все сервисы в цепочке вызовов. Позволяет связать логи Auth и Tasks для одного пользовательского запроса и быстро найти, на каком этапе произошла ошибка.

**3. Какие статусы нужно вернуть клиенту при невалидном токене?**

`401 Unauthorized` — стандартный HTTP-статус, означающий что клиент не прошёл аутентификацию.

**4. Чем опасно "делить одну БД" между сервисами?**

Общая БД создаёт жёсткую связность между сервисами: изменение схемы в одном сервисе ломает другой. Это противоречит принципу независимости микросервисов, усложняет масштабирование и делает невозможным независимый деплой.



# Практика 2

---

## Тема: gRPC — создание простого микросервиса, вызовы методов

---

## Структура проекта

```
tech-ip-sem2/
├── proto/
│   ├── auth.proto                    # Контракт gRPC
│   └── auth/
│       ├── auth.pb.go                # Сгенерированный код
│       └── auth_grpc.pb.go           # Сгенерированный код
├── services/
│   ├── auth/
│   │   ├── cmd/auth/
│   │   │   └── main.go              # HTTP + gRPC сервер
│   │   └── internal/
│   │       ├── grpc/
│   │       │   └── server.go        # gRPC сервер Auth
│   │       ├── http/
│   │       │   └── handler.go
│   │       └── service/
│   │           └── auth.go
│   └── tasks/
│       ├── cmd/tasks/
│       │   └── main.go
│       └── internal/
│           ├── http/
│           │   └── handler.go
│           ├── service/
│           │   └── task.go
│           └── client/authclient/
│               ├── http.go          # HTTP клиент (ПЗ1)
│               └── grpc.go          # gRPC клиент (ПЗ2)
├── shared/
│   ├── middleware/
│   │   ├── requestid.go
│   │   └── logging.go
│   └── httpx/
│       └── client.go
├── docs/
│   └── api.md
├── go.mod
└── README.md
```

---

## 1. Контракт — proto файл

**`proto/auth.proto`:**

```protobuf
syntax = "proto3";

package auth;

option go_package = "github.com/omnik/tech-ip-sem2/proto/auth";

service AuthService {
  rpc Verify(VerifyRequest) returns (VerifyResponse);
}

message VerifyRequest {
  string token = 1;
}

message VerifyResponse {
  bool valid = 1;
  string subject = 2;
  string error = 3;
}
```

`.proto` файл является контрактом — он формально описывает API (сервисы, методы, структуры сообщений). По нему генерируется код для клиента и сервера: обе стороны обязаны следовать описанным типам и сигнатурам.

---

## 2. Команды генерации кода

```powershell
# Установка плагинов
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Генерация Go кода из proto
protoc --go_out=. --go_opt=module=github.com/omnik/tech-ip-sem2 \
       --go-grpc_out=. --go-grpc_opt=module=github.com/omnik/tech-ip-sem2 \
       proto/auth.proto
```

Сгенерированные файлы находятся в `proto/auth/`:
- `auth.pb.go` — структуры сообщений
- `auth_grpc.pb.go` — интерфейсы сервера и клиента

---

## 3. Схема взаимодействия

```mermaid
sequenceDiagram
    participant C as Client
    participant T as Tasks Service (HTTP)
    participant A as Auth Service (gRPC)

    C->>T: HTTP запрос с Authorization: Bearer token
    T->>A: gRPC Verify(token) — deadline 2с
    A-->>T: VerifyResponse(valid, subject) или Unauthenticated
    T-->>C: 200/201/204 или 401/503
```

---

## 4. Маппинг ошибок gRPC → HTTP

| Ситуация | gRPC код | HTTP статус | Описание |
|---|---|---|---|
| Невалидный токен | `Unauthenticated` | 401 Unauthorized | Auth вернул что токен невалиден |
| Auth недоступен | `Unavailable` | 503 Service Unavailable | Сервис не отвечает |
| Превышен deadline | `DeadlineExceeded` | 503 Service Unavailable | Таймаут gRPC вызова (2 секунды) |
| Внутренняя ошибка | `Internal` | 503 Service Unavailable | Ошибка на стороне Auth |

---

## 5. Скриншоты Postman

### 5.1 POST /v1/auth/login — получение токена (200 OK)
![login](image/pz2_p1.png)

### 5.2 POST /v1/tasks — создание задачи через gRPC verify (201 Created)
![create task](image/pz2_p2.png)

### 5.3 GET /v1/tasks — список задач (200 OK)
![list tasks](image/pz2_p3.png)

### 5.4 GET /v1/tasks без токена (401 Unauthorized)
![unauthorized](image/pz2_p4.png)

### 5.5 GET /v1/tasks при недоступном Auth (503 Service Unavailable)
![auth unavailable](image/pz2_p5.png)

---

## 6. Логи с подтверждением gRPC вызовов

### Логи Auth Service — gRPC Verify
![auth grpc logs](image/pz2_log1.png)

### Логи Tasks Service — Calling Auth gRPC verify
![tasks grpc logs](image/pz2_log2.png)

---

## 7. Инструкция запуска

### Требования
- Go 1.22+
- protoc 29.3+
- protoc-gen-go, protoc-gen-go-grpc

### Клонирование
```bash
git clone https://github.com/omnikk/pz2.1.git
cd pz2.1
go mod download
```

### Запуск Auth Service — Терминал 1
```powershell
$env:AUTH_PORT="8081"
$env:AUTH_GRPC_PORT="50051"
go run ./services/auth/cmd/auth
```

### Запуск Tasks Service в режиме gRPC — Терминал 2
```powershell
$env:TASKS_PORT="8082"
$env:AUTH_GRPC_ADDR="localhost:50051"
$env:AUTH_MODE="grpc"
go run ./services/tasks/cmd/tasks
```

### Запуск Tasks Service в режиме HTTP (ПЗ1) — Терминал 2
```powershell
$env:TASKS_PORT="8082"
$env:AUTH_BASE_URL="http://localhost:8081"
$env:AUTH_MODE="http"
go run ./services/tasks/cmd/tasks
```

### Переменные окружения

| Переменная | Сервис | По умолчанию | Описание |
|---|---|---|---|
| AUTH_PORT | Auth | 8081 | Порт HTTP сервера |
| AUTH_GRPC_PORT | Auth | 50051 | Порт gRPC сервера |
| TASKS_PORT | Tasks | 8082 | Порт HTTP сервера |
| AUTH_MODE | Tasks | grpc | Режим: `grpc` или `http` |
| AUTH_GRPC_ADDR | Tasks | localhost:50051 | Адрес Auth gRPC |
| AUTH_BASE_URL | Tasks | http://localhost:8081 | Адрес Auth HTTP |

---

## 8. Ответы на контрольные вопросы

**1. Что такое .proto и почему он считается контрактом?**

Файл `.proto` — формальное описание API на языке Protocol Buffers: сервисы, методы, структуры сообщений. Он является контрактом потому что по нему генерируется код для клиента и сервера — обе стороны обязаны следовать описанным типам и сигнатурам. Изменение контракта требует согласования обеих сторон.

**2. Что такое deadline в gRPC и чем он полезен?**

Deadline — абсолютное время до которого вызов должен завершиться. В отличие от таймаута, deadline автоматически пробрасывается через цепочку вызовов. Если время истекло, gRPC отменяет запрос и возвращает `DeadlineExceeded`. В нашем случае deadline = 2 секунды: если Auth не ответил за это время, Tasks возвращает клиенту 503.

**3. Почему "exactly-once" не даётся просто так даже в RPC?**

Сетевые сбои делают невозможным гарантировать однократное выполнение из коробки: запрос мог дойти до сервера, но ответ потерялся, и клиент повторит вызов. Для exactly-once нужна идемпотентность операций и дедупликация на стороне сервера.

**4. Как обеспечивать совместимость при расширении .proto?**

- Добавлять новые поля с новыми номерами (не переиспользовать старые)
- Не удалять и не переименовывать существующие поля
- Использовать `reserved` для защиты удалённых номеров полей от повторного использования
- Старые клиенты просто игнорируют неизвестные поля

