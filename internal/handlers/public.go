package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"clinic/internal/auth"
	"clinic/internal/store"
)

func (s *Server) PublicHome(w http.ResponseWriter, r *http.Request) {
	featured, err := s.Store.ListFeaturedServices(r.Context(), 6)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	doctors, _ := s.Store.ListDoctors(r.Context())
	if len(doctors) > 3 {
		doctors = doctors[:3]
	}
	s.render(w, "public_home.html", pageData{
		Title:    "Клиника",
		Active:   "home",
		Services: featured,
		Doctors:  doctors,
	})
}

func (s *Server) PublicServices(w http.ResponseWriter, r *http.Request) {
	list, err := s.Store.ListPublicServices(r.Context(), true)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	s.render(w, "public_services.html", pageData{
		Title:    "Услуги",
		Active:   "services",
		Services: list,
	})
}

func (s *Server) PublicService(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	svc, err := s.Store.GetPublicServiceBySlug(r.Context(), slug)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if !svc.IsPublished {
		http.NotFound(w, r)
		return
	}
	s.render(w, "public_service.html", pageData{
		Title:   svc.Title,
		Active:  "services",
		Service: svc,
	})
}

func (s *Server) PublicDoctors(w http.ResponseWriter, r *http.Request) {
	doctors, err := s.Store.ListDoctors(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	s.render(w, "public_doctors.html", pageData{
		Title:   "Врачи",
		Active:  "doctors",
		Doctors: doctors,
	})
}

func (s *Server) PublicContacts(w http.ResponseWriter, r *http.Request) {
	s.render(w, "public_contacts.html", pageData{Title: "Контакты", Active: "contacts"})
}

func (s *Server) PublicAppointmentGet(w http.ResponseWriter, r *http.Request) {
	list, _ := s.Store.ListPublicServices(r.Context(), true)
	doctors, _ := s.Store.ListDoctors(r.Context())
	form := map[string]string{"service_slug": r.URL.Query().Get("service")}
	s.render(w, "public_appointment.html", pageData{
		Title:    "Запись",
		Active:   "appointment",
		Services: list,
		Doctors:  doctors,
		Form:     form,
		Flash:    flashOK(r),
		FlashError: flashErr(r),
	})
}

func (s *Server) PublicAppointmentPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		redirectErr(w, r, "/appointment", "Ошибка формы")
		return
	}
	name := strings.TrimSpace(r.FormValue("full_name"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	if name == "" || phone == "" {
		redirectErr(w, r, "/appointment", "Укажите имя и телефон")
		return
	}
	in := store.AppointmentRequestInput{
		FullName:    name,
		Phone:       phone,
		ServiceSlug: strings.TrimSpace(r.FormValue("service_slug")),
		DoctorNote:  strings.TrimSpace(r.FormValue("doctor_note")),
		Comment:     strings.TrimSpace(r.FormValue("comment")),
	}
	if d := strings.TrimSpace(r.FormValue("preferred_date")); d != "" {
		t, err := time.ParseInLocation("2006-01-02", d, time.Local)
		if err == nil {
			in.PreferredDate = &t
		}
	}
	if _, err := s.Store.CreateAppointmentRequest(r.Context(), in); err != nil {
		redirectErr(w, r, "/appointment", "Не удалось сохранить заявку")
		return
	}
	redirectOK(w, r, "/appointment", "Заявка принята. Администратор перезвонит для подтверждения.")
}

func (s *Server) AdminSiteServices(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	list, err := s.Store.ListPublicServices(r.Context(), false)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	s.render(w, "admin_site_services.html", pageData{
		Title:      "Услуги сайта",
		Active:     "site_services",
		Session:    sess,
		Services:   list,
		Flash:      flashOK(r),
		FlashError: flashErr(r),
	})
}

func (s *Server) AdminSiteServiceNewGet(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	s.render(w, "admin_site_service_form.html", pageData{
		Title:   "Новая услуга",
		Active:  "site_services",
		Session: sess,
		Form: map[string]string{
			"sort_order":   "0",
			"is_published": "1",
			"image_url":    "/static/img/service-placeholder.jpg",
		},
		FlashError: flashErr(r),
	})
}

func (s *Server) AdminSiteServiceCreatePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		redirectErr(w, r, "/admin/site-services/new", "Ошибка формы")
		return
	}
	in, err := parseServiceForm(r)
	if err != nil {
		redirectErr(w, r, "/admin/site-services/new", err.Error())
		return
	}
	if _, err := s.Store.CreatePublicService(r.Context(), in); err != nil {
		if errors.Is(err, store.ErrConflict) {
			redirectErr(w, r, "/admin/site-services/new", "Услуга с таким кодом уже есть")
			return
		}
		redirectErr(w, r, "/admin/site-services/new", "Не сохранено")
		return
	}
	redirectOK(w, r, "/admin/site-services", "Услуга добавлена")
}

func (s *Server) AdminSiteServiceEditGet(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	id, _ := strconv.Atoi(r.PathValue("id"))
	svc, err := s.Store.GetPublicServiceByID(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	form := map[string]string{
		"slug":         svc.Slug,
		"title":        svc.Title,
		"summary":      svc.Summary,
		"body":         svc.Body,
		"image_url":    svc.ImageURL,
		"sort_order":   strconv.Itoa(svc.SortOrder),
		"is_featured":  boolToForm(svc.IsFeatured),
		"is_published": boolToForm(svc.IsPublished),
	}
	if svc.PriceFrom.Valid {
		form["price_from"] = strconv.FormatInt(svc.PriceFrom.Int64, 10)
	}
	s.render(w, "admin_site_service_form.html", pageData{
		Title:   "Редактировать услугу",
		Active:  "site_services",
		Session: sess,
		Service: svc,
		Form:    form,
	})
}

func (s *Server) AdminSiteServiceEditPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if err := r.ParseForm(); err != nil {
		redirectErr(w, r, "/admin/site-services", "Ошибка формы")
		return
	}
	in, err := parseServiceForm(r)
	if err != nil {
		redirectErr(w, r, "/admin/site-services/"+strconv.Itoa(id)+"/edit", err.Error())
		return
	}
	if err := s.Store.UpdatePublicService(r.Context(), id, in); err != nil {
		redirectErr(w, r, "/admin/site-services/"+strconv.Itoa(id)+"/edit", "Не сохранено")
		return
	}
	redirectOK(w, r, "/admin/site-services", "Услуга обновлена")
}

func (s *Server) AdminSiteRequests(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFromContext(r.Context())
	list, err := s.Store.ListAppointmentRequests(r.Context())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	s.render(w, "admin_site_requests.html", pageData{
		Title:      "Заявки с сайта",
		Active:     "site_requests",
		Session:    sess,
		Requests:   list,
		Flash:      flashOK(r),
		FlashError: flashErr(r),
	})
}

func (s *Server) AdminSiteRequestStatusPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if err := r.ParseForm(); err != nil {
		redirectErr(w, r, "/admin/site-requests", "Ошибка формы")
		return
	}
	if err := s.Store.UpdateAppointmentRequestStatus(r.Context(), id, r.FormValue("status")); err != nil {
		redirectErr(w, r, "/admin/site-requests", "Не обновлено")
		return
	}
	redirectOK(w, r, "/admin/site-requests", "Статус обновлён")
}

func parseServiceForm(r *http.Request) (store.PublicServiceInput, error) {
	slug := strings.TrimSpace(r.FormValue("slug"))
	title := strings.TrimSpace(r.FormValue("title"))
	if slug == "" || title == "" {
		return store.PublicServiceInput{}, errors.New("укажите код страницы и название")
	}
	img := strings.TrimSpace(r.FormValue("image_url"))
	if img == "" {
		img = "/static/img/service-placeholder.jpg"
	}
	if strings.HasPrefix(img, "http://") {
		return store.PublicServiceInput{}, errors.New("укажите защищённую ссылку https://… или путь /static/img/…")
	}
	if strings.HasPrefix(img, "https://") || strings.HasPrefix(img, "/static/") {
		// ok
	} else if strings.HasPrefix(img, "/") {
		// ok absolute path on same host
	} else {
		return store.PublicServiceInput{}, errors.New("фото: ссылка https://… или путь /static/img/…")
	}
	in := store.PublicServiceInput{
		Slug:        slug,
		Title:       title,
		Summary:     strings.TrimSpace(r.FormValue("summary")),
		Body:        strings.TrimSpace(r.FormValue("body")),
		ImageURL:    img,
		IsFeatured:  r.FormValue("is_featured") == "1" || r.FormValue("is_featured") == "on",
		IsPublished: r.FormValue("is_published") == "1" || r.FormValue("is_published") == "on",
	}
	if so := strings.TrimSpace(r.FormValue("sort_order")); so != "" {
		n, err := strconv.Atoi(so)
		if err != nil {
			return store.PublicServiceInput{}, errors.New("некорректный порядок")
		}
		in.SortOrder = n
	}
	if pf := strings.TrimSpace(r.FormValue("price_from")); pf != "" {
		n, err := strconv.Atoi(pf)
		if err != nil {
			return store.PublicServiceInput{}, errors.New("некорректная цена")
		}
		in.PriceFrom = &n
	}
	return in, nil
}

func boolToForm(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
