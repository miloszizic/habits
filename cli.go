package habits

import (
	"fmt"
	"os"
	"strings"
)

// Hard coding the file name for the destination file that contains the habits
const habitsFileName = ".habits.json"

func RunCli() {
	// Define an items list
	l := &List{}
	//Parse the habit
	habitName := strings.Join(os.Args[1:], " ")
	// Use the Get method to read habits from file
	if err := l.Get(habitsFileName); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			fmt.Printf("failed to read %s: %s\n", habitsFileName, err)
		}
		os.Exit(1)
	}
	if len(os.Args) == 1 {
		fmt.Println("You are tracking following habits: ")
		for _, item := range *l {
			fmt.Println(item.Name)
		}
		return
	}
	//Making a decision
	i, found := l.Find(habitName)
	if !found {
		err := l.Add(habitName)
		if err != nil {
			fmt.Println("failed to add the habit with error:", err)
		}
	}
	if found {
		l.DecisionsHandler(i, Now())
	}
	// Save the new habit to the file
	if err := l.Save(habitsFileName); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			fmt.Printf("failed to save to file %s with following error: %s\n", habitsFileName, err)
		}
		os.Exit(1)
	}
}
