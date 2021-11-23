package habits

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

//Habit struct has all habit attributes
type Habit struct {
	Name      string
	LastCheck time.Time // last date checked
	Streak    int       //increment on next date
	Done      bool
}

//Now passes the current system time
var Now = time.Now

//List is a list of Habits
type List []Habit

//Add method is adding a new habit to the list of Habits
func (l *List) Add(w io.Writer, name string) error {
	ls := *l
	t := Habit{
		Name:      name,
		Done:      false,
		LastCheck: Now().Round(time.Hour),
		Streak:    1,
	}
	*l = append(ls, t)
	fmt.Fprintf(w, "Good luck with your new %s habit. Don't forget to do it again tomorrow.\n", name)
	return nil
}

//Delete method deletes a Habit from the list of Habits
func (l *List) Delete(i int) error {
	ls := *l
	if i < 0 || i > len(ls) {
		return fmt.Errorf("item %d does not exist", i)
	}
	*l = append(ls[:i], ls[i+1:]...)

	return nil
}

//Save method encodes the List as JSON and saves it
//using the provided file name
func (l *List) Save(filename string) error {
	js, err := json.Marshal(l)
	if err != nil {
		log.Printf(" marshaling to JSON failed : %v \n", err)
	}
	return ioutil.WriteFile(filename, js, 0644)
}

//LastCheckDays method checks  for number of days current date and
func (l List) LastCheckDays(now time.Time, h Habit) (int, error) {
	for _, v := range l {
		if v.Name == h.Name {
			//days := int(v.LastCheck.Sub(now).Hours() / 24)
			days := int(now.Sub(v.LastCheck).Hours() / 24)
			if days >= 0 {
				return days, nil
			}
		}
	}
	return -1, errors.New("the dates can't be checked ") //error or -1
}

//Find method returns true and habit if it finds it
func (l List) Find(name string) (int, bool) {
	for i, habit := range l {
		if habit.Name == name {
			return i, true
		}
	}
	return -1, false
}
func (l *List) Break(now time.Time, i int) {
	ls := *l
	ls[i].LastCheck = now //Proper pointer dereference ?
	ls[i].Streak = 1
}

//UpdateYesterday method is updating the habit if last check was yesterday.
func (l *List) UpdateYesterday(now time.Time, i int) {
	ls := *l
	ls[i].LastCheck = now //Proper pointer dereference ?
	ls[i].Streak++
}

//DecisionsHandler makes decisions based on the number of days between today's date
//and date habit was last checked
func (l *List) DecisionsHandler(w io.Writer, i int, now time.Time) {
	ls := *l
	days, _ := l.LastCheckDays(now, ls[i])
	switch {
	case days >= 0 && ls[i].Done:
		fmt.Fprintf(w, "You already finished the %v habit.\n", ls[i].Name)
	case days == 0:
		fmt.Fprintf(w, "Nice work: you've done the habit '%s' for %v days in a row Now.\n", ls[i].Name, ls[i].Streak)
	case days == 1 && ls[i].Streak == 29:
		l.UpdateYesterday(now, i)
		ls[i].Done = true
		fmt.Fprintf(w, "Congratulations, this is your %dth day for '%s' habit. You finished successfully!!\n", ls[i].Streak, ls[i].Name)
	case days == 1 && ls[i].Streak > 15:
		l.UpdateYesterday(now, i)
		fmt.Fprintf(w, "You're currently on a %d-day streak for '%s'. Stick to it!\n", ls[i].Streak, ls[i].Name)
	case days == 1:
		l.UpdateYesterday(now, i)
		fmt.Fprintf(w, "Nice work: you've done the habit '%s' for %v days in a row Now.\n", ls[i].Name, ls[i].Streak)
	case days >= 3:
		l.Break(now, i)
		fmt.Fprintf(w, "You last did the habit '%s' %d days ago, so you're starting a new streak today. Good luck!\n", ls[i].Name, days)
	}
}

// Get method opens the provided file name, decodes
// the JSON data and parses it into a List
func (l *List) Get(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("reading file failed: %v\n", err)
		}
	}
	return json.Unmarshal(file, l)
}
