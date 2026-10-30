// Package service holds user validation and business rules.
package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/HeroicHorizon/user-management/backend/internal/apperror"
	"github.com/HeroicHorizon/user-management/backend/internal/model"
	"github.com/HeroicHorizon/user-management/backend/internal/repository"
)

const (
	DefaultPageSize = 10
	MaxPageSize     = 100
	maxFieldLen     = 100
)

var (
	emailRe = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9\-]+(\.[A-Za-z0-9\-]+)*\.[A-Za-z]{2,}$`)
	phoneRe = regexp.MustCompile(`^\+?[0-9][0-9 \-]{5,19}$`)
)

type UserService struct {
	repo repository.UserRepository
	now  func() time.Time
}

func NewUserService(repo repository.UserRepository, now func() time.Time) *UserService {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &UserService{repo: repo, now: now}
}

func (s *UserService) Create(ctx context.Context, in model.UserInput) (*model.User, error) {
	in, err := validate(in)
	if err != nil {
		return nil, err
	}
	t := s.now()
	u := &model.User{
		FirstName: in.FirstName, LastName: in.LastName, Email: in.Email, Phone: in.Phone,
		Department: in.Department, Designation: in.Designation,
		Status: model.StatusActive, CreatedAt: t, UpdatedAt: t,
	}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Get(ctx context.Context, id string) (*model.User, error) {
	oid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, oid)
}

func (s *UserService) Update(ctx context.Context, id string, in model.UserInput) (*model.User, error) {
	oid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	in, err = validate(in)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		ID: oid, FirstName: in.FirstName, LastName: in.LastName, Email: in.Email, Phone: in.Phone,
		Department: in.Department, Designation: in.Designation, UpdatedAt: s.now(),
	}
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) UpdateStatus(ctx context.Context, id string, status string) (*model.User, error) {
	oid, err := parseID(id)
	if err != nil {
		return nil, err
	}
	st := model.Status(strings.TrimSpace(status))
	if !st.Valid() {
		return nil, &apperror.ValidationError{Fields: map[string]string{"status": "must be ACTIVE or INACTIVE"}}
	}
	return s.repo.UpdateStatus(ctx, oid, st, s.now())
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	oid, err := parseID(id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, oid)
}

func (s *UserService) List(ctx context.Context, q model.ListQuery) (*model.ListResult, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = DefaultPageSize
	}
	if q.PageSize > MaxPageSize {
		q.PageSize = MaxPageSize
	}
	if q.Status != "" && !q.Status.Valid() {
		return nil, &apperror.ValidationError{Fields: map[string]string{"status": "must be ACTIVE or INACTIVE"}}
	}
	search := ""
	if t := strings.TrimSpace(q.Search); t != "" {
		search = regexp.QuoteMeta(t)
	}
	items, total, err := s.repo.List(ctx, int64(q.Page-1)*int64(q.PageSize), int64(q.PageSize), search, q.Status)
	if err != nil {
		return nil, err
	}
	return &model.ListResult{Items: items, Page: q.Page, PageSize: q.PageSize, Total: total}, nil
}

func (s *UserService) Stats(ctx context.Context) (model.Stats, error) {
	return s.repo.Stats(ctx)
}

func parseID(id string) (bson.ObjectID, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return bson.NilObjectID, apperror.ErrInvalidID
	}
	return oid, nil
}

func validate(in model.UserInput) (model.UserInput, error) {
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Phone = strings.TrimSpace(in.Phone)
	in.Department = strings.TrimSpace(in.Department)
	in.Designation = strings.TrimSpace(in.Designation)

	f := map[string]string{}
	required := map[string]string{
		"firstName": in.FirstName, "lastName": in.LastName, "email": in.Email,
		"department": in.Department, "designation": in.Designation,
	}
	for k, v := range required {
		if v == "" {
			f[k] = "is required"
		} else if len(v) > maxFieldLen && k != "email" {
			f[k] = "is too long"
		}
	}
	if in.Email != "" && (len(in.Email) > 254 || !emailRe.MatchString(in.Email)) {
		f["email"] = "must be a valid email address"
	}
	if in.Phone != "" && !phoneRe.MatchString(in.Phone) {
		f["phone"] = "must be a valid phone number"
	}
	if len(f) > 0 {
		return in, &apperror.ValidationError{Fields: f}
	}
	return in, nil
}
