package controllers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/miloszizic/habits/views"

	"github.com/miloszizic/habits/store"
)

type Handlers struct {
	Store     store.HabitStore
	Templates struct {
		New Template
	}
}

func (h Handlers) Home(w http.ResponseWriter, r *http.Request) {
	habits, err := h.Store.AllHabits()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	h.Templates.New.Execute(w, habits)

}

func (h Handlers) Habit(w http.ResponseWriter, r *http.Request) {
	h.Templates.New.Execute(w, nil)
}
func (h Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	habitName := r.FormValue("name")
	habit := store.Habit{Name: habitName}
	exist, err := h.Store.GetHabit(habitName)
	if err != nil {
		if err != sql.ErrNoRows {
			vd.Alert = &views.Alert{
				Level:   views.AlertLvlSuccess,
				Message: fmt.Sprintf("You successfully created a %s Habit", habitName),
			}
			h.Store.Add(habit)
			h.Templates.New.Execute(w, vd)
			return
		} else {
			vd.Alert = &views.Alert{
				Level:   views.AlertLvlError,
				Message: err.Error(),
			}
			h.Templates.New.Execute(w, vd)
			return
		}
	}
	if exist != nil {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlError,
			Message: "Habit already exists",
		}
		h.Templates.New.Execute(w, vd)
	}
}
