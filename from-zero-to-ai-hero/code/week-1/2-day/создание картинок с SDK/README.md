# Сервис генерации рекламных баннеров на Go (с SDK)

Пример проекта, где генерация изображений выполняется через Go SDK `github.com/sashabaranov/go-openai`, а не через ручные HTTP-запросы.

## Структура

- `backend` — Go API (`POST /api/generate`, `GET /health`) с использованием SDK
- `frontend` — статический интерфейс (HTML/CSS/JS)

## Настройка backend

1. Перейдите в папку:
   ```bash
   cd backend
   ```
2. Заполните переменные окружения:
   - `OPENAI_API_KEY` — ключ API (обязательно)
   - `OPENAI_BASE_URL` — опционально, если нужен совместимый proxy/endpoint
   - `OPENAI_IMAGE_MODEL` — по умолчанию `dall-e-3`
   - `PORT` — по умолчанию `8081`

Пример для PowerShell:
```powershell
$env:OPENAI_API_KEY="your_api_key"
$env:PORT="8081"
```

## Запуск backend

```bash
go run .
```

Проверка:
```bash
curl http://localhost:8081/health
```

## Запуск frontend

Откройте `frontend/index.html` в браузере.

Frontend отправляет запросы на:
- `http://localhost:8081/api/generate` при запуске через `file://`
- `/api/generate` при запуске за reverse proxy.

## API

### POST `/api/generate`

Тело:
```json
{
  "message": "Весенняя распродажа до 50%",
  "audience": "Студенты"
}
```

Ответ:
```json
{
  "result": {
    "imageUrl": "",
    "imageBase64": "iVBORw0KGgoAAA...",
    "prompt": "Создай рекламный баннер..."
  }
}
```
