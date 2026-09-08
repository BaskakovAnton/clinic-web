# Деплой clinic на Amvera

Два проекта: **PostgreSQL** и **веб (Go)**. Код: ветка с мордой, вход `cmd/web`, образ из `Dockerfile`.

Официально: [Go + PostgreSQL](https://docs.amvera.ru/general/examples/go-postgresql.html).

## 1. PostgreSQL (отдельный проект)

1. Создать сервис PostgreSQL в [cloud.amvera.ru](https://cloud.amvera.ru/projects).
2. Сохранить из «Инфо»: user, password, dbname, host вида `amvera-<user>-cnpg-<name>-rw`, port `5432`.
3. Подключиться снаружи/туннелем и накатить **только** `sborka/` (не `roles_etalon.sql` с `postgres`):

```text
01_schema.sql → 02_views.sql → 03_roles.sql → 04_seed.sql
GRANT CONNECT ON DATABASE clinic_db TO u_admin, u_registrar, u_doctor;
```

4. Рекомендуется отдельный пул-пользователь `clinic_app` с широкими GRANT (как admin) и своим паролем.
5. Пароли `u_*` на проде — **не** оставлять учебный `ClinicDemo1!`, если URL публичный.
6. Проверка: counts patients 3 / staff 4 / appointments 7 / visits 2 / visit_orders 5.

## 2. Веб-приложение

1. Создать приложение, привязать GitHub `goshva/clinic`, ветку с Docker/`amvera.yml`.
2. В Secrets / Variables:

| Имя | Пример |
|-----|--------|
| `PORT` | `80` |
| `DATABASE_URL` | `postgres://clinic_app:…@amvera-…-cnpg-…-rw:5432/clinic_db?sslmode=disable` |
| `SESSION_SECRET` | длинная случайная строка ≥32 |
| `APP_ENV` | `production` |
| `SECURE_COOKIE` | `true` |

3. Дождаться статуса «Успешно развернуто»; логи сборки и runtime.
4. Включить доменное имя в настройках проекта.
5. Проверка: `https://<домен>/healthz` → `ok`.

## 3. Приёмка

- Логин `u_registrar` / `u_doctor` / `u_admin`
- Запись слота, визит, дашборд админа
- Чужие разделы → 403

## 4. Важно

- Слушает `0.0.0.0:$PORT` (иначе 502).
- В образ входят `templates/` и `static/`.
- Внутренний host БД `…-rw`, не публичный, для связи app↔db внутри Amvera.
- `.env` в git не коммитить.
