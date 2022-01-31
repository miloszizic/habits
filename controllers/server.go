package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/miloszizic/habits/templates"

	"github.com/miloszizic/habits/store"
	"github.com/miloszizic/habits/views"
)

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
	// THe http server
	server := &http.Server{Addr: ":3000", Handler: service()}
	fmt.Printf("started habit service on port %v\n", server.Addr)
	//// Trying to set k8s core maxprocs
	//if _, err := maxprocs.Set(); err != nil {
	//	fmt.Println("maxprocs: %w", err)
	//}
	//g := runtime.GOMAXPROCS(0)
	//fmt.Printf("starting habit service with build: [%s] and [%d] of available CPU cores.\n", build, g)

	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// Run the server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()

}

func service() http.Handler {
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

	return r
}
