package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

type Patient struct {
	ID              int
	CardNumber      string
	FullName        string
	BirthDate       time.Time
	InsuranceType   string
	InsuranceNumber string
	PassportData    string
	Address         sql.NullString
	Contacts        sql.NullString
}

type PatientInput struct {
	CardNumber      string
	FullName        string
	BirthDate       time.Time
	InsuranceType   string
	InsuranceNumber string
	PassportData    string
	Address         string
	Contacts        string
}

type FreeSlot struct {
	AppointmentID int
	DoctorName    string
	Specialty     string
	Department    string
	Office        sql.NullString
	StartAt       time.Time
}

type AppointmentRow struct {
	ID            int
	DoctorID      int
	DoctorName    string
	Specialty     sql.NullString
	Office        sql.NullString
	StartAt       time.Time
	Status        string
	PatientID     sql.NullInt64
	PatientName   sql.NullString
	PatientCard   sql.NullString
	HasVisit      bool
}

type Doctor struct {
	ID        int
	FullName  string
	Specialty sql.NullString
}

type Staff struct {
	ID           int
	FullName     string
	StaffKind    string
	Specialty    sql.NullString
	Department   string
	WorkSchedule sql.NullString
	Office       sql.NullString
}

type StaffInput struct {
	FullName     string
	StaffKind    string
	Specialty    string
	Department   string
	WorkSchedule string
	Office       string
}

type CardRow struct {
	PatientID        int
	CardNumber       string
	PatientName      string
	BirthDate        time.Time
	InsuranceType    string
	InsuranceNumber  string
	VisitID          sql.NullInt64
	VisitAt          sql.NullTime
	DoctorName       sql.NullString
	DoctorSpecialty  sql.NullString
	Complaints       sql.NullString
	DiagnosisICD10   sql.NullString
	AppointmentID    sql.NullInt64
	OrderID          sql.NullInt64
	OrderKind        sql.NullString
	OrderDescription sql.NullString
	IsPrescription   sql.NullBool
	IsPreferential   sql.NullBool
}

type Visit struct {
	ID             int
	PatientID      int
	DoctorID       int
	VisitAt        time.Time
	Complaints     sql.NullString
	DiagnosisICD10 string
	AppointmentID  sql.NullInt64
	DoctorName     string
	PatientName    string
}

type VisitOrder struct {
	ID             int
	VisitID        int
	OrderKind      string
	Description    string
	IsPrescription bool
	IsPreferential bool
}

type WorkloadRow struct {
	DoctorID     int
	DoctorName   string
	Specialty    sql.NullString
	Department   string
	PatientsSeen int
}

type PreferentialRow struct {
	OrderID        int
	VisitAt        time.Time
	CardNumber     string
	PatientName    string
	DoctorName     string
	OrderKind      string
	Description    string
	IsPrescription bool
	IsPreferential bool
}

type DashboardStats struct {
	Patients       int
	StaffDoctors   int
	StaffNurses    int
	SlotsFree      int
	SlotsBooked    int
	Visits         int
	Preferential   int
}

type Store struct {
	DB *sql.DB
}

func scanPatient(scanner interface {
	Scan(dest ...any) error
}) (Patient, error) {
	var p Patient
	err := scanner.Scan(
		&p.ID, &p.CardNumber, &p.FullName, &p.BirthDate, &p.InsuranceType, &p.InsuranceNumber,
		&p.PassportData, &p.Address, &p.Contacts,
	)
	return p, err
}

const patientCols = `id, card_number, full_name, birth_date, insurance_type, insurance_number, passport_data, address, contacts`

func (s *Store) ListPatients(ctx context.Context, q string) ([]Patient, error) {
	q = strings.TrimSpace(q)
	var (
		rows *sql.Rows
		err  error
	)
	if q == "" {
		rows, err = s.DB.QueryContext(ctx, `SELECT `+patientCols+` FROM patients ORDER BY id`)
	} else {
		like := "%" + q + "%"
		rows, err = s.DB.QueryContext(ctx, `
			SELECT `+patientCols+`
			FROM patients
			WHERE card_number ILIKE $1 OR full_name ILIKE $1 OR insurance_number ILIKE $1 OR COALESCE(contacts,'') ILIKE $1
			ORDER BY id`, like)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Patient
	for rows.Next() {
		p, err := scanPatient(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetPatient(ctx context.Context, id int) (Patient, error) {
	p, err := scanPatient(s.DB.QueryRowContext(ctx, `SELECT `+patientCols+` FROM patients WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Patient{}, ErrNotFound
	}
	return p, err
}

func (s *Store) CreatePatient(ctx context.Context, in PatientInput) (int, error) {
	var id int
	err := s.DB.QueryRowContext(ctx, `
		INSERT INTO patients (card_number, full_name, birth_date, insurance_type, insurance_number, passport_data, address, contacts)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id`,
		in.CardNumber, in.FullName, in.BirthDate, in.InsuranceType, in.InsuranceNumber,
		in.PassportData, nullIfEmpty(in.Address), nullIfEmpty(in.Contacts),
	).Scan(&id)
	return id, err
}

func (s *Store) UpdatePatient(ctx context.Context, id int, in PatientInput) error {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE patients SET
			card_number=$1, full_name=$2, birth_date=$3, insurance_type=$4, insurance_number=$5,
			passport_data=$6, address=$7, contacts=$8
		WHERE id=$9`,
		in.CardNumber, in.FullName, in.BirthDate, in.InsuranceType, in.InsuranceNumber,
		in.PassportData, nullIfEmpty(in.Address), nullIfEmpty(in.Contacts), id,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListPatientAppointments(ctx context.Context, patientID int) ([]AppointmentRow, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT a.id, a.doctor_id, s.full_name, s.specialty, s.office, a.start_at, a.status,
		       a.patient_id, p.full_name, p.card_number,
		       EXISTS(SELECT 1 FROM visits v WHERE v.appointment_id = a.id) AS has_visit
		FROM appointments a
		JOIN staff s ON s.id = a.doctor_id
		LEFT JOIN patients p ON p.id = a.patient_id
		WHERE a.patient_id = $1
		ORDER BY a.start_at DESC`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAppointments(rows)
}

func (s *Store) ListAppointmentsDay(ctx context.Context, day time.Time, doctorID int) ([]AppointmentRow, error) {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	end := start.Add(24 * time.Hour)
	q := `
		SELECT a.id, a.doctor_id, s.full_name, s.specialty, s.office, a.start_at, a.status,
		       a.patient_id, p.full_name, p.card_number,
		       EXISTS(SELECT 1 FROM visits v WHERE v.appointment_id = a.id) AS has_visit
		FROM appointments a
		JOIN staff s ON s.id = a.doctor_id
		LEFT JOIN patients p ON p.id = a.patient_id
		WHERE a.start_at >= $1 AND a.start_at < $2`
	args := []any{start, end}
	if doctorID > 0 {
		q += ` AND a.doctor_id = $3`
		args = append(args, doctorID)
	}
	q += ` ORDER BY a.start_at, s.full_name`
	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAppointments(rows)
}

func scanAppointments(rows *sql.Rows) ([]AppointmentRow, error) {
	var out []AppointmentRow
	for rows.Next() {
		var a AppointmentRow
		if err := rows.Scan(
			&a.ID, &a.DoctorID, &a.DoctorName, &a.Specialty, &a.Office, &a.StartAt, &a.Status,
			&a.PatientID, &a.PatientName, &a.PatientCard, &a.HasVisit,
		); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) GetAppointment(ctx context.Context, id int) (AppointmentRow, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT a.id, a.doctor_id, s.full_name, s.specialty, s.office, a.start_at, a.status,
		       a.patient_id, p.full_name, p.card_number,
		       EXISTS(SELECT 1 FROM visits v WHERE v.appointment_id = a.id) AS has_visit
		FROM appointments a
		JOIN staff s ON s.id = a.doctor_id
		LEFT JOIN patients p ON p.id = a.patient_id
		WHERE a.id = $1`, id)
	var a AppointmentRow
	err := row.Scan(
		&a.ID, &a.DoctorID, &a.DoctorName, &a.Specialty, &a.Office, &a.StartAt, &a.Status,
		&a.PatientID, &a.PatientName, &a.PatientCard, &a.HasVisit,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return AppointmentRow{}, ErrNotFound
	}
	return a, err
}

func (s *Store) ListFreeSlots(ctx context.Context, specialty string) ([]FreeSlot, error) {
	q := `
		SELECT appointment_id, doctor_name, specialty, department, office, start_at
		FROM v_free_slots_by_specialty`
	args := []any{}
	if specialty != "" {
		q += ` WHERE specialty = $1`
		args = append(args, specialty)
	}
	q += ` ORDER BY start_at, doctor_name`
	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FreeSlot
	for rows.Next() {
		var sl FreeSlot
		if err := rows.Scan(&sl.AppointmentID, &sl.DoctorName, &sl.Specialty, &sl.Department, &sl.Office, &sl.StartAt); err != nil {
			return nil, err
		}
		out = append(out, sl)
	}
	return out, rows.Err()
}

func (s *Store) ListSpecialties(ctx context.Context) ([]string, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT DISTINCT specialty FROM staff
		WHERE staff_kind = 'doctor' AND specialty IS NOT NULL
		ORDER BY specialty`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var spt string
		if err := rows.Scan(&spt); err != nil {
			return nil, err
		}
		out = append(out, spt)
	}
	return out, rows.Err()
}

func (s *Store) BookSlot(ctx context.Context, appointmentID, patientID int) error {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE appointments SET status = 'booked', patient_id = $1
		WHERE id = $2 AND status = 'free'`, patientID, appointmentID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrConflict
	}
	return nil
}

func (s *Store) CancelSlot(ctx context.Context, appointmentID int) error {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE appointments SET status = 'free', patient_id = NULL
		WHERE id = $1 AND status = 'booked'
		  AND NOT EXISTS (SELECT 1 FROM visits v WHERE v.appointment_id = $1)`, appointmentID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrConflict
	}
	return nil
}

func (s *Store) RescheduleSlot(ctx context.Context, fromID, toID, patientID int) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx, `
		UPDATE appointments SET status = 'free', patient_id = NULL
		WHERE id = $1 AND status = 'booked' AND patient_id = $2
		  AND NOT EXISTS (SELECT 1 FROM visits v WHERE v.appointment_id = $1)`, fromID, patientID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrConflict
	}
	res, err = tx.ExecContext(ctx, `
		UPDATE appointments SET status = 'booked', patient_id = $1
		WHERE id = $2 AND status = 'free'`, patientID, toID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrConflict
	}
	return tx.Commit()
}

func (s *Store) CreateSlotsBatch(ctx context.Context, doctorID int, start time.Time, count, stepMin int) (int, error) {
	if count < 1 || count > 48 || stepMin < 5 {
		return 0, fmt.Errorf("некорректные параметры слотов")
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	created := 0
	for i := 0; i < count; i++ {
		t := start.Add(time.Duration(i*stepMin) * time.Minute)
		res, err := tx.ExecContext(ctx, `
			INSERT INTO appointments (doctor_id, start_at, status, patient_id)
			VALUES ($1, $2, 'free', NULL)
			ON CONFLICT (doctor_id, start_at) DO NOTHING`, doctorID, t)
		if err != nil {
			return 0, err
		}
		n, _ := res.RowsAffected()
		created += int(n)
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return created, nil
}

func (s *Store) CreateSlotsWeek(ctx context.Context, doctorIDs []int, weekStart time.Time, dayStartHour, dayEndHour, stepMin int) (int, error) {
	if stepMin < 5 || dayEndHour <= dayStartHour {
		return 0, fmt.Errorf("некорректные параметры")
	}
	total := 0
	for _, docID := range doctorIDs {
		for d := 0; d < 5; d++ { // Mon-Fri
			day := weekStart.AddDate(0, 0, d)
			start := time.Date(day.Year(), day.Month(), day.Day(), dayStartHour, 0, 0, 0, weekStart.Location())
			end := time.Date(day.Year(), day.Month(), day.Day(), dayEndHour, 0, 0, 0, weekStart.Location())
			n := 0
			for t := start; t.Before(end); t = t.Add(time.Duration(stepMin) * time.Minute) {
				n++
			}
			c, err := s.CreateSlotsBatch(ctx, docID, start, n, stepMin)
			if err != nil {
				return total, err
			}
			total += c
		}
	}
	return total, nil
}

func (s *Store) ListDoctors(ctx context.Context) ([]Doctor, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, full_name, specialty FROM staff
		WHERE staff_kind = 'doctor' ORDER BY full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Doctor
	for rows.Next() {
		var d Doctor
		if err := rows.Scan(&d.ID, &d.FullName, &d.Specialty); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) ListStaff(ctx context.Context) ([]Staff, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, full_name, staff_kind, specialty, department, work_schedule, office
		FROM staff ORDER BY staff_kind, full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Staff
	for rows.Next() {
		var st Staff
		if err := rows.Scan(&st.ID, &st.FullName, &st.StaffKind, &st.Specialty, &st.Department, &st.WorkSchedule, &st.Office); err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, rows.Err()
}

func (s *Store) GetStaff(ctx context.Context, id int) (Staff, error) {
	var st Staff
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, full_name, staff_kind, specialty, department, work_schedule, office
		FROM staff WHERE id = $1`, id).
		Scan(&st.ID, &st.FullName, &st.StaffKind, &st.Specialty, &st.Department, &st.WorkSchedule, &st.Office)
	if errors.Is(err, sql.ErrNoRows) {
		return Staff{}, ErrNotFound
	}
	return st, err
}

func (s *Store) CreateStaff(ctx context.Context, in StaffInput) (int, error) {
	var id int
	var spec any
	if in.StaffKind == "doctor" {
		spec = in.Specialty
	} else {
		spec = nil
	}
	err := s.DB.QueryRowContext(ctx, `
		INSERT INTO staff (full_name, staff_kind, specialty, department, work_schedule, office)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		in.FullName, in.StaffKind, spec, in.Department, nullIfEmpty(in.WorkSchedule), nullIfEmpty(in.Office),
	).Scan(&id)
	return id, err
}

func (s *Store) UpdateStaff(ctx context.Context, id int, in StaffInput) error {
	var spec any
	if in.StaffKind == "doctor" {
		spec = in.Specialty
	} else {
		spec = nil
	}
	res, err := s.DB.ExecContext(ctx, `
		UPDATE staff SET full_name=$1, staff_kind=$2, specialty=$3, department=$4, work_schedule=$5, office=$6
		WHERE id=$7`,
		in.FullName, in.StaffKind, spec, in.Department, nullIfEmpty(in.WorkSchedule), nullIfEmpty(in.Office), id,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteStaff(ctx context.Context, id int) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM staff WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) PatientCard(ctx context.Context, patientID int) ([]CardRow, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT patient_id, card_number, patient_name, birth_date, insurance_type, insurance_number,
		       visit_id, visit_at, doctor_name, doctor_specialty, complaints, diagnosis_icd10,
		       appointment_id, order_id, order_kind, order_description, is_prescription, is_preferential
		FROM v_patient_medical_card
		WHERE patient_id = $1
		ORDER BY visit_at DESC NULLS LAST, visit_id DESC NULLS LAST, order_id`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CardRow
	for rows.Next() {
		var r CardRow
		if err := rows.Scan(
			&r.PatientID, &r.CardNumber, &r.PatientName, &r.BirthDate, &r.InsuranceType, &r.InsuranceNumber,
			&r.VisitID, &r.VisitAt, &r.DoctorName, &r.DoctorSpecialty, &r.Complaints, &r.DiagnosisICD10,
			&r.AppointmentID, &r.OrderID, &r.OrderKind, &r.OrderDescription, &r.IsPrescription, &r.IsPreferential,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

type CreateVisitInput struct {
	PatientID      int
	DoctorID       int
	Complaints     string
	DiagnosisICD10 string
	AppointmentID  *int
	OrderKind      string
	OrderDesc      string
	IsPrescription bool
	IsPreferential bool
}

func (s *Store) CreateVisit(ctx context.Context, in CreateVisitInput) (int, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	var visitID int
	err = tx.QueryRowContext(ctx, `
		INSERT INTO visits (patient_id, doctor_id, complaints, diagnosis_icd10, appointment_id)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		in.PatientID, in.DoctorID, nullIfEmpty(in.Complaints), in.DiagnosisICD10, nullableInt(in.AppointmentID),
	).Scan(&visitID)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(in.OrderKind) != "" && strings.TrimSpace(in.OrderDesc) != "" {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO visit_orders (visit_id, order_kind, description, is_prescription, is_preferential)
			VALUES ($1,$2,$3,$4,$5)`,
			visitID, in.OrderKind, in.OrderDesc, in.IsPrescription, in.IsPreferential)
		if err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return visitID, nil
}

func (s *Store) GetVisit(ctx context.Context, id int) (Visit, error) {
	var v Visit
	err := s.DB.QueryRowContext(ctx, `
		SELECT v.id, v.patient_id, v.doctor_id, v.visit_at, v.complaints, v.diagnosis_icd10, v.appointment_id,
		       s.full_name, p.full_name
		FROM visits v
		JOIN staff s ON s.id = v.doctor_id
		JOIN patients p ON p.id = v.patient_id
		WHERE v.id = $1`, id).
		Scan(&v.ID, &v.PatientID, &v.DoctorID, &v.VisitAt, &v.Complaints, &v.DiagnosisICD10, &v.AppointmentID,
			&v.DoctorName, &v.PatientName)
	if errors.Is(err, sql.ErrNoRows) {
		return Visit{}, ErrNotFound
	}
	return v, err
}

func (s *Store) UpdateVisit(ctx context.Context, id int, complaints, diagnosis string) error {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE visits SET complaints = $1, diagnosis_icd10 = $2 WHERE id = $3`,
		nullIfEmpty(complaints), diagnosis, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListVisitOrders(ctx context.Context, visitID int) ([]VisitOrder, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, visit_id, order_kind, description, is_prescription, is_preferential
		FROM visit_orders WHERE visit_id = $1 ORDER BY id`, visitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VisitOrder
	for rows.Next() {
		var o VisitOrder
		if err := rows.Scan(&o.ID, &o.VisitID, &o.OrderKind, &o.Description, &o.IsPrescription, &o.IsPreferential); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *Store) AddVisitOrder(ctx context.Context, visitID int, kind, desc string, prescription, preferential bool) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO visit_orders (visit_id, order_kind, description, is_prescription, is_preferential)
		VALUES ($1,$2,$3,$4,$5)`, visitID, kind, desc, prescription, preferential)
	return err
}

func (s *Store) UpdateVisitOrder(ctx context.Context, id int, kind, desc string, prescription, preferential bool) error {
	res, err := s.DB.ExecContext(ctx, `
		UPDATE visit_orders SET order_kind=$1, description=$2, is_prescription=$3, is_preferential=$4
		WHERE id=$5`, kind, desc, prescription, preferential, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) Workload(ctx context.Context) ([]WorkloadRow, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT doctor_id, doctor_name, specialty, department, patients_seen
		FROM v_doctor_workload ORDER BY patients_seen DESC, doctor_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WorkloadRow
	for rows.Next() {
		var r WorkloadRow
		if err := rows.Scan(&r.DoctorID, &r.DoctorName, &r.Specialty, &r.Department, &r.PatientsSeen); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) Preferential(ctx context.Context) ([]PreferentialRow, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT order_id, visit_at, card_number, patient_name, doctor_name,
		       order_kind, description, is_prescription, is_preferential
		FROM v_preferential_orders ORDER BY visit_at DESC, order_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PreferentialRow
	for rows.Next() {
		var r PreferentialRow
		if err := rows.Scan(
			&r.OrderID, &r.VisitAt, &r.CardNumber, &r.PatientName, &r.DoctorName,
			&r.OrderKind, &r.Description, &r.IsPrescription, &r.IsPreferential,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) Dashboard(ctx context.Context) (DashboardStats, error) {
	var d DashboardStats
	err := s.DB.QueryRowContext(ctx, `
		SELECT
			(SELECT count(*) FROM patients),
			(SELECT count(*) FROM staff WHERE staff_kind='doctor'),
			(SELECT count(*) FROM staff WHERE staff_kind='nurse'),
			(SELECT count(*) FROM appointments WHERE status='free'),
			(SELECT count(*) FROM appointments WHERE status='booked'),
			(SELECT count(*) FROM visits),
			(SELECT count(*) FROM visit_orders WHERE is_preferential)`).
		Scan(&d.Patients, &d.StaffDoctors, &d.StaffNurses, &d.SlotsFree, &d.SlotsBooked, &d.Visits, &d.Preferential)
	return d, err
}

func nullIfEmpty(s string) sql.NullString {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullableInt(p *int) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*p), Valid: true}
}
