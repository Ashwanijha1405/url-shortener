package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Ashwanijha1405/url-shortener/internal/generator"
	"github.com/Ashwanijha1405/url-shortener/internal/validator"
)

type CreateURLRequest struct {
	URL string `json:"url"`
}

type CreateURLResponse struct {
	ShortCode string `json:"short_code"`
}

func CreateURL(w http.ResponseWriter, r *http.Request) {
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

	response := CreateURLResponse{
		ShortCode: shortCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
    }

	json.NewEncoder(w).Encode(response)
}

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("shortCode")

	if shortCode == "" {
		http.Error(w, "short code is required", http.StatusBadRequest)
		return
	}

	// Temporary mapping.
	// This will eventually come from the database.
	url := "https://google.com"

	http.Redirect(w, r, url, http.StatusFound)
}
