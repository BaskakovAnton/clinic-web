--
-- PostgreSQL database dump
--

\restrict TXgqrE0UtyrW1k5SQkeCiYLQyA1tMEvLrueKFYWGdVnIlh1SviIt9ARbzGiRWdJ

-- Dumped from database version 16.10
-- Dumped by pg_dump version 16.10

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

ALTER TABLE IF EXISTS ONLY public.visits DROP CONSTRAINT IF EXISTS visits_patient_id_fkey;
ALTER TABLE IF EXISTS ONLY public.visits DROP CONSTRAINT IF EXISTS visits_doctor_id_fkey;
ALTER TABLE IF EXISTS ONLY public.visits DROP CONSTRAINT IF EXISTS visits_appointment_id_fkey;
ALTER TABLE IF EXISTS ONLY public.visit_orders DROP CONSTRAINT IF EXISTS visit_orders_visit_id_fkey;
ALTER TABLE IF EXISTS ONLY public.appointments DROP CONSTRAINT IF EXISTS appointments_patient_id_fkey;
ALTER TABLE IF EXISTS ONLY public.appointments DROP CONSTRAINT IF EXISTS appointments_doctor_id_fkey;
DROP TRIGGER IF EXISTS trg_visits_doctor_only ON public.visits;
DROP TRIGGER IF EXISTS trg_appointments_doctor_only ON public.appointments;
DROP INDEX IF EXISTS public.visits_patient_idx;
DROP INDEX IF EXISTS public.visits_doctor_idx;
DROP INDEX IF EXISTS public.visit_orders_visit_idx;
DROP INDEX IF EXISTS public.visit_orders_preferential_idx;
DROP INDEX IF EXISTS public.appointments_status_idx;
DROP INDEX IF EXISTS public.appointments_start_at_idx;
ALTER TABLE IF EXISTS ONLY public.visits DROP CONSTRAINT IF EXISTS visits_pkey;
ALTER TABLE IF EXISTS ONLY public.visits DROP CONSTRAINT IF EXISTS visits_appointment_id_key;
ALTER TABLE IF EXISTS ONLY public.visit_orders DROP CONSTRAINT IF EXISTS visit_orders_pkey;
ALTER TABLE IF EXISTS ONLY public.staff DROP CONSTRAINT IF EXISTS staff_pkey;
ALTER TABLE IF EXISTS ONLY public.patients DROP CONSTRAINT IF EXISTS patients_pkey;
ALTER TABLE IF EXISTS ONLY public.patients DROP CONSTRAINT IF EXISTS patients_card_number_uq;
ALTER TABLE IF EXISTS ONLY public.appointments DROP CONSTRAINT IF EXISTS appointments_pkey;
ALTER TABLE IF EXISTS ONLY public.appointments DROP CONSTRAINT IF EXISTS appointments_doctor_time_uq;
DROP VIEW IF EXISTS public.v_preferential_orders;
DROP VIEW IF EXISTS public.v_patient_medical_card;
DROP TABLE IF EXISTS public.visit_orders;
DROP VIEW IF EXISTS public.v_free_slots_by_specialty;
DROP VIEW IF EXISTS public.v_doctor_workload;
DROP TABLE IF EXISTS public.visits;
DROP TABLE IF EXISTS public.staff;
DROP TABLE IF EXISTS public.patients;
DROP TABLE IF EXISTS public.appointments;
DROP FUNCTION IF EXISTS public.visits_doctor_only();
DROP FUNCTION IF EXISTS public.appointments_doctor_only();
--
-- Name: appointments_doctor_only(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.appointments_doctor_only() RETURNS trigger
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


ALTER FUNCTION public.appointments_doctor_only() OWNER TO postgres;

--
-- Name: visits_doctor_only(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.visits_doctor_only() RETURNS trigger
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


ALTER FUNCTION public.visits_doctor_only() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: appointments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.appointments (
    id integer NOT NULL,
    doctor_id integer NOT NULL,
    start_at timestamp without time zone NOT NULL,
    status character varying(16) NOT NULL,
    patient_id integer,
    CONSTRAINT appointments_status_chk CHECK (((status)::text = ANY ((ARRAY['free'::character varying, 'booked'::character varying])::text[]))),
    CONSTRAINT appointments_status_patient_chk CHECK (((((status)::text = 'free'::text) AND (patient_id IS NULL)) OR (((status)::text = 'booked'::text) AND (patient_id IS NOT NULL))))
);


ALTER TABLE public.appointments OWNER TO postgres;

--
-- Name: TABLE appointments; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.appointments IS 'Слоты приёма: свободен / занят';


--
-- Name: appointments_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.appointments ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.appointments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: patients; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.patients (
    id integer NOT NULL,
    card_number character varying(32) NOT NULL,
    full_name character varying(200) NOT NULL,
    birth_date date NOT NULL,
    insurance_type character varying(8) NOT NULL,
    insurance_number character varying(64) NOT NULL,
    passport_data character varying(200) NOT NULL,
    address character varying(300),
    contacts character varying(200),
    CONSTRAINT patients_insurance_type_chk CHECK (((insurance_type)::text = ANY ((ARRAY['OMS'::character varying, 'DMS'::character varying])::text[])))
);


ALTER TABLE public.patients OWNER TO postgres;

--
-- Name: TABLE patients; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.patients IS 'Картотека пациентов (амбулаторные карты)';


--
-- Name: patients_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.patients ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.patients_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: staff; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.staff (
    id integer NOT NULL,
    full_name character varying(200) NOT NULL,
    staff_kind character varying(16) NOT NULL,
    specialty character varying(100),
    department character varying(100) NOT NULL,
    work_schedule character varying(200),
    office character varying(32),
    CONSTRAINT staff_doctor_specialty_chk CHECK (((((staff_kind)::text = 'doctor'::text) AND (specialty IS NOT NULL)) OR ((staff_kind)::text = 'nurse'::text))),
    CONSTRAINT staff_kind_chk CHECK (((staff_kind)::text = ANY ((ARRAY['doctor'::character varying, 'nurse'::character varying])::text[])))
);


ALTER TABLE public.staff OWNER TO postgres;

--
-- Name: TABLE staff; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.staff IS 'Сотрудники: врачи и медсёстры';


--
-- Name: staff_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.staff ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.staff_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: visits; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.visits (
    id integer NOT NULL,
    patient_id integer NOT NULL,
    doctor_id integer NOT NULL,
    visit_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    complaints text,
    diagnosis_icd10 character varying(16) NOT NULL,
    appointment_id integer
);


ALTER TABLE public.visits OWNER TO postgres;

--
-- Name: TABLE visits; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.visits IS 'Посещения: жалобы, диагноз МКБ-10; связь с записью необязательна';


--
-- Name: v_doctor_workload; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.v_doctor_workload AS
 SELECT s.id AS doctor_id,
    s.full_name AS doctor_name,
    s.specialty,
    s.department,
    count(v.id) AS patients_seen
   FROM (public.staff s
     LEFT JOIN public.visits v ON ((v.doctor_id = s.id)))
  WHERE ((s.staff_kind)::text = 'doctor'::text)
  GROUP BY s.id, s.full_name, s.specialty, s.department;


ALTER VIEW public.v_doctor_workload OWNER TO postgres;

--
-- Name: VIEW v_doctor_workload; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON VIEW public.v_doctor_workload IS 'Расчёт нагрузки на врачей (число принятых пациентов)';


--
-- Name: v_free_slots_by_specialty; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.v_free_slots_by_specialty AS
 SELECT a.id AS appointment_id,
    s.full_name AS doctor_name,
    s.specialty,
    s.department,
    s.office,
    a.start_at
   FROM (public.appointments a
     JOIN public.staff s ON ((s.id = a.doctor_id)))
  WHERE (((a.status)::text = 'free'::text) AND ((s.staff_kind)::text = 'doctor'::text));


ALTER VIEW public.v_free_slots_by_specialty OWNER TO postgres;

--
-- Name: VIEW v_free_slots_by_specialty; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON VIEW public.v_free_slots_by_specialty IS 'Поиск свободных слотов к нужному специалисту';


--
-- Name: visit_orders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.visit_orders (
    id integer NOT NULL,
    visit_id integer NOT NULL,
    order_kind character varying(16) NOT NULL,
    description character varying(500) NOT NULL,
    is_prescription boolean DEFAULT false NOT NULL,
    is_preferential boolean DEFAULT false NOT NULL,
    CONSTRAINT visit_orders_kind_chk CHECK (((order_kind)::text = ANY ((ARRAY['medication'::character varying, 'procedure'::character varying, 'test'::character varying])::text[])))
);


ALTER TABLE public.visit_orders OWNER TO postgres;

--
-- Name: TABLE visit_orders; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.visit_orders IS 'Назначения: лекарства / процедуры / анализы; рецепт и льгота';


--
-- Name: v_patient_medical_card; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.v_patient_medical_card AS
 SELECT p.id AS patient_id,
    p.card_number,
    p.full_name AS patient_name,
    p.birth_date,
    p.insurance_type,
    p.insurance_number,
    v.id AS visit_id,
    v.visit_at,
    d.full_name AS doctor_name,
    d.specialty AS doctor_specialty,
    v.complaints,
    v.diagnosis_icd10,
    v.appointment_id,
    o.id AS order_id,
    o.order_kind,
    o.description AS order_description,
    o.is_prescription,
    o.is_preferential
   FROM (((public.patients p
     LEFT JOIN public.visits v ON ((v.patient_id = p.id)))
     LEFT JOIN public.staff d ON ((d.id = v.doctor_id)))
     LEFT JOIN public.visit_orders o ON ((o.visit_id = v.id)));


ALTER VIEW public.v_patient_medical_card OWNER TO postgres;

--
-- Name: VIEW v_patient_medical_card; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON VIEW public.v_patient_medical_card IS 'Выдача истории болезни пациента (электронная карта)';


--
-- Name: v_preferential_orders; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.v_preferential_orders AS
 SELECT o.id AS order_id,
    v.visit_at,
    p.card_number,
    p.full_name AS patient_name,
    d.full_name AS doctor_name,
    o.order_kind,
    o.description,
    o.is_prescription,
    o.is_preferential
   FROM (((public.visit_orders o
     JOIN public.visits v ON ((v.id = o.visit_id)))
     JOIN public.patients p ON ((p.id = v.patient_id)))
     JOIN public.staff d ON ((d.id = v.doctor_id)))
  WHERE (o.is_preferential = true);


ALTER VIEW public.v_preferential_orders OWNER TO postgres;

--
-- Name: VIEW v_preferential_orders; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON VIEW public.v_preferential_orders IS 'Учёт льготных рецептов и медикаментов';


--
-- Name: visit_orders_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.visit_orders ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.visit_orders_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: visits_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

ALTER TABLE public.visits ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.visits_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Data for Name: appointments; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.appointments (id, doctor_id, start_at, status, patient_id) FROM stdin;
1	1	2026-09-10 09:00:00	free	\N
2	1	2026-09-10 09:30:00	booked	1
3	1	2026-09-10 10:00:00	free	\N
4	2	2026-09-10 11:00:00	free	\N
5	2	2026-09-10 11:30:00	booked	2
6	3	2026-09-11 09:00:00	free	\N
7	3	2026-09-11 09:30:00	free	\N
\.


--
-- Data for Name: patients; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.patients (id, card_number, full_name, birth_date, insurance_type, insurance_number, passport_data, address, contacts) FROM stdin;
1	A-1001	Смирнов Алексей Николаевич	1985-03-12	OMS	7700 123456	4500 123456, выдан ОВД	г. Москва, ул. Ленина, 1	+7-900-111-22-33
2	A-1002	Кузнецова Ольга Игоревна	1992-07-21	DMS	ДМС-998877	4501 654321, выдан ОВД	г. Москва, ул. Мира, 5	+7-900-222-33-44
3	A-1003	Волков Игорь Петрович	1978-11-03	OMS	7700 654321	4502 111222, выдан ОВД	г. Москва, пр. Мира, 10	+7-900-333-44-55
\.


--
-- Data for Name: staff; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.staff (id, full_name, staff_kind, specialty, department, work_schedule, office) FROM stdin;
1	Иванова Анна Сергеевна	doctor	терапевт	Терапевтическое	Пн–Пт 09:00–15:00	101
2	Петров Дмитрий Игоревич	doctor	хирург	Хирургическое	Пн–Пт 10:00–16:00	205
3	Сидорова Елена Викторовна	doctor	окулист	Офтальмология	Вт, Чт 09:00–14:00	310
4	Козлова Мария Павловна	nurse	\N	Терапевтическое	Пн–Пт 08:00–16:00	101
\.


--
-- Data for Name: visit_orders; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.visit_orders (id, visit_id, order_kind, description, is_prescription, is_preferential) FROM stdin;
1	1	medication	Ибупрофен 200 мг при головной боли	t	f
2	1	test	Общий анализ крови	f	f
3	2	medication	Урсодезоксихолевая кислота 250 мг	t	t
4	2	procedure	УЗИ органов брюшной полости	f	f
5	2	medication	Спазмолитик по схеме (льготный)	t	t
\.


--
-- Data for Name: visits; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.visits (id, patient_id, doctor_id, visit_at, complaints, diagnosis_icd10, appointment_id) FROM stdin;
1	1	1	2026-09-10 09:35:00	Головная боль, слабость	G44.2	2
2	3	2	2026-09-09 12:00:00	Боль в правом боку	K80.2	\N
\.


--
-- Name: appointments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.appointments_id_seq', 7, true);


--
-- Name: patients_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.patients_id_seq', 3, true);


--
-- Name: staff_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.staff_id_seq', 4, true);


--
-- Name: visit_orders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.visit_orders_id_seq', 5, true);


--
-- Name: visits_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.visits_id_seq', 2, true);


--
-- Name: appointments appointments_doctor_time_uq; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.appointments
    ADD CONSTRAINT appointments_doctor_time_uq UNIQUE (doctor_id, start_at);


--
-- Name: appointments appointments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.appointments
    ADD CONSTRAINT appointments_pkey PRIMARY KEY (id);


--
-- Name: patients patients_card_number_uq; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT patients_card_number_uq UNIQUE (card_number);


--
-- Name: patients patients_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.patients
    ADD CONSTRAINT patients_pkey PRIMARY KEY (id);


--
-- Name: staff staff_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.staff
    ADD CONSTRAINT staff_pkey PRIMARY KEY (id);


--
-- Name: visit_orders visit_orders_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.visit_orders
    ADD CONSTRAINT visit_orders_pkey PRIMARY KEY (id);


--
-- Name: visits visits_appointment_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.visits
    ADD CONSTRAINT visits_appointment_id_key UNIQUE (appointment_id);


--
-- Name: visits visits_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.visits
    ADD CONSTRAINT visits_pkey PRIMARY KEY (id);


--
-- Name: appointments_start_at_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX appointments_start_at_idx ON public.appointments USING btree (start_at);


--
-- Name: appointments_status_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX appointments_status_idx ON public.appointments USING btree (status);


--
-- Name: visit_orders_preferential_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX visit_orders_preferential_idx ON public.visit_orders USING btree (is_preferential) WHERE (is_preferential = true);


--
-- Name: visit_orders_visit_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX visit_orders_visit_idx ON public.visit_orders USING btree (visit_id);


--
-- Name: visits_doctor_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX visits_doctor_idx ON public.visits USING btree (doctor_id);


--
-- Name: visits_patient_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX visits_patient_idx ON public.visits USING btree (patient_id);


--
-- Name: appointments trg_appointments_doctor_only; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_appointments_doctor_only BEFORE INSERT OR UPDATE OF doctor_id ON public.appointments FOR EACH ROW EXECUTE FUNCTION public.appointments_doctor_only();


--
-- Name: visits trg_visits_doctor_only; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_visits_doctor_only BEFORE INSERT OR UPDATE OF doctor_id ON public.visits FOR EACH ROW EXECUTE FUNCTION public.visits_doctor_only();


--
-- Name: appointments appointments_doctor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.appointments
    ADD CONSTRAINT appointments_doctor_id_fkey FOREIGN KEY (doctor_id) REFERENCES public.staff(id) ON DELETE RESTRICT;


--
-- Name: appointments appointments_patient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.appointments
    ADD CONSTRAINT appointments_patient_id_fkey FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE RESTRICT;


--
-- Name: visit_orders visit_orders_visit_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.visit_orders
    ADD CONSTRAINT visit_orders_visit_id_fkey FOREIGN KEY (visit_id) REFERENCES public.visits(id) ON DELETE CASCADE;


--
-- Name: visits visits_appointment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.visits
    ADD CONSTRAINT visits_appointment_id_fkey FOREIGN KEY (appointment_id) REFERENCES public.appointments(id) ON DELETE SET NULL;


--
-- Name: visits visits_doctor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.visits
    ADD CONSTRAINT visits_doctor_id_fkey FOREIGN KEY (doctor_id) REFERENCES public.staff(id) ON DELETE RESTRICT;


--
-- Name: visits visits_patient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.visits
    ADD CONSTRAINT visits_patient_id_fkey FOREIGN KEY (patient_id) REFERENCES public.patients(id) ON DELETE RESTRICT;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: pg_database_owner
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO clinic_admin;
GRANT USAGE ON SCHEMA public TO clinic_registrar;
GRANT USAGE ON SCHEMA public TO clinic_doctor;


--
-- Name: FUNCTION appointments_doctor_only(); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON FUNCTION public.appointments_doctor_only() TO clinic_admin;


--
-- Name: FUNCTION visits_doctor_only(); Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON FUNCTION public.visits_doctor_only() TO clinic_admin;


--
-- Name: TABLE appointments; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.appointments TO clinic_admin;
GRANT SELECT,INSERT,UPDATE ON TABLE public.appointments TO clinic_registrar;
GRANT SELECT ON TABLE public.appointments TO clinic_doctor;


--
-- Name: SEQUENCE appointments_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.appointments_id_seq TO clinic_admin;
GRANT SELECT,USAGE ON SEQUENCE public.appointments_id_seq TO clinic_registrar;
GRANT SELECT,USAGE ON SEQUENCE public.appointments_id_seq TO clinic_doctor;


--
-- Name: TABLE patients; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.patients TO clinic_admin;
GRANT SELECT,INSERT,UPDATE ON TABLE public.patients TO clinic_registrar;
GRANT SELECT ON TABLE public.patients TO clinic_doctor;


--
-- Name: SEQUENCE patients_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.patients_id_seq TO clinic_admin;
GRANT SELECT,USAGE ON SEQUENCE public.patients_id_seq TO clinic_registrar;
GRANT SELECT,USAGE ON SEQUENCE public.patients_id_seq TO clinic_doctor;


--
-- Name: TABLE staff; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.staff TO clinic_admin;
GRANT SELECT ON TABLE public.staff TO clinic_registrar;
GRANT SELECT ON TABLE public.staff TO clinic_doctor;


--
-- Name: SEQUENCE staff_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.staff_id_seq TO clinic_admin;
GRANT SELECT,USAGE ON SEQUENCE public.staff_id_seq TO clinic_registrar;
GRANT SELECT,USAGE ON SEQUENCE public.staff_id_seq TO clinic_doctor;


--
-- Name: TABLE visits; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.visits TO clinic_admin;
GRANT SELECT,INSERT,UPDATE ON TABLE public.visits TO clinic_doctor;


--
-- Name: TABLE v_doctor_workload; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.v_doctor_workload TO clinic_admin;
GRANT SELECT ON TABLE public.v_doctor_workload TO clinic_registrar;
GRANT SELECT ON TABLE public.v_doctor_workload TO clinic_doctor;


--
-- Name: TABLE v_free_slots_by_specialty; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.v_free_slots_by_specialty TO clinic_admin;
GRANT SELECT ON TABLE public.v_free_slots_by_specialty TO clinic_registrar;
GRANT SELECT ON TABLE public.v_free_slots_by_specialty TO clinic_doctor;


--
-- Name: TABLE visit_orders; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.visit_orders TO clinic_admin;
GRANT SELECT,INSERT,UPDATE ON TABLE public.visit_orders TO clinic_doctor;


--
-- Name: TABLE v_patient_medical_card; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.v_patient_medical_card TO clinic_admin;
GRANT SELECT ON TABLE public.v_patient_medical_card TO clinic_doctor;


--
-- Name: TABLE v_preferential_orders; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON TABLE public.v_preferential_orders TO clinic_admin;
GRANT SELECT ON TABLE public.v_preferential_orders TO clinic_doctor;


--
-- Name: SEQUENCE visit_orders_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.visit_orders_id_seq TO clinic_admin;
GRANT SELECT,USAGE ON SEQUENCE public.visit_orders_id_seq TO clinic_registrar;
GRANT SELECT,USAGE ON SEQUENCE public.visit_orders_id_seq TO clinic_doctor;


--
-- Name: SEQUENCE visits_id_seq; Type: ACL; Schema: public; Owner: postgres
--

GRANT ALL ON SEQUENCE public.visits_id_seq TO clinic_admin;
GRANT SELECT,USAGE ON SEQUENCE public.visits_id_seq TO clinic_registrar;
GRANT SELECT,USAGE ON SEQUENCE public.visits_id_seq TO clinic_doctor;


--
-- Name: DEFAULT PRIVILEGES FOR SEQUENCES; Type: DEFAULT ACL; Schema: public; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON SEQUENCES TO clinic_admin;


--
-- Name: DEFAULT PRIVILEGES FOR TABLES; Type: DEFAULT ACL; Schema: public; Owner: postgres
--

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT ALL ON TABLES TO clinic_admin;


--
-- PostgreSQL database dump complete
--

\unrestrict TXgqrE0UtyrW1k5SQkeCiYLQyA1tMEvLrueKFYWGdVnIlh1SviIt9ARbzGiRWdJ

