-- clinic_db · 03_roles.sql
-- Роли и учебные пользователи + GRANT
-- Запускать после 01_schema.sql и 02_views.sql
-- Внимание: пароли учебные, только для курсовой/демо

BEGIN;

-- Роли (группы прав)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'clinic_admin') THEN
        CREATE ROLE clinic_admin NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'clinic_registrar') THEN
        CREATE ROLE clinic_registrar NOLOGIN;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'clinic_doctor') THEN
        CREATE ROLE clinic_doctor NOLOGIN;
    END IF;
END
$$;

-- Учебные логины (пароль: ClinicDemo1!)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'u_admin') THEN
        CREATE ROLE u_admin LOGIN PASSWORD 'ClinicDemo1!' IN ROLE clinic_admin;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'u_registrar') THEN
        CREATE ROLE u_registrar LOGIN PASSWORD 'ClinicDemo1!' IN ROLE clinic_registrar;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'u_doctor') THEN
        CREATE ROLE u_doctor LOGIN PASSWORD 'ClinicDemo1!' IN ROLE clinic_doctor;
    END IF;
END
$$;

-- Забрать лишнее у PUBLIC (если есть)
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO clinic_admin, clinic_registrar, clinic_doctor;

-- ---------------------------------------------------------------------------
-- clinic_admin — полный доступ к объектам схемы
-- ---------------------------------------------------------------------------
GRANT ALL ON ALL TABLES IN SCHEMA public TO clinic_admin;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO clinic_admin;
GRANT ALL ON ALL FUNCTIONS IN SCHEMA public TO clinic_admin;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL ON TABLES TO clinic_admin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL ON SEQUENCES TO clinic_admin;

-- ---------------------------------------------------------------------------
-- clinic_registrar — картотека и расписание
-- ---------------------------------------------------------------------------
GRANT SELECT ON staff TO clinic_registrar;
GRANT SELECT, INSERT, UPDATE ON patients TO clinic_registrar;
GRANT SELECT, INSERT, UPDATE ON appointments TO clinic_registrar;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO clinic_registrar;

GRANT SELECT ON v_free_slots_by_specialty TO clinic_registrar;
GRANT SELECT ON v_doctor_workload TO clinic_registrar;

-- ---------------------------------------------------------------------------
-- clinic_doctor — чтение картотеки/слотов; запись визитов и назначений
-- ---------------------------------------------------------------------------
GRANT SELECT ON patients TO clinic_doctor;
GRANT SELECT ON staff TO clinic_doctor;
GRANT SELECT ON appointments TO clinic_doctor;
GRANT SELECT, INSERT, UPDATE ON visits TO clinic_doctor;
GRANT SELECT, INSERT, UPDATE ON visit_orders TO clinic_doctor;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO clinic_doctor;

GRANT SELECT ON v_free_slots_by_specialty TO clinic_doctor;
GRANT SELECT ON v_patient_medical_card TO clinic_doctor;
GRANT SELECT ON v_doctor_workload TO clinic_doctor;
GRANT SELECT ON v_preferential_orders TO clinic_doctor;

COMMIT;

-- Подключение к БД (выполнить отдельно, если ещё не выдано):
-- GRANT CONNECT ON DATABASE clinic_db TO u_admin, u_registrar, u_doctor;
