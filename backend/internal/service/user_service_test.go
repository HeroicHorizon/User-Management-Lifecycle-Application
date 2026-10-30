package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/HeroicHorizon/user-management/backend/internal/apperror"
	"github.com/HeroicHorizon/user-management/backend/internal/model"
	"github.com/HeroicHorizon/user-management/backend/internal/testutil"
)

func valid() model.UserInput {
	return model.UserInput{
		FirstName: "John", LastName: "Doe", Email: "john.doe@example.com",
		Phone: "+911234567890", Department: "Engineering", Designation: "Engineer",
	}
}

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func newSvc() (*UserService, *testutil.MemRepo, *clock) {
	r := testutil.NewMemRepo()
	c := &clock{t: time.Date(2026, 10, 30, 10, 0, 0, 0, time.UTC)}
	return NewUserService(r, c.now), r, c
}

func TestCreate_Valid(t *testing.T) {
	s, r, c := newSvc()
	u, err := s.Create(context.Background(), valid())
	if err != nil {
		t.Fatal(err)
	}
	if u.Status != model.StatusActive || !u.CreatedAt.Equal(c.t) || !u.UpdatedAt.Equal(c.t) {
		t.Fatalf("unexpected user %+v", u)
	}
	if len(r.Users) != 1 {
		t.Fatal("not persisted")
	}
}

func TestCreate_MissingFields(t *testing.T) {
	s, r, _ := newSvc()
	_, err := s.Create(context.Background(), model.UserInput{})
	var ve *apperror.ValidationError
	if !errors.As(err, &ve) || !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	for _, f := range []string{"firstName", "lastName", "email", "department", "designation"} {
		if ve.Fields[f] == "" {
			t.Errorf("missing error for %s", f)
		}
	}
	if len(r.Users) != 0 {
		t.Fatal("record created")
	}
}

func TestCreate_InvalidEmail(t *testing.T) {
	s, _, _ := newSvc()
	for _, e := range []string{"nope", "a@b", "a b@c.com", "@x.com"} {
		in := valid()
		in.Email = e
		if _, err := s.Create(context.Background(), in); !errors.Is(err, apperror.ErrValidation) {
			t.Errorf("%q: expected validation error, got %v", e, err)
		}
	}
}

func TestCreate_InvalidPhone(t *testing.T) {
	s, _, _ := newSvc()
	in := valid()
	in.Phone = "abc"
	if _, err := s.Create(context.Background(), in); !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("got %v", err)
	}
}

func TestCreate_DuplicateEmail(t *testing.T) {
	s, _, _ := newSvc()
	if _, err := s.Create(context.Background(), valid()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create(context.Background(), valid()); !errors.Is(err, apperror.ErrDuplicateEmail) {
		t.Fatalf("got %v", err)
	}
}

func TestGet_InvalidID(t *testing.T) {
	s, _, _ := newSvc()
	if _, err := s.Get(context.Background(), "xyz"); !errors.Is(err, apperror.ErrInvalidID) {
		t.Fatalf("got %v", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	s, _, _ := newSvc()
	if _, err := s.Get(context.Background(), "64b7f0c2a1b2c3d4e5f60718"); !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestGet_Found(t *testing.T) {
	s, _, _ := newSvc()
	u, _ := s.Create(context.Background(), valid())
	got, err := s.Get(context.Background(), u.ID.Hex())
	if err != nil || got.Email != "john.doe@example.com" {
		t.Fatalf("got %v %v", got, err)
	}
}

func TestUpdate_RefreshesUpdatedAt(t *testing.T) {
	s, _, c := newSvc()
	u, _ := s.Create(context.Background(), valid())
	c.t = c.t.Add(time.Hour)
	in := valid()
	in.FirstName = "Jane"
	got, err := s.Update(context.Background(), u.ID.Hex(), in)
	if err != nil {
		t.Fatal(err)
	}
	if got.FirstName != "Jane" || !got.UpdatedAt.Equal(c.t) || !got.CreatedAt.Equal(u.CreatedAt) {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestUpdate_DuplicateEmail(t *testing.T) {
	s, _, _ := newSvc()
	s.Create(context.Background(), valid())
	in2 := valid()
	in2.Email = "other@example.com"
	u2, _ := s.Create(context.Background(), in2)
	in2.Email = "john.doe@example.com"
	if _, err := s.Update(context.Background(), u2.ID.Hex(), in2); !errors.Is(err, apperror.ErrDuplicateEmail) {
		t.Fatalf("got %v", err)
	}
}

func TestUpdate_NotFoundAndBadID(t *testing.T) {
	s, _, _ := newSvc()
	if _, err := s.Update(context.Background(), "bad", valid()); !errors.Is(err, apperror.ErrInvalidID) {
		t.Fatalf("got %v", err)
	}
	if _, err := s.Update(context.Background(), "64b7f0c2a1b2c3d4e5f60718", valid()); !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestUpdateStatus_ActiveInactive(t *testing.T) {
	s, _, _ := newSvc()
	u, _ := s.Create(context.Background(), valid())
	got, err := s.UpdateStatus(context.Background(), u.ID.Hex(), "INACTIVE")
	if err != nil || got.Status != model.StatusInactive {
		t.Fatalf("got %v %v", got, err)
	}
	got, err = s.UpdateStatus(context.Background(), u.ID.Hex(), "ACTIVE")
	if err != nil || got.Status != model.StatusActive {
		t.Fatalf("got %v %v", got, err)
	}
}

func TestUpdateStatus_Unsupported(t *testing.T) {
	s, _, _ := newSvc()
	u, _ := s.Create(context.Background(), valid())
	if _, err := s.UpdateStatus(context.Background(), u.ID.Hex(), "BANNED"); !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("got %v", err)
	}
}

func TestDelete(t *testing.T) {
	s, r, _ := newSvc()
	u, _ := s.Create(context.Background(), valid())
	if err := s.Delete(context.Background(), u.ID.Hex()); err != nil {
		t.Fatal(err)
	}
	if len(r.Users) != 0 {
		t.Fatal("not deleted")
	}
}

func TestDelete_NotFound(t *testing.T) {
	s, _, _ := newSvc()
	if err := s.Delete(context.Background(), "64b7f0c2a1b2c3d4e5f60718"); !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
	if err := s.Delete(context.Background(), "x"); !errors.Is(err, apperror.ErrInvalidID) {
		t.Fatalf("got %v", err)
	}
}

func TestList_PaginationClamp(t *testing.T) {
	s, r, _ := newSvc()
	res, err := s.List(context.Background(), model.ListQuery{Page: -3, PageSize: 1000})
	if err != nil || res.Page != 1 || res.PageSize != MaxPageSize {
		t.Fatalf("got %+v %v", res, err)
	}
	res, _ = s.List(context.Background(), model.ListQuery{})
	if res.PageSize != DefaultPageSize {
		t.Fatalf("default page size %d", res.PageSize)
	}
	s.List(context.Background(), model.ListQuery{Page: 3, PageSize: 5})
	if r.LastSkip != 10 || r.LastLimit != 5 {
		t.Fatalf("skip/limit %d/%d", r.LastSkip, r.LastLimit)
	}
}

func TestList_Pages(t *testing.T) {
	s, _, _ := newSvc()
	for _, e := range []string{"a@x.com", "b@x.com", "c@x.com"} {
		in := valid()
		in.Email = e
		s.Create(context.Background(), in)
	}
	res, _ := s.List(context.Background(), model.ListQuery{Page: 2, PageSize: 2})
	if res.Total != 3 || len(res.Items) != 1 {
		t.Fatalf("got %+v", res)
	}
}

func TestList_SearchEscapesRegex(t *testing.T) {
	s, r, _ := newSvc()
	s.List(context.Background(), model.ListQuery{Search: " a.*(b "})
	if r.LastSearch != `a\.\*\(b` {
		t.Fatalf("search %q", r.LastSearch)
	}
}

func TestList_FilterAndSearch(t *testing.T) {
	s, _, _ := newSvc()
	u, _ := s.Create(context.Background(), valid())
	in := valid()
	in.Email, in.FirstName = "z@x.com", "Zed"
	s.Create(context.Background(), in)
	s.UpdateStatus(context.Background(), u.ID.Hex(), "INACTIVE")

	res, _ := s.List(context.Background(), model.ListQuery{Status: model.StatusInactive})
	if res.Total != 1 || res.Items[0].Email != "john.doe@example.com" {
		t.Fatalf("got %+v", res)
	}
	res, _ = s.List(context.Background(), model.ListQuery{Search: "zed"})
	if res.Total != 1 {
		t.Fatalf("search got %+v", res)
	}
	if _, err := s.List(context.Background(), model.ListQuery{Status: "X"}); !errors.Is(err, apperror.ErrValidation) {
		t.Fatalf("got %v", err)
	}
}

func TestStats(t *testing.T) {
	s, _, _ := newSvc()
	u, _ := s.Create(context.Background(), valid())
	in := valid()
	in.Email = "b@x.com"
	s.Create(context.Background(), in)
	s.UpdateStatus(context.Background(), u.ID.Hex(), "INACTIVE")
	st, err := s.Stats(context.Background())
	if err != nil || st.Total != 2 || st.Active != 1 || st.Inactive != 1 {
		t.Fatalf("got %+v %v", st, err)
	}
}

type ctxKey struct{}

func TestContextPropagated(t *testing.T) {
	s, r, _ := newSvc()
	ctx := context.WithValue(context.Background(), ctxKey{}, "v")
	s.Create(ctx, valid())
	if r.LastCtx.Value(ctxKey{}) != "v" {
		t.Fatal("context not propagated to repository")
	}
	cctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := s.Stats(cctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}
