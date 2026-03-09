package service

import (
	"fmt"
	"sync"
	"time"
)

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Done        bool   `json:"done"`
}

type TaskService struct {
	mu     sync.RWMutex
	tasks  map[string]Task
	nextID int
}

func New() *TaskService {
	return &TaskService{
		tasks: make(map[string]Task),
	}
}

func (s *TaskService) generateID() string {
	s.nextID++
	return fmt.Sprintf("t_%03d_%d", s.nextID, time.Now().UnixNano()%1000)
}

func (s *TaskService) Create(title, description, dueDate string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Task{
		ID:          s.generateID(),
		Title:       title,
		Description: description,
		DueDate:     dueDate,
		Done:        false,
	}
	s.tasks[t.ID] = t
	return t
}

func (s *TaskService) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result
}

func (s *TaskService) Get(id string) (Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *TaskService) Update(id string, title *string, done *bool) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	if !ok {
		return Task{}, false
	}
	if title != nil {
		t.Title = *title
	}
	if done != nil {
		t.Done = *done
	}
	s.tasks[id] = t
	return t, true
}

func (s *TaskService) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return false
	}
	delete(s.tasks, id)
	return true
}
