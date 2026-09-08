package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"clinic/internal/auth"
	"clinic/internal/store"
)

func (s *Server) resolveDoctorID(w http.ResponseWriter, r *http.Request, doctors []store.Doctor) int {
	if v := r.URL.Query().Get("doctor_id"); v != "" {
		id, _ := strconv.Atoi(v)
		if id > 0 {
			setDoctorCookie(w, id)
			return id
		}
	}
	if id := doctorCookie(r); id > 0 {
		return id
	}
	if len(doctors) > 0 {
		setDoctorCookie(w, doctors[0].ID)
		return doctors[0].ID
	}
	return 0
}

func (s *Server) DoctorDay(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	doctors, err := s.Store.ListDoctors(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	doctorID := s.resolveDoctorID(w, r, doctors)
	day, err := parseDate(r.URL.Query().Get("date"))
	if err != nil {
		redirectErr(w, r, "/doctor/day", "Некорректная дата")
		return
	}
	// default seed day if today empty? keep today / query
	if r.URL.Query().Get("date") == "" {
		// prefer a day with data for demo: 2026-09-10 if exists
		apptsToday, _ := s.Store.ListAppointmentsDay(r.Context(), day, doctorID)
		if len(apptsToday) == 0 {
			demo, _ := time.ParseInLocation("2006-01-02", "2026-09-10", time.Local)
			if demoAppts, err := s.Store.ListAppointmentsDay(r.Context(), demo, doctorID); err == nil && len(demoAppts) > 0 {
				day = demo
			}
		}
	}
	appts, err := s.Store.ListAppointmentsDay(r.Context(), day, doctorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "doctor_day.html", pageData{
		Title: "Мой день", Active: "day", Session: sess,
		Appointments: appts, Doctors: doctors, DoctorID: doctorID,
		Date: day.Format("2006-01-02"),
		Flash: flashOK(r), FlashError: flashErr(r),
	})
}

func (s *Server) DoctorPatients(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	patients, err := s.Store.ListPatients(r.Context(), q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "doctor_patients.html", pageData{
		Title: "Пациенты", Active: "doc_patients", Session: sess,
		Patients: patients, Query: q,
	})
}

func (s *Server) DoctorSchedule(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	doctors, err := s.Store.ListDoctors(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	doctorID := s.resolveDoctorID(w, r, doctors)
	day, err := parseDate(r.URL.Query().Get("date"))
	if err != nil {
		redirectErr(w, r, "/doctor/schedule", "Некорректная дата")
		return
	}
	appts, err := s.Store.ListAppointmentsDay(r.Context(), day, doctorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "doctor_schedule.html", pageData{
		Title: "Моё расписание", Active: "doc_schedule", Session: sess,
		Appointments: appts, Doctors: doctors, DoctorID: doctorID,
		Date: day.Format("2006-01-02"),
	})
}

func (s *Server) DoctorPatient(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	patient, err := s.Store.GetPatient(r.Context(), id)
	if err == store.ErrNotFound {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	card, err := s.Store.PatientCard(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	doctors, err := s.Store.ListDoctors(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	doctorID := s.resolveDoctorID(w, r, doctors)
	form := map[string]string{"doctor_id": strconv.Itoa(doctorID)}
	if appt := r.URL.Query().Get("appointment_id"); appt != "" {
		form["appointment_id"] = appt
		if a, err := s.Store.GetAppointment(r.Context(), mustAtoi(appt)); err == nil {
			form["doctor_id"] = strconv.Itoa(a.DoctorID)
		}
	}
	s.render(w, "doctor_patient.html", pageData{
		Title: "Карта пациента", Active: "doc_patients", Session: sess,
		Patient: patient, CardRows: card, Doctors: doctors, DoctorID: doctorID,
		Form: form, Flash: flashOK(r), FlashError: flashErr(r),
	})
}

func mustAtoi(s string) int { n, _ := strconv.Atoi(s); return n }

func (s *Server) CreateVisit(w http.ResponseWriter, r *http.Request) {
	patientID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	redirect := fmt.Sprintf("/doctor/patients/%d", patientID)
	if err := r.ParseForm(); err != nil {
		redirectErr(w, r, redirect, "Ошибка формы")
		return
	}
	doctorID, err := strconv.Atoi(r.FormValue("doctor_id"))
	if err != nil {
		redirectErr(w, r, redirect, "Выберите врача")
		return
	}
	diagnosis := strings.TrimSpace(r.FormValue("diagnosis_icd10"))
	if diagnosis == "" {
		redirectErr(w, r, redirect, "Укажите код МКБ-10")
		return
	}
	var appt *int
	if v := strings.TrimSpace(r.FormValue("appointment_id")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			redirectErr(w, r, redirect, "Некорректный appointment_id")
			return
		}
		appt = &n
	}
	visitID, err := s.Store.CreateVisit(r.Context(), store.CreateVisitInput{
		PatientID: patientID, DoctorID: doctorID,
		Complaints: r.FormValue("complaints"), DiagnosisICD10: diagnosis,
		AppointmentID: appt, OrderKind: r.FormValue("order_kind"),
		OrderDesc: r.FormValue("order_description"),
		IsPrescription: r.FormValue("is_prescription") == "on",
		IsPreferential: r.FormValue("is_preferential") == "on",
	})
	if err != nil {
		redirectErr(w, r, redirect, err.Error())
		return
	}
	redirectOK(w, r, fmt.Sprintf("/doctor/visits/%d", visitID), "Визит сохранён")
}

func (s *Server) DoctorVisitGet(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	visit, err := s.Store.GetVisit(r.Context(), id)
	if err == store.ErrNotFound {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orders, err := s.Store.ListVisitOrders(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "doctor_visit.html", pageData{
		Title: "Визит", Active: "day", Session: sess,
		Visit: visit, Orders: orders,
		Flash: flashOK(r), FlashError: flashErr(r),
	})
}

func (s *Server) DoctorVisitUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	path := fmt.Sprintf("/doctor/visits/%d", id)
	_ = r.ParseForm()
	diagnosis := strings.TrimSpace(r.FormValue("diagnosis_icd10"))
	if diagnosis == "" {
		redirectErr(w, r, path, "Укажите МКБ-10")
		return
	}
	if err := s.Store.UpdateVisit(r.Context(), id, r.FormValue("complaints"), diagnosis); err != nil {
		redirectErr(w, r, path, err.Error())
		return
	}
	redirectOK(w, r, path, "Визит обновлён")
}

func (s *Server) DoctorOrderCreate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	path := fmt.Sprintf("/doctor/visits/%d", id)
	_ = r.ParseForm()
	kind := r.FormValue("order_kind")
	desc := strings.TrimSpace(r.FormValue("order_description"))
	if kind == "" || desc == "" {
		redirectErr(w, r, path, "Тип и описание обязательны")
		return
	}
	if err := s.Store.AddVisitOrder(r.Context(), id, kind, desc,
		r.FormValue("is_prescription") == "on", r.FormValue("is_preferential") == "on"); err != nil {
		redirectErr(w, r, path, err.Error())
		return
	}
	redirectOK(w, r, path, "Назначение добавлено")
}

func (s *Server) DoctorOrderUpdate(w http.ResponseWriter, r *http.Request) {
	visitID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	orderID, err := strconv.Atoi(r.PathValue("orderID"))
	if err != nil {
		http.Error(w, "bad order id", http.StatusBadRequest)
		return
	}
	path := fmt.Sprintf("/doctor/visits/%d", visitID)
	_ = r.ParseForm()
	kind := r.FormValue("order_kind")
	desc := strings.TrimSpace(r.FormValue("order_description"))
	if kind == "" || desc == "" {
		redirectErr(w, r, path, "Тип и описание обязательны")
		return
	}
	if err := s.Store.UpdateVisitOrder(r.Context(), orderID, kind, desc,
		r.FormValue("is_prescription") == "on", r.FormValue("is_preferential") == "on"); err != nil {
		redirectErr(w, r, path, err.Error())
		return
	}
	redirectOK(w, r, path, "Назначение обновлено")
}

func (s *Server) ReportsWorkload(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	rows, err := s.Store.Workload(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "reports_workload.html", pageData{
		Title: "Нагрузка врачей", Active: "workload", Session: sess, Workload: rows,
	})
}

func (s *Server) ReportsPreferential(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	rows, err := s.Store.Preferential(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "reports_preferential.html", pageData{
		Title: "Льготные назначения", Active: "preferential", Session: sess, Preferential: rows,
	})
}
