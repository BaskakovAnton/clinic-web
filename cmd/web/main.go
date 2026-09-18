package main

import (
	"log"
	"net/http"

	"clinic/internal/auth"
	"clinic/internal/config"
	"clinic/internal/db"
	"clinic/internal/handlers"
	"clinic/internal/store"
	apptemplates "clinic/internal/templates"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	sqlDB, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer sqlDB.Close()

	tmpl, err := apptemplates.New("templates")
	if err != nil {
		log.Fatalf("templates: %v", err)
	}

	authMgr := auth.NewManager(cfg.SessionSecret, cfg.DatabaseURL, cfg.SecureCookie)
	srv := &handlers.Server{
		Auth:  authMgr,
		Store: &store.Store{DB: sqlDB},
		Tmpl:  tmpl,
		DB:    sqlDB,
	}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("GET /healthz", srv.Healthz)

	// Public client site (no auth)
	mux.HandleFunc("GET /{$}", srv.PublicHome)
	mux.HandleFunc("GET /services", srv.PublicServices)
	mux.HandleFunc("GET /services/{slug}", srv.PublicService)
	mux.HandleFunc("GET /doctors", srv.PublicDoctors)
	mux.HandleFunc("GET /contacts", srv.PublicContacts)
	mux.HandleFunc("GET /appointment", srv.PublicAppointmentGet)
	mux.HandleFunc("POST /appointment", srv.PublicAppointmentPost)

	mux.HandleFunc("GET /login", srv.LoginGet)
	mux.HandleFunc("POST /login", srv.LoginPost)
	mux.Handle("POST /logout", authMgr.RequireAuth(http.HandlerFunc(srv.Logout)))
	mux.Handle("GET /app", authMgr.RequireAuth(http.HandlerFunc(srv.Home)))

	reg := auth.RequireRoles(auth.RoleRegistrar, auth.RoleAdmin)
	doc := auth.RequireRoles(auth.RoleDoctor, auth.RoleAdmin)
	adm := auth.RequireRoles(auth.RoleAdmin)
	authN := authMgr.RequireAuth

	// registrar
	mux.Handle("GET /registrar/patients", authN(reg(http.HandlerFunc(srv.RegistrarPatients))))
	mux.Handle("GET /registrar/patients/new", authN(reg(http.HandlerFunc(srv.RegistrarPatientNewGet))))
	mux.Handle("POST /registrar/patients", authN(reg(http.HandlerFunc(srv.RegistrarPatientCreate))))
	mux.Handle("GET /registrar/patients/{id}", authN(reg(http.HandlerFunc(srv.RegistrarPatientGet))))
	mux.Handle("GET /registrar/patients/{id}/edit", authN(reg(http.HandlerFunc(srv.RegistrarPatientEditGet))))
	mux.Handle("POST /registrar/patients/{id}", authN(reg(http.HandlerFunc(srv.RegistrarPatientUpdate))))
	mux.Handle("GET /registrar/slots", authN(reg(http.HandlerFunc(srv.RegistrarSlots))))
	mux.Handle("POST /registrar/slots/batch", authN(reg(http.HandlerFunc(srv.CreateSlotsBatch))))
	mux.Handle("POST /registrar/slots/{id}/book", authN(reg(http.HandlerFunc(srv.BookSlot))))
	mux.Handle("POST /registrar/slots/{id}/cancel", authN(reg(http.HandlerFunc(srv.CancelSlot))))
	mux.Handle("POST /registrar/slots/{id}/reschedule", authN(reg(http.HandlerFunc(srv.RescheduleSlot))))
	mux.Handle("GET /registrar/schedule", authN(reg(http.HandlerFunc(srv.RegistrarSchedule))))

	// doctor
	mux.Handle("GET /doctor/day", authN(doc(http.HandlerFunc(srv.DoctorDay))))
	mux.Handle("GET /doctor/patients", authN(doc(http.HandlerFunc(srv.DoctorPatients))))
	mux.Handle("GET /doctor/schedule", authN(doc(http.HandlerFunc(srv.DoctorSchedule))))
	mux.Handle("GET /doctor/patients/{id}", authN(doc(http.HandlerFunc(srv.DoctorPatient))))
	mux.Handle("POST /doctor/patients/{id}/visits", authN(doc(http.HandlerFunc(srv.CreateVisit))))
	mux.Handle("GET /doctor/visits/{id}", authN(doc(http.HandlerFunc(srv.DoctorVisitGet))))
	mux.Handle("POST /doctor/visits/{id}", authN(doc(http.HandlerFunc(srv.DoctorVisitUpdate))))
	mux.Handle("POST /doctor/visits/{id}/orders", authN(doc(http.HandlerFunc(srv.DoctorOrderCreate))))
	mux.Handle("POST /doctor/visits/{id}/orders/{orderID}", authN(doc(http.HandlerFunc(srv.DoctorOrderUpdate))))
	mux.Handle("GET /reports/workload", authN(doc(http.HandlerFunc(srv.ReportsWorkload))))
	mux.Handle("GET /reports/preferential", authN(doc(http.HandlerFunc(srv.ReportsPreferential))))

	// admin
	mux.Handle("GET /admin", authN(adm(http.HandlerFunc(srv.AdminDashboard))))
	mux.Handle("GET /admin/roles", authN(adm(http.HandlerFunc(srv.AdminRoles))))
	mux.Handle("GET /admin/staff", authN(adm(http.HandlerFunc(srv.AdminStaff))))
	mux.Handle("GET /admin/staff/new", authN(adm(http.HandlerFunc(srv.AdminStaffNewGet))))
	mux.Handle("POST /admin/staff", authN(adm(http.HandlerFunc(srv.AdminStaffCreate))))
	mux.Handle("GET /admin/staff/{id}/edit", authN(adm(http.HandlerFunc(srv.AdminStaffEditGet))))
	mux.Handle("POST /admin/staff/{id}", authN(adm(http.HandlerFunc(srv.AdminStaffUpdate))))
	mux.Handle("POST /admin/staff/{id}/delete", authN(adm(http.HandlerFunc(srv.AdminStaffDelete))))
	mux.Handle("GET /admin/mass-slots", authN(adm(http.HandlerFunc(srv.AdminMassSlots))))
	mux.Handle("POST /admin/mass-slots", authN(adm(http.HandlerFunc(srv.AdminMassSlotsCreate))))
	mux.Handle("GET /admin/site-services", authN(adm(http.HandlerFunc(srv.AdminSiteServices))))
	mux.Handle("GET /admin/site-services/new", authN(adm(http.HandlerFunc(srv.AdminSiteServiceNewGet))))
	mux.Handle("POST /admin/site-services", authN(adm(http.HandlerFunc(srv.AdminSiteServiceCreatePost))))
	mux.Handle("GET /admin/site-services/{id}/edit", authN(adm(http.HandlerFunc(srv.AdminSiteServiceEditGet))))
	mux.Handle("POST /admin/site-services/{id}", authN(adm(http.HandlerFunc(srv.AdminSiteServiceEditPost))))
	mux.Handle("GET /admin/site-requests", authN(adm(http.HandlerFunc(srv.AdminSiteRequests))))
	mux.Handle("POST /admin/site-requests/{id}/status", authN(adm(http.HandlerFunc(srv.AdminSiteRequestStatusPost))))

	addr := "0.0.0.0:" + cfg.Port
	log.Printf("Сервер запущен на http://%s (публичный сайт + /login для персонала)", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
