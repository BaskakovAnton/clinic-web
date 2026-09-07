-- clinic_db · 04_seed.sql
-- Демо-данные для защиты (учебные, не реальные ПДн)
-- Запускать после 01–03

BEGIN;

-- Очистка при повторном прогоне (порядок из-за FK)
TRUNCATE visit_orders, visits, appointments, patients, staff
    RESTART IDENTITY CASCADE;

-- Сотрудники
INSERT INTO staff (full_name, staff_kind, specialty, department, work_schedule, office)
VALUES
    ('Иванова Анна Сергеевна',  'doctor', 'терапевт',  'Терапевтическое', 'Пн–Пт 09:00–15:00', '101'),
    ('Петров Дмитрий Игоревич', 'doctor', 'хирург',    'Хирургическое',   'Пн–Пт 10:00–16:00', '205'),
    ('Сидорова Елена Викторовна','doctor','окулист',   'Офтальмология',   'Вт, Чт 09:00–14:00', '310'),
    ('Козлова Мария Павловна',  'nurse',  NULL,        'Терапевтическое', 'Пн–Пт 08:00–16:00', '101');

-- Пациенты
INSERT INTO patients (
    card_number, full_name, birth_date,
    insurance_type, insurance_number, passport_data, address, contacts
)
VALUES
    ('A-1001', 'Смирнов Алексей Николаевич', '1985-03-12',
     'OMS', '7700 123456', '4500 123456, выдан ОВД', 'г. Москва, ул. Ленина, 1', '+7-900-111-22-33'),
    ('A-1002', 'Кузнецова Ольга Игоревна', '1992-07-21',
     'DMS', 'ДМС-998877', '4501 654321, выдан ОВД', 'г. Москва, ул. Мира, 5', '+7-900-222-33-44'),
    ('A-1003', 'Волков Игорь Петрович', '1978-11-03',
     'OMS', '7700 654321', '4502 111222, выдан ОВД', 'г. Москва, пр. Мира, 10', '+7-900-333-44-55');

-- Слоты: свободные и занятые
-- doctor ids: 1 терапевт, 2 хирург, 3 окулист
INSERT INTO appointments (doctor_id, start_at, status, patient_id)
VALUES
    (1, '2026-09-10 09:00:00', 'free',   NULL),
    (1, '2026-09-10 09:30:00', 'booked', 1),
    (1, '2026-09-10 10:00:00', 'free',   NULL),
    (2, '2026-09-10 11:00:00', 'free',   NULL),
    (2, '2026-09-10 11:30:00', 'booked', 2),
    (3, '2026-09-11 09:00:00', 'free',   NULL),
    (3, '2026-09-11 09:30:00', 'free',   NULL);

-- Визит по записи (appointment_id = 2 → пациент 1, терапевт)
INSERT INTO visits (patient_id, doctor_id, visit_at, complaints, diagnosis_icd10, appointment_id)
VALUES
    (1, 1, '2026-09-10 09:35:00',
     'Головная боль, слабость', 'G44.2', 2);

-- Визит без предварительной записи (пациент 3, хирург)
INSERT INTO visits (patient_id, doctor_id, visit_at, complaints, diagnosis_icd10, appointment_id)
VALUES
    (3, 2, '2026-09-09 12:00:00',
     'Боль в правом боку', 'K80.2', NULL);

-- Назначения (в т.ч. льготные)
INSERT INTO visit_orders (visit_id, order_kind, description, is_prescription, is_preferential)
VALUES
    (1, 'medication', 'Ибупрофен 200 мг при головной боли', TRUE,  FALSE),
    (1, 'test',       'Общий анализ крови',                 FALSE, FALSE),
    (2, 'medication', 'Урсодезоксихолевая кислота 250 мг',  TRUE,  TRUE),
    (2, 'procedure',  'УЗИ органов брюшной полости',       FALSE, FALSE),
    (2, 'medication', 'Спазмолитик по схеме (льготный)',   TRUE,  TRUE);

COMMIT;
