# Показ полной базы (10–15 минут)

Это **ORIGINAL** — полное учебное наполнение.

Подключение: pgAdmin → **`clinic_db`** (под `postgres` или `u_admin`).  
Перед новым запросом очищай Query Tool.

Если данные «поплыли» после тестов — восстановите эталон или выполните `sborka/04_seed.sql`.

---

## 1. Таблицы

Слева: `Schemas → public → Tables`  
`patients`, `staff`, `appointments`, `visits`, `visit_orders`.

**Показываем:** полный состав учёта поликлиники.

---

## 2. Штат (врачи и медсестра)

```sql
SELECT id, full_name, staff_kind, specialty, department, office
FROM staff
ORDER BY id;
```

**Показываем:** 3 врача + медсестра; слоты только у врачей.

---

## 3. Свободные слоты

```sql
SELECT * FROM v_free_slots_by_specialty
WHERE specialty = 'терапевт'
ORDER BY start_at;
```

**Показываем:** поиск свободного времени к специалисту.

---

## 4. Запись пациента

```sql
UPDATE appointments
SET status = 'booked', patient_id = 3
WHERE id = (
    SELECT appointment_id FROM v_free_slots_by_specialty
    WHERE specialty = 'терапевт'
    ORDER BY start_at LIMIT 1
)
RETURNING id, status, patient_id;
```

Если 0 строк — уберите фильтр по специальности во вложенном запросе.

**Показываем:** запись на приём.

---

## 5. Карта и льготы

```sql
SELECT patient_name, visit_at, doctor_name, diagnosis_icd10,
       order_kind, order_description, is_preferential
FROM v_patient_medical_card
WHERE card_number = 'A-1001'
ORDER BY visit_at, order_id;

SELECT * FROM v_preferential_orders;
```

**Показываем:** электронная карта и льготные назначения.

---

## 6. Нагрузка врачей

```sql
SELECT doctor_name, specialty, patients_seen
FROM v_doctor_workload
ORDER BY patients_seen DESC, doctor_name;
```

**Показываем:** нагрузку (ожидается 3 врача в списке).

---

## 7. Права доступа

```sql
SET ROLE clinic_doctor;
UPDATE staff SET office = '999' WHERE id = 1;
```

Ожидается ошибка «нет доступа». Затем:

```sql
RESET ROLE;
```

**Показываем:** врач не меняет штат.

---

## 8. Бэкап

ПКМ по `clinic_db` → **Backup…** → сохранить файл.

**Показываем:** резервное копирование.

---

## Про реальные данные

В пояснительной: на реальных данных — шифрование базы и защита текстовых полей с солью. В учебной БД это не включали. Текст: `zashchita/Fragment_bezopasnost.md`.

Для защиты вуза с готовыми фразами: папка `zashchita/Shpargalka.md`.
