package habits_test

import (
	"database/sql"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/miloszizic/habits"
)

//Mocks time.Now method for testing purposes
var Now = func() time.Time {
	return time.Date(2021, 12, 15, 17, 8, 0, 0, time.UTC)
}

//vars for testing purposes
var sameDay = time.Date(2021, 12, 15, 17, 8, 0, 0, time.UTC)
var threeDays = time.Date(2021, 12, 11, 18, 9, 0, 0, time.UTC)
var dayBefore = time.Date(2021, 12, 14, 15, 9, 0, 0, time.UTC)

var seedData = []habits.Habit{
	{Name: "k8s", LastCheck: sameDay, Streak: 4},
	{Name: "piano", LastCheck: threeDays, Streak: 4},
	{Name: "code", LastCheck: threeDays, Streak: 4},
	{Name: "Go", LastCheck: dayBefore, Streak: 4},
	{Name: "docker", LastCheck: dayBefore, Streak: 16},
	{Name: "SQL", LastCheck: dayBefore, Streak: 29, Done: false},
	{Name: "NoSQL", LastCheck: sameDay, Streak: 30, Done: true},
}

const sqlInsert = `INSERT INTO habits (name, last_check, streak, done) VALUES (?,?,?,?)`

//Seed is adding testing data to the database
func Seed(db *sql.DB, h []habits.Habit) {
	var doneI int
	for _, v := range h {
		if v.Done {
			doneI = 1
		} else {
			doneI = 0
		}
		_, err := db.Exec(sqlInsert, v.Name, v.LastCheck, v.Streak, doneI)
		if err != nil {
			fmt.Printf("seed execute failed: %v", err)
		}
	}
}

func TestLastCheckDays(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := t.TempDir() + "test.db"
	t.Logf("db file: %s", dbFile) //for checking the temp name
	//Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	Seed(store.DB, seedData)
	// Testing
	want := 3
	habit, _ := store.GetHabit("piano")
	got := habit.LastCheckDays(Now())
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := t.TempDir() + "test.db"
	//Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	// Testing
	want := "piano"
	store.Add(want)
	dbHabit, _ := store.GetHabit("piano")
	got := dbHabit.Name
	if got != want {
		t.Errorf("expected %q, got %q instead.", want, got)
	}

}
func TestGetOne(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := t.TempDir() + "test.db"
	//Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	Seed(store.DB, seedData)
	// Testing
	want := habits.Habit{ID: 4, Name: "Go", LastCheck: dayBefore, Streak: 4}
	got, _ := store.GetHabit("Go")
	if got != want {
		t.Errorf("expected %v, got %v instead.", want, got)
	}
}

//
func TestGetOneInvalid(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := t.TempDir() + "test.db"
	//Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	Seed(store.DB, seedData)
	// Testing
	habit, found := store.GetHabit("Biking")
	if found {
		t.Errorf("expected nil value for habit got %v", habit)
	}
}

func TestGetAllSeed(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := t.TempDir() + "test.db"
	//Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	Seed(store.DB, seedData)
	// Testing
	want := []habits.Habit{
		{ID: 1, Name: "k8s", LastCheck: sameDay, Streak: 4},
		{ID: 2, Name: "piano", LastCheck: threeDays, Streak: 4},
		{ID: 3, Name: "code", LastCheck: threeDays, Streak: 4},
		{ID: 4, Name: "Go", LastCheck: dayBefore, Streak: 4},
		{ID: 5, Name: "docker", LastCheck: dayBefore, Streak: 16},
		{ID: 6, Name: "SQL", LastCheck: dayBefore, Streak: 29, Done: false},
		{ID: 7, Name: "NoSQL", LastCheck: sameDay, Streak: 30, Done: true},
	}
	got := store.AllHabits()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestBreak(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := t.TempDir() + "test.db"
	//Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	Seed(store.DB, seedData)
	// Testing
	want := habits.Habit{ID: 4, Name: "Go", LastCheck: Now(), Streak: 1}
	habit, _ := store.GetHabit("Go")
	store.Break(habit, Now())
	got, _ := store.GetHabit("Go")
	if got != want {
		t.Errorf("expected %v, got %v instead.", want, got)
	}
}
func TestUpdateYesterday(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := t.TempDir() + "test.db"
	//Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	Seed(store.DB, seedData)
	// Testing
	want := habits.Habit{ID: 4, Name: "Go", LastCheck: Now(), Streak: 5}
	habit, _ := store.GetHabit("Go")
	store.UpdateYesterday(habit, Now())
	got, _ := store.GetHabit("Go")
	if got != want {
		t.Errorf("expected %v, got %v instead.", want, got)
	}
}

func TestDone(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := t.TempDir() + "test.db"
	//Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	Seed(store.DB, seedData)
	// Testing
	want := habits.Habit{ID: 4, Name: "Go", LastCheck: Now(), Streak: 4, Done: true}
	habit, _ := store.GetHabit("Go")
	store.Done(habit, Now())
	got, _ := store.GetHabit("Go")
	if got != want {
		t.Errorf("expected %v, got %v instead.", want, got)
	}
}

func TestDecisionsHandler(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := t.TempDir() + "test.db"
	//Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	Seed(store.DB, seedData)
	//Testing
	want := []habits.Habit{
		{ID: 1, Name: "k8s", LastCheck: Now(), Streak: 4},
		{ID: 2, Name: "piano", LastCheck: Now(), Streak: 1},
		{ID: 3, Name: "code", LastCheck: Now(), Streak: 1},
		{ID: 4, Name: "Go", LastCheck: Now(), Streak: 5},
		{ID: 5, Name: "docker", LastCheck: Now(), Streak: 17},
		{ID: 6, Name: "SQL", LastCheck: Now(), Streak: 30, Done: true},
		{ID: 7, Name: "NoSQL", LastCheck: sameDay, Streak: 30, Done: true},
	}
	habitNames := []string{"k8s", "piano", "code", "Go", "docker", "SQL", "NoSQL"}

	for _, habitName := range habitNames {
		habit, _ := store.GetHabit(habitName)
		days := habit.LastCheckDays(Now())
		store.PerformHabit(&habit, days, Now())
	}
	got := store.AllHabits()

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
