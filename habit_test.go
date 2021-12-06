package habits_test

import (
	"io"
	"io/ioutil"
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

//tempFile makes a temp file for the database
func tempFile() string {
	file, _ := ioutil.TempFile("", "*.db")
	//require.NoError(err)
	file.Close()
	return file.Name()
}

func TestLastCheckDays(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := t.TempDir() + "test.db"
	t.Logf("db file: %s", dbFile) //for checking the temp name
	//Making a store
	store := habits.FromSQLite(dbFile)
	dbTest.Output = io.Discard
	dbTest.Seed(seedData)
	// Testing
	want := 3
	habit, _ := dbTest.GetOne("piano")
	got := habit.LastCheckDays(Now())
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestAdd(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := tempFile()
	//Making a store
	dbTest := habits.FromSQLite(dbFile)
	dbTest.Output = io.Discard
	// Testing
	want := "piano"
	dbTest.Add(want)
	dbHabit, _ := dbTest.GetOne("piano")
	got := dbHabit.Name
	if got != want {
		t.Errorf("expected %q, got %q instead.", want, got)
	}

}
func TestGetOne(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := tempFile()
	//Making a store
	dbTest := habits.FromSQLite(dbFile)
	dbTest.Output = io.Discard
	dbTest.Seed(seedData)
	// Testing
	want := habits.Habit{ID: 4, Name: "Go", LastCheck: dayBefore, Streak: 4}
	got, _ := dbTest.GetOne("Go")
	if got != want {
		t.Errorf("expected %v, got %v instead.", want, got)
	}
}

//
func TestGetOneInvalid(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := tempFile()
	//Making a store
	dbTest := habits.FromSQLite(dbFile)
	dbTest.Output = io.Discard
	dbTest.Seed(seedData)
	// Testing
	habit, found := dbTest.GetOne("Biking")
	if found {
		t.Errorf("expected nil value for habit got %v", habit)
	}
}

func TestGetAllSeed(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := tempFile()
	//Making a store
	dbTest := habits.FromSQLite(dbFile)
	dbTest.Output = io.Discard
	dbTest.Seed(seedData)
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
	got := dbTest.GetAll()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestBreak(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := tempFile()
	//Making a store
	dbTest := habits.FromSQLite(dbFile)
	dbTest.Output = io.Discard
	dbTest.Seed(seedData)
	// Testing
	want := habits.Habit{ID: 4, Name: "Go", LastCheck: Now(), Streak: 1}
	habit, _ := dbTest.GetOne("Go")
	dbTest.Break(habit, Now())
	got, _ := dbTest.GetOne("Go")
	if got != want {
		t.Errorf("expected %v, got %v instead.", want, got)
	}
}
func TestUpdateYesterday(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := tempFile()
	//Making a store
	dbTest := habits.FromSQLite(dbFile)
	dbTest.Output = io.Discard
	dbTest.Seed(seedData)
	// Testing
	want := habits.Habit{ID: 4, Name: "Go", LastCheck: Now(), Streak: 5}
	habit, _ := dbTest.GetOne("Go")
	dbTest.UpdateYesterday(habit, Now())
	got, _ := dbTest.GetOne("Go")
	if got != want {
		t.Errorf("expected %v, got %v instead.", want, got)
	}
}

func TestDone(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := tempFile()
	//Making a store
	dbTest := habits.FromSQLite(dbFile)
	dbTest.Output = io.Discard
	dbTest.Seed(seedData)
	// Testing
	want := habits.Habit{ID: 4, Name: "Go", LastCheck: Now(), Streak: 4, Done: true}
	habit, _ := dbTest.GetOne("Go")
	dbTest.Done(habit, Now())
	got, _ := dbTest.GetOne("Go")
	if got != want {
		t.Errorf("expected %v, got %v instead.", want, got)
	}
}

func TestDecisionsHandler(t *testing.T) {
	t.Parallel()
	//Making temp db
	dbFile := tempFile()
	//Making a store
	dbTest := habits.FromSQLite(dbFile)
	dbTest.Output = io.Discard
	dbTest.Seed(seedData)
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
		habit, _ := dbTest.GetOne(habitName)
		days := habit.LastCheckDays(Now())
		dbTest.DecisionsHandler(&habit, days, Now())
	}
	got := dbTest.GetAll()

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}
