-- clinic_db · 05_demo_queries.sql
-- Готовые запросы для защиты (копировать в Query Tool)

-- ========== 1. Свободные слоты к специалисту (например, терапевт) ==========
SELECT *
FROM v_free_slots_by_specialty
WHERE specialty = 'терапевт';

-- Все свободные слоты:
-- SELECT * FROM v_free_slots_by_specialty;

-- ========== 2. Электронная карта пациента ==========
SELECT *
FROM v_patient_medical_card
WHERE card_number = 'A-1001';

-- ========== 3. Нагрузка на врачей ==========
SELECT * FROM v_doctor_workload;

-- ========== 4. Льготные рецепты и медикаменты ==========
SELECT * FROM v_preferential_orders;

-- ========== Картотека и штат (обзор) ==========
SELECT id, card_number, full_name, insurance_type FROM patients ORDER BY id;
SELECT id, full_name, staff_kind, specialty, department, office FROM staff ORDER BY id;

-- ========== Расписание ==========
SELECT a.id, s.full_name AS doctor, a.start_at, a.status, p.full_name AS patient
FROM appointments a
JOIN staff s ON s.id = a.doctor_id
LEFT JOIN patients p ON p.id = a.patient_id
ORDER BY a.start_at;

-- ========== Проверка прав (выполнять под разными логинами) ==========
-- Под u_doctor — должно работать:
--   INSERT INTO visits ... / INSERT INTO visit_orders ...
-- Под u_doctor — должно ПАДАТЬ:
--   UPDATE staff SET office = '999' WHERE id = 1;
-- Под u_registrar — должно работать:
--   UPDATE appointments SET status = 'booked', patient_id = 3 WHERE id = 1 AND status = 'free';
-- Под u_registrar — должно ПАДАТЬ:
--   INSERT INTO visits (patient_id, doctor_id, visit_at, complaints, diagnosis_icd10)
--   VALUES (1, 1, now(), 'тест', 'J06.9');
