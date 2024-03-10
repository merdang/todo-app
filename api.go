package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type APIServer struct {
	listenAddr string
	store      Storage
	metrics    *PrometheusMetric
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	metrics, err := NewPrometheusMetric()
	if err != nil {
		log.Fatal("Failed tp initialize Prometheus metrics: ", err)
	}

	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
		metrics:    metrics,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/api", makeHTTPHandleFunc(s.handleTask))
	router.HandleFunc("/api/{id}", makeHTTPHandleFunc(s.handleTaskByID))
	router.HandleFunc("/metrics", makeHTTPHandleFunc(s.handleMetrics))
	router.HandleFunc("/alive", makeHTTPHandleFunc(s.handleLiveness))

	s.metrics.Init()

	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleLiveness(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusOK)
	return nil
}

func (s *APIServer) handleMetrics(w http.ResponseWriter, r *http.Request) error {
	promhttp.Handler().ServeHTTP(w, r)
	return nil
}

func (s *APIServer) handleTask(w http.ResponseWriter, r *http.Request) error {
	url := r.URL.Path
	status := http.StatusOK
	s.metrics.uriCounter.WithLabelValues(r.Method, url, strconv.Itoa(status)).Inc()
	start := time.Now()

	defer func() {
		duration := time.Since(start).Milliseconds()
		s.metrics.responseTimeHistogram.WithLabelValues(r.Method, url).Observe(float64(duration))
	}()

	if r.Method == "GET" {
		return s.handleGetAllTask(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateTask(w, r)
	}
	if r.Method == "DELETE" {
		return s.handleDeleteAll(w, r)
	}
	return nil
}

func (s *APIServer) handleTaskByID(w http.ResponseWriter, r *http.Request) error {
	url := r.URL.Path
	status := http.StatusOK
	s.metrics.uriCounter.WithLabelValues(r.Method, url, strconv.Itoa(status)).Inc()
	start := time.Now()

	defer func() {
		duration := time.Since(start).Milliseconds()
		s.metrics.responseTimeHistogram.WithLabelValues(r.Method, url).Observe(float64(duration))
	}()

	if r.Method == "GET" {
		id, err := getID(r)
		if err != nil {
			return err
		}
		task, err := s.store.GetTaskByID(id)
		if err != nil {
			return err
		}
		return WriteJSON(w, http.StatusOK, task)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteTask(w, r)
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleGetAllTask(w http.ResponseWriter, r *http.Request) error {

	update := new(UpdateTaskReq)
	if err := json.NewDecoder(r.Body).Decode(update); err != nil {
		return err
	}
	if update.Update != "" {
		id := update.Update
		err := s.handleUpdateTask(w, r, id)
		if err != nil {
			return err
		}
	}

	tasks, err := s.store.GetTasks()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, tasks)
}

func (s *APIServer) handleCreateTask(w http.ResponseWriter, r *http.Request) error {
	createTaskReq := new(CreateTaskReqest)
	if err := json.NewDecoder(r.Body).Decode(createTaskReq); err != nil {
		return err
	}

	task := NewTask(createTaskReq.TaskDesc)

	if err := s.store.CreateTask(task); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, task)
}

func (s *APIServer) handleUpdateTask(w http.ResponseWriter, r *http.Request, idStr string) error {
	if r.Method == "GET" {
		//id, err := getID(r)
		//if err != nil {
		//	return err
		//}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return fmt.Errorf("invalid id given %s", idStr)
		}

		if err := s.store.UpdateTask(id); err != nil {
			return err
		}
		return nil
		//return WriteJSON(w, http.StatusOK, map[string]int{"changed ": id})
	}

	return fmt.Errorf("method not allowed %s", r.Method)
}

func (s *APIServer) handleDeleteTask(w http.ResponseWriter, r *http.Request) error {

	id, err := getID(r)
	if err != nil {
		return err
	}

	if err := s.store.DeleteTask(id); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int{"deleted: ": id})
}

func (s *APIServer) handleDeleteAll(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}
