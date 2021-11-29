package habits_test

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/miloszizic/habits"
)

//Mocks time.Now method for testing purposes
var Now = func() time.Time {
	return time.Date(2021, 12, 15, 17, 8, 0, 0, time.UTC)
}

// Output is Default terminal

//TestAdd tests adding a habit to a slice of habits
func TestAdd(t *testing.T) {
	t.Parallel()
	store := habits.Store{
		Output: io.Discard,
	}
	want := "piano"
	store.Add(want)
	got := store.Habits[0].Name
	if got != want {
		t.Errorf("expected %q, got %q instead.", want, got)
	}

}

// TestDelete tests the Delete method of the Store type
func TestDelete(t *testing.T) {
	t.Parallel()
	store := habits.Store{
		Habits: []habits.Habit{
			{Name: "piano", Streak: 4},
			{Name: "code", Streak: 4},
		},

		Output: io.Discard,
	}
	want := []habits.Habit{
		{Name: "code", Streak: 4},
	}
	err := store.Delete(0)
	if err != nil {
		t.Fatalf("failed to delete habit from the list: %v", err)
	}
	if !cmp.Equal(want, store.Habits) {
		t.Error(cmp.Diff(want, store.Habits))
	}
}

//TestDeleteInvalid tests a scenario where the slice don't have the habit'
func TestDeleteInvalid(t *testing.T) {
	t.Parallel()
	store := habits.Store{
		Habits: []habits.Habit{
			{Name: "piano"},
			{Name: "100daysOfGo"},
			{Name: "devops"},
		},
		Output: io.Discard,
	}
	err := store.Delete(4)
	if err == nil {
		t.Errorf("expected error for nonexistent habit got nil")
	}
}

//TestSaveGet tests saving and retrieving data from file
func TestSaveGet(t *testing.T) {
	want := []habits.Habit{
		{Name: "piano"},
		{Name: "100daysOfGo"},
		{Name: "devops"},
	}
	store := habits.Store{
		Habits: want,
	}

	tf, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("Error creating temp file: %s", err)
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("removeing temp file failed with error: %s", err)
		}
	}(tf.Name())

	if err := store.Save(tf.Name()); err != nil {
		t.Fatalf("Error saving list to file: %s", err)
	}
	store2 := habits.Store{}
	if err := store2.Get(tf.Name()); err != nil {
		t.Fatalf("Error getting list from file: %s", err)
	}
	if !cmp.Equal(store.Habits, store2.Habits) {
		t.Error(cmp.Diff(store.Habits, store2.Habits))
	}
}
func TestGetFailure(t *testing.T) {
	store := habits.Store{
		Output: io.Discard,
	}
	if err := store.Get("fail.txt"); err == nil {
		t.Fatal("reading nonexistent file should give an error, got nil")
	}
}

//TestLastCheckDays tests the LastCheckDays method
func TestLastCheckDays(t *testing.T) {
	t.Parallel()
	mokLastCheck := time.Date(2021, 12, 13, 17, 8, 0, 0, time.UTC)
	store := habits.Store{
		Habits: []habits.Habit{
			{Name: "piano", LastCheck: mokLastCheck, Streak: 1},
		},
		Output: io.Discard,
	}
	habitName := "piano"
	want := 2
	habit, _ := store.Find(habitName)
	got, err := store.LastCheckDays(Now(), *habit)
	if err != nil {
		t.Fatalf("got an error while checking days: %v", err)
	}
	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}
func TestLastCheckDaysInvalid(t *testing.T) {
	t.Parallel()
	mokLastCheck := time.Date(2021, 12, 17, 17, 8, 0, 0, time.UTC)
	store := habits.Store{
		Habits: []habits.Habit{
			{Name: "piano", LastCheck: mokLastCheck, Streak: 1},
		},
		Output: io.Discard,
	}
	habitName := "piano"
	habit, _ := store.Find(habitName)
	_, err := store.LastCheckDays(Now(), *habit)
	if err == nil {
		t.Fatal(" expected error got nil")
	}
}

//TestFind tests the Find method
func TestFind(t *testing.T) {
	t.Parallel()
	store := habits.Store{
		Habits: []habits.Habit{
			{Name: "k8s", LastCheck: Now(), Streak: 4},
			{Name: "piano", LastCheck: Now(), Streak: 4},
		},
		Output: io.Discard,
	}
	_, found := store.Find("piano")
	if !found {
		t.Error("find should had found a piano but failed with error")
	}
}
func TestFindInvalidElement(t *testing.T) {
	store := habits.Store{
		Habits: []habits.Habit{
			{Name: "k8s", LastCheck: Now(), Streak: 4},
			{Name: "piano", LastCheck: Now(), Streak: 4},
		},
		Output: io.Discard,
	}
	_, b := store.Find("go")
	if b {
		t.Error("expected an error as (-1) got nothing")
	}
}

//TestBreak tests the habit brake
func TestBreak(t *testing.T) {
	t.Parallel()
	mokLastCheck := time.Date(2021, 12, 17, 17, 8, 0, 0, time.UTC)
	store := habits.Store{
		Habits: []habits.Habit{
			{Name: "piano", LastCheck: mokLastCheck, Streak: 4},
		},
		Output: io.Discard,
	}
	want := habits.Store{
		Habits: []habits.Habit{
			{Name: "piano", LastCheck: Now(), Streak: 1},
		},
		Output: io.Discard,
	}
	habit, _ := store.Find("piano")
	habit.Break(Now())
	if !cmp.Equal(want, store) {
		t.Error(cmp.Diff(want, store))
	}

}
func TestUpdateYesterday(t *testing.T) {
	t.Parallel()
	mokLastCheck := time.Date(2021, 12, 14, 17, 8, 0, 0, time.UTC)
	store := habits.Store{
		Habits: []habits.Habit{
			{Name: "piano", LastCheck: mokLastCheck, Streak: 1},
		},
		Output: io.Discard,
	}
	want := habits.Store{
		Habits: []habits.Habit{
			{Name: "piano", LastCheck: Now(), Streak: 2},
		},
		Output: io.Discard,
	}
	habit, _ := store.Find("piano")
	habit.UpdateYesterday(Now())
	if !cmp.Equal(want, store) {
		t.Error(cmp.Diff(want, store))
	}
}

func TestDecisionsHandler(t *testing.T) {
	//time.Date(2021, 12, 15, 17, 8, 0, 0, time.UTC)
	t.Parallel()

	sameDay := time.Date(2021, 12, 15, 17, 8, 0, 0, time.UTC)
	threeDays := time.Date(2021, 12, 11, 18, 9, 0, 0, time.UTC)
	dayBefore := time.Date(2021, 12, 14, 15, 9, 0, 0, time.UTC)

	store := habits.Store{
		Habits: []habits.Habit{
			{Name: "k8s", LastCheck: sameDay, Streak: 4, Output: io.Discard},
			{Name: "piano", LastCheck: threeDays, Streak: 4, Output: io.Discard},
			{Name: "code", LastCheck: threeDays, Streak: 4, Output: io.Discard},
			{Name: "Go", LastCheck: dayBefore, Streak: 4, Output: io.Discard},
			{Name: "docker", LastCheck: dayBefore, Streak: 16, Output: io.Discard},
			{Name: "SQL", LastCheck: dayBefore, Streak: 29, Done: false, Output: io.Discard},
			{Name: "NoSQL", LastCheck: sameDay, Streak: 30, Done: true, Output: io.Discard},
		},
		Output: io.Discard,
	}
	want := habits.Store{
		Habits: []habits.Habit{
			{Name: "k8s", LastCheck: Now(), Streak: 4, Output: io.Discard},
			{Name: "piano", LastCheck: Now(), Streak: 1, Output: io.Discard},
			{Name: "code", LastCheck: Now(), Streak: 1, Output: io.Discard},
			{Name: "Go", LastCheck: Now(), Streak: 5, Output: io.Discard},
			{Name: "docker", LastCheck: Now(), Streak: 17, Output: io.Discard},
			{Name: "SQL", LastCheck: Now(), Streak: 30, Done: true, Output: io.Discard},
			{Name: "NoSQL", LastCheck: sameDay, Streak: 30, Done: true, Output: io.Discard},
		},
		Output: io.Discard,
	}
	habitNames := []string{"k8s", "piano", "code", "Go", "docker", "SQL", "NoSQL"}

	for _, habitName := range habitNames {
		habit, _ := store.Find(habitName)
		days, _ := store.LastCheckDays(Now(), *habit)
		habit.DecisionsHandler(Now(), days)
	}

	if !cmp.Equal(want, store) {
		t.Error(cmp.Diff(want, store))
	}
}
