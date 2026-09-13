/**
 * Capture all clinic UI pages for client review.
 * Usage: node capture-screens.mjs
 */
import { chromium } from "playwright";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const OUT = path.join(__dirname, "screens");
const BASE = "http://127.0.0.1:8080";
const PASS = "ClinicDemo1!";

const pages = [
  { id: "01-login", title: "Вход", path: "/login", auth: false },
  { id: "02-admin-dashboard", title: "Админ · Дашборд", path: "/admin", role: "u_admin" },
  { id: "03-admin-staff", title: "Админ · Персонал", path: "/admin/staff", role: "u_admin" },
  { id: "04-admin-staff-new", title: "Админ · Новый сотрудник", path: "/admin/staff/new", role: "u_admin" },
  { id: "05-admin-staff-edit", title: "Админ · Редактирование сотрудника", path: "/admin/staff/__STAFF__/edit", role: "u_admin" },
  { id: "06-admin-roles", title: "Админ · Роли", path: "/admin/roles", role: "u_admin" },
  { id: "07-admin-mass-slots", title: "Админ · Массовые слоты", path: "/admin/mass-slots", role: "u_admin" },
  { id: "08-registrar-patients", title: "Регистратура · Пациенты", path: "/registrar/patients", role: "u_admin" },
  { id: "09-registrar-patient-new", title: "Регистратура · Новый пациент", path: "/registrar/patients/new", role: "u_admin" },
  { id: "10-registrar-patient", title: "Регистратура · Карточка пациента", path: "/registrar/patients/__PATIENT__", role: "u_admin" },
  { id: "11-registrar-patient-edit", title: "Регистратура · Редактирование пациента", path: "/registrar/patients/__PATIENT__/edit", role: "u_admin" },
  { id: "12-registrar-slots", title: "Регистратура · Свободные слоты", path: "/registrar/slots", role: "u_admin" },
  { id: "13-registrar-schedule", title: "Регистратура · Расписание", path: "/registrar/schedule?date=2026-09-10", role: "u_admin" },
  { id: "14-doctor-day", title: "Врач · Мой день", path: "/doctor/day?date=2026-09-10", role: "u_admin" },
  { id: "15-doctor-patients", title: "Врач · Карты пациентов", path: "/doctor/patients", role: "u_admin" },
  { id: "16-doctor-patient", title: "Врач · Карта пациента", path: "/doctor/patients/__PATIENT__", role: "u_admin" },
  { id: "17-doctor-schedule", title: "Врач · Моё расписание", path: "/doctor/schedule", role: "u_admin" },
  { id: "18-doctor-visit", title: "Врач · Визит", path: "/doctor/visits/__VISIT__", role: "u_admin" },
  { id: "19-reports-workload", title: "Отчёт · Нагрузка", path: "/reports/workload", role: "u_admin" },
  { id: "20-reports-preferential", title: "Отчёт · Льготные назначения", path: "/reports/preferential", role: "u_admin" },
  { id: "21-registrar-home", title: "Роль регистратура · Пациенты", path: "/registrar/patients", role: "u_registrar" },
  { id: "22-doctor-home", title: "Роль врач · Мой день", path: "/doctor/day?date=2026-09-10", role: "u_doctor" },
];

fs.mkdirSync(OUT, { recursive: true });

async function login(page, login) {
  await page.goto(`${BASE}/login`, { waitUntil: "networkidle" });
  await page.fill("#login", login);
  await page.fill("#password", PASS);
  await Promise.all([
    page.waitForURL((url) => !url.pathname.endsWith("/login"), { timeout: 15000 }),
    page.click('button[type="submit"]'),
  ]);
}

async function logout(page) {
  const btn = page.locator('form[action="/logout"] button');
  if (await btn.count()) {
    await Promise.all([
      page.waitForURL("**/login", { timeout: 10000 }).catch(() => {}),
      btn.first().click(),
    ]);
  } else {
    await page.goto(`${BASE}/login`, { waitUntil: "networkidle" });
  }
}

const browser = await chromium.launch({
  headless: true,
  executablePath: "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
});
const context = await browser.newContext({
  viewport: { width: 1440, height: 900 },
  deviceScaleFactor: 1,
});
const page = await context.newPage();

// resolve ids while logged in as admin
await login(page, "u_admin");
await page.goto(`${BASE}/admin/staff`, { waitUntil: "networkidle" });
const staffHref = await page.locator('a[href*="/admin/staff/"][href$="/edit"]').first().getAttribute("href");
const staffId = (staffHref || "/admin/staff/1/edit").match(/staff\/(\d+)/)?.[1] || "1";
await page.goto(`${BASE}/registrar/patients`, { waitUntil: "networkidle" });
const patientHref = await page.locator('a[href^="/registrar/patients/"]').first().getAttribute("href");
const patientId = (patientHref || "/registrar/patients/1").match(/patients\/(\d+)/)?.[1] || "1";
let visitId = "1";
await page.goto(`${BASE}/doctor/patients/${patientId}`, { waitUntil: "networkidle" });
if (await page.locator('a[href^="/doctor/visits/"]').count()) {
  const visitHref = await page.locator('a[href^="/doctor/visits/"]').first().getAttribute("href");
  visitId = visitHref?.match(/visits\/(\d+)/)?.[1] || "1";
} else {
  await page.goto(`${BASE}/doctor/day?date=2026-09-10`, { waitUntil: "networkidle" });
  if (await page.locator('a[href^="/doctor/visits/"]').count()) {
    const vh = await page.locator('a[href^="/doctor/visits/"]').first().getAttribute("href");
    visitId = vh?.match(/visits\/(\d+)/)?.[1] || "1";
  }
}

console.log({ staffId, patientId, visitId });

let currentRole = "u_admin";
const results = [];

for (const item of pages) {
  const urlPath = item.path
    .replace("__STAFF__", staffId)
    .replace("__PATIENT__", patientId)
    .replace("__VISIT__", visitId);

  try {
    if (item.auth === false) {
      if (currentRole) {
        await logout(page);
        currentRole = null;
      }
      await page.goto(`${BASE}${urlPath}`, { waitUntil: "networkidle" });
    } else {
      if (currentRole !== item.role) {
        if (currentRole) await logout(page);
        await login(page, item.role);
        currentRole = item.role;
      }
      await page.goto(`${BASE}${urlPath}`, { waitUntil: "networkidle" });
    }
    await page.waitForTimeout(400);
    const file = `${item.id}.png`;
    await page.screenshot({ path: path.join(OUT, file), fullPage: true });
    results.push({ ...item, path: urlPath, file, ok: true });
    console.log("OK", item.id, urlPath);
  } catch (err) {
    results.push({ ...item, path: urlPath, file: null, ok: false, error: String(err) });
    console.error("FAIL", item.id, err.message);
  }
}

fs.writeFileSync(path.join(OUT, "manifest.json"), JSON.stringify(results, null, 2), "utf8");
await browser.close();
console.log("done", results.filter((r) => r.ok).length, "/", results.length);
