package transport

import (
	"alexdenkk/labs/internal/lab/domain"
	"alexdenkk/labs/pkg/middleware"
	"alexdenkk/labs/pkg/token/jwt"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
)

type API struct {
	service    domain.LabService
	middleware *middleware.Middleware
}

func New(
	service domain.LabService,
	mw *middleware.Middleware,
) *API {
	return &API{
		service:    service,
		middleware: mw,
	}
}

func (api *API) RegisterEndpoints(r *mux.Router) {
	sub := r.PathPrefix("/lab").Subrouter()

	sub.Handle("/self/", api.middleware.Authorization(http.HandlerFunc(api.GetForSelf))).Methods(http.MethodGet)
	sub.Handle("/{id}/", api.middleware.Authorization(http.HandlerFunc(api.Get))).Methods(http.MethodGet)
	sub.Handle("/", api.middleware.Authorization(http.HandlerFunc(api.Generate))).Methods(http.MethodPost)
	sub.Handle("/{id}/", api.middleware.Authorization(http.HandlerFunc(api.Delete))).Methods(http.MethodDelete)
}

func (api *API) GetForSelf(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*jwt.Claims)

	labs, err := api.service.GetForSelf(r.Context(), claims)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(labs)
}

func (api *API) Get(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*jwt.Claims)

	id, err := uuid.Parse(mux.Vars(r)["id"])

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	lab, err := api.service.Get(r.Context(), id, claims)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lab)
}

func (api *API) Generate(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*jwt.Claims)

	var lab domain.Lab

	if err := json.NewDecoder(r.Body).Decode(&lab); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	err := api.service.Generate(r.Context(), lab, claims)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "lab generated",
	})
}

func (api *API) Delete(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*jwt.Claims)

	id, err := uuid.Parse(mux.Vars(r)["id"])

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	err = api.service.Delete(r.Context(), id, claims)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "lab deleted",
	})
}
