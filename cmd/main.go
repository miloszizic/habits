package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/miloszizic/habits/controllers"
	"github.com/miloszizic/habits/store"
)

func main() {

	store, err := store.FromSQLite("./habits.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "opening %q database: %v\n", err, store)
		os.Exit(1)
	}
	r := chi.NewRouter()
	srv := controllers.Server{Store: store}

	r.Get("/", srv.Home)

	r.Get("/habit", srv.Habit)
	r.Post("/habit", srv.Create)

	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)

}
