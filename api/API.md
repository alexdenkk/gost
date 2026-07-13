# Labs API — документация

Все эндпоинты возвращают JSON.

> **Базовый URL:** `http(s)://127.0.0.1:8080`

---

## Аутентификация

### JWT Access Token
Для эндпоинтов, помеченных как **Auth**, требуется заголовок:

- `Authorization: Bearer <access_token>`

Внутренняя middleware проверяет, что строка в заголовке имеет минимум 8 символов, затем берёт токен как `Authorization[7:]` (т.е. ожидается префикс `Bearer ` длиной 7 символов).

Если токен невалидный — ответ `403 Forbidden` с телом:
```json
{"error":"not authorized"}
```

### JWT Refresh Token
Refresh-токен используется только для эндпоинта `POST /user/refresh/`.

---

## Общие ответы об ошибках

Коды ошибок в текущей реализации:
- `500 Internal Server Error` — при любых ошибках в обработчиках (парсинг/валидация/сервис/репозиторий).
- `403 Forbidden` — ошибка авторизации по access token.

Формат тела при `500`:
```json
{"error":"<message>"}
```

---

## Health

### `GET /health/`
Проверка доступности.

**Ответ 200**:
```json
{"message":"i love brainfuck!"}
```

---

## Пользователи (User)

### `POST /user/signup/`
Регистрация пользователя.

**Body (JSON)** (ожидается объект `map[string]string`):
- `email` (string)
- `password` (string)

**Поведение**:
- пароль хэшируется bcrypt (`bcrypt.DefaultCost`)
- создаётся запись пользователя в БД

**Ответ 200**:
```json
{"message":"user signed up"}
```

**Ошибки** (500):
```json
{"error":"user with this email already exists"}
```

---

### `POST /user/signin/`
Вход по email/паролю.

**Body (JSON)**:
- `email` (string)
- `password` (string)

**Ответ 200**:
```json
{
  "access_token": "<jwt>",
  "refresh_token": "<jwt>"
}
```

**Ошибки (500)**:
```json
{"error":"invalid email or password"}
```

---

### `POST /user/refresh/`
Обновление access token по refresh token.

**Body (JSON)**:
- `refresh_token` (string)

**Ответ 200**:
```json
{
  "access_token": "<jwt>",
  "refresh_token": "<jwt>"
}
```

**Ошибки (500)**:
```json
{"error":"error while generating token"}
```

---

### `GET /user/self/`  (Auth)
Получение текущего пользователя.

Требует заголовок `Authorization: Bearer <access_token>`.

**Ответ 200**:
```json
{
  "id": "<uuid>",
  "email": "<email>"
}
```

> Поля `PasswordHash` в JSON не возвращаются (`json:"-"`).

---

## Лабы (Lab)

### `POST /lab/`  (Auth)
Запуск генерации лаб.

Эндпоинт создаёт сущность `Lab`, вызывает внешний **agent API** два раза (formatter и generator), пишет файлы в `./files/`, затем запускает `./typst compile` и сохраняет результат.

**Важно:** структура входной JSON соответствует `domain.Lab`, но эндпоинт принимает `domain.Lab` целиком:

`domain.Lab` (как ожидается для JSON):
- `text` (string) — используется как `Message` в formatter
- `images` (array of Image)
  - `description` (string)
  - `base_64` (string) — base64 данных изображения

Дополнительные поля (например метаданные `department`, `student`, и т.д.) пробрасываются в шаблон `template.typ` через `tmpl.Execute(&sb, lab)`.

**Body (JSON) — пример**:
```json
{
  "text": "...",
  "images": [
    {"description": "img1", "base_64": "<base64>"}
  ],
  "department": "...",
  "professor": "...",
  "course": "...",
  "group": "...",
  "student": "...",
  "lab_number": "...",
  "lab_title": "..."
}
```

**Ответ 200**:
```json
{"message":"lab generated"}
```

**Ошибки (500)** — возвращается `{"error":"<err>"}`.

---

### `GET /lab/self/`  (Auth)
Получить список лаб текущего пользователя.

**Ответ 200**:
```json
[
  {
    "id": "<uuid>",
    "user_id": "<uuid>",
    "filename": "<string>",
    "date": "...",
    "department": "...",
    "professor": "...",
    "course": "...",
    "group": "...",
    "student": "...",
    "lab_number": "...",
    "lab_title": "...",
    "images": [
      {"description":"...","base_64":"..."}
    ]
  }
]
```

> В БД фактически сохраняются `Lab` и связанные поля GORM-модели. `Text` не хранится в базе (`gorm:"-"`), но может приходить/не приходить в зависимости от того, где и как сформирована структура.

---

### `GET /lab/{id}/`  (Auth)
Получить конкретную лабу по ID.

- `id` — UUID в пути.

Проверка доступа: лаба должна принадлежать текущему пользователю (`lab.UserID == claims.UserID`).

**Ответ 200**:
```json
{
  "id": "<uuid>",
  "user_id": "<uuid>",
  "filename": "<string>",
  "date": "..."
  
  // прочие поля Lab
}
```

**Ошибки (500)**:
```json
{"error":"lab not found"}
```

---

### `DELETE /lab/{id}/`  (Auth)
Удалить лабу по ID.

- `id` — UUID в пути.

Проверка доступа аналогична `GET /lab/{id}/`.

**Ответ 200**:
```json
{"message":"lab deleted"}
```

**Ошибки (500)**:
```json
{"error":"lab not found"}
```

---

## Служебные статические файлы

### `GET /files/{filename}`
Файлы из директории `./files` обслуживаются напрямую через `http.FileServer`.

Эти файлы создаются во время генерации лаб:
- изображения: `./files/<uuid>.png`
- исходник шаблона `.typ`: `./files/<lab_uuid>.typ`
- результат компиляции: зависит от `typst compile` (обычно `.pdf`).

---

## Интеграция с внешним Agent API (внутреннее описание)

Генерация использует `domain.AgentAdapter`:

### Формирование запросов
Метод `POST <BaseURL><accessID>/call` с заголовками:
- `Authorization: Bearer <AccessToken>`
- `x-proxy-source: ` (пустое значение)
- `Content-Type: application/json`

В теле запроса отправляется `domain.AgentRequest`:
```json
{
  "message": "<string>",
  "parent_message_id": "<string>",
  "file_ids": ["<string>"],
  "metadata": {}
}
```

Валидация ответа: ожидается HTTP `200 OK`, затем JSON декодируется в `domain.AgentResponse`:
```json
{
  "message": "<string>",
  "id": "<string>"
}
```

---

## Структуры данных (DTO)

### `User` (для JSON)
```json
{
  "id": "<uuid>",
  "email": "<string>"
}
```

### `Lab`
```json
{
  "id": "<uuid>",
  "user_id": "<uuid>",
  "filename": "<string>",
  "date": "<string>",
  "department": "<string>",
  "professor": "<string>",
  "course": "<string>",
  "group": "<string>",
  "student": "<string>",
  "lab_number": "<string>",
  "lab_title": "<string>",
  "text": "<string>",
  "images": [
    {"description":"<string>","base_64":"<base64>"}
  ]
}
```

> `text` и `images` имеют `gorm:"-"` — вне JSON, но включаются при генерации.

---

## Примечания по обработке ошибок и HTTP-кодам

- Большинство ошибок во входных/сервисных сценариях возвращаются как `500`.
- Ошибка авторизации возвращает `403`.
- Эндпоинты не документируют `400`/`401` — в текущем коде они не используются.

