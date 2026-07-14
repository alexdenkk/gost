# TODO — auth check + refresh daemon

- [ ] `frontend/web/html/auth.html`: добавить проверку `GET /user/self/` при наличии access token; если `ok` → редирект на `/lk/list/`
- [ ] `frontend/web/html/list.html`: добавить `GET /user/self/` в начало; если не авторизован → редирект на `/auth/`; refresh-daemon запускать только после успешной проверки
- [ ] `frontend/web/html/generate.html`: добавить `GET /user/self/` в начало; если не авторизован → редирект на `/auth/`; refresh-daemon запускать только после успешной проверки
- [ ] Исправить редиректы на страницах `list.html`/`generate.html` (вместо `auth.html` использовать `/auth/`)
- [ ] Проверить ручные сценарии (без токенов/с токенами/протухшие токены)
