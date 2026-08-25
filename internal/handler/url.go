package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Ashwanijha1405/url-shortener/internal/generator"
	"github.com/Ashwanijha1405/url-shortener/internal/repository"
	"github.com/Ashwanijha1405/url-shortener/internal/validator"
)

type Handler struct {
	repo repository.URLRepository
}

type CreateURLRequest struct {
	URL string `json:"url"`
}

type CreateURLResponse struct {
	ShortCode string `json:"short_code"`
}

func NewHandler(repo repository.URLRepository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) CreateURL(w http.ResponseWriter, r *http.Request) {
	var req CreateURLRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	if err := validator.ValidateURL(req.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortCode, err := generator.Generate(generator.DefaultLength)
	if err != nil {
		http.Error(w, "failed to generate short code", http.StatusInternalServerError)
		return
	}

	if err := h.repo.Create(r.Context(), shortCode, req.URL); err != nil {
		http.Error(w, "failed to create short URL", http.StatusInternalServerError)
		return
	}

	response := CreateURLResponse{
		ShortCode: shortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")

	if shortCode == "" {
		http.Error(w, "short code is required", http.StatusBadRequest)
		return
	}

	originalURL, err := h.repo.GetByShortCode(r.Context(), shortCode)
	if err != nil {
		http.Error(w, "short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}
