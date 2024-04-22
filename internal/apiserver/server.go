// Package apiserver provides a HTTP API server.
//
// It implements CRUD operations for managing numbers.
//
//	Schemes: http
//	Host: localhost:8080
//	BasePath: /
//	Version: 1.0.0
//	License: MIT http://opensource.org/licenses/MIT
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
package apiserver

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"sync"
	"tz/internal/model"
	"tz/internal/store"
)

// Server represents the HTTP API server.
type Server struct {
	router *mux.Router
	logger *slog.Logger
	store  store.Store
}

func NewServer(lg *slog.Logger, st store.Store) *Server {
	s := &Server{
		router: mux.NewRouter(),
		logger: lg,
		store:  st,
	}
	s.configureStore()
	s.configureRouter()
	return s
}

func (s *Server) configureStore() {
	err := s.store.Migrate()
	if err != nil {
		s.logger.Error("Failed to migrate store", "error", err)
	} else {
		s.logger.Debug("Successfully migrated store")
	}
}

func (s *Server) configureRouter() {
	s.router.HandleFunc("/create", s.handleCreate).Methods("POST")
	s.router.HandleFunc("/", s.handleStart).Methods("GET")
	s.router.HandleFunc("/delete", s.handleDelete).Methods("POST")
	s.router.HandleFunc("/list", s.handleList).Methods("GET")
	s.router.HandleFunc("/update", s.handleUpdate).Methods("POST")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	s.respond(w, r, http.StatusOK, "Start page")
}

// handleCreate handles the creation of numbers.
//
// swagger:operation POST /create
//
// ---
// produces:
// - application/json
// parameters:
//   - name: regNums
//     in: body
//     description: Numbers to be created.
//     required: true
//     schema:
//     type: array
//     items:
//     type: string
//
// responses:
//
//	'200':
//	  description: Number(s) created successfully.
//	'400':
//	  description: Invalid request payload.
//	'500':
//	  description: Internal server error.
func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Num []string `json:"regNums"`
	}
	req := &request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("decode request body error: %v", err)
		s.error(w, r, http.StatusBadRequest, err)
		return
	}
	s.logger.Debug("handleCreate", "request", req)
	n := &model.Number{
		Number: req.Num,
	}

	wg := &sync.WaitGroup{}
	url := "http://api.url/info?regNum="
	for _, v := range req.Num {
		v := v
		go func() {
			wg.Add(1)
			err := s.store.Create(v)
			if err != nil {
				s.logger.Error("create error", "num", v, "err", err)
				wg.Done()
				return
			}
			s.logger.Info("create number success", "num", v)

			//Задание 2+3
			response, err := http.Get(url + v)
			if err != nil {
				s.logger.Error("get url %s error: %v", url+v, err)
				wg.Done()
				return
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusOK {
				wg.Done()
				s.logger.Error("create error", "num", v, "enriching error", errors.New("response error"))
				return
			}
			info := make(map[string]interface{})
			err = json.NewDecoder(response.Body).Decode(&info)
			if err != nil {
				s.logger.Error("decode response body error: %v", err)
				wg.Done()
				return
			}
			err = s.store.CreateEnriched(info)
			if err != nil {
				s.logger.Error("enriching error", "error", err)
				wg.Done()
				return
			}
			s.logger.Info("enriching success", "num", v)
			wg.Done()
			//
		}()
	}

	wg.Wait()
	s.respond(w, r, http.StatusOK, n)
}

// handleDelete handles the deletion of numbers.
//
// swagger:operation POST /delete
//
// ---
// produces:
// - application/json
// parameters:
//   - name: ids
//     in: body
//     description: IDs of numbers to be deleted.
//     required: true
//     schema:
//     type: array
//     items:
//     type: string
//
// responses:
//
//	'200':
//	  description: Numbers deleted successfully.
//	'400':
//	  description: Invalid request payload.
//	'500':
//	  description: Internal server error.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Ids []string `json:"ids"`
	}
	req := &request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("decode request body error: %v", err)
		s.error(w, r, http.StatusBadRequest, err)
		return
	}
	s.logger.Debug("handleDelete", "request", req)
	wg := &sync.WaitGroup{}
	for _, v := range req.Ids {
		v := v
		go func() {
			wg.Add(1)
			err := s.store.Delete(v)
			if err != nil {
				s.logger.Error("delete %s error: %v", v, err)
				s.error(w, r, http.StatusInternalServerError, err)
				wg.Done()
				return
			}
			wg.Done()
			s.logger.Info("delete %s success", v)
		}()
	}

	wg.Wait()
	s.respond(w, r, http.StatusOK, req.Ids)
}

// handleList handles the retrieval of numbers based on different criteria.
//
// swagger:operation GET /list
//
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: query
//     description: ID of the number to retrieve.
//     type: string
//   - name: prefix
//     in: query
//     description: Prefix of the numbers to retrieve.
//     type: string
//   - name: region
//     in: query
//     description: Region of the numbers to retrieve.
//     type: string
//
// responses:
//
//	'200':
//	  description: Number(s) retrieved successfully.
//	'400':
//	  description: Invalid request parameters.
//	'500':
//	  description: Internal server error.
func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	prefix := r.URL.Query().Get("prefix")
	region := r.URL.Query().Get("region")
	s.logger.Debug("handleList", "id", id, "prefix", prefix, "region", region)

	if id != "" {
		number, err := s.store.GetById(id)
		if err != nil {
			s.logger.Error("get[GetById] %s error: %v", id, err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		s.respond(w, r, http.StatusOK, number)
		s.logger.Info("get[GetById] %s success", id)
		return
	} else if prefix != "" {
		numbers, err := s.store.GetByPrefix(prefix)
		if err != nil {
			s.logger.Error("get[GetByPrefix] %s error: %v", prefix, err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		s.respond(w, r, http.StatusOK, numbers)
		s.logger.Info("get[GetByPrefix] success", "prefix", prefix)
		return
	} else if region != "" {
		numbers, err := s.store.GetByRegion(region)
		if err != nil {
			s.logger.Error("get[GetByRegion] %s error: %v", region, err)
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		s.respond(w, r, http.StatusOK, numbers)
		s.logger.Info("get[GetByRegion] success", "prefix", region)
		return
	}

	err := errors.New("no args respond")
	s.error(w, r, http.StatusBadRequest, err)
	s.logger.Error("get error: %v", err)
}

// handleUpdate handles the update of a number.
//
// swagger:operation POST /update
//
// ---
// produces:
// - application/json
// parameters:
//   - name: body
//     in: body
//     description: Request body containing ID of the number to update and its new value.
//     required: true
//     schema:
//     "$ref": "#/definitions/UpdateRequest"
//
// responses:
//
//	'200':
//	  description: Number updated successfully.
//	'400':
//	  description: Invalid request payload.
//	'500':
//	  description: Internal server error.
func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Id     string `json:"id"`
		NewNum string `json:"newNum"`
	}
	req := &request{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("decode request body error: %v", err)
		s.error(w, r, http.StatusBadRequest, err)
		return
	}
	s.logger.Debug("handleUpdate", "request", req)

	id, err := s.store.Update(req.Id, req.NewNum)
	if err != nil {
		s.logger.Error("update error: %v", err)
		s.error(w, r, http.StatusInternalServerError, errors.New("update error"))
		return
	}

	s.logger.Info("update success", "id", id)
	s.respond(w, r, http.StatusOK, id)
}
