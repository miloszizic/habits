package habits

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Hard coding the file name for the destination file that contains the habits
const habitsFileName = ".habits.json"

func RunCli() {
	// Output is Default terminal
	var w io.Writer = os.Stdout
	// Define an items list
	s := Store{}
	//Parse the habit
	habitName := strings.Join(os.Args[1:], " ")
	// Use the Get method to read habits from file
	if err := s.Get(habitsFileName); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			fmt.Printf("failed to read %s: %s\n", habitsFileName, err)
		}
		os.Exit(1) // Can it stay here ? Recommended in the main.go
	}
	if len(os.Args) == 1 {
		fmt.Println("You are tracking following habits: ")
		fmt.Println(s)
		for _, item := range s.Habits {
			if item.Done {
				continue
			}
			fmt.Println(item.Name)
		}
		return
	}
	//Making a decision
	habit, found := s.Find(habitName)
	if !found {
		s.Add(habitName)
	}
	if found {
		days, _ := s.LastCheckDays(Now(), *habit)
		habit.DecisionsHandler(w, days, Now())
	}
	// Save the new habit to the file
	if err := s.Save(habitsFileName); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			fmt.Printf("failed to save to file %s with following error: %s\n", habitsFileName, err)
		}
		os.Exit(1) // Can it stay here ? Recommended in the main.go
	}
}
