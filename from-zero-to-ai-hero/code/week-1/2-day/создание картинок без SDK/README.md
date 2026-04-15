# Сервис генерации рекламных баннеров (без SDK)

Учебный проект на Go и JavaScript:
- бэкенд принимает текст сообщения и аудиторию;
- вызывает API генерации изображений через `net/http` без SDK;
- фронтенд отображает форму и список всех сгенерированных изображений.

## Структура

- `backend` — Go API (`POST /api/generate`, `GET /health`)
- `frontend` — статический интерфейс (HTML/CSS/JS)
- `docker-compose.yml` — запуск frontend и backend в контейнерах

## Настройка backend

1. Перейдите в папку:
   ```bash
   cd backend
   ```
2. Создайте `.env` из примера:
   - скопируйте `backend/.env.example` в `backend/.env`;
   - заполните `GIGACHAT_AUTH_TOKEN`.
3. Экспортируйте переменные окружения (в PowerShell):
   ```powershell
   $env:GIGACHAT_AUTH_TOKEN="ВАШ_ТОКЕН"
   $env:PORT="8080"
   ```

Опционально:
- `GIGACHAT_IMAGE_API_URL` — URL API генерации изображений;
- `GIGACHAT_IMAGE_MODEL` — модель генерации.
- `GIGACHAT_INSECURE_SKIP_VERIFY` — `true/false`, отключение TLS-проверки сертификата (только для dev).

## Запуск backend

```bash
go run .
```

Сервис будет доступен на `http://localhost:8080`.

Проверка здоровья:
```bash
curl http://localhost:8080/health
```

## Запуск frontend

Откройте файл `frontend/index.html` в браузере.

Далее:
1. Введите рекламное сообщение.
2. Выберите аудиторию.
3. Нажмите `Отправить`.
4. Новое сгенерированное изображение появится в галерее ниже.

## Docker Compose

1. В корне проекта создайте `.env` из примера:
   ```bash
   cp .env.example .env
   ```
2. Заполните `GIGACHAT_AUTH_TOKEN` в `.env`.
   - если в сети подменяется TLS-сертификат (корпоративный прокси), временно можно выставить:
     - `GIGACHAT_INSECURE_SKIP_VERIFY=true`
   - для production оставляйте `false` и установите корректные корневые сертификаты.
3. Запустите контейнеры:
   ```bash
   docker compose up --build
   ```
4. Откройте `http://localhost:8080`.

Остановка:
```bash
docker compose down
```

## API

### POST `/api/generate`

Тело запроса:
```json
{
  "message": "Скидки до 40% на курсы программирования",
  "audience": "Студенты"
}
```

Успешный ответ:
```json
{
  "result": {
    "imageUrl": "https://...",
    "imageBase64": "iVBORw0KGgoAAA...",
    "prompt": "Создай рекламный баннер..."
  }
}
```

Если `imageUrl` отсутствует, фронтенд использует `imageBase64`.

## Проверка сценариев

- Успешная генерация: корректный токен и непустые `message`/`audience`.
- Ошибка токена: API вернет ошибку, backend отдаст `502` с описанием.
- Пустой текст: backend вернет `400` и сообщение `message is required`.
- Недоступный внешний API: backend вернет `502`.
