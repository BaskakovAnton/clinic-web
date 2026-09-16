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

func (s *Server) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	stats, err := s.Store.Dashboard(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newReq, _ := s.Store.CountAppointmentRequestsByStatus(r.Context(), "new")
	s.render(w, "admin_dashboard.html", pageData{
		Title: "Админ", Active: "admin", Session: sess, Stats: stats,
		NewRequests: newReq,
		Flash: flashOK(r), FlashError: flashErr(r),
	})
}

func (s *Server) AdminRoles(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	s.render(w, "admin_roles.html", pageData{
		Title: "Роли доступа", Active: "roles", Session: sess,
	})
}

func (s *Server) AdminStaff(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	list, err := s.Store.ListStaff(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "admin_staff.html", pageData{
		Title: "Персонал", Active: "staff", Session: sess, StaffList: list,
		Flash: flashOK(r), FlashError: flashErr(r),
	})
}

func (s *Server) AdminStaffNewGet(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	s.render(w, "admin_staff_form.html", pageData{
		Title: "Новый сотрудник", Active: "staff", Session: sess,
		Form: map[string]string{"staff_kind": "doctor"}, FlashError: flashErr(r),
	})
}

func (s *Server) AdminStaffCreate(w http.ResponseWriter, r *http.Request) {
	in, msg := staffFromForm(r)
	if msg != "" {
		redirectErr(w, r, "/admin/staff/new", msg)
		return
	}
	_, err := s.Store.CreateStaff(r.Context(), in)
	if err != nil {
		redirectErr(w, r, "/admin/staff/new", err.Error())
		return
	}
	redirectOK(w, r, "/admin/staff", "Сотрудник создан")
}

func (s *Server) AdminStaffEditGet(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	st, err := s.Store.GetStaff(r.Context(), id)
	if err == store.ErrNotFound {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "admin_staff_form.html", pageData{
		Title: "Редактирование сотрудника", Active: "staff", Session: sess,
		Staff: st, Form: staffToForm(st), FlashError: flashErr(r),
	})
}

func (s *Server) AdminStaffUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	path := fmt.Sprintf("/admin/staff/%d/edit", id)
	in, msg := staffFromForm(r)
	if msg != "" {
		redirectErr(w, r, path, msg)
		return
	}
	if err := s.Store.UpdateStaff(r.Context(), id, in); err != nil {
		redirectErr(w, r, path, err.Error())
		return
	}
	redirectOK(w, r, "/admin/staff", "Сохранено")
}

func (s *Server) AdminStaffDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := s.Store.DeleteStaff(r.Context(), id); err != nil {
		redirectErr(w, r, "/admin/staff", err.Error())
		return
	}
	redirectOK(w, r, "/admin/staff", "Удалено")
}

func staffFromForm(r *http.Request) (store.StaffInput, string) {
	if err := r.ParseForm(); err != nil {
		return store.StaffInput{}, "Ошибка формы"
	}
	in := store.StaffInput{
		FullName:     strings.TrimSpace(r.FormValue("full_name")),
		StaffKind:    r.FormValue("staff_kind"),
		Specialty:    strings.TrimSpace(r.FormValue("specialty")),
		Department:   strings.TrimSpace(r.FormValue("department")),
		WorkSchedule: strings.TrimSpace(r.FormValue("work_schedule")),
		Office:       strings.TrimSpace(r.FormValue("office")),
	}
	if in.FullName == "" || in.Department == "" {
		return in, "ФИО и отделение обязательны"
	}
	if in.StaffKind != "doctor" && in.StaffKind != "nurse" {
		return in, "Тип: doctor или nurse"
	}
	if in.StaffKind == "doctor" && in.Specialty == "" {
		return in, "У врача нужна специальность"
	}
	return in, ""
}

func staffToForm(st store.Staff) map[string]string {
	m := map[string]string{
		"full_name": st.FullName, "staff_kind": st.StaffKind, "department": st.Department,
	}
	if st.Specialty.Valid {
		m["specialty"] = st.Specialty.String
	}
	if st.WorkSchedule.Valid {
		m["work_schedule"] = st.WorkSchedule.String
	}
	if st.Office.Valid {
		m["office"] = st.Office.String
	}
	return m
}

func (s *Server) AdminMassSlots(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	doctors, err := s.Store.ListDoctors(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "admin_mass_slots.html", pageData{
		Title: "Массовые слоты", Active: "mass", Session: sess, Doctors: doctors,
		Flash: flashOK(r), FlashError: flashErr(r),
		Form: map[string]string{
			"week_start":    nextMonday().Format("2006-01-02"),
			"day_start":     "9",
			"day_end":       "15",
			"step_min":      "30",
		},
	})
}

func nextMonday() time.Time {
	now := time.Now()
	off := (8 - int(now.Weekday())) % 7
	if off == 0 {
		off = 7
	}
	d := now.AddDate(0, 0, off)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

func (s *Server) AdminMassSlotsCreate(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	weekStart, err := time.ParseInLocation("2006-01-02", r.FormValue("week_start"), time.Local)
	if err != nil {
		redirectErr(w, r, "/admin/mass-slots", "Некорректная дата недели")
		return
	}
	// normalize to Monday
	for weekStart.Weekday() != time.Monday {
		weekStart = weekStart.AddDate(0, 0, -1)
	}
	dayStart, _ := strconv.Atoi(r.FormValue("day_start"))
	dayEnd, _ := strconv.Atoi(r.FormValue("day_end"))
	step, _ := strconv.Atoi(r.FormValue("step_min"))
	if step == 0 {
		step = 30
	}
	var ids []int
	for _, v := range r.Form["doctor_ids"] {
		id, err := strconv.Atoi(v)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		redirectErr(w, r, "/admin/mass-slots", "Выберите хотя бы одного врача")
		return
	}
	n, err := s.Store.CreateSlotsWeek(r.Context(), ids, weekStart, dayStart, dayEnd, step)
	if err != nil {
		redirectErr(w, r, "/admin/mass-slots", err.Error())
		return
	}
	redirectOK(w, r, "/admin/mass-slots", fmt.Sprintf("Создано слотов: %d", n))
}
