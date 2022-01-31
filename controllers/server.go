package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/go-chi/chi/v5"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/miloszizic/habits/templates"

	"github.com/miloszizic/habits/store"
	"github.com/miloszizic/habits/views"
)

var build = "develop"

type Server struct {
	Store     store.HabitStore
	Templates struct {
		New Template
	}
	Data views.Data
}

// Home handler is handling the home page
func (s Server) Home(w http.ResponseWriter, _ *http.Request) {
	habits, err := s.Store.AllHabits()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.Templates.New = views.Must(views.ParseFS(templates.Files, "home.gohtml", "*.layout.gohtml"))
	s.Templates.New.Execute(w, habits)

}

// Habit handler handles the get method for add habit page
func (s Server) Habit(w http.ResponseWriter, _ *http.Request) {
	s.Templates.New = views.Must(views.ParseFS(templates.Files, "habit.gohtml", "*.layout.gohtml"))
	s.Templates.New.Execute(w, nil)
}

// Create handler creates new habit or files with user alert
func (s Server) Create(w http.ResponseWriter, r *http.Request) {
	habitName := r.FormValue("name")
	habit := store.Habit{Name: habitName}
	s.Templates.New = views.Must(views.ParseFS(templates.Files, "habit.gohtml", "*.layout.gohtml"))
	exist, err := s.Store.GetHabit(habitName)
	if errors.Unwrap(err) == sql.ErrNoRows || exist == nil {
		s.Data.Alert = &views.Alert{
			Color:   views.AlertLvlSuccess,
			Message: fmt.Sprintf("You successfully created a %s Habit", habitName),
		}
		s.Store.Add(habit)
		s.Templates.New.Execute(w, s.Data)
	}
	if exist != nil {
		s.Data.Alert = &views.Alert{
			Color:   views.AlertLvlError,
			Message: "Habit already exists",
		}
		s.Templates.New.Execute(w, s.Data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// Delete handler deletes the habit
func (s *Server) Delete(w http.ResponseWriter, r *http.Request) {
	habitName := r.FormValue("delete")
	err := s.Store.DeleteHabitByName(habitName)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	habits, err := s.Store.AllHabits()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.Templates.New = views.Must(views.ParseFS(templates.Files, "home.gohtml", "*.layout.gohtml"))
	s.Templates.New.Execute(w, habits)
}

// PerformHabit handler performs the habit and return a massage
func (s *Server) PerformHabit(w http.ResponseWriter, r *http.Request) {
	habitName := r.FormValue("perform")
	habit, err := s.Store.GetHabit(habitName)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	days := s.Store.LastCheckDays(*habit)
	massage := s.Store.PerformHabit(*habit, days)
	s.Data.Alert = &views.Alert{
		Color:   views.AlertLvlNeutral,
		Message: massage,
	}
	s.Templates.New = views.Must(views.ParseFS(templates.Files, "perform.gohtml", "*.layout.gohtml"))
	s.Templates.New.Execute(w, s.Data)
}
func RunHTTP() {

	store, err := store.FromSQLite("./habits.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "opening %q database: %v\n", err, store)
		os.Exit(1)
	}
	r := chi.NewRouter()
	srv := Server{Store: store}

	r.Get("/", srv.Home)
	r.Post("/", srv.Delete)
	r.Post("/perform", srv.PerformHabit)

	r.Get("/habit", srv.Habit)
	r.Post("/habit", srv.Create)

	if _, err := maxprocs.Set(); err != nil {
		fmt.Println("maxprocs: %w", err)
		os.Exit(1)
	}
	g := runtime.GOMAXPROCS(0)
	fmt.Printf("starting habit service with build: [%s] and [%d] of available CPU cores.\n", build, g)
	//defer log.Println("service ended")
	//
	//shutdown := make(chan os.Signal, 1)
	//signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	//<-shutdown
	//log.Println("stopping habit service")
	http.ListenAndServe(":3000", r)
}
