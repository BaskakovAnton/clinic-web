-- clinic_db · 02_views.sql
-- Четыре задачи методички (п. 55)
-- Запускать после 01_schema.sql

BEGIN;

-- 1. Свободные слоты к специалисту
CREATE OR REPLACE VIEW v_free_slots_by_specialty AS
SELECT
    a.id            AS appointment_id,
    s.full_name     AS doctor_name,
    s.specialty,
    s.department,
    s.office,
    a.start_at
FROM appointments a
JOIN staff s ON s.id = a.doctor_id
WHERE a.status = 'free'
  AND s.staff_kind = 'doctor';

COMMENT ON VIEW v_free_slots_by_specialty IS
    'Поиск свободных слотов к нужному специалисту';

-- 2. Электронная карта / история болезни
CREATE OR REPLACE VIEW v_patient_medical_card AS
SELECT
    p.id                 AS patient_id,
    p.card_number,
    p.full_name          AS patient_name,
    p.birth_date,
    p.insurance_type,
    p.insurance_number,
    v.id                 AS visit_id,
    v.visit_at,
    d.full_name          AS doctor_name,
    d.specialty          AS doctor_specialty,
    v.complaints,
    v.diagnosis_icd10,
    v.appointment_id,
    o.id                 AS order_id,
    o.order_kind,
    o.description        AS order_description,
    o.is_prescription,
    o.is_preferential
FROM patients p
LEFT JOIN visits v ON v.patient_id = p.id
LEFT JOIN staff d ON d.id = v.doctor_id
LEFT JOIN visit_orders o ON o.visit_id = v.id;

COMMENT ON VIEW v_patient_medical_card IS
    'Выдача истории болезни пациента (электронная карта)';

-- 3. Нагрузка на врачей
CREATE OR REPLACE VIEW v_doctor_workload AS
SELECT
    s.id            AS doctor_id,
    s.full_name     AS doctor_name,
    s.specialty,
    s.department,
    COUNT(v.id)     AS patients_seen
FROM staff s
LEFT JOIN visits v ON v.doctor_id = s.id
WHERE s.staff_kind = 'doctor'
GROUP BY s.id, s.full_name, s.specialty, s.department;

COMMENT ON VIEW v_doctor_workload IS
    'Расчёт нагрузки на врачей (число принятых пациентов)';

-- 4. Льготные рецепты и медикаменты
CREATE OR REPLACE VIEW v_preferential_orders AS
SELECT
    o.id                 AS order_id,
    v.visit_at,
    p.card_number,
    p.full_name          AS patient_name,
    d.full_name          AS doctor_name,
    o.order_kind,
    o.description,
    o.is_prescription,
    o.is_preferential
FROM visit_orders o
JOIN visits v ON v.id = o.visit_id
JOIN patients p ON p.id = v.patient_id
JOIN staff d ON d.id = v.doctor_id
WHERE o.is_preferential = TRUE;

COMMENT ON VIEW v_preferential_orders IS
    'Учёт льготных рецептов и медикаментов';

COMMIT;
