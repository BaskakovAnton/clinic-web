-- clinic_db · 08_staff_photo.sql
-- Пол и фото врача: отдельная таблица (u_admin может CREATE; ALTER staff требует владельца).
-- Идемпотентно.

BEGIN;

CREATE TABLE IF NOT EXISTS staff_photo (
    staff_id  INTEGER PRIMARY KEY REFERENCES staff (id) ON DELETE CASCADE,
    gender    CHAR(1),
    image_url TEXT NOT NULL DEFAULT '',
    CONSTRAINT staff_photo_gender_chk
        CHECK (gender IS NULL OR gender IN ('f', 'm'))
);

COMMENT ON TABLE staff_photo IS 'Пол и URL фото сотрудника (заглушка по gender, если URL пуст)';
COMMENT ON COLUMN staff_photo.gender IS 'Пол: f / m — для заглушки аватара';
COMMENT ON COLUMN staff_photo.image_url IS 'URL фото; пусто → заглушка по gender';

GRANT ALL ON TABLE staff_photo TO clinic_admin;
GRANT SELECT ON TABLE staff_photo TO clinic_registrar;
GRANT SELECT ON TABLE staff_photo TO clinic_doctor;

-- Учебный seed: пол по демо-ФИО
INSERT INTO staff_photo (staff_id, gender, image_url)
SELECT s.id, v.gender, ''
FROM staff s
JOIN (VALUES
    ('Иванова Анна Сергеевна', 'f'),
    ('Сидорова Елена Викторовна', 'f'),
    ('Козлова Мария Павловна', 'f'),
    ('Петров Дмитрий Игоревич', 'm')
) AS v(full_name, gender) ON v.full_name = s.full_name
ON CONFLICT (staff_id) DO UPDATE
SET gender = COALESCE(staff_photo.gender, EXCLUDED.gender);

COMMIT;
