package habits

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	sqlSchema = `
		CREATE TABLE IF NOT EXISTS "habits" (
	   		"ID" INTEGER PRIMARY KEY AUTOINCREMENT,
			"name" TEXT NOT NULL,
			"last_check" DATETIME NOT NULL,
			"streak" INTEGER,
			"done" INTEGER
	);
	`
	sqlInsert    = `INSERT INTO habits (name, last_check, streak, done) VALUES (?,?,?,?)`
	sqlGetAll    = `SELECT ID, name, last_check, streak, done FROM habits`
	sqlGetOne    = `SELECT ID, name, last_check, streak, done FROM habits WHERE name=?;`
	sqlBreak     = `UPDATE habits set last_check=?,streak=1 WHERE name=?`
	sqlYesterday = `UPDATE habits set last_check=?,streak=? WHERE name=?`
	sqlDone      = `UPDATE habits set last_check=?,done=? WHERE name=?`
)

//Now passes the current system time
var Now = time.Now

//Habit struct has all habit attributes
type Habit struct {
	ID        int
	Name      string
	LastCheck time.Time
	Streak    int
	Done      bool
	Output    io.Writer
}

//Store is a struct of all Store properties
type Store struct {
	Habits []Habit
	Output io.Writer
	DB     *sql.DB
}

//FromSQLite is checking for scheme to prepare it, if it doesn't exist
//and returns a store with connection
func FromSQLite(dbFIle string) *Store {
	db, _ := sql.Open("sqlite3", dbFIle)
	stmt, _ := db.Prepare(sqlSchema)
	stmt.Exec()
	return &Store{
		DB: db,
	}
}

//Seed is adding testing data to the database
func (s *Store) Seed(h []Habit) {
	var doneI int
	for _, v := range h {
		if v.Done {
			doneI = 1
		} else {
			doneI = 0
		}
		_, err := s.DB.Exec(sqlInsert, v.Name, v.LastCheck, v.Streak, doneI)
		if err != nil {
			fmt.Printf("seed execute failed: %v", err)
		}
	}
}

//Print as Store method is wrapping Fprintf so that is not needed to specify
//the default output every time
func (s Store) Print(massage string, params ...interface{}) {
	if s.Output == nil {
		fmt.Fprintf(os.Stdout, massage, params...)
	} else {
		fmt.Fprintf(s.Output, massage, params...)
	}
}

//LastCheckDays method checks  for number of days current date and
func (h Habit) LastCheckDays(time time.Time) int {
	days := int(time.Sub(h.LastCheck).Hours() / 24)
	if days >= 0 {
		return days
	}
	return -1
}

//Add method is adding a new habit to the table of Habits
func (s *Store) Add(name string) {
	_, err := s.DB.Exec(sqlInsert, name, Now(), 1, 0)
	if err != nil {
		fmt.Printf("execute failed: %v", err)
	}
	s.Print("Good luck with your new '%s' habit. Don't forget to do it again tomorrow.", name)
}

//GetOne takes habit name and returns a habit if it finds one
func (s *Store) GetOne(name string) (Habit, bool) {
	row := s.DB.QueryRow(sqlGetOne, name)
	h := Habit{}
	var b bool
	err := row.Scan(&h.ID, &h.Name, &h.LastCheck, &h.Streak, &h.Done)
	if errors.Is(err, sql.ErrNoRows) {
		return Habit{}, b
	}
	return h, true
}

//GetAll lists all Habits in the database
func (s *Store) GetAll() []Habit {
	habits := []Habit{}
	rows, err := s.DB.Query(sqlGetAll)
	if err != nil {
		fmt.Printf("query error: %v\n", err)
	}
	defer rows.Close()
	var done int
	habit := Habit{}
	for rows.Next() {
		err := rows.Scan(&habit.ID, &habit.Name, &habit.LastCheck, &habit.Streak, &done)
		if err != nil {
			fmt.Printf("scan error: %v\n", err)
		}
		if done != 1 {
			habit.Done = false
		} else {
			habit.Done = true
		}
		habits = append(habits, habit)
	}
	return habits
}

//Break will restart a streak on a habit based on the LastCheck time
func (s *Store) Break(habit Habit, time time.Time) {
	_, err := s.DB.Exec(sqlBreak, time, habit.Name)
	if err != nil {
		fmt.Printf(" failed to execute brake on habit with error: %v\n", err)
	}
}

//UpdateYesterday changes the last checked date
func (s *Store) UpdateYesterday(habit Habit, time time.Time) {
	habit.Streak++
	_, err := s.DB.Exec(sqlYesterday, time, habit.Streak, habit.Name)
	if err != nil {
		fmt.Printf(" failed to execute last checked date and streak on habit with error: %v\n", err)
	}
}

//Done changes the habit status to done
func (s *Store) Done(habit Habit, time time.Time) {
	_, err := s.DB.Exec(sqlDone, time, 1, habit.Name)
	if err != nil {
		fmt.Printf(" failed to execute done on habit with error: %v\n", err)
	}
}

//DecisionsHandler makes a dissection based on days between current time and last checked date
func (s *Store) DecisionsHandler(h *Habit, days int, time time.Time) {
	switch {
	case days >= 0 && h.Done:
		s.Print("You already finished the %v habit.\n", h.Name)
	case days == 0:
		s.Print("Nice work: you've done the habit '%s' for %v days in a row Now.\n", h.Name, h.Streak+1)
	case days == 1 && h.Streak == 29:
		s.UpdateYesterday(*h, time)
		s.Done(*h, time)
		s.Print("Congratulations, this is your %dth day for '%s' habit. You finished successfully!!\n", h.Streak+1, h.Name)
	case days == 1 && h.Streak > 15:
		s.UpdateYesterday(*h, time)
		s.Print("You're currently on a %d-day streak for '%s'. Stick to it!\n", h.Streak+1, h.Name)
	case days == 1:
		s.UpdateYesterday(*h, time)
		s.Print("Nice work: you've done the habit '%s' for %v days in a row Now.\n", h.Name, h.Streak+1)
	case days >= 3:
		s.Break(*h, time)
		s.Print("You last did the habit '%s' %d days ago, so you're starting a new streak today. Good luck!\n", h.Name, days)
	}
}
