package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/", MainPage)
	router.HandleFunc("/auth/", AuthPage)
	router.HandleFunc("/lk/generate/", GeneratePage)
	router.HandleFunc("/lk/list/", ListPage)

	router.HandleFunc("/logo/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/static/logo.png")
	})

	server := &http.Server{
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		Addr:         os.Getenv("HOST") + ":" + os.Getenv("PORT"),
		Handler:      router,
	}

	server.ListenAndServe()
}

func MainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/html/index.html")
}

func AuthPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/html/auth.html")
}

func GeneratePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/html/generate.html")
}

func ListPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/html/list.html")
}
