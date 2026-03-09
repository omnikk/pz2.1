package httphandler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/omnik/tech-ip-sem2/services/tasks/internal/client/authclient"
	"github.com/omnik/tech-ip-sem2/services/tasks/internal/service"
	"github.com/omnik/tech-ip-sem2/shared/middleware"
)

type Handler struct {
	svc        *service.TaskService
	authClient *authclient.Client
}

func New(svc *service.TaskService, authClient *authclient.Client) *Handler {
	return &Handler{svc: svc, authClient: authClient}
}

func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tasks", h.tasksCollection)
	mux.HandleFunc("/v1/tasks/", h.taskItem)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// checkAuth извлекает токен и проверяет его через Auth service.
func (h *Handler) checkAuth(ctx context.Context, r *http.Request) (string, bool) {
	rid := r.Header.Get(middleware.RequestIDHeader)
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" || token == authHeader {
		return rid, false
	}

	subject, err := h.authClient.Verify(ctx, token, rid)
	if err != nil {
		log.Printf("[%s] auth verify error: %v", rid, err)
		return rid, false
	}

	log.Printf("[%s] auth ok, subject=%s", rid, subject)
	return rid, true
}

// POST /v1/tasks  |  GET /v1/tasks
func (h *Handler) tasksCollection(w http.ResponseWriter, r *http.Request) {
	rid, ok := h.checkAuth(r.Context(), r)
	if !ok {
		log.Printf("[%s] unauthorized", rid)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.createTask(w, r, rid)
	case http.MethodGet:
		h.listTasks(w, r, rid)
	default:
		http.NotFound(w, r)
	}
}

// GET /v1/tasks/{id}  |  PATCH /v1/tasks/{id}  |  DELETE /v1/tasks/{id}
func (h *Handler) taskItem(w http.ResponseWriter, r *http.Request) {
	rid, ok := h.checkAuth(r.Context(), r)
	if !ok {
		log.Printf("[%s] unauthorized", rid)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing task id"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r, rid, id)
	case http.MethodPatch:
		h.updateTask(w, r, rid, id)
	case http.MethodDelete:
		h.deleteTask(w, r, rid, id)
	default:
		http.NotFound(w, r)
	}
}

type createRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request, rid string) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	task := h.svc.Create(req.Title, req.Description, req.DueDate)
	log.Printf("[%s] task created: %s", rid, task.ID)
	writeJSON(w, http.StatusCreated, task)
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request, rid string) {
	tasks := h.svc.List()
	log.Printf("[%s] list tasks: %d items", rid, len(tasks))
	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request, rid, id string) {
	task, ok := h.svc.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}
	log.Printf("[%s] get task: %s", rid, id)
	writeJSON(w, http.StatusOK, task)
}

type updateRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request, rid, id string) {
	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	task, ok := h.svc.Update(id, req.Title, req.Done)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}
	log.Printf("[%s] task updated: %s", rid, id)
	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request, rid, id string) {
	if !h.svc.Delete(id) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}
	log.Printf("[%s] task deleted: %s", rid, id)
	w.WriteHeader(http.StatusNoContent)
}
