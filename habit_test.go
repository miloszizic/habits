package habits_test

import (
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

//TestAdd tests adding a habit to a slice of habits
func TestAdd(t *testing.T) {
	t.Parallel()
	l := habits.List{}
	habitName := "piano"
	//os.Stdout = nil
	err := l.Add(habitName)
	if err != nil {
		t.Errorf("adding habit failed with error: %v", err)
	}
	if l[0].Name != habitName {
		t.Errorf("expected %q, got %q instead.", habitName, l[0].Name)
	}

}

// TestDelete tests the Delete method of the List type
func TestDelete(t *testing.T) {
	t.Parallel()
	l := habits.List{
		{Name: "piano", Streak: 4},
		{Name: "code", Streak: 4},
	}
	want := habits.List{
		{Name: "code", Streak: 4},
	}

	err := l.Delete("piano")
	if err != nil {
		t.Fatalf("failed to delete habit from the list: %v", err)
	}
	if !cmp.Equal(want, l) {
		t.Error(cmp.Diff(want, l))
	}
}

//TestDeleteInvalid tests a scenario where the slice don't have the habit'
func TestDeleteInvalid(t *testing.T) {
	t.Parallel()
	l := habits.List{
		{Name: "piano"},
		{Name: "100daysOfGo"},
		{Name: "devops"},
	}
	err := l.Delete("java")
	if err == nil {
		t.Errorf("expected error for nonexistent habit got nil")
	}
}

//TestSaveGet tests saving and retrieving data from file
func TestSaveGet(t *testing.T) {
	l1 := habits.List{
		{Name: "piano", Streak: 4},
		{Name: "code", Streak: 4},
	}
	l2 := habits.List{}

	tf, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("Error creating temp file: %s", err)
	}
	// removing the tmp file
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			t.Fatalf("removeing temp file failed with error: %s", err)
		}
	}(tf.Name())

	if err := l1.Save(tf.Name()); err != nil {
		t.Fatalf("Error saving list to file: %s", err)
	}
	if err := l2.Get(tf.Name()); err != nil {
		t.Fatalf("Error getting list from file: %s", err)
	}
	if !cmp.Equal(l1, l2) {
		t.Error(cmp.Diff(l1, l2))
	}
}

//TestLastCheckDays tests the LastCheckDays method
func TestLastCheckDays(t *testing.T) {
	t.Parallel()
	mokLastCheck := time.Date(2021, 12, 13, 17, 8, 0, 0, time.UTC)
	l := habits.List{
		{Name: "piano", LastCheck: mokLastCheck, Streak: 1},
	}
	habitName := "piano"
	want := 2
	i, _ := l.Find(habitName)
	got, err := l.LastCheckDays(Now(), l[i])
	if err != nil {
		t.Fatalf("got an error while checking days: %v", err)
	}
	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}

//TestLastCheckDaysNegative is testing checking from future time
func TestLastCheckDaysNegative(t *testing.T) {
	t.Parallel()
	mokLastCheck := time.Date(2021, 12, 17, 17, 8, 0, 0, time.UTC)
	l := habits.List{
		{Name: "piano", LastCheck: mokLastCheck, Streak: 1},
	}
	habitName := "piano"
	i, _ := l.Find(habitName)
	got, err := l.LastCheckDays(Now(), l[i])
	if err == nil {
		t.Errorf("expexted error for habit from future got: %v", got)
	}
}

//TestFind tests the Find method
func TestFind(t *testing.T) {
	t.Parallel()
	l := habits.List{
		{Name: "k8s", LastCheck: Now(), Streak: 4},
		{Name: "piano", LastCheck: Now(), Streak: 4},
	}
	_, found := l.Find("piano")
	if !found {
		t.Error("find should had found a piano but failed with error")
	}
}
func TestFindInvalidElement(t *testing.T) {
	l := habits.List{
		{Name: "k8s", LastCheck: Now(), Streak: 4},
		{Name: "piano", LastCheck: Now(), Streak: 4},
	}
	i, _ := l.Find("go")
	if i != -1 {
		t.Error("expected an error as (-1) got nothing")
	}
}

//TestBreak tests the habit brake
func TestBreak(t *testing.T) {
	t.Parallel()
	mokLastCheck := time.Date(2021, 12, 17, 17, 8, 0, 0, time.UTC)
	l := habits.List{
		{Name: "piano", LastCheck: mokLastCheck, Streak: 4},
	}
	want := habits.List{
		{Name: "piano", LastCheck: Now(), Streak: 1},
	}
	i, _ := l.Find("piano")
	l.Break(Now(), i)
	if !cmp.Equal(want, l) {
		t.Error(cmp.Diff(want, l))
	}
	if !cmp.Equal(want, l) {
		t.Error(cmp.Diff(want, l))
	}
}
func TestUpdateYesterday(t *testing.T) {
	t.Parallel()
	mokLastCheck := time.Date(2021, 12, 14, 17, 8, 0, 0, time.UTC)
	l := habits.List{
		{Name: "piano", LastCheck: mokLastCheck, Streak: 1},
	}
	want := habits.List{
		{Name: "piano", LastCheck: Now(), Streak: 2},
	}
	i, _ := l.Find("piano")
	l.UpdateYesterday(Now(), i)
	if !cmp.Equal(want, l) {
		t.Error(cmp.Diff(want, l))
	}
}

func TestDecisionsHandler(t *testing.T) {
	//time.Date(2021, 12, 15, 17, 8, 0, 0, time.UTC)
	t.Parallel()
	k8sLastCheck := time.Date(2021, 12, 15, 17, 8, 0, 0, time.UTC)
	codeLastCheck := time.Date(2021, 12, 11, 18, 9, 0, 0, time.UTC)
	goLastCheck := time.Date(2021, 12, 14, 15, 9, 0, 0, time.UTC)
	l := habits.List{
		{Name: "k8s", LastCheck: k8sLastCheck, Streak: 4},
		{Name: "piano", LastCheck: codeLastCheck, Streak: 4},
		{Name: "code", LastCheck: codeLastCheck, Streak: 4},
		{Name: "Go", LastCheck: goLastCheck, Streak: 4},
		{Name: "docker", LastCheck: goLastCheck, Streak: 16},
	}
	want := habits.List{
		{Name: "k8s", LastCheck: Now(), Streak: 4},
		{Name: "piano", LastCheck: Now(), Streak: 1},
		{Name: "code", LastCheck: Now(), Streak: 1},
		{Name: "Go", LastCheck: Now(), Streak: 5},
		{Name: "docker", LastCheck: Now(), Streak: 17},
	}
	habitNames := []string{"k8s", "piano", "code", "Go", "docker"}
	//os.Stdout = nil
	for _, name := range habitNames {
		i, _ := l.Find(name)
		l.DecisionsHandler(i, Now())
	}

	if !cmp.Equal(want, l) {
		t.Error(cmp.Diff(want, l))
	}
}
