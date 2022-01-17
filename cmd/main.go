package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/miloszizic/habits/templates"

	"github.com/miloszizic/habits/views"

	"github.com/miloszizic/habits/controllers"

	"github.com/miloszizic/habits/store"

	"github.com/go-chi/chi/v5"
)

func main() {
	//habits.RunCLI()
	store, err := store.FromSQLite("./habits.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "opening %q database: %v\n", err, store)
		os.Exit(1)
	}
	r := chi.NewRouter()
	handler := controllers.Handlers{Store: store}

	handler.Templates.New, err = views.ParseFS(templates.Files, "home.gohtml", "tailwind.gohtml", "alerts.gohtml")
	if err != nil {
		fmt.Errorf("error parsing %w", err)
	}
	r.Get("/", handler.Home)
	handler.Templates.New, err = views.ParseFS(templates.Files, "habit.gohtml", "tailwind.gohtml", "alerts.gohtml")
	if err != nil {
		fmt.Errorf("error parsing %w", err)
	}
	r.Get("/habit", handler.Habit)

	handler.Templates.New, err = views.ParseFS(templates.Files, "submit.gohtml", "tailwind.gohtml", "alerts.gohtml")
	if err != nil {
		fmt.Errorf("error parsing %w", err)
	}
	r.Post("/create", handler.Create)

	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
