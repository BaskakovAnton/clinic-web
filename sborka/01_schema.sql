-- clinic_db · 01_schema.sql
-- Каркас поликлиники (курсовая, п. 55)
-- Запуск: Query Tool в pgAdmin, подключение к базе clinic_db

BEGIN;

-- ---------------------------------------------------------------------------
-- patients — картотека
-- ---------------------------------------------------------------------------
CREATE TABLE patients (
    id                INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    card_number       VARCHAR(32)  NOT NULL,
    full_name         VARCHAR(200) NOT NULL,
    birth_date        DATE         NOT NULL,
    insurance_type    VARCHAR(8)   NOT NULL,
    insurance_number  VARCHAR(64)  NOT NULL,
    passport_data     VARCHAR(200) NOT NULL,
    address           VARCHAR(300),
    contacts          VARCHAR(200),
    CONSTRAINT patients_card_number_uq UNIQUE (card_number),
    CONSTRAINT patients_insurance_type_chk
        CHECK (insurance_type IN ('OMS', 'DMS'))
);

COMMENT ON TABLE patients IS 'Картотека пациентов (амбулаторные карты)';

-- ---------------------------------------------------------------------------
-- staff — врачи и медсёстры
-- ---------------------------------------------------------------------------
CREATE TABLE staff (
    id             INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    full_name      VARCHAR(200) NOT NULL,
    staff_kind     VARCHAR(16)  NOT NULL,
    specialty      VARCHAR(100),
    department     VARCHAR(100) NOT NULL,
    work_schedule  VARCHAR(200),
    office         VARCHAR(32),
    CONSTRAINT staff_kind_chk CHECK (staff_kind IN ('doctor', 'nurse')),
    CONSTRAINT staff_doctor_specialty_chk
        CHECK (
            (staff_kind = 'doctor' AND specialty IS NOT NULL)
            OR (staff_kind = 'nurse')
        )
);

COMMENT ON TABLE staff IS 'Сотрудники: врачи и медсёстры';
-- Пол и фото: см. staff_photo (08_staff_photo.sql)

-- ---------------------------------------------------------------------------
-- appointments — расписание приёма (слоты)
-- ---------------------------------------------------------------------------
CREATE TABLE appointments (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    doctor_id   INTEGER      NOT NULL REFERENCES staff (id) ON DELETE RESTRICT,
    start_at    TIMESTAMP    NOT NULL,
    status      VARCHAR(16)  NOT NULL,
    patient_id  INTEGER      REFERENCES patients (id) ON DELETE RESTRICT,
    CONSTRAINT appointments_status_chk CHECK (status IN ('free', 'booked')),
    CONSTRAINT appointments_status_patient_chk CHECK (
        (status = 'free'   AND patient_id IS NULL)
        OR (status = 'booked' AND patient_id IS NOT NULL)
    ),
    CONSTRAINT appointments_doctor_time_uq UNIQUE (doctor_id, start_at)
);

COMMENT ON TABLE appointments IS 'Слоты приёма: свободен / занят';

CREATE INDEX appointments_status_idx ON appointments (status);
CREATE INDEX appointments_start_at_idx ON appointments (start_at);

-- Слот только у врача (не у медсестры)
CREATE OR REPLACE FUNCTION appointments_doctor_only()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_kind VARCHAR(16);
BEGIN
    SELECT staff_kind INTO v_kind FROM staff WHERE id = NEW.doctor_id;
    IF v_kind IS NULL THEN
        RAISE EXCEPTION 'Сотрудник id=% не найден', NEW.doctor_id;
    END IF;
    IF v_kind <> 'doctor' THEN
        RAISE EXCEPTION 'Слот приёма можно создать только для врача (staff_kind=doctor)';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_appointments_doctor_only
    BEFORE INSERT OR UPDATE OF doctor_id ON appointments
    FOR EACH ROW
    EXECUTE PROCEDURE appointments_doctor_only();

-- ---------------------------------------------------------------------------
-- visits — факт посещения
-- ---------------------------------------------------------------------------
CREATE TABLE visits (
    id               INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    patient_id       INTEGER      NOT NULL REFERENCES patients (id) ON DELETE RESTRICT,
    doctor_id        INTEGER      NOT NULL REFERENCES staff (id) ON DELETE RESTRICT,
    visit_at         TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    complaints       TEXT,
    diagnosis_icd10  VARCHAR(16)  NOT NULL,
    appointment_id   INTEGER      UNIQUE REFERENCES appointments (id) ON DELETE SET NULL
);

COMMENT ON TABLE visits IS 'Посещения: жалобы, диагноз МКБ-10; связь с записью необязательна';

CREATE INDEX visits_patient_idx ON visits (patient_id);
CREATE INDEX visits_doctor_idx ON visits (doctor_id);

CREATE OR REPLACE FUNCTION visits_doctor_only()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    v_kind VARCHAR(16);
BEGIN
    SELECT staff_kind INTO v_kind FROM staff WHERE id = NEW.doctor_id;
    IF v_kind IS NULL OR v_kind <> 'doctor' THEN
        RAISE EXCEPTION 'Посещение оформляет только врач (staff_kind=doctor)';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_visits_doctor_only
    BEFORE INSERT OR UPDATE OF doctor_id ON visits
    FOR EACH ROW
    EXECUTE PROCEDURE visits_doctor_only();

-- ---------------------------------------------------------------------------
-- visit_orders — назначения и рецепты (в т.ч. льготные)
-- ---------------------------------------------------------------------------
CREATE TABLE visit_orders (
    id               INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    visit_id         INTEGER      NOT NULL REFERENCES visits (id) ON DELETE CASCADE,
    order_kind       VARCHAR(16)  NOT NULL,
    description      VARCHAR(500) NOT NULL,
    is_prescription  BOOLEAN      NOT NULL DEFAULT FALSE,
    is_preferential  BOOLEAN      NOT NULL DEFAULT FALSE,
    CONSTRAINT visit_orders_kind_chk
        CHECK (order_kind IN ('medication', 'procedure', 'test'))
);

COMMENT ON TABLE visit_orders IS 'Назначения: лекарства / процедуры / анализы; рецепт и льгота';

CREATE INDEX visit_orders_visit_idx ON visit_orders (visit_id);
CREATE INDEX visit_orders_preferential_idx ON visit_orders (is_preferential)
    WHERE is_preferential = TRUE;

COMMIT;
