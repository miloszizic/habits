package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/miloszizic/habits/templates"

	"github.com/miloszizic/habits/store"
	"github.com/miloszizic/habits/views"
)

type Server struct {
	Store     store.HabitStore
	Templates struct {
		New Template
	}
}

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

func (s Server) Habit(w http.ResponseWriter, _ *http.Request) {
	s.Templates.New = views.Must(views.ParseFS(templates.Files, "habit.gohtml", "*.layout.gohtml"))
	s.Templates.New.Execute(w, nil)
}
func (s Server) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	habitName := r.FormValue("name")
	habit := store.Habit{Name: habitName}
	s.Templates.New = views.Must(views.ParseFS(templates.Files, "habit.gohtml", "*.layout.gohtml"))
	exist, err := s.Store.GetHabit(habitName)
	if errors.Unwrap(err) == sql.ErrNoRows || exist == nil {
		vd.Alert = &views.Alert{
			Color:   views.AlertLvlSuccess,
			Message: fmt.Sprintf("You successfully created a %s Habit", habitName),
		}
		s.Store.Add(habit)
		s.Templates.New.Execute(w, vd)
	}
	if exist != nil {
		vd.Alert = &views.Alert{
			Color:   views.AlertLvlError,
			Message: "Habit already exists",
		}
		s.Templates.New.Execute(w, vd)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
func (s *Server) Delete(w http.ResponseWriter, r *http.Request) {
	habitName := r.FormValue("delete")
	s.Store.DeleteHabitByName(habitName)
	habits, err := s.Store.AllHabits()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.Templates.New = views.Must(views.ParseFS(templates.Files, "home.gohtml", "*.layout.gohtml"))
	s.Templates.New.Execute(w, habits)
}
func (s *Server) PerformHabit(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	habitName := r.FormValue("perform")
	habit, err := s.Store.GetHabit(habitName)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	days := s.Store.LastCheckDays(*habit)
	massage := s.Store.PerformHabit(*habit, days)
	vd.Alert = &views.Alert{
		Color:   views.AlertLvlNeutral,
		Message: massage,
	}
	s.Templates.New = views.Must(views.ParseFS(templates.Files, "perform.gohtml", "*.layout.gohtml"))
	s.Templates.New.Execute(w, vd)
}
