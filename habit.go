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

//Store is a list of Habits

type Store struct {
	Habits []Habit
	Output io.Writer
}

func PrintStore() Store {
	return Store{
		Output: os.Stdout,
	}
}
func (s Store) Print(massage string) {
	fmt.Fprintln(s.Output, massage)
}

//func (p PrintStore) Print(text string, a ...interface{}) {
//	fmt.Fprintf( text, a...)
//}

//Add method is adding a new habit to the list of Habits
func (s *Store) Add(name string) {
	s.Habits = append(s.Habits, Habit{
		Name:      name,
		Done:      false,
		LastCheck: Now().Round(time.Hour),
		Streak:    1,
	})
	s.Print("Good luck with your new" + name + "habit. Don't forget to do it again tomorrow.\n")
	//Print("Good luck with your new %v habit. Don't forget to do it again tomorrow.\n", name)
	//fmt.Fprintf(os.Stdout, "Good luck with your new %s habit. Don't forget to do it again tomorrow.\n", name)
}

//Delete method deletes a Habit from the list of Habits
func (s *Store) Delete(i int) error {
	if i < 0 || i > len(s.Habits) {
		return fmt.Errorf("item %d does not exist", i)
	}
	s.Habits = append(s.Habits[:i], s.Habits[i+1:]...)
	return nil
}

//Save method encodes the Store as JSON and saves it
//using the provided file name
func (s *Store) Save(filename string) error {
	js, err := json.Marshal(s.Habits)
	if err != nil {
		log.Printf(" marshaling to JSON failed : %v \n", err)
	}
	return ioutil.WriteFile(filename, js, 0644)
}

//LastCheckDays method checks  for number of days current date and
func (s Store) LastCheckDays(now time.Time, h Habit) (int, error) {
	for _, v := range s.Habits {
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
func (s *Store) Find(name string) (*Habit, bool) {
	for i, habit := range s.Habits {
		if habit.Name == name {
			return &s.Habits[i], true
		}
	}
	return nil, false
}
func (h *Habit) Break(now time.Time) {
	h.LastCheck = now
	h.Streak = 1
}

//UpdateYesterday method is updating the habit if last check was yesterday.
func (h *Habit) UpdateYesterday(now time.Time) {
	h.LastCheck = now
	h.Streak++
}

//DecisionsHandler makes decisions based on the number of days between today's date
//and date habit was last checked
func (h *Habit) DecisionsHandler(w io.Writer, days int, now time.Time) {
	switch {
	case days >= 0 && h.Done:
		fmt.Fprintf(w, "You already finished the %v habit.\n", h.Name)
	case days == 0:
		fmt.Fprintf(w, "Nice work: you've done the habit '%s' for %v days in a row Now.\n", h.Name, h.Streak)
	case days == 1 && h.Streak == 29:
		h.UpdateYesterday(now)
		h.Done = true
		fmt.Fprintf(w, "Congratulations, this is your %dth day for '%s' habit. You finished successfully!!\n", h.Streak, h.Name)
	case days == 1 && h.Streak > 15:
		h.UpdateYesterday(now)
		fmt.Fprintf(w, "You're currently on a %d-day streak for '%s'. Stick to it!\n", h.Streak, h.Name)
	case days == 1:
		h.UpdateYesterday(now)
		fmt.Fprintf(w, "Nice work: you've done the habit '%s' for %v days in a row Now.\n", h.Name, h.Streak)
	case days >= 3:
		h.Break(now)
		fmt.Fprintf(w, "You last did the habit '%s' %d days ago, so you're starting a new streak today. Good luck!\n", h.Name, days)
	}
}

// Get method opens the provided file name, decodes
// the JSON data and parses it into a Store
func (s *Store) Get(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("reading file failed: %v\n", err)
		}
	}
	return json.Unmarshal(file, &s.Habits)
}
