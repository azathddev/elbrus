# Сервис генерации рекламных баннеров на Go (с SDK)

Пример проекта, где генерация изображений выполняется через Go SDK `github.com/tigusigalpa/gigachat-go`, а не через ручные HTTP-запросы.

## Структура

- `backend` — Go API (`POST /api/generate`, `GET /health`) с использованием SDK
- `frontend` — статический интерфейс (HTML/CSS/JS)

## Настройка backend

1. Перейдите в папку:
   ```bash
   cd backend
   ```
2. Заполните переменные окружения:
   - `GIGACHAT_AUTH_KEY` — Base64 ключ (`ClientID:ClientSecret`) (обязательно, если не указаны `GIGACHAT_CLIENT_ID`/`GIGACHAT_CLIENT_SECRET`)
   - `GIGACHAT_CLIENT_ID` + `GIGACHAT_CLIENT_SECRET` — альтернативный вариант, ключ соберется автоматически
   - `GIGACHAT_IMAGE_MODEL` — по умолчанию `GigaChat-2-Max`
   - `GIGACHAT_CA_CERT_PATH` — путь к PEM сертификату корпоративного корневого CA (опционально)
   - `GIGACHAT_INSECURE_SKIP_VERIFY` — `true/false`, отключение проверки TLS (только для dev)
   - `PORT` — по умолчанию `8081`

Пример для PowerShell:
```powershell
$env:GIGACHAT_AUTH_KEY="base64_clientid_colon_clientsecret"
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
    "prompt": "Нарисуй рекламный баннер..."
  }
}
```
