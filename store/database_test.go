package store_test

//
//import (
//	"database/sql"
//	"fmt"
//	"io"
//	"os"
//	"testing"
//	"time"
//
//	"github.com/miloszizic/habits"
//
//	"github.com/google/go-cmp/cmp"
//)
//
//const (
//	defaultMySqlURL  = "tester:secret@tcp(127.0.0.1:3307)/habits?parseTime=true"
//	defaultSqliteURL = "test.db"
//)
//
//var (
//	testMySqlURL       string
//	testSqliteURL      string
//	today              = fakeNow()
//	notQuiteTwoDays    = time.Date(2021, 10, 13, 18, 9, 0, 0, time.UTC)
//	dayBeforeYesterday = time.Date(2021, 10, 13, 06, 37, 0, 0, time.UTC)
//	yesterday          = time.Date(2021, 10, 14, 15, 9, 0, 0, time.UTC)
//	seedData           = []habits.Habit{
//		{Name: "k8s", LastPerformed: today, Streak: 4},
//		{Name: "piano", LastPerformed: dayBeforeYesterday, Streak: 4},
//		{Name: "code", LastPerformed: dayBeforeYesterday, Streak: 4},
//		{Name: "Go", LastPerformed: yesterday, Streak: 4},
//		{Name: "docker", LastPerformed: yesterday, Streak: 16},
//		{Name: "SQL", LastPerformed: yesterday, Streak: 29},
//		{Name: "NoSQL", LastPerformed: today, Streak: 30},
//	}
//)
//
//// Mocks time.Now method for testing purposes
//func fakeNow() time.Time {
//	return time.Date(2021, 10, 15, 17, 8, 0, 0, time.UTC)
//}
//
//// Seed is adding testing data to the database
//func Seed(db *sql.DB, h []habits.Habit) {
//	for _, v := range h {
//		_, err := db.Exec("INSERT INTO habits (name, LastPerformed, streak) VALUES (?,?,?)", v.Name, v.LastPerformed, v.Streak)
//		if err != nil {
//			fmt.Printf("seed execute failed: %v", err)
//		}
//	}
//}
//
////init will check environmental variables, if nil it will set the default
//func init() {
//	testMySqlURL = os.Getenv("MYSQL_URL")
//	testSqliteURL = os.Getenv("SQLITE_URL")
//	if testMySqlURL == "" {
//		testMySqlURL = defaultMySqlURL
//	}
//	if testSqliteURL == "" {
//		testSqliteURL = defaultSqliteURL
//	}
//	fmt.Println("Using")
//}
//
////resetMySqlDB will clean the content and restart auto-increment
//// MySQL database before running the next test
//func resetMySqlDB(t *testing.T, sqlDB *sql.DB) {
//	_, err := sqlDB.Exec("TRUNCATE TABLE habits")
//	if err != nil {
//		t.Fatalf("restarting AUTO_INCREMENT failed with err= %v; want nil", err)
//	}
//}
//
////resetSQLiteDB will clean the content and restart auto-increment
//// Sqlite3 database before running the next test
//func resetSQLiteDB(t *testing.T, sqlDB *sql.DB) {
//	_, err := sqlDB.Exec("DELETE FROM `sqlite_sequence` WHERE `name` ='habits'")
//	if err != nil {
//		t.Fatalf("restarting AUTO_INCREMENT failed with err= %v; want nil", err)
//	}
//	_, err = sqlDB.Exec("DELETE FROM habits")
//	if err != nil {
//		t.Fatalf("DELETE FROM habits err = %v; want nil", err)
//	}
//
//}
//func TestSqliteDatabase(t *testing.T) {
//
//	store, err := habits.FromSQLite(t.TempDir() + testSqliteURL)
//	if err != nil {
//		t.Fatalf("FromSQLite() err = %v; want %v", err, nil)
//	}
//	store.Output = io.Discard
//	store.Now = fakeNow()
//	defer store.Close()
//
//	tests := []struct {
//		testName string
//		testCase func(*testing.T, *habits.DBStore)
//	}{
//		{"LastCheckDays", testLastCheckDays},
//		{"AddingHabit", testAdd},
//		{"GettingOneHabit", testGetOne},
//		{"GettingAllHabits", testGetAllSeedHabits},
//		{"PerformIncreasesStreakIfDoneYesterday", testPerformIncreasesStreakIfDoneYesterday},
//		{"PerformResetsStreakIfDoneBeforeYesterday", testPerformResetsStreakIfDoneBeforeYesterday},
//		{"PerformHabit", testPerformHabit},
//	}
//	for _, tc := range tests {
//		t.Run(tc.testName, func(t *testing.T) {
//			resetSQLiteDB(t, store.DB)
//			tc.testCase(t, store)
//		})
//	}
//}
//func TestMySqlDatabase(t *testing.T) {
//	store, err := habits.FromMySQL(testMySqlURL)
//	if err != nil {
//		t.Fatalf("FromMySql() err = %v; want %v", err, nil)
//	}
//	store.Output = io.Discard
//	store.Now = fakeNow()
//	defer store.Close()
//	tests := map[string]func(*testing.T, *habits.DBStore){
//		"LastCheckDays":                            testLastCheckDays,
//		"AddingHabit":                              testAdd,
//		"GettingOneHabit":                          testGetOne,
//		"GettingAllHabits":                         testGetAllSeedHabits,
//		"PerformIncreasesStreakIfDoneYesterday":    testPerformIncreasesStreakIfDoneYesterday,
//		"PerformResetsStreakIfDoneBeforeYesterday": testPerformResetsStreakIfDoneBeforeYesterday,
//		"PerformHabit":                             testPerformHabit,
//	}
//	for name, tc := range tests {
//		t.Run(name, func(t *testing.T) {
//			resetMySqlDB(t, store.DB)
//			tc(t, store)
//		})
//	}
//}
//
//func testLastCheckDays(t *testing.T, store *habits.DBStore) {
//	tcs := []struct {
//		date time.Time
//		want int
//	}{
//		{today, 0},
//		{yesterday, 1},
//		{notQuiteTwoDays, 2},
//		{dayBeforeYesterday, 2},
//	}
//	for _, tc := range tcs {
//		got := store.LastCheckDays(habits.Habit{
//			LastPerformed: tc.date,
//		})
//		if got != tc.want {
//			t.Errorf("%s: want %v, got %v", tc.date, tc.want, got)
//		}
//	}
//}
//
//func testAdd(t *testing.T, store *habits.DBStore) {
//
//	store.Create(habits.Habit{Name: "piano"})
//	want := &habits.Habit{ID: 1, Name: "piano", LastPerformed: fakeNow(), Streak: 0}
//	got, err := store.GetHabit("piano")
//	if err != nil {
//		t.Error(err)
//	}
//
//	if !cmp.Equal(want, got) {
//		t.Error(cmp.Diff(want, got))
//	}
//}
//func testGetOne(t *testing.T, store *habits.DBStore) {
//	Seed(store.DB, seedData)
//	// Testing
//	want := &habits.Habit{ID: 4, Name: "Go", LastPerformed: yesterday, Streak: 4}
//	got, err := store.GetHabit("Go")
//	if err != nil {
//		t.Error(err)
//	}
//	if !cmp.Equal(want, got) {
//		t.Error(cmp.Diff(want, got))
//	}
//}
//func testGetAllSeedHabits(t *testing.T, store *habits.DBStore) {
//	Seed(store.DB, seedData)
//	// Testing
//	want := []habits.Habit{
//		{ID: 1, Name: "k8s", LastPerformed: today, Streak: 4},
//		{ID: 2, Name: "piano", LastPerformed: dayBeforeYesterday, Streak: 4},
//		{ID: 3, Name: "code", LastPerformed: dayBeforeYesterday, Streak: 4},
//		{ID: 4, Name: "Go", LastPerformed: yesterday, Streak: 4},
//		{ID: 5, Name: "docker", LastPerformed: yesterday, Streak: 16},
//		{ID: 6, Name: "SQL", LastPerformed: yesterday, Streak: 29},
//		{ID: 7, Name: "NoSQL", LastPerformed: today, Streak: 30},
//	}
//	got := store.AllHabits()
//	if !cmp.Equal(want, got) {
//		t.Error(cmp.Diff(want, got))
//	}
//}
//func testPerformIncreasesStreakIfDoneYesterday(t *testing.T, store *habits.DBStore) {
//	store.Create(habits.Habit{
//		Name:          "Go",
//		LastPerformed: yesterday,
//		Streak:        4,
//	})
//	habit, err := store.GetHabit("Go")
//	if err != nil {
//		t.Error(err)
//	}
//	store.Perform(*habit)
//	updatedHabit, err := store.GetHabit("Go")
//	if err != nil {
//		t.Error(err)
//	}
//	want := 5
//	got := updatedHabit.Streak
//	if got != want {
//		t.Errorf("expected %v, got %v instead.", want, got)
//	}
//}
//func testPerformResetsStreakIfDoneBeforeYesterday(t *testing.T, store *habits.DBStore) {
//	habit := habits.Habit{
//		Name:          "Go",
//		LastPerformed: dayBeforeYesterday,
//		Streak:        4,
//	}
//	store.Create(habit)
//	store.Perform(habit)
//	updatedHabit, err := store.GetHabit("Go")
//	if err != nil {
//		t.Error(err)
//	}
//	want := 1
//	got := updatedHabit.Streak
//	if !cmp.Equal(want, got) {
//		t.Error(cmp.Diff(want, got))
//	}
//}
//func testPerformHabit(t *testing.T, store *habits.DBStore) {
//	Seed(store.DB, seedData)
//	// Testing
//	want := []habits.Habit{
//		{ID: 1, Name: "k8s", LastPerformed: today, Streak: 4},
//		{ID: 2, Name: "piano", LastPerformed: today, Streak: 1},
//		{ID: 3, Name: "code", LastPerformed: today, Streak: 1},
//		{ID: 4, Name: "Go", LastPerformed: today, Streak: 5},
//		{ID: 5, Name: "docker", LastPerformed: today, Streak: 17},
//		{ID: 6, Name: "SQL", LastPerformed: today, Streak: 30},
//		{ID: 7, Name: "NoSQL", LastPerformed: today, Streak: 30},
//	}
//	habitNames := []string{"k8s", "piano", "code", "Go", "docker", "SQL", "NoSQL"}
//
//	for _, habitName := range habitNames {
//		habit, err := store.GetHabit(habitName)
//		if err != nil {
//			t.Error(err)
//		}
//		days := store.LastCheckDays(*habit)
//		store.PerformHabit(*habit, days)
//	}
//	got := store.AllHabits()
//
//	if !cmp.Equal(want, got) {
//		t.Error(cmp.Diff(want, got))
//	}
//}
