package transport

import (
	"alexdenkk/labs/internal/user/domain"
	"alexdenkk/labs/pkg/middleware"
	"alexdenkk/labs/pkg/token/jwt"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type API struct {
	service    domain.UserService
	middleware *middleware.Middleware
}

func New(
	service domain.UserService,
	mw *middleware.Middleware,
) *API {
	return &API{
		service:    service,
		middleware: mw,
	}
}

func (api *API) RegisterEndpoints(r *mux.Router) {
	sub := r.PathPrefix("/user").Subrouter()

	sub.HandleFunc("/signin/", api.SignIn).Methods(http.MethodPost)
	sub.HandleFunc("/signup/", api.SignUp).Methods(http.MethodPost)
	sub.HandleFunc("/refresh/", api.RefreshToken).Methods(http.MethodPost)
	sub.Handle("/self/", api.middleware.Authorization(http.HandlerFunc(api.GetSelf))).Methods(http.MethodGet)
}

func (api *API) GetSelf(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*jwt.Claims)

	// Вызов сервиса
	user, err := api.service.GetSelf(r.Context(), claims)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (api *API) SignIn(w http.ResponseWriter, r *http.Request) {
	var data map[string]string

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	accessToken, refreshToken, err := api.service.SignIn(r.Context(), data["email"], data["password"])

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (api *API) SignUp(w http.ResponseWriter, r *http.Request) {
	var data map[string]string

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	err := api.service.SignUp(
		r.Context(),
		data["email"],
		data["password"],
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "user signed up",
	})
}

func (api *API) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var data map[string]string

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	accessToken, refreshToken, err := api.service.RefreshToken(r.Context(), data["refresh_token"])

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
