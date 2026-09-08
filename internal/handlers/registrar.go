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

func (s *Server) RegistrarPatients(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	patients, err := s.Store.ListPatients(r.Context(), q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "registrar_patients.html", pageData{
		Title: "Пациенты", Active: "patients", Session: sess,
		Patients: patients, Query: q, Flash: flashOK(r), FlashError: flashErr(r),
	})
}

func (s *Server) RegistrarPatientNewGet(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	s.render(w, "registrar_patient_form.html", pageData{
		Title: "Новый пациент", Active: "patients", Session: sess,
		Form: map[string]string{"insurance_type": "OMS"}, FlashError: flashErr(r),
	})
}

func (s *Server) RegistrarPatientCreate(w http.ResponseWriter, r *http.Request) {
	in, errMsg := patientFromForm(r)
	if errMsg != "" {
		redirectErr(w, r, "/registrar/patients/new", errMsg)
		return
	}
	id, err := s.Store.CreatePatient(r.Context(), in)
	if err != nil {
		redirectErr(w, r, "/registrar/patients/new", err.Error())
		return
	}
	redirectOK(w, r, fmt.Sprintf("/registrar/patients/%d", id), "Пациент создан")
}

func (s *Server) RegistrarPatientGet(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	p, err := s.Store.GetPatient(r.Context(), id)
	if err == store.ErrNotFound {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	appts, err := s.Store.ListPatientAppointments(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	free, err := s.Store.ListFreeSlots(r.Context(), "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "registrar_patient.html", pageData{
		Title: "Карточка пациента", Active: "patients", Session: sess,
		Patient: p, Appointments: appts, Slots: free,
		Flash: flashOK(r), FlashError: flashErr(r),
	})
}

func (s *Server) RegistrarPatientEditGet(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	p, err := s.Store.GetPatient(r.Context(), id)
	if err == store.ErrNotFound {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "registrar_patient_form.html", pageData{
		Title: "Редактирование пациента", Active: "patients", Session: sess,
		Patient: p, Form: patientToForm(p), FlashError: flashErr(r),
	})
}

func (s *Server) RegistrarPatientUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	in, errMsg := patientFromForm(r)
	path := fmt.Sprintf("/registrar/patients/%d/edit", id)
	if errMsg != "" {
		redirectErr(w, r, path, errMsg)
		return
	}
	if err := s.Store.UpdatePatient(r.Context(), id, in); err != nil {
		redirectErr(w, r, path, err.Error())
		return
	}
	redirectOK(w, r, fmt.Sprintf("/registrar/patients/%d", id), "Сохранено")
}

func patientFromForm(r *http.Request) (store.PatientInput, string) {
	if err := r.ParseForm(); err != nil {
		return store.PatientInput{}, "Ошибка формы"
	}
	bd, err := time.Parse("2006-01-02", r.FormValue("birth_date"))
	if err != nil {
		return store.PatientInput{}, "Некорректная дата рождения"
	}
	ins := r.FormValue("insurance_type")
	if ins != "OMS" && ins != "DMS" {
		return store.PatientInput{}, "Тип полиса: OMS или DMS"
	}
	in := store.PatientInput{
		CardNumber:      strings.TrimSpace(r.FormValue("card_number")),
		FullName:        strings.TrimSpace(r.FormValue("full_name")),
		BirthDate:       bd,
		InsuranceType:   ins,
		InsuranceNumber: strings.TrimSpace(r.FormValue("insurance_number")),
		PassportData:    strings.TrimSpace(r.FormValue("passport_data")),
		Address:         strings.TrimSpace(r.FormValue("address")),
		Contacts:        strings.TrimSpace(r.FormValue("contacts")),
	}
	if in.CardNumber == "" || in.FullName == "" || in.InsuranceNumber == "" || in.PassportData == "" {
		return in, "Заполните обязательные поля"
	}
	return in, ""
}

func patientToForm(p store.Patient) map[string]string {
	m := map[string]string{
		"card_number": p.CardNumber, "full_name": p.FullName,
		"birth_date": p.BirthDate.Format("2006-01-02"),
		"insurance_type": p.InsuranceType, "insurance_number": p.InsuranceNumber,
		"passport_data": p.PassportData,
	}
	if p.Address.Valid {
		m["address"] = p.Address.String
	}
	if p.Contacts.Valid {
		m["contacts"] = p.Contacts.String
	}
	return m
}

func (s *Server) RegistrarSlots(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	specialty := strings.TrimSpace(r.URL.Query().Get("specialty"))
	slots, err := s.Store.ListFreeSlots(r.Context(), specialty)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	specialties, err := s.Store.ListSpecialties(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	patients, err := s.Store.ListPatients(r.Context(), "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	doctors, err := s.Store.ListDoctors(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "registrar_slots.html", pageData{
		Title: "Свободные слоты", Active: "slots", Session: sess,
		Slots: slots, Specialties: specialties, Specialty: specialty,
		Patients: patients, Doctors: doctors,
		Flash: flashOK(r), FlashError: flashErr(r),
	})
}

func (s *Server) BookSlot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	_ = r.ParseForm()
	back := r.FormValue("return")
	if back == "" {
		back = "/registrar/slots"
	}
	patientID, err := strconv.Atoi(r.FormValue("patient_id"))
	if err != nil {
		redirectErr(w, r, back, "Выберите пациента")
		return
	}
	err = s.Store.BookSlot(r.Context(), id, patientID)
	if err == store.ErrConflict {
		redirectErr(w, r, back, "Слот уже занят")
		return
	}
	if err != nil {
		redirectErr(w, r, back, err.Error())
		return
	}
	redirectOK(w, r, back, "Пациент записан")
}

func (s *Server) CancelSlot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	_ = r.ParseForm()
	back := r.FormValue("return")
	if back == "" {
		back = "/registrar/schedule"
	}
	err = s.Store.CancelSlot(r.Context(), id)
	if err == store.ErrConflict {
		redirectErr(w, r, back, "Нельзя отменить: слот свободен или уже есть визит")
		return
	}
	if err != nil {
		redirectErr(w, r, back, err.Error())
		return
	}
	redirectOK(w, r, back, "Запись отменена")
}

func (s *Server) RescheduleSlot(w http.ResponseWriter, r *http.Request) {
	fromID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	_ = r.ParseForm()
	back := r.FormValue("return")
	if back == "" {
		back = "/registrar/schedule"
	}
	toID, err := strconv.Atoi(r.FormValue("to_id"))
	if err != nil {
		redirectErr(w, r, back, "Выберите новый слот")
		return
	}
	patientID, err := strconv.Atoi(r.FormValue("patient_id"))
	if err != nil {
		redirectErr(w, r, back, "Нет пациента")
		return
	}
	err = s.Store.RescheduleSlot(r.Context(), fromID, toID, patientID)
	if err == store.ErrConflict {
		redirectErr(w, r, back, "Перенос невозможен (занято / есть визит)")
		return
	}
	if err != nil {
		redirectErr(w, r, back, err.Error())
		return
	}
	redirectOK(w, r, back, "Запись перенесена")
}

func (s *Server) RegistrarSchedule(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	day, err := parseDate(r.URL.Query().Get("date"))
	if err != nil {
		redirectErr(w, r, "/registrar/schedule", "Некорректная дата")
		return
	}
	doctorID, _ := strconv.Atoi(r.URL.Query().Get("doctor_id"))
	appts, err := s.Store.ListAppointmentsDay(r.Context(), day, doctorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	doctors, err := s.Store.ListDoctors(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	patients, err := s.Store.ListPatients(r.Context(), "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	free, err := s.Store.ListFreeSlots(r.Context(), "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "registrar_schedule.html", pageData{
		Title: "Расписание", Active: "schedule", Session: sess,
		Appointments: appts, Doctors: doctors, Patients: patients, Slots: free,
		Date: day.Format("2006-01-02"), DoctorID: doctorID,
		Flash: flashOK(r), FlashError: flashErr(r),
	})
}

func (s *Server) CreateSlotsBatch(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	back := r.FormValue("return")
	if back == "" {
		back = "/registrar/slots"
	}
	doctorID, err := strconv.Atoi(r.FormValue("doctor_id"))
	if err != nil {
		redirectErr(w, r, back, "Выберите врача")
		return
	}
	start, err := time.ParseInLocation("2006-01-02T15:04", r.FormValue("start_at"), time.Local)
	if err != nil {
		start, err = time.ParseInLocation("2006-01-02T15:04:05", r.FormValue("start_at"), time.Local)
	}
	if err != nil {
		redirectErr(w, r, back, "Некорректное время начала")
		return
	}
	count, _ := strconv.Atoi(r.FormValue("count"))
	step, _ := strconv.Atoi(r.FormValue("step_min"))
	if step == 0 {
		step = 30
	}
	n, err := s.Store.CreateSlotsBatch(r.Context(), doctorID, start, count, step)
	if err != nil {
		redirectErr(w, r, back, err.Error())
		return
	}
	redirectOK(w, r, back, fmt.Sprintf("Создано слотов: %d", n))
}
