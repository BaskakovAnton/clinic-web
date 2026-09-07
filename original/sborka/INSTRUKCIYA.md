# Сборка базы clinic_db с нуля (ORIGINAL)

Полное учебное наполнение. Урезанное демо: `../../demo/sborka/`.

| Порядок | Файл | Что делает |
|---------|------|------------|
| 1 | `01_schema.sql` | Таблицы, ограничения, триггеры |
| 2 | `02_views.sql` | Отчёты |
| 3 | `03_roles.sql` | Роли и логины |
| 4 | `04_seed.sql` | **Полные** учебные данные |
| 5 | `05_demo_queries.sql` | Примеры запросов (не обязателен) |

Дамп: **`../etalon_bd/`**.

Для сдачи поднимайте эталон или собирайте по этой инструкции заново — не старую базу после долгих тестов.

---

## Что нужно заранее

1. **PostgreSQL** и **pgAdmin**.  
2. Сервер запущен.  
3. Пароль **`postgres`** — от **установки** PostgreSQL (на каждом ПК свой).

---

## Шаг 1. Подключение

Host `localhost`, Port `5432`, User `postgres`, свой пароль установки.

---

## Шаг 2. Пустая база

Создайте `clinic_db`. Если уже есть и нужна чистая — удалите и создайте снова.

---

## Шаг 3. Скрипты по порядку

Query Tool на `clinic_db`:

```text
01_schema.sql → 02_views.sql → 03_roles.sql → 04_seed.sql
```

Ошибка «уже существует» → удалите базу, начните с шага 2.

---

## Шаг 4. GRANT CONNECT

```sql
GRANT CONNECT ON DATABASE clinic_db TO u_admin, u_registrar, u_doctor;
```

---

## Шаг 5. Проверка

```sql
SELECT 'patients' AS t, count(*)::text AS n FROM patients
UNION ALL SELECT 'staff', count(*)::text FROM staff
UNION ALL SELECT 'appointments', count(*)::text FROM appointments
UNION ALL SELECT 'visits', count(*)::text FROM visits
UNION ALL SELECT 'visit_orders', count(*)::text FROM visit_orders
ORDER BY 1;
```

Ожидаемо (полное): patients **3**, staff **4**, appointments **7**, visits **2**, visit_orders **5**.

```sql
SELECT * FROM v_free_slots_by_specialty LIMIT 5;
```

---

## Учебные логины

Пароль **только для них:** `ClinicDemo1!`  
Не путать с паролем `postgres`.

| Логин | Кто |
|-------|-----|
| `u_admin` | Админ |
| `u_registrar` | Регистратура |
| `u_doctor` | Врач |
