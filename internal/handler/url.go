package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Ashwanijha1405/url-shortener/internal/middleware"
	"github.com/Ashwanijha1405/url-shortener/internal/service"
)

const defaultMaxBodyBytes = 10240 // 10KB

type Handler struct {
	svc          service.URLService
	maxBodyBytes int64
}

type HandlerOption func(*Handler)

func WithMaxBodyBytes(n int64) HandlerOption {
	return func(h *Handler) {
		if n > 0 {
			h.maxBodyBytes = n
		}
	}
}

func NewHandler(svc service.URLService, opts ...HandlerOption) *Handler {
	h := &Handler{
		svc:          svc,
		maxBodyBytes: defaultMaxBodyBytes,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

type CreateURLRequest struct {
	URL string `json:"url"`
}

type CreateURLResponse struct {
	ShortCode string `json:"short_code"`
	Cache     string `json:"cache,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func (h *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodyBytes)

	var req CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.svc.CreateShortURLWithMetadata(r.Context(), req.URL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrConflict) {
			http.Error(w, "unable to generate unique short code, please try again", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := CreateURLResponse{
		ShortCode: result.ShortCode,
		Cache:     result.Cache,
		RequestID: middleware.GetRequestID(r.Context()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")
	if shortCode == "" {
		http.Error(w, "short code is required", http.StatusBadRequest)
		return
	}

	originalURL, err := h.svc.ResolveURL(r.Context(), shortCode)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "short URL not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
