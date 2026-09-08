-- Amvera bootstrap after sborka 01–04
-- DB name on Amvera is clinicdb (not clinic_db).

GRANT CONNECT ON DATABASE clinicdb TO u_admin, u_registrar, u_doctor;
GRANT CONNECT ON DATABASE clinicdb TO clinicapp;

GRANT USAGE ON SCHEMA public TO clinicapp;
GRANT ALL ON ALL TABLES IN SCHEMA public TO clinicapp;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO clinicapp;
GRANT ALL ON ALL FUNCTIONS IN SCHEMA public TO clinicapp;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO clinicapp;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO clinicapp;

-- Teaching passwords (idempotent reset)
ALTER ROLE u_admin WITH LOGIN PASSWORD 'ClinicDemo1!';
ALTER ROLE u_registrar WITH LOGIN PASSWORD 'ClinicDemo1!';
ALTER ROLE u_doctor WITH LOGIN PASSWORD 'ClinicDemo1!';
