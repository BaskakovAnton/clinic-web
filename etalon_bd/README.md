# Эталон ORIGINAL (полное наполнение)

| Файл | Зачем |
|------|--------|
| `clinic_db_etalon.dump` | Restore в pgAdmin (основной способ) |
| `clinic_db_etalon.sql` | Текстовый дамп |
| `roles_etalon.sql` | Учебные роли, если их ещё нет |

Counts: patients **3**, staff **4**, appointments **7**, visits **2**, visit_orders **5**.

Урезанный демо-эталон: `../../demo/etalon_bd/`.

---

## Пароли

- **`postgres`** — пароль при установке PostgreSQL.  
- **`u_*`** — учебный пароль `ClinicDemo1!`

---

## Восстановление

Под **`postgres`**:

1. Создайте пустую `clinic_db` (старую при необходимости удалите).  
2. ПКМ → **Restore…** → `clinic_db_etalon.dump`.  
3. Выполните `roles_etalon.sql` (если роли уже есть — «already exists» нормально).  
4. Выполните:

```sql
GRANT CONNECT ON DATABASE clinic_db TO u_admin, u_registrar, u_doctor;
```

5. Проверка:

```sql
SELECT count(*) FROM patients;  -- 3
SELECT count(*) FROM staff;     -- 4
```

Сборка скриптами: `../sborka/INSTRUKCIYA.md`.
