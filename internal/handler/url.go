package handler

import (
	"encoding/json"
	"net/http"

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

	if err := validator.ValidateURL(req.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := CreateURLResponse{
		ShortCode: "abc123",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

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