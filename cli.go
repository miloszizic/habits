package habits

import (
	"fmt"
	"os"
	"strings"
)

const DBFile = "./data.db"

func RunCli() {

	habitName := strings.Join(os.Args[1:], " ")
	store := FromSQLite(DBFile)
	if len(os.Args) == 1 {
		fmt.Println("You are tracking following habits: ")
		for _, habit := range store.AllHabits() {
			if habit.Done {
				continue
			}
			fmt.Println(habit.Name)
		}
		return
	}
	habit, found := store.GetHabit(habitName)
	if !found {
		store.Add(habitName)
	}
	days := habit.LastCheckDays(Now())
	fmt.Println(days)
	store.PerformHabit(&habit, days, Now())
}
