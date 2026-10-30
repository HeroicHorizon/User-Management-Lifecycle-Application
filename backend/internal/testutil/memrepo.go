// Package testutil provides an in-memory UserRepository for unit tests.
package testutil

import (
	"context"
	"regexp"
	"sort"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/HeroicHorizon/user-management/backend/internal/apperror"
	"github.com/HeroicHorizon/user-management/backend/internal/model"
)

type MemRepo struct {
	mu    sync.Mutex
	Users map[bson.ObjectID]model.User
	// Err, when set, is returned by every call (e.g. apperror.ErrUnavailable).
	Err error
	// LastCtx records the context passed to the latest call.
	LastCtx context.Context
	// LastSearch / LastStatus / LastSkip / LastLimit record the latest List arguments.
	LastSearch string
	LastStatus model.Status
	LastSkip   int64
	LastLimit  int64
}

func NewMemRepo() *MemRepo { return &MemRepo{Users: map[bson.ObjectID]model.User{}} }

func (m *MemRepo) enter(ctx context.Context) error {
	m.mu.Lock()
	m.LastCtx = ctx
	if m.Err != nil {
		return m.Err
	}
	return ctx.Err()
}

func (m *MemRepo) emailTaken(email string, except bson.ObjectID) bool {
	for id, u := range m.Users {
		if u.Email == email && id != except {
			return true
		}
	}
	return false
}

func (m *MemRepo) Create(ctx context.Context, u *model.User) error {
	err := m.enter(ctx)
	defer m.mu.Unlock()
	if err != nil {
		return err
	}
	if m.emailTaken(u.Email, bson.NilObjectID) {
		return apperror.ErrDuplicateEmail
	}
	u.ID = bson.NewObjectID()
	m.Users[u.ID] = *u
	return nil
}

func (m *MemRepo) GetByID(ctx context.Context, id bson.ObjectID) (*model.User, error) {
	err := m.enter(ctx)
	defer m.mu.Unlock()
	if err != nil {
		return nil, err
	}
	u, ok := m.Users[id]
	if !ok {
		return nil, apperror.ErrNotFound
	}
	return &u, nil
}

func (m *MemRepo) Update(ctx context.Context, u *model.User) error {
	err := m.enter(ctx)
	defer m.mu.Unlock()
	if err != nil {
		return err
	}
	old, ok := m.Users[u.ID]
	if !ok {
		return apperror.ErrNotFound
	}
	if m.emailTaken(u.Email, u.ID) {
		return apperror.ErrDuplicateEmail
	}
	u.Status, u.CreatedAt = old.Status, old.CreatedAt
	m.Users[u.ID] = *u
	return nil
}

func (m *MemRepo) UpdateStatus(ctx context.Context, id bson.ObjectID, s model.Status, at time.Time) (*model.User, error) {
	err := m.enter(ctx)
	defer m.mu.Unlock()
	if err != nil {
		return nil, err
	}
	u, ok := m.Users[id]
	if !ok {
		return nil, apperror.ErrNotFound
	}
	u.Status, u.UpdatedAt = s, at
	m.Users[id] = u
	return &u, nil
}

func (m *MemRepo) Delete(ctx context.Context, id bson.ObjectID) error {
	err := m.enter(ctx)
	defer m.mu.Unlock()
	if err != nil {
		return err
	}
	if _, ok := m.Users[id]; !ok {
		return apperror.ErrNotFound
	}
	delete(m.Users, id)
	return nil
}

func (m *MemRepo) List(ctx context.Context, skip, limit int64, search string, status model.Status) ([]model.User, int64, error) {
	err := m.enter(ctx)
	defer m.mu.Unlock()
	m.LastSearch, m.LastStatus, m.LastSkip, m.LastLimit = search, status, skip, limit
	if err != nil {
		return nil, 0, err
	}
	var rx *regexp.Regexp
	if search != "" {
		rx = regexp.MustCompile("(?i)" + search)
	}
	all := []model.User{}
	for _, u := range m.Users {
		if status != "" && u.Status != status {
			continue
		}
		if rx != nil && !rx.MatchString(u.FirstName) && !rx.MatchString(u.LastName) && !rx.MatchString(u.Email) {
			continue
		}
		all = append(all, u)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Email < all[j].Email })
	total := int64(len(all))
	if skip > total {
		skip = total
	}
	end := skip + limit
	if end > total {
		end = total
	}
	return all[skip:end], total, nil
}

func (m *MemRepo) Stats(ctx context.Context) (model.Stats, error) {
	err := m.enter(ctx)
	defer m.mu.Unlock()
	var s model.Stats
	if err != nil {
		return s, err
	}
	for _, u := range m.Users {
		s.Total++
		if u.Status == model.StatusActive {
			s.Active++
		} else {
			s.Inactive++
		}
	}
	return s, nil
}
