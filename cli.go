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
			fmt.Println(habit.Name)
		}
		return
	}
	habit, err := store.GetHabit(habitName)
	if err != nil {
		store.Print("searching for record returned an error:%v", err)
	}
	if habit == nil {
		store.Add(Habit{Name: habitName})
		return
	}
	days := store.LastCheckDays(*habit)
	fmt.Println(days)
	store.PerformHabit(*habit, days)
}
