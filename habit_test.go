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

// Mocks time.Now method for testing purposes
var fakeNow = func() time.Time {
	return time.Date(2021, 10, 15, 17, 8, 0, 0, time.UTC)
}

// vars for testing purposes
var (
	today              = fakeNow()
	notQuiteTwoDays    = time.Date(2021, 10, 13, 18, 9, 0, 0, time.UTC)
	dayBeforeYesterday = time.Date(2021, 10, 13, 06, 37, 0, 0, time.UTC)
	yesterday          = time.Date(2021, 10, 14, 15, 9, 0, 0, time.UTC)
)

var seedData = []habits.Habit{
	{Name: "k8s", LastPerformed: today, Streak: 4},
	{Name: "piano", LastPerformed: dayBeforeYesterday, Streak: 4},
	{Name: "code", LastPerformed: dayBeforeYesterday, Streak: 4},
	{Name: "Go", LastPerformed: yesterday, Streak: 4},
	{Name: "docker", LastPerformed: yesterday, Streak: 16},
	{Name: "SQL", LastPerformed: yesterday, Streak: 29},
	{Name: "NoSQL", LastPerformed: today, Streak: 30},
}

const sqlInsert = `INSERT INTO habits (name, LastPerformed, streak) VALUES (?,?,?)`

// Seed is adding testing data to the database
func Seed(db *sql.DB, h []habits.Habit) {
	for _, v := range h {
		_, err := db.Exec(sqlInsert, v.Name, v.LastPerformed, v.Streak)
		if err != nil {
			fmt.Printf("seed execute failed: %v", err)
		}
	}
}

func TestLastCheckDays(t *testing.T) {
	t.Parallel()
	// Making temp db
	dbFile := t.TempDir() + "test.db"
	t.Logf("db file: %s", dbFile) // for checking the temp name
	// Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	store.Now = fakeNow
	tcs := []struct {
		date time.Time
		want int
	}{
		{today, 0},
		{yesterday, 1},
		{notQuiteTwoDays, 2},
		{dayBeforeYesterday, 2},
	}
	for _, tc := range tcs {
		got := store.LastCheckDays(habits.Habit{
			LastPerformed: tc.date,
		})
		if got != tc.want {
			t.Errorf("%s: want %v, got %v", tc.date, tc.want, got)
		}
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()
	// Making temp db
	dbFile := t.TempDir() + "test.db"
	// Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	store.Now = fakeNow
	// Testing

	store.Add(habits.Habit{Name: "piano"})
	want := &habits.Habit{ID: 1, Name: "piano", LastPerformed: fakeNow(), Streak: 0}
	got, err := store.GetHabit("piano")
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestGetOne(t *testing.T) {
	t.Parallel()
	// Making temp db
	dbFile := t.TempDir() + "test.db"
	// Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	Seed(store.DB, seedData)
	// Testing
	want := &habits.Habit{ID: 4, Name: "Go", LastPerformed: yesterday, Streak: 4}
	got, err := store.GetHabit("Go")
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

//
func TestAddingHabitIfNotExiting(t *testing.T) {
	t.Parallel()
	// Making temp db
	dbFile := t.TempDir() + "test.db"
	// Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	store.Now = fakeNow
	// Testing
	_, err := store.GetHabit("Biking")
	if err == nil {
		fmt.Println("searching for non existing record should return an err, but got nil")
	}

}

func TestGetAllSeed(t *testing.T) {
	t.Parallel()
	// Making temp db
	dbFile := t.TempDir() + "test.db"
	// Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	Seed(store.DB, seedData)
	// Testing
	want := []habits.Habit{
		{ID: 1, Name: "k8s", LastPerformed: today, Streak: 4},
		{ID: 2, Name: "piano", LastPerformed: dayBeforeYesterday, Streak: 4},
		{ID: 3, Name: "code", LastPerformed: dayBeforeYesterday, Streak: 4},
		{ID: 4, Name: "Go", LastPerformed: yesterday, Streak: 4},
		{ID: 5, Name: "docker", LastPerformed: yesterday, Streak: 16},
		{ID: 6, Name: "SQL", LastPerformed: yesterday, Streak: 29},
		{ID: 7, Name: "NoSQL", LastPerformed: today, Streak: 30},
	}
	got := store.AllHabits()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestPerformIncreasesStreakIfDoneYesterday(t *testing.T) {
	t.Parallel()
	// Making temp db
	dbFile := t.TempDir() + "test.db"
	// Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	store.Now = fakeNow
	store.Add(habits.Habit{
		Name:          "Go",
		LastPerformed: yesterday,
		Streak:        4,
	})
	habit, err := store.GetHabit("Go")
	if err != nil {
		t.Error(err)
	}
	store.Perform(*habit)
	updatedHabit, err := store.GetHabit("Go")
	if err != nil {
		t.Error(err)
	}
	want := 5
	got := updatedHabit.Streak
	if got != want {
		t.Errorf("expected %v, got %v instead.", want, got)
	}
}

func TestPerformResetsStreakIfDoneBeforeYesterday(t *testing.T) {
	t.Parallel()
	// Making temp db
	dbFile := t.TempDir() + "test.db"
	// Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	habit := habits.Habit{
		Name:          "Go",
		LastPerformed: dayBeforeYesterday,
		Streak:        4,
	}
	store.Add(habit)
	store.Now = fakeNow
	store.Perform(habit)
	updatedHabit, err := store.GetHabit("Go")
	if err != nil {
		t.Error(err)
	}
	want := 1
	got := updatedHabit.Streak
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestPerformHabit(t *testing.T) {
	t.Parallel()
	// Making temp db
	dbFile := t.TempDir() + "test.db"
	// Making a store
	store := habits.FromSQLite(dbFile)
	store.Output = io.Discard
	store.Now = fakeNow
	Seed(store.DB, seedData)
	// Testing
	want := []habits.Habit{
		{ID: 1, Name: "k8s", LastPerformed: today, Streak: 4},
		{ID: 2, Name: "piano", LastPerformed: today, Streak: 1},
		{ID: 3, Name: "code", LastPerformed: today, Streak: 1},
		{ID: 4, Name: "Go", LastPerformed: today, Streak: 5},
		{ID: 5, Name: "docker", LastPerformed: today, Streak: 17},
		{ID: 6, Name: "SQL", LastPerformed: today, Streak: 30},
		{ID: 7, Name: "NoSQL", LastPerformed: today, Streak: 30},
	}
	habitNames := []string{"k8s", "piano", "code", "Go", "docker", "SQL", "NoSQL"}

	for _, habitName := range habitNames {
		habit, err := store.GetHabit(habitName)
		if err != nil {
			t.Error(err)
		}
		days := store.LastCheckDays(*habit)
		store.PerformHabit(*habit, days)
	}
	got := store.AllHabits()

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
