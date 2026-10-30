package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/HeroicHorizon/user-management/backend/internal/apperror"
	"github.com/HeroicHorizon/user-management/backend/internal/service"
	"github.com/HeroicHorizon/user-management/backend/internal/testutil"
)

const (
	origin    = "http://localhost:3000"
	missingID = "64b7f0c2a1b2c3d4e5f60718"
	body      = `{"firstName":"John","lastName":"Doe","email":"john@example.com","phone":"+911234567890","department":"Eng","designation":"Dev"}`
)

func setup(maxBody int64) (http.Handler, *testutil.MemRepo) {
	repo := testutil.NewMemRepo()
	h := NewUserHandler(service.NewUserService(repo, nil))
	return Chain(h.Routes(), origin, maxBody), repo
}

func do(h http.Handler, method, path, b string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(b))
	return doReq(h, req)
}

func doReq(h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

type resp struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Details map[string]string `json:"details"`
	} `json:"error"`
}

func parse(t *testing.T, rec *httptest.ResponseRecorder) resp {
	t.Helper()
	var r resp
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("bad json %q: %v", rec.Body.String(), err)
	}
	return r
}

func create(t *testing.T, h http.Handler, b string) string {
	t.Helper()
	rec := do(h, "POST", "/api/v1/users", b)
	if rec.Code != 201 {
		t.Fatalf("create: %d %s", rec.Code, rec.Body)
	}
	var u struct{ ID string }
	json.Unmarshal(parse(t, rec).Data, &u)
	return u.ID
}

func TestPostUser_201(t *testing.T) {
	h, repo := setup(1 << 20)
	rec := do(h, "POST", "/api/v1/users", body)
	if rec.Code != 201 || !parse(t, rec).Success {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	if len(repo.Users) != 1 {
		t.Fatal("not stored")
	}
}

func TestPostUser_409(t *testing.T) {
	h, _ := setup(1 << 20)
	create(t, h, body)
	rec := do(h, "POST", "/api/v1/users", body)
	if rec.Code != 409 {
		t.Fatalf("%d", rec.Code)
	}
	if m := parse(t, rec).Error.Message; m != "A user with this email already exists." {
		t.Fatalf("message %q", m)
	}
}

func TestPostUser_400Validation(t *testing.T) {
	h, repo := setup(1 << 20)
	rec := do(h, "POST", "/api/v1/users", `{"email":"bad"}`)
	if rec.Code != 400 {
		t.Fatalf("%d", rec.Code)
	}
	e := parse(t, rec).Error
	if e.Details["firstName"] == "" || e.Details["email"] == "" {
		t.Fatalf("details %v", e.Details)
	}
	if len(repo.Users) != 0 {
		t.Fatal("record created")
	}
}

func TestPostUser_MalformedJSON(t *testing.T) {
	h, _ := setup(1 << 20)
	for _, b := range []string{`{`, `{"bogus":1}`, body + body} {
		if rec := do(h, "POST", "/api/v1/users", b); rec.Code != 400 {
			t.Errorf("%q: %d", b, rec.Code)
		}
	}
}

func TestGetUser_400BadID(t *testing.T) {
	h, _ := setup(1 << 20)
	if rec := do(h, "GET", "/api/v1/users/not-an-id", ""); rec.Code != 400 {
		t.Fatalf("%d", rec.Code)
	}
}

func TestGetUser_404(t *testing.T) {
	h, _ := setup(1 << 20)
	if rec := do(h, "GET", "/api/v1/users/"+missingID, ""); rec.Code != 404 {
		t.Fatalf("%d", rec.Code)
	}
}

func TestGetUser_200(t *testing.T) {
	h, _ := setup(1 << 20)
	id := create(t, h, body)
	rec := do(h, "GET", "/api/v1/users/"+id, "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "john@example.com") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestPutUser_200(t *testing.T) {
	h, _ := setup(1 << 20)
	id := create(t, h, body)
	upd := strings.Replace(body, "John", "Jane", 1)
	rec := do(h, "PUT", "/api/v1/users/"+id, upd)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Jane") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	if got := do(h, "GET", "/api/v1/users/"+id, ""); !strings.Contains(got.Body.String(), "Jane") {
		t.Fatal("not persisted")
	}
}

func TestPutUser_409(t *testing.T) {
	h, _ := setup(1 << 20)
	create(t, h, body)
	id2 := create(t, h, strings.Replace(body, "john@", "other@", 1))
	if rec := do(h, "PUT", "/api/v1/users/"+id2, body); rec.Code != 409 {
		t.Fatalf("%d", rec.Code)
	}
}

func TestPutUser_404And400(t *testing.T) {
	h, _ := setup(1 << 20)
	if rec := do(h, "PUT", "/api/v1/users/"+missingID, body); rec.Code != 404 {
		t.Fatalf("%d", rec.Code)
	}
	if rec := do(h, "PUT", "/api/v1/users/zzz", body); rec.Code != 400 {
		t.Fatalf("%d", rec.Code)
	}
}

func TestPatchStatus_400(t *testing.T) {
	h, _ := setup(1 << 20)
	id := create(t, h, body)
	if rec := do(h, "PATCH", "/api/v1/users/"+id+"/status", `{"status":"BANNED"}`); rec.Code != 400 {
		t.Fatalf("%d", rec.Code)
	}
}

func TestPatchStatus_200(t *testing.T) {
	h, _ := setup(1 << 20)
	id := create(t, h, body)
	rec := do(h, "PATCH", "/api/v1/users/"+id+"/status", `{"status":"INACTIVE"}`)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"INACTIVE"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	rec = do(h, "PATCH", "/api/v1/users/"+id+"/status", `{"status":"ACTIVE"}`)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"ACTIVE"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	if rec := do(h, "PATCH", "/api/v1/users/"+missingID+"/status", `{"status":"ACTIVE"}`); rec.Code != 404 {
		t.Fatalf("%d", rec.Code)
	}
}

func TestDelete_404(t *testing.T) {
	h, _ := setup(1 << 20)
	if rec := do(h, "DELETE", "/api/v1/users/"+missingID, ""); rec.Code != 404 {
		t.Fatalf("%d", rec.Code)
	}
}

func TestDelete_200(t *testing.T) {
	h, repo := setup(1 << 20)
	id := create(t, h, body)
	if rec := do(h, "DELETE", "/api/v1/users/"+id, ""); rec.Code != 200 {
		t.Fatalf("%d", rec.Code)
	}
	if len(repo.Users) != 0 {
		t.Fatal("not deleted")
	}
	if rec := do(h, "GET", "/api/v1/users/"+id, ""); rec.Code != 404 {
		t.Fatalf("%d", rec.Code)
	}
}

func TestList_QueryParams(t *testing.T) {
	h, repo := setup(1 << 20)
	for _, e := range []string{"a", "b", "c"} {
		create(t, h, strings.Replace(body, "john@", e+"@", 1))
	}
	rec := do(h, "GET", "/api/v1/users?page=2&pageSize=2", "")
	var res struct {
		Items []map[string]any
		Page  int
		Total int
	}
	json.Unmarshal(parse(t, rec).Data, &res)
	if rec.Code != 200 || res.Page != 2 || res.Total != 3 || len(res.Items) != 1 {
		t.Fatalf("%d %+v", rec.Code, res)
	}
	do(h, "GET", "/api/v1/users?search=a.b&status=ACTIVE", "")
	if repo.LastSearch != `a\.b` || repo.LastStatus != "ACTIVE" {
		t.Fatalf("search=%q status=%q", repo.LastSearch, repo.LastStatus)
	}
	if rec := do(h, "GET", "/api/v1/users?status=NOPE", ""); rec.Code != 400 {
		t.Fatalf("%d", rec.Code)
	}
}

func TestList_EmptyHasEmptyItems(t *testing.T) {
	h, _ := setup(1 << 20)
	rec := do(h, "GET", "/api/v1/users", "")
	if !strings.Contains(rec.Body.String(), `"items":[]`) {
		t.Fatalf("%s", rec.Body)
	}
}

func TestStatsEndpoint(t *testing.T) {
	h, _ := setup(1 << 20)
	id := create(t, h, body)
	create(t, h, strings.Replace(body, "john@", "b@", 1))
	do(h, "PATCH", "/api/v1/users/"+id+"/status", `{"status":"INACTIVE"}`)
	rec := do(h, "GET", "/api/v1/users/stats", "")
	var s struct{ Total, Active, Inactive int }
	json.Unmarshal(parse(t, rec).Data, &s)
	if rec.Code != 200 || s.Total != 2 || s.Active != 1 || s.Inactive != 1 {
		t.Fatalf("%d %+v", rec.Code, s)
	}
}

func TestUnavailable_503_NoLeak(t *testing.T) {
	h, repo := setup(1 << 20)
	repo.Err = apperror.ErrUnavailable
	for _, c := range [][2]string{{"GET", "/api/v1/users"}, {"GET", "/api/v1/users/stats"}, {"POST", "/api/v1/users"}} {
		rec := do(h, c[0], c[1], body)
		if rec.Code != 503 {
			t.Fatalf("%v: %d", c, rec.Code)
		}
		s := strings.ToLower(rec.Body.String())
		for _, bad := range []string{"mongodb://", "goroutine", "panic", ".go:", "27017", "localhost"} {
			if strings.Contains(s, bad) {
				t.Fatalf("response leaks %q: %s", bad, s)
			}
		}
	}
}

func TestUnknownError_500Generic(t *testing.T) {
	h, repo := setup(1 << 20)
	repo.Err = errString("mongodb://user:pw@host:27017 exploded")
	rec := do(h, "GET", "/api/v1/users", "")
	if rec.Code != 500 || strings.Contains(rec.Body.String(), "mongodb://") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func TestOversizedBody_Rejected(t *testing.T) {
	h, repo := setup(64)
	rec := do(h, "POST", "/api/v1/users", body)
	if rec.Code != 413 {
		t.Fatalf("%d", rec.Code)
	}
	if len(repo.Users) != 0 {
		t.Fatal("created")
	}
}

func TestCORS_OnlyConfiguredOrigin(t *testing.T) {
	h, _ := setup(1 << 20)
	req := httptest.NewRequest("GET", "/healthz", nil)
	req.Header.Set("Origin", origin)
	if got := doReq(h, req).Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("allowed origin got %q", got)
	}
	req = httptest.NewRequest("GET", "/healthz", nil)
	req.Header.Set("Origin", "http://evil.example")
	if got := doReq(h, req).Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("foreign origin got %q", got)
	}
	req = httptest.NewRequest("OPTIONS", "/api/v1/users", nil)
	req.Header.Set("Origin", origin)
	rec := doReq(h, req)
	if rec.Code != 204 || rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("preflight %d", rec.Code)
	}
}

func TestCanceledContext(t *testing.T) {
	h, repo := setup(1 << 20)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest("GET", "/api/v1/users", nil).WithContext(ctx)
	rec := doReq(h, req)
	if rec.Code != 499 {
		t.Fatalf("%d", rec.Code)
	}
	if repo.LastCtx == nil || repo.LastCtx.Err() == nil {
		t.Fatal("request context was not propagated to the repository")
	}
}

func TestPanicRecovered(t *testing.T) {
	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }), origin, 10)
	rec := do(h, "GET", "/", "")
	if rec.Code != 500 || strings.Contains(rec.Body.String(), "boom") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
