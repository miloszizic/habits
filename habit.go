package habits

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

//Now passes the current system time
var Now = time.Now

//Habit struct has all habit attributes
type Habit struct {
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

//Print as Habit method is wrapping Fprintf so that is not needed to specify
//the default print output every time
func (h Habit) Print(massage string, params ...interface{}) {
	if h.Output == nil {
		fmt.Fprintf(os.Stdout, massage, params...)
	} else {
		fmt.Fprintf(h.Output, massage, params...)
	}
}

//Add method is adding a new habit to the list of Habits
func (s *Store) Add(name string) {
	s.Habits = append(s.Habits, Habit{
		Name:      name,
		Done:      false,
		LastCheck: Now().Round(time.Hour),
		Streak:    1,
	})
	s.Print("Good luck with your new %v habit. Don't forget to do it again tomorrow.\n", name)
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
		fmt.Printf(" marshaling to JSON failed : %v \n", err)
	}
	return ioutil.WriteFile(filename, js, 0644)
}

// Get method opens the provided file name, decodes
// the JSON data and parses it into a Store
func (s *Store) Load(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.Print("reading file failed: %v\n", err)
		}
	}
	return json.Unmarshal(file, &s.Habits)
}

//LastCheckDays method checks  for number of days current date and
func (s Store) LastCheckDays(time time.Time, h Habit) (int, error) {
	for _, v := range s.Habits {
		if v.Name == h.Name {
			days := int(time.Sub(v.LastCheck).Hours() / 24)
			if days >= 0 {
				return days, nil
			}
		}
	}
	return -1, errors.New("the dates can't be checked ")
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

//Break will restart a streak based on the LastCheck time
func (h *Habit) Break(time time.Time) {
	h.LastCheck = time
	h.Streak = 1
}

//UpdateYesterday method is updating the habit if last check was yesterday.
func (h *Habit) UpdateYesterday(time time.Time) {
	h.LastCheck = time
	h.Streak++
}

//DecisionsHandler makes decisions based on the number of days between today's date
//and date habit was last checked
func (h *Habit) DecisionsHandler(time time.Time, days int) {
	switch {
	case days >= 0 && h.Done:
		h.Print("You already finished the %v habit.\n", h.Name)
	case days == 0:
		h.Print("Nice work: you've done the habit '%s' for %v days in a row Now.\n", h.Name, h.Streak)
	case days == 1 && h.Streak == 29:
		h.UpdateYesterday(time)
		h.Done = true
		h.Print("Congratulations, this is your %dth day for '%s' habit. You finished successfully!!\n", h.Streak, h.Name)
	case days == 1 && h.Streak > 15:
		h.UpdateYesterday(time)
		h.Print("You're currently on a %d-day streak for '%s'. Stick to it!\n", h.Streak, h.Name)
	case days == 1:
		h.UpdateYesterday(time)
		h.Print("Nice work: you've done the habit '%s' for %v days in a row Now.\n", h.Name, h.Streak)
	case days >= 3:
		h.Break(time)
		h.Print("You last did the habit '%s' %d days ago, so you're starting a new streak today. Good luck!\n", h.Name, days)
	}
}
