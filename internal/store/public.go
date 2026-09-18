package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type PublicService struct {
	ID          int
	Slug        string
	Title       string
	Summary     string
	Body        string
	PriceFrom   sql.NullInt64
	ImageURL    string
	SortOrder   int
	IsFeatured  bool
	IsPublished bool
}

type PublicServiceInput struct {
	Slug        string
	Title       string
	Summary     string
	Body        string
	PriceFrom   *int
	ImageURL    string
	SortOrder   int
	IsFeatured  bool
	IsPublished bool
}

type AppointmentRequest struct {
	ID            int
	FullName      string
	Phone         string
	ServiceSlug   sql.NullString
	PreferredDate sql.NullTime
	DoctorNote    sql.NullString
	Comment       string
	Status        string
	CreatedAt     time.Time
}

type AppointmentRequestInput struct {
	FullName      string
	Phone         string
	ServiceSlug   string
	PreferredDate *time.Time
	DoctorNote    string
	Comment       string
}

func (s *Store) ListPublicServices(ctx context.Context, publishedOnly bool) ([]PublicService, error) {
	q := `
SELECT id, slug, title, summary, body, price_from, image_url, sort_order, is_featured, is_published
FROM public_services`
	if publishedOnly {
		q += ` WHERE is_published = true`
	}
	q += ` ORDER BY sort_order, title`
	rows, err := s.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PublicService
	for rows.Next() {
		var it PublicService
		if err := rows.Scan(&it.ID, &it.Slug, &it.Title, &it.Summary, &it.Body, &it.PriceFrom, &it.ImageURL, &it.SortOrder, &it.IsFeatured, &it.IsPublished); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Store) ListFeaturedServices(ctx context.Context, limit int) ([]PublicService, error) {
	rows, err := s.DB.QueryContext(ctx, `
SELECT id, slug, title, summary, body, price_from, image_url, sort_order, is_featured, is_published
FROM public_services
WHERE is_published = true AND is_featured = true
ORDER BY sort_order
LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PublicService
	for rows.Next() {
		var it PublicService
		if err := rows.Scan(&it.ID, &it.Slug, &it.Title, &it.Summary, &it.Body, &it.PriceFrom, &it.ImageURL, &it.SortOrder, &it.IsFeatured, &it.IsPublished); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Store) GetPublicServiceBySlug(ctx context.Context, slug string) (PublicService, error) {
	var it PublicService
	err := s.DB.QueryRowContext(ctx, `
SELECT id, slug, title, summary, body, price_from, image_url, sort_order, is_featured, is_published
FROM public_services WHERE slug = $1`, slug).
		Scan(&it.ID, &it.Slug, &it.Title, &it.Summary, &it.Body, &it.PriceFrom, &it.ImageURL, &it.SortOrder, &it.IsFeatured, &it.IsPublished)
	if err == sql.ErrNoRows {
		return PublicService{}, ErrNotFound
	}
	return it, err
}

func (s *Store) GetPublicServiceByID(ctx context.Context, id int) (PublicService, error) {
	var it PublicService
	err := s.DB.QueryRowContext(ctx, `
SELECT id, slug, title, summary, body, price_from, image_url, sort_order, is_featured, is_published
FROM public_services WHERE id = $1`, id).
		Scan(&it.ID, &it.Slug, &it.Title, &it.Summary, &it.Body, &it.PriceFrom, &it.ImageURL, &it.SortOrder, &it.IsFeatured, &it.IsPublished)
	if err == sql.ErrNoRows {
		return PublicService{}, ErrNotFound
	}
	return it, err
}

func (s *Store) CreatePublicService(ctx context.Context, in PublicServiceInput) (int, error) {
	img := strings.TrimSpace(in.ImageURL)
	if img == "" {
		img = "/static/img/service-placeholder.jpg"
	}
	var price any
	if in.PriceFrom != nil {
		price = *in.PriceFrom
	}
	var id int
	err := s.DB.QueryRowContext(ctx, `
INSERT INTO public_services (slug, title, summary, body, price_from, image_url, sort_order, is_featured, is_published)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING id`,
		in.Slug, in.Title, in.Summary, in.Body, price, img, in.SortOrder, in.IsFeatured, in.IsPublished,
	).Scan(&id)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") || strings.Contains(err.Error(), "duplicate") {
			return 0, ErrConflict
		}
		return 0, err
	}
	return id, nil
}

func (s *Store) UpsertPublicService(ctx context.Context, in PublicServiceInput) error {
	img := strings.TrimSpace(in.ImageURL)
	if img == "" {
		img = "/static/img/service-placeholder.jpg"
	}
	var price any
	if in.PriceFrom != nil {
		price = *in.PriceFrom
	}
	_, err := s.DB.ExecContext(ctx, `
INSERT INTO public_services (slug, title, summary, body, price_from, image_url, sort_order, is_featured, is_published)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
ON CONFLICT (slug) DO UPDATE SET
  title = EXCLUDED.title,
  summary = EXCLUDED.summary,
  body = EXCLUDED.body,
  price_from = EXCLUDED.price_from,
  image_url = EXCLUDED.image_url,
  sort_order = EXCLUDED.sort_order,
  is_featured = EXCLUDED.is_featured,
  is_published = EXCLUDED.is_published
`, in.Slug, in.Title, in.Summary, in.Body, price, img, in.SortOrder, in.IsFeatured, in.IsPublished)
	return err
}

func (s *Store) UpdatePublicService(ctx context.Context, id int, in PublicServiceInput) error {
	img := strings.TrimSpace(in.ImageURL)
	if img == "" {
		img = "/static/img/service-placeholder.jpg"
	}
	var price any
	if in.PriceFrom != nil {
		price = *in.PriceFrom
	}
	res, err := s.DB.ExecContext(ctx, `
UPDATE public_services SET
  slug=$2, title=$3, summary=$4, body=$5, price_from=$6, image_url=$7,
  sort_order=$8, is_featured=$9, is_published=$10
WHERE id=$1`, id, in.Slug, in.Title, in.Summary, in.Body, price, img, in.SortOrder, in.IsFeatured, in.IsPublished)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CreateAppointmentRequest(ctx context.Context, in AppointmentRequestInput) (int, error) {
	var id int
	var pref any
	if in.PreferredDate != nil {
		pref = *in.PreferredDate
	}
	var slug any
	if strings.TrimSpace(in.ServiceSlug) != "" {
		slug = strings.TrimSpace(in.ServiceSlug)
	}
	var doctor any
	if strings.TrimSpace(in.DoctorNote) != "" {
		doctor = strings.TrimSpace(in.DoctorNote)
	}
	err := s.DB.QueryRowContext(ctx, `
INSERT INTO appointment_requests (full_name, phone, service_slug, preferred_date, doctor_note, comment)
VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		strings.TrimSpace(in.FullName), strings.TrimSpace(in.Phone), slug, pref, doctor, strings.TrimSpace(in.Comment)).Scan(&id)
	return id, err
}

func (s *Store) ListAppointmentRequests(ctx context.Context) ([]AppointmentRequest, error) {
	rows, err := s.DB.QueryContext(ctx, `
SELECT id, full_name, phone, service_slug, preferred_date, doctor_note, comment, status, created_at
FROM appointment_requests
ORDER BY created_at DESC
LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AppointmentRequest
	for rows.Next() {
		var it AppointmentRequest
		if err := rows.Scan(&it.ID, &it.FullName, &it.Phone, &it.ServiceSlug, &it.PreferredDate, &it.DoctorNote, &it.Comment, &it.Status, &it.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Store) UpdateAppointmentRequestStatus(ctx context.Context, id int, status string) error {
	switch status {
	case "new", "called", "done", "cancelled":
	default:
		return fmt.Errorf("invalid status")
	}
	res, err := s.DB.ExecContext(ctx, `UPDATE appointment_requests SET status=$2 WHERE id=$1`, id, status)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CountAppointmentRequestsByStatus(ctx context.Context, status string) (int, error) {
	var n int
	err := s.DB.QueryRowContext(ctx, `SELECT count(*) FROM appointment_requests WHERE status=$1`, status).Scan(&n)
	return n, err
}
