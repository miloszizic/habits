//go:build integration

package store_test

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/miloszizic/habits/store"
)

const (
	defaultMySqlURL  = "tester:secret@tcp(127.0.0.1:3307)/habits?parseTime=true"
	defaultSqliteURL = "test.db"
)

var (
	testMySqlURL       string
	testSqliteURL      string
	today              = fakeNow()
	notQuiteTwoDays    = time.Date(2021, 10, 13, 18, 9, 0, 0, time.UTC)
	dayBeforeYesterday = time.Date(2021, 10, 13, 06, 37, 0, 0, time.UTC)
	yesterday          = time.Date(2021, 10, 14, 15, 9, 0, 0, time.UTC)
	seedData           = []store.Habit{
		{Name: "k8s", LastPerformed: today, Streak: 4},
		{Name: "piano", LastPerformed: dayBeforeYesterday, Streak: 4},
		{Name: "code", LastPerformed: dayBeforeYesterday, Streak: 4},
		{Name: "Go", LastPerformed: yesterday, Streak: 4},
		{Name: "docker", LastPerformed: yesterday, Streak: 16},
		{Name: "SQL", LastPerformed: yesterday, Streak: 29},
		{Name: "NoSQL", LastPerformed: today, Streak: 30},
	}
)

func TestMySQLAndSQLLiteDatabase(t *testing.T) {
	storeMySQL, err := store.FromMySQL(testMySqlURL)
	if err != nil {
		t.Fatalf("FromMySql() err = %v; want %v", err, nil)
	}
	storeSQLite, err := store.FromSQLite(testSqliteURL)
	if err != nil {
		t.Fatalf("FromMySql() err = %v; want %v", err, nil)
	}
	storeMySQL.Output = io.Discard
	storeMySQL.Now = fakeNow()
	defer storeMySQL.Close()
	storeSQLite.Output = io.Discard
	storeSQLite.Now = fakeNow()
	defer storeSQLite.Close()
	tests := map[string]func(*testing.T, *store.DBStore){
		"LastCheckDays":                            testLastCheckDays,
		"AddingAndGettingHabit":                    testAddAndGetOne,
		"PerformIncreasesStreakIfDoneYesterday":    testPerformIncreasesStreakIfDoneYesterday,
		"PerformResetsStreakIfDoneBeforeYesterday": testPerformResetsStreakIfDoneBeforeYesterday,
		"SeedAndPerformHabit":                      testSeedAndPerformHabit,
		"DeleteHabitByName":                        testDeleteHabitByName,
		"GetAllHabits":                             testGetAllHabits,
	}

	for name, tc := range tests {
		resetMySqlDB(t, storeMySQL.DB)
		resetSQLiteDB(t, storeSQLite.DB)
		t.Run(name, func(t *testing.T) {
			tc(t, storeMySQL)
			tc(t, storeSQLite)
		})
	}

}
func testDeleteHabitByName(t *testing.T, storeDB *store.DBStore) {
	storeDB.Add(store.Habit{
		Name: "Go",
	})
	storeDB.DeleteHabitByName("Go")
	_, err := storeDB.GetHabit("Go")
	if errors.Unwrap(err) != sql.ErrNoRows {
		t.Errorf("wanted no rows, got %v", err)
	}
}
func testLastCheckDays(t *testing.T, dbStore *store.DBStore) {
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
		got := dbStore.LastCheckDays(store.Habit{
			LastPerformed: tc.date,
		})
		if got != tc.want {
			t.Errorf("%s: want %v, got %v", tc.date, tc.want, got)
		}
	}
}

func testAddAndGetOne(t *testing.T, dbStore *store.DBStore) {
	dbStore.Add(store.Habit{Name: "CCNA"})
	want := &store.Habit{Name: "CCNA", LastPerformed: today, Streak: 0}
	got, err := dbStore.GetHabit("CCNA")
	if err != nil {
		t.Fatalf("got an error getting Habit: %v", err)
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(store.Habit{}, "ID")) {
		t.Error(cmp.Diff(want, got))
	}
}

func testPerformIncreasesStreakIfDoneYesterday(t *testing.T, dbStore *store.DBStore) {
	dbStore.Add(store.Habit{
		Name:          "Sqlite3",
		LastPerformed: yesterday,
		Streak:        4,
	})
	habit, err := dbStore.GetHabit("Sqlite3")
	if err != nil {
		t.Error(err)
	}
	dbStore.Perform(*habit)
	updatedHabit, err := dbStore.GetHabit("Sqlite3")
	if err != nil {
		t.Error(err)
	}
	want := 5
	got := updatedHabit.Streak
	if got != want {
		t.Errorf("expected %v, got %v instead.", want, got)
	}
}
func testPerformResetsStreakIfDoneBeforeYesterday(t *testing.T, dbStore *store.DBStore) {
	habit := store.Habit{
		Name:          "Cycling",
		LastPerformed: dayBeforeYesterday,
		Streak:        4,
	}
	dbStore.Add(habit)
	dbStore.Perform(habit)
	updatedHabit, err := dbStore.GetHabit("Cycling")
	if err != nil {
		t.Error(err)
	}
	want := 1
	got := updatedHabit.Streak
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(store.Habit{}, "ID")) {
		t.Error(cmp.Diff(want, got))
	}
}
func testGetAllHabits(t *testing.T, dbStore *store.DBStore) {
	Seed(dbStore.DB, seedData)
	// Testing
	want := []store.Habit{
		{Name: "k8s", LastPerformed: today, Streak: 4},
		{Name: "piano", LastPerformed: dayBeforeYesterday, Streak: 4},
		{Name: "code", LastPerformed: dayBeforeYesterday, Streak: 4},
		{Name: "Go", LastPerformed: yesterday, Streak: 4},
		{Name: "docker", LastPerformed: yesterday, Streak: 16},
		{Name: "SQL", LastPerformed: yesterday, Streak: 29},
		{Name: "NoSQL", LastPerformed: today, Streak: 30},
	}
	got, err := dbStore.AllHabits()
	if err != nil {
		t.Error(err)
	}
	//Todo: figure out why this isn't needed

	//less := func(a, b string) bool { return a < b }
	//differ := cmp.Diff(want, got, cmpopts.SortSlices(less), cmpopts.IgnoreFields(store.Habit{}, "ID"))
	//if cmp.Diff(want, got, cmpopts.SortSlices(less), cmpopts.IgnoreFields(store.Habit{}, "ID")) != "" {
	//	t.Errorf("wanted no difference got: %v,", differ)
	//}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(store.Habit{}, "ID")) {
		t.Error(cmp.Diff(want, got))
	}
}
func testSeedAndPerformHabit(t *testing.T, dbStore *store.DBStore) {
	Seed(dbStore.DB, seedData)
	// Testing
	want := []store.Habit{
		{Name: "k8s", LastPerformed: today, Streak: 4},
		{Name: "piano", LastPerformed: today, Streak: 1},
		{Name: "code", LastPerformed: today, Streak: 1},
		{Name: "Go", LastPerformed: today, Streak: 5},
		{Name: "docker", LastPerformed: today, Streak: 17},
		{Name: "SQL", LastPerformed: today, Streak: 30},
		{Name: "NoSQL", LastPerformed: today, Streak: 30},
	}
	habitNames := []string{"k8s", "piano", "code", "Go", "docker", "SQL", "NoSQL"}

	for _, habitName := range habitNames {
		habit, err := dbStore.GetHabit(habitName)
		if err != nil {
			t.Fatalf("got an error getting Habit: %v", err)
		}
		days := dbStore.LastCheckDays(*habit)
		dbStore.PerformHabit(*habit, days)
	}
	var got []store.Habit
	for _, habitName := range habitNames {
		habit, err := dbStore.GetHabit(habitName)
		if err != nil {
			t.Fatalf("got an error getting Habit: %v", err)
		}
		got = append(got, *habit)
	}
	if !cmp.Equal(want, got, cmpopts.IgnoreFields(store.Habit{}, "ID")) {
		t.Error(cmp.Diff(want, got))
	}
}

// Mocks time.Now method for testing purposes
func fakeNow() time.Time {
	return time.Date(2021, 10, 15, 17, 8, 0, 0, time.UTC)
}

// Seed is adding testing data to the database
func Seed(db *sql.DB, h []store.Habit) {
	for _, v := range h {
		_, err := db.Exec("INSERT INTO habits (name, LastPerformed, streak) VALUES (?,?,?)", v.Name, v.LastPerformed, v.Streak)
		if err != nil {
			fmt.Printf("seed execute failed: %v", err)
		}
	}
}

//init will check environmental variables, if nil it will set the default
func init() {
	testMySqlURL = os.Getenv("MYSQL_URL")
	testSqliteURL = os.Getenv("SQLITE_URL")
	if testMySqlURL == "" {
		testMySqlURL = defaultMySqlURL
	}
	if testSqliteURL == "" {
		testSqliteURL = defaultSqliteURL
	}
	fmt.Println("Using")
}

//resetMySqlDB will clean the content and restart auto-increment
// MySQL database before running the next test
func resetMySqlDB(t *testing.T, sqlDB *sql.DB) {
	_, err := sqlDB.Exec("TRUNCATE TABLE habits")
	if err != nil {
		t.Fatalf("restarting AUTO_INCREMENT failed with err= %v; want nil", err)
	}
}

//resetSQLiteDB will clean the content and restart auto-increment
// Sqlite3 database before running the next test
func resetSQLiteDB(t *testing.T, sqlDB *sql.DB) {
	_, err := sqlDB.Exec("DELETE FROM `sqlite_sequence` WHERE `name` ='habits'")
	if err != nil {
		t.Fatalf("restarting AUTO_INCREMENT failed with err= %v; want nil", err)
	}
	_, err = sqlDB.Exec("DELETE FROM habits")
	if err != nil {
		t.Fatalf("DELETE FROM habits err = %v; want nil", err)
	}
}
