package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/HeroicHorizon/user-management/backend/internal/model"
)

// UserService is the business layer the handler depends on.
type UserService interface {
	Create(ctx context.Context, in model.UserInput) (*model.User, error)
	Get(ctx context.Context, id string) (*model.User, error)
	Update(ctx context.Context, id string, in model.UserInput) (*model.User, error)
	UpdateStatus(ctx context.Context, id string, status string) (*model.User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, q model.ListQuery) (*model.ListResult, error)
	Stats(ctx context.Context) (model.Stats, error)
}

type UserHandler struct {
	svc UserService
}

func NewUserHandler(svc UserService) *UserHandler { return &UserHandler{svc: svc} }

// Routes registers all endpoints.
func (h *UserHandler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeData(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/v1/users", h.create)
	mux.HandleFunc("GET /api/v1/users", h.list)
	mux.HandleFunc("GET /api/v1/users/stats", h.stats)
	mux.HandleFunc("GET /api/v1/users/{id}", h.get)
	mux.HandleFunc("PUT /api/v1/users/{id}", h.update)
	mux.HandleFunc("PATCH /api/v1/users/{id}/status", h.updateStatus)
	mux.HandleFunc("DELETE /api/v1/users/{id}", h.delete)
	return mux
}

func decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return errBodyTooLarge
		}
		return errBadBody
	}
	if _, err := dec.Token(); err != io.EOF {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return errBodyTooLarge
		}
		return errBadBody
	}
	return nil
}

func (h *UserHandler) create(w http.ResponseWriter, r *http.Request) {
	var in model.UserInput
	if err := decode(r, &in); err != nil {
		writeError(w, r, err)
		return
	}
	u, err := h.svc.Create(r.Context(), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, u)
}

func (h *UserHandler) list(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	page, _ := strconv.Atoi(qs.Get("page"))
	size, _ := strconv.Atoi(qs.Get("pageSize"))
	res, err := h.svc.List(r.Context(), model.ListQuery{
		Page: page, PageSize: size, Search: qs.Get("search"), Status: model.Status(qs.Get("status")),
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, res)
}

func (h *UserHandler) stats(w http.ResponseWriter, r *http.Request) {
	s, err := h.svc.Stats(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, s)
}

func (h *UserHandler) get(w http.ResponseWriter, r *http.Request) {
	u, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, u)
}

func (h *UserHandler) update(w http.ResponseWriter, r *http.Request) {
	var in model.UserInput
	if err := decode(r, &in); err != nil {
		writeError(w, r, err)
		return
	}
	u, err := h.svc.Update(r.Context(), r.PathValue("id"), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, u)
}

func (h *UserHandler) updateStatus(w http.ResponseWriter, r *http.Request) {
	var in model.StatusInput
	if err := decode(r, &in); err != nil {
		writeError(w, r, err)
		return
	}
	u, err := h.svc.UpdateStatus(r.Context(), r.PathValue("id"), in.Status)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, u)
}

func (h *UserHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"message": "User deleted."})
}
