package httphandler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/omnik/tech-ip-sem2/services/auth/internal/service"
)

type Handler struct {
	svc *service.AuthService
}

func New(svc *service.AuthService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/login", h.login)
	mux.HandleFunc("/v1/auth/verify", h.verify)
	return mux
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type verifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject,omitempty"`
	Error   string `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	token, ok := h.svc.Login(req.Username, req.Password)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	})
}

func (h *Handler) verify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" || token == authHeader {
		writeJSON(w, http.StatusUnauthorized, verifyResponse{Valid: false, Error: "unauthorized"})
		return
	}

	subject, valid := h.svc.Verify(token)
	if !valid {
		writeJSON(w, http.StatusUnauthorized, verifyResponse{Valid: false, Error: "unauthorized"})
		return
	}

	writeJSON(w, http.StatusOK, verifyResponse{Valid: true, Subject: subject})
}
