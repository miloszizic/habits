package store

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

//currentTime returns system time
func currentTime() time.Time {
	return time.Now()
}

const (
	sqlite3Schema = `
		CREATE TABLE IF NOT EXISTS "habits" (
	   		"ID" INTEGER PRIMARY KEY AUTOINCREMENT,
			"name" TEXT NOT NULL,
			"LastPerformed" DATETIME NOT NULL,
			"streak" INTEGER
	);`

	mySqlSchema = `
	CREATE TABLE IF NOT EXISTS habits (
  		ID INT PRIMARY KEY NOT NULL AUTO_INCREMENT,
  		name TEXT NOT NULL,
 		LastPerformed DATETIME NOT NULL,
  		streak INT NOT NULL
)
	`
)

type HabitStore interface {
	LastCheckDays(h Habit) int
	Add(habit Habit)
	AllHabits() ([]Habit, error)
	Perform(habit Habit)
	PerformHabit(h Habit, days int)
	GetHabit(name string) (*Habit, error)
}

type DBStore struct {
	Habits []Habit
	Output io.Writer
	DB     *sql.DB
	Now    time.Time
}

func (s *DBStore) Close() {
	s.DB.Close()
}

// FromMySQL  is checking for scheme to prepare it, if it doesn't exist
// and returns a DBStore with connection
func FromMySQL(source string) (*DBStore, error) {
	db, err := sql.Open("mysql", source)
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare(mySqlSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema with error: %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to execute schema with error: %v", err)

	}
	return &DBStore{
		DB:  db,
		Now: currentTime(),
	}, nil
}

// FromSQLite  is checking for scheme to prepare it, if it doesn't exist
// and returns a DBStore with connection
func FromSQLite(source string) (*DBStore, error) {
	db, err := sql.Open("sqlite3", source)
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare(sqlite3Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare schema with error: %v", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("failed to execute schema with error: %v", err)
	}
	return &DBStore{
		DB:  db,
		Now: currentTime(),
	}, nil
}

// Print as DBStore method is wrapping Fprintf so that is not needed to specify
// the default output every time
func (s DBStore) Print(massage string, params ...interface{}) {
	if s.Output == nil {
		fmt.Fprintf(os.Stdout, massage, params...)
	} else {
		fmt.Fprintf(s.Output, massage, params...)
	}
}

// LastCheckDays method checks  for number of days current date and
func (s DBStore) LastCheckDays(h Habit) int {
	lastPerformedCalendarDay := h.LastPerformed.Truncate(24 * time.Hour)
	nowCalendarDay := s.Now.Truncate(24 * time.Hour)
	return int(nowCalendarDay.Sub(lastPerformedCalendarDay).Hours()) / 24
}

// Add method is adding a habit to the table of Habits
func (s *DBStore) Add(habit Habit) {
	_, err := s.DB.Exec(
		`INSERT INTO habits (name, LastPerformed, streak) VALUES (?,?,?)`,
		habit.Name,
		s.Now,
		habit.Streak,
	)
	if err != nil {
		fmt.Printf("execute failed: %v", err)
	}
	s.Print("Good luck with your new '%s' habit. Don't forget to do it again tomorrow.\n", habit.Name)
}

// GetHabit takes habit name and returns a habit if it finds one
func (s *DBStore) GetHabit(name string) (*Habit, error) {
	row := s.DB.QueryRow(`SELECT ID, name, LastPerformed, streak FROM habits WHERE name=?;`, name)
	h := Habit{}
	err := row.Scan(&h.ID, &h.Name, &h.LastPerformed, &h.Streak)
	if err != nil {
		return nil, fmt.Errorf("failed to get Habit with error: %v", err)
	}
	return &h, nil
}

// AllHabits lists all Habits in the database
func (s *DBStore) AllHabits() ([]Habit, error) {
	allHabits := []Habit{}
	rows, err := s.DB.Query(`SELECT ID, name, LastPerformed, streak FROM habits`)
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
		allHabits = append(allHabits, habit)
	}
	return allHabits, nil
}

// Perform changes the last checked date
func (s *DBStore) Perform(habit Habit) {
	if s.LastCheckDays(habit) > 1 {
		habit.Streak = 1
	} else {
		habit.Streak++
	}
	_, err := s.DB.Exec(`UPDATE habits set LastPerformed=?,streak=? WHERE name=?`, s.Now, habit.Streak, habit.Name)
	if err != nil {
		fmt.Printf(" failed to execute last checked date and streak on habit with error: %v\n", err)
	}
}

// PerformHabit makes a dissection based on days between current time and last checked date
func (s *DBStore) PerformHabit(h Habit, days int) {
	switch {
	case days == 0:
		s.Print("Nice work: you've done the habit '%s' for %v days in a row Now.\n", h.Name, h.Streak)
	case days == 1 && h.Streak > 15:
		s.Perform(h)
		s.Print("You're currently on a %d-day streak for '%s'. Stick to it!\n", h.Streak+1, h.Name)
	case days == 1:
		s.Perform(h)
		s.Print("Nice work: you've done the habit '%s' for %v days in a row Now.\n", h.Name, h.Streak+1)
	case days >= 2:
		s.Perform(h)
		s.Print("You last did the habit '%s' %d days ago, so you're starting a new streak today. Good luck!\n", h.Name, days)
	}
}
