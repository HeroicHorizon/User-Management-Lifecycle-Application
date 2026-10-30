package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/HeroicHorizon/user-management/backend/internal/apperror"
	"github.com/HeroicHorizon/user-management/backend/internal/model"
)

// newRepo connects to MONGO_TEST_URI (skipping when unset) and uses a
// throwaway database that is dropped at the end of the test.
func newRepo(t *testing.T) *MongoUserRepository {
	t.Helper()
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping MongoDB repository test")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	db := client.Database(fmt.Sprintf("usermgmt_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = db.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	r := NewMongoUserRepository(db)
	if err := r.EnsureIndexes(context.Background()); err != nil {
		t.Fatal(err)
	}
	return r
}

func mk(email, first string, st model.Status, at time.Time) *model.User {
	return &model.User{FirstName: first, LastName: "Doe", Email: email, Department: "D", Designation: "X",
		Status: st, CreatedAt: at, UpdatedAt: at}
}

func TestRepoCRUD(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)
	u := mk("a@x.com", "A", model.StatusActive, now)
	if err := r.Create(ctx, u); err != nil || u.ID.IsZero() {
		t.Fatalf("create %v", err)
	}
	got, err := r.GetByID(ctx, u.ID)
	if err != nil || got.Email != "a@x.com" {
		t.Fatalf("get %v %v", got, err)
	}
	u.FirstName, u.UpdatedAt = "B", now.Add(time.Hour)
	if err := r.Update(ctx, u); err != nil || u.FirstName != "B" || u.Status != model.StatusActive {
		t.Fatalf("update %v %+v", err, u)
	}
	st, err := r.UpdateStatus(ctx, u.ID, model.StatusInactive, now.Add(2*time.Hour))
	if err != nil || st.Status != model.StatusInactive {
		t.Fatalf("status %v", err)
	}
	if err := r.Delete(ctx, u.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := r.GetByID(ctx, u.ID); !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
	if err := r.Delete(ctx, u.ID); !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
	if err := r.Update(ctx, &model.User{ID: bson.NewObjectID(), Email: "q@x.com"}); !errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestRepoUniqueEmailIndex(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()
	if err := r.Create(ctx, mk("a@x.com", "A", model.StatusActive, now)); err != nil {
		t.Fatal(err)
	}
	if err := r.Create(ctx, mk("a@x.com", "B", model.StatusActive, now)); !errors.Is(err, apperror.ErrDuplicateEmail) {
		t.Fatalf("got %v", err)
	}
	b := mk("b@x.com", "B", model.StatusActive, now)
	r.Create(ctx, b)
	b.Email = "a@x.com"
	if err := r.Update(ctx, b); !errors.Is(err, apperror.ErrDuplicateEmail) {
		t.Fatalf("got %v", err)
	}
}

func TestRepoListSearchFilterPaginate(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	base := time.Now().UTC().Truncate(time.Millisecond)
	for i := 0; i < 5; i++ {
		st := model.StatusActive
		if i%2 == 1 {
			st = model.StatusInactive
		}
		r.Create(ctx, mk(fmt.Sprintf("u%d@x.com", i), fmt.Sprintf("Name%d", i), st, base.Add(time.Duration(i)*time.Minute)))
	}
	items, total, err := r.List(ctx, 0, 2, "", "")
	if err != nil || total != 5 || len(items) != 2 || items[0].Email != "u4@x.com" {
		t.Fatalf("page1 %v %d %+v", err, total, items)
	}
	items, _, _ = r.List(ctx, 4, 2, "", "")
	if len(items) != 1 {
		t.Fatalf("last page %d", len(items))
	}
	_, total, _ = r.List(ctx, 0, 10, "", model.StatusInactive)
	if total != 2 {
		t.Fatalf("inactive %d", total)
	}
	_, total, _ = r.List(ctx, 0, 10, "name3", "")
	if total != 1 {
		t.Fatalf("search %d", total)
	}
	_, total, _ = r.List(ctx, 0, 10, regexp.QuoteMeta("u1@x.com"), "")
	if total != 1 {
		t.Fatalf("email search %d", total)
	}
	_, total, _ = r.List(ctx, 0, 10, regexp.QuoteMeta(".*"), "")
	if total != 0 {
		t.Fatalf("escaped regex matched %d", total)
	}
}

func TestRepoStats(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	now := time.Now().UTC()
	r.Create(ctx, mk("a@x.com", "A", model.StatusActive, now))
	r.Create(ctx, mk("b@x.com", "B", model.StatusInactive, now))
	r.Create(ctx, mk("c@x.com", "C", model.StatusInactive, now))
	s, err := r.Stats(ctx)
	if err != nil || s.Total != 3 || s.Active != 1 || s.Inactive != 2 {
		t.Fatalf("%v %+v", err, s)
	}
}

func TestRepoContextCancel(t *testing.T) {
	r := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := r.Stats(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

// Does not need MongoDB.
func TestMapErr(t *testing.T) {
	if mapErr(nil) != nil {
		t.Fatal("nil")
	}
	if !errors.Is(mapErr(mongo.ErrNoDocuments), apperror.ErrNotFound) {
		t.Fatal("no documents")
	}
	if !errors.Is(mapErr(context.DeadlineExceeded), apperror.ErrUnavailable) {
		t.Fatal("deadline")
	}
	if !errors.Is(mapErr(errors.New("server selection error: host:27017")), apperror.ErrUnavailable) {
		t.Fatal("server selection")
	}
	if !errors.Is(mapErr(context.Canceled), context.Canceled) {
		t.Fatal("canceled")
	}
	we := mongo.WriteException{WriteErrors: mongo.WriteErrors{{Code: 11000}}}
	if !errors.Is(mapErr(we), apperror.ErrDuplicateEmail) {
		t.Fatal("duplicate")
	}
}
