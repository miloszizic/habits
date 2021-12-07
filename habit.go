package habits

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	sqlSchema = `
		CREATE TABLE IF NOT EXISTS "habits" (
	   		"ID" INTEGER PRIMARY KEY AUTOINCREMENT,
			"name" TEXT NOT NULL,
			"last_check" DATETIME NOT NULL,
			"streak" INTEGER
	);
	`
	sqlGetAll    = `SELECT ID, name, last_check, streak FROM habits`
	sqlGetOne    = `SELECT ID, name, last_check, streak FROM habits WHERE name=?;`
	sqlBreak     = `UPDATE habits set last_check=?,streak=1 WHERE name=?`
	sqlYesterday = `UPDATE habits set last_check=?,streak=? WHERE name=?`
)

// Habit struct has all habit attributes
type Habit struct {
	ID            int
	Name          string
	LastPerformed time.Time
	Streak        int
	Output        io.Writer
}

// Store is a struct of all Store properties
type Store struct {
	Habits []Habit
	Output io.Writer
	DB     *sql.DB
	Now    func() time.Time
}

// FromSQLite is checking for scheme to prepare it, if it doesn't exist
// and returns a store with connection
func FromSQLite(dbFIle string) *Store {
	db, _ := sql.Open("sqlite3", dbFIle)
	stmt, err := db.Prepare(sqlSchema)
	if err != nil {
		fmt.Printf("failed to prepare schema with error: %v\n", err)
	}
	_, err = stmt.Exec()
	if err != nil {
		fmt.Printf("failed to execute schema with error: %v\n", err)
	}
	return &Store{
		DB: db,
		Now: func() time.Time {
			return time.Now()
		},
	}
}

// Print as Store method is wrapping Fprintf so that is not needed to specify
// the default output every time
func (s Store) Print(massage string, params ...interface{}) {
	if s.Output == nil {
		fmt.Fprintf(os.Stdout, massage, params...)
	} else {
		fmt.Fprintf(s.Output, massage, params...)
	}
}

// LastCheckDays method checks  for number of days current date and
func (s Store) LastCheckDays(h Habit) int {
	lastPerformedCalendarDay := h.LastPerformed.Truncate(24 * time.Hour)
	nowCalendarDay := s.Now().Truncate(24 * time.Hour)
	return int(nowCalendarDay.Sub(lastPerformedCalendarDay).Hours()) / 24
}

// Add method is adding a habit to the table of Habits
func (s *Store) Add(habit Habit) {
	_, err := s.DB.Exec(
		`INSERT INTO habits (name, last_check, streak) VALUES (?,?,?)`,
		habit.Name,
		habit.LastPerformed,
		habit.Streak,
	)
	if err != nil {
		fmt.Printf("execute failed: %v", err)
	}
	s.Print("Good luck with your new '%s' habit. Don't forget to do it again tomorrow.", habit.Name)
}

// GetHabit takes habit name and returns a habit if it finds one
func (s *Store) GetHabit(name string) (Habit, bool) {
	row := s.DB.QueryRow(sqlGetOne, name)
	h := Habit{}
	var b bool
	err := row.Scan(&h.ID, &h.Name, &h.LastPerformed, &h.Streak)
	if errors.Is(err, sql.ErrNoRows) {
		return Habit{}, b
	}
	return h, true
}

// AllHabits lists all Habits in the database
func (s *Store) AllHabits() []Habit {
	habits := []Habit{}
	rows, err := s.DB.Query(sqlGetAll)
	if err != nil {
		fmt.Printf("query error: %v\n", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Printf("error closing rows: %v\n", err)
		}
	}(rows)
	habit := Habit{}
	for rows.Next() {
		err := rows.Scan(&habit.ID, &habit.Name, &habit.LastPerformed, &habit.Streak)
		if err != nil {
			fmt.Printf("scan error: %v\n", err)
		}
		habits = append(habits, habit)
	}
	return habits
}

// Perform changes the last checked date
func (s *Store) Perform(habit Habit) {
	if s.LastCheckDays(habit) > 1 {
		habit.Streak = 1
	} else {
		habit.Streak++
	}
	_, err := s.DB.Exec(sqlYesterday, s.Now(), habit.Streak, habit.Name)
	if err != nil {
		fmt.Printf(" failed to execute last checked date and streak on habit with error: %v\n", err)
	}
}

// PerformHabit makes a dissection based on days between current time and last checked date
func (s *Store) PerformHabit(h *Habit, days int) {
	switch {
	case days == 0:
		s.Print("Nice work: you've done the habit '%s' for %v days in a row Now.\n", h.Name, h.Streak+1)
	case days == 1 && h.Streak > 15:
		s.Perform(*h)
		s.Print("You're currently on a %d-day streak for '%s'. Stick to it!\n", h.Streak+1, h.Name)
	case days == 1:
		s.Perform(*h)
		s.Print("Nice work: you've done the habit '%s' for %v days in a row Now.\n", h.Name, h.Streak+1)
	case days >= 2:
		s.Perform(*h)
		s.Print("You last did the habit '%s' %d days ago, so you're starting a new streak today. Good luck!\n", h.Name, days)
	}
}
