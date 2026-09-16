-- clinic_db · 07_public_site.sql
-- Публичный сайт: каталог услуг и заявки на запись
-- Идемпотентный seed: ON CONFLICT (slug) DO UPDATE

BEGIN;

-- Needed on PostgreSQL 15+ (public CREATE revoked from non-owners).
-- Harmless if already granted; requires a role that can grant on schema public.
GRANT USAGE, CREATE ON SCHEMA public TO clinic_admin;

CREATE TABLE IF NOT EXISTS public_services (
  id SERIAL PRIMARY KEY,
  slug TEXT UNIQUE NOT NULL,
  title TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  body TEXT NOT NULL DEFAULT '',
  price_from INT,
  image_url TEXT NOT NULL DEFAULT '/static/img/service-placeholder.jpg',
  sort_order INT NOT NULL DEFAULT 0,
  is_featured BOOLEAN NOT NULL DEFAULT false,
  is_published BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS appointment_requests (
  id SERIAL PRIMARY KEY,
  full_name TEXT NOT NULL,
  phone TEXT NOT NULL,
  service_slug TEXT,
  preferred_date DATE,
  doctor_note TEXT,
  comment TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new','called','done','cancelled')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO public_services (
  slug, title, summary, body, price_from, image_url, sort_order, is_featured, is_published
) VALUES
  (
    'therapy',
    'Терапия — первичный приём',
    'Осмотр терапевта, сбор анамнеза и план обследования.',
    'Первичный приём врача-терапевта: жалобы, осмотр, направление на анализы и консультации узких специалистов.',
    2500,
    '/static/img/service-placeholder.jpg',
    10,
    false,
    true
  ),
  (
    'checkup',
    'Комплексный Check-up',
    'Комплексные программы обследования за один день.',
    'Программы check-up с лабораторной и инструментальной диагностикой, заключением терапевта и рекомендациями.',
    8900,
    '/static/img/service-checkup.jpg',
    20,
    true,
    true
  ),
  (
    'cardio',
    'Кардиология',
    'Диагностика сердца и сосудов, холтер, СМАД.',
    'Приём кардиолога, ЭКГ, холтеровское мониторирование, СМАД и подбор терапии.',
    3200,
    '/static/img/service-cardio.jpg',
    30,
    true,
    true
  ),
  (
    'neuro',
    'Неврология',
    'Головная боль, головокружение, боли в спине и шее.',
    'Консультация невролога, неврологический осмотр и направление на МРТ/КТ при необходимости.',
    3000,
    '/static/img/service-neuro.jpg',
    40,
    false,
    true
  ),
  (
    'endo',
    'Эндокринология',
    'Щитовидная железа, сахарный диабет, гормональный фон.',
    'Приём эндокринолога, интерпретация анализов и коррекция терапии.',
    3100,
    '/static/img/service-placeholder.jpg',
    50,
    false,
    true
  ),
  (
    'gastro',
    'Гастроэнтерология',
    'Приём, ФГДС, колоноскопия.',
    'Диагностика и лечение заболеваний ЖКТ, подготовка к эндоскопии и наблюдение после процедур.',
    3300,
    '/static/img/service-gastro.jpg',
    60,
    true,
    true
  ),
  (
    'gynecology',
    'Акушерство и гинекология',
    'Плановый приём, УЗИ и ведение беременности.',
    'Консультация гинеколога, диагностика и индивидуальный план наблюдения.',
    3400,
    '/static/img/service-placeholder.jpg',
    70,
    false,
    true
  ),
  (
    'urology',
    'Урология',
    'Консультация уролога и диагностика.',
    'Приём уролога, УЗИ мочевыводящих путей и подбор лечения.',
    3200,
    '/static/img/service-placeholder.jpg',
    80,
    false,
    true
  ),
  (
    'lor',
    'Оториноларингология (ЛОР)',
    'Заболевания уха, горла и носа.',
    'Осмотр ЛОР-врача, эндоскопия и лечение острых и хронических состояний.',
    2900,
    '/static/img/service-placeholder.jpg',
    90,
    false,
    true
  ),
  (
    'ophthalmology',
    'Офтальмология',
    'Проверка зрения и заболевания глаз.',
    'Приём офтальмолога, визометрия и рекомендации по коррекции и лечению.',
    2800,
    '/static/img/service-placeholder.jpg',
    100,
    false,
    true
  ),
  (
    'trauma',
    'Травматология и ортопедия',
    'Суставы, позвоночник, травмы опорно-двигательного аппарата.',
    'Консультация травматолога-ортопеда, план обследования и реабилитации.',
    3500,
    '/static/img/service-placeholder.jpg',
    110,
    false,
    true
  ),
  (
    'allergy',
    'Аллергология и иммунология',
    'Аллергии, астма, иммунный статус.',
    'Приём аллерголога-иммунолога, подбор обследования и терапии.',
    3100,
    '/static/img/service-placeholder.jpg',
    120,
    false,
    true
  ),
  (
    'phleb',
    'Флебология',
    'Варикоз, сосуды нижних конечностей.',
    'Диагностика и лечение заболеваний вен, УЗДС и малоинвазивные методы.',
    3600,
    '/static/img/service-phleb.jpg',
    130,
    false,
    true
  ),
  (
    'uzi',
    'Ультразвуковая диагностика (УЗИ)',
    'Экспертный класс, все основные зоны.',
    'УЗИ органов брюшной полости, малого таза, щитовидной железы, сосудов и мягких тканей.',
    2200,
    '/static/img/service-uzi.jpg',
    140,
    true,
    true
  ),
  (
    'ct',
    'Компьютерная томография (КТ)',
    'КТ-исследования основных зон.',
    'Компьютерная томография с описанием врача-рентгенолога и выдачей заключения.',
    5500,
    '/static/img/service-placeholder.jpg',
    150,
    false,
    true
  ),
  (
    'mri',
    'Магнитно-резонансная томография (МРТ)',
    'МРТ головного мозга, позвоночника и суставов.',
    'МРТ-диагностика с экспертным описанием и рекомендациями по дальнейшему обследованию.',
    6900,
    '/static/img/service-mri.jpg',
    160,
    false,
    true
  ),
  (
    'lab',
    'Лабораторная диагностика',
    'Анализы крови, мочи и биохимия.',
    'Широкий спектр лабораторных исследований с быстрой выдачей результатов.',
    900,
    '/static/img/service-lab.jpg',
    170,
    false,
    true
  ),
  (
    'endoscopy',
    'Эндоскопия (ФГДС / колоноскопия)',
    'Диагностическая и лечебная эндоскопия.',
    'ФГДС и колоноскопия под местной или медикаментозной седацией, с биопсией при необходимости.',
    7500,
    '/static/img/service-gastro.jpg',
    180,
    false,
    true
  ),
  (
    'day-hospital',
    'Дневной стационар / IV-терапия',
    'Инфузионная терапия и наблюдение в дневном стационаре.',
    'Капельницы, послеоперационное наблюдение и курсовое лечение в комфортных условиях.',
    4500,
    '/static/img/service-placeholder.jpg',
    190,
    false,
    true
  ),
  (
    'cosmetology',
    'Косметология',
    'Эстетические и аппаратные процедуры.',
    'Консультация косметолога и подбор процедур по состоянию кожи.',
    4000,
    '/static/img/service-placeholder.jpg',
    200,
    false,
    true
  ),
  (
    'dental',
    'Стоматология терапевтическая',
    'Лечение кариеса, каналов и профессиональная гигиена.',
    'Терапевтический приём стоматолога, лечение и профилактика заболеваний зубов и дёсен.',
    2700,
    '/static/img/service-dental.jpg',
    210,
    true,
    true
  ),
  (
    'dental-aesthetic',
    'Эстетическая стоматология (виниры, отбеливание)',
    'Виниры, отбеливание и эстетическая коррекция улыбки.',
    'Индивидуальный план эстетического лечения: отбеливание, виниры и реставрации.',
    12000,
    '/static/img/service-dental.jpg',
    220,
    false,
    true
  ),
  (
    'dental-implant',
    'Имплантация / протезирование зубов',
    'Имплантация и ортопедическое восстановление.',
    'Планирование имплантации, протезирование и сопровождение на всех этапах.',
    35000,
    '/static/img/service-dental.jpg',
    230,
    false,
    true
  ),
  (
    'massage',
    'Массаж и восстановительная медицина',
    'Реабилитация, массаж и восстановление после нагрузок.',
    'Курсы массажа и восстановительные программы по назначению врача.',
    2500,
    '/static/img/service-placeholder.jpg',
    240,
    false,
    true
  ),
  (
    'surgery-ambulatory',
    'Амбулаторная хирургия',
    'Операции одного дня, быстрое восстановление.',
    'Малоинвазивные амбулаторные вмешательства с наблюдением в дневном стационаре.',
    15000,
    '/static/img/service-surgery.jpg',
    250,
    true,
    true
  ),
  (
    'surgery-abdominal',
    'Абдоминальная хирургия',
    'Плановые операции на органах брюшной полости.',
    'Консультация хирурга, предоперационная подготовка и послеоперационное наблюдение.',
    45000,
    '/static/img/service-surgery.jpg',
    260,
    false,
    true
  ),
  (
    'surgery-gynecology',
    'Оперативная гинекология',
    'Малоинвазивные гинекологические операции.',
    'Плановые оперативные вмешательства с подготовкой и сопровождением.',
    40000,
    '/static/img/service-surgery.jpg',
    270,
    false,
    true
  ),
  (
    'surgery-urology',
    'Оперативная урология',
    'Плановые урологические вмешательства.',
    'Консультация оперирующего уролога и план хирургического лечения.',
    42000,
    '/static/img/service-surgery.jpg',
    280,
    false,
    true
  ),
  (
    'home-care',
    'Помощь на дому',
    'Выезд врача и медсестры на дом.',
    'Вызов специалиста на дом для осмотра, забора анализов и назначения лечения.',
    5000,
    '/static/img/service-placeholder.jpg',
    290,
    false,
    true
  ),
  (
    'telemedicine',
    'Телемедицина / онлайн-консультация',
    'Онлайн-консультация врача без визита в клинику.',
    'Дистанционный приём с разбором анализов и рекомендациями по дальнейшим шагам.',
    1800,
    '/static/img/service-placeholder.jpg',
    300,
    false,
    true
  )
ON CONFLICT (slug) DO UPDATE SET
  title = EXCLUDED.title,
  summary = EXCLUDED.summary,
  body = EXCLUDED.body,
  price_from = EXCLUDED.price_from,
  image_url = EXCLUDED.image_url,
  sort_order = EXCLUDED.sort_order,
  is_featured = EXCLUDED.is_featured,
  is_published = EXCLUDED.is_published;

GRANT ALL ON TABLE public_services, appointment_requests TO clinic_admin;
GRANT ALL ON SEQUENCE public_services_id_seq, appointment_requests_id_seq TO clinic_admin;

COMMIT;
