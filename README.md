# Auth-Service-Test-BackDev

## Описание

Auth-Service-Test-BackDev — это сервис аутентификации, реализованный на языке Go с использованием JWT и PostgreSQL. Сервис предоставляет два REST API маршрута для генерации и обновления пар токенов доступа (Access Token) и обновления (Refresh Token). Проект включает в себя контейнеризацию с использованием Docker и покрыт тестами для обеспечения надежности.

## Используемые технологии

- **Go**: Основной язык программирования.
- **JWT (JSON Web Tokens)**: Для создания и валидации токенов доступа.
- **PostgreSQL**: Система управления базами данных.
- **Gin**: Веб-фреймворк для создания REST API.
- **Docker**: Контейнеризация приложения и базы данных.
- **Goose**: Инструмент для управления миграциями базы данных.
- **Testify**: Фреймворк для написания тестов.
- **Bcrypt**: Для хэширования Refresh токенов.

## Функционал

- **Генерация токенов**: Предоставляет пару Access и Refresh токенов для указанного пользователя.
- **Обновление токенов**: Позволяет обновить пару токенов, используя действительный Refresh токен.
- **Безопасность**:
  - Access токен представляет собой JWT, подписанный с использованием алгоритма HS512.
  - Refresh токен генерируется как случайная строка, кодируется в base64 и хранится в базе данных исключительно в виде bcrypt хеша.
  - Связь между Access и Refresh токенами обеспечивается их совместным хранением и проверкой.
  - В payload токенов включается IP-адрес клиента для дополнительной проверки безопасности. При изменении IP адреса отправляется предупреждение на электронную почту пользователя (реализовано с использованием моковых данных).

## Структура проекта
```bash
auth-service-test-BackDev/
├── cmd/
│   ├── main.go
│   └── .env
├── docker-compose.yaml
├── Dockerfile
├── Makefile
├── go.mod
├── internal/
│   ├── db/
│   │   └── postgres.go
│   ├── handlers/
│   │   ├── auth.go
│   │   └── auth_test.go
│   └── utils/
│       ├── hashing.go
│       ├── jwt.go
│       └── utils_test.go
├── migrations/
│   └── 20241203220540_migration_users_table.sql
└── .env

## Установка и запуск

### Предварительные требования

- **Docker** и **Docker Compose** установлены на вашей машине.
- **Go** установлен (для разработки и тестирования).

### Настройка переменных окружения

Создайте файл `.env` в корне проекта и заполните следующие переменные:

```env
PG_DATABASE_NAME=your_database_name
PG_USER=your_database_user
PG_PASSWORD=your_database_password
PG_PORT=5432
MIGRATION_DIR=./migrations
DATABASE_URL=postgres://your_database_user:your_database_password@pg:5432/your_database_name?sslmode=disable
SECRET_KEY=your_secret_key
PORT=8080

## API Маршруты

### Генерация токенов

- **URL:** `/auth/generate-tokens`
- **Метод:** `POST`
- **Параметры запроса:**
  - `user_id` (UUID) — идентификатор пользователя.
- **Ответ:**
  - `access_token`: JWT токен.
  - `refresh_token`: Base64 закодированный токен.

**Пример запроса:**

```bash
POST http://localhost:8080/auth/generate-tokens?user_id=123e4567-e89b-12d3-a456-426614174000
Content-Type: application/json

{
  "access_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "dGhpc19pcyBhX3JlZnJlc2hfZG9hdA=="
}

### Обновление токенов

- **URL:** `/auth/refresh-tokens`
- **Метод:** `POST`
- **Тело запроса (JSON):**
  - `access_token`: Текущий Access токен.
  - `refresh_token`: Текущий Refresh токен.
- **Ответ:**
  - `access_token`: Новый JWT токен.
  - `refresh_token`: Новый Base64 закодированный токен.

**Пример запроса:**

```bash
POST http://localhost:8080/auth/refresh-tokens
Content-Type: application/json

{
  "access_token": "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "dGhpc19pcyBhX3JlZnJlc2hfZG9hdA=="
}

## Миграции базы данных

Миграции управляются с помощью **Goose**. Файлы миграций находятся в директории `migrations/`.

### Применение миграций:

**Накат миграций:**
```bash
make local-migration-up

**Откат миграций:**
```bash
make local-migration-down

**Проверка статуса миграций**
```bash
make local-migration-status

## Docker

Проект настроен для использования Docker. В `docker-compose.yaml` определены два сервиса:

- **pg**: PostgreSQL база данных.
- **app**: Сам сервис аутентификации.

## Makefile

В проекте используется Makefile для упрощения управления миграциями и установкой зависимостей.

## Авторы

- **Ваше Имя** — [titoffon](https://github.com/titoffon)
