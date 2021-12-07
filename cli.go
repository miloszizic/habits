package habits

import (
	"fmt"
	"os"
	"strings"
	"time"
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
	habit, found := store.GetHabit(habitName)
	if !found {
		store.Add(Habit{
			Name:          habitName,
			LastPerformed: time.Now(),
		})
	}
	days := store.LastCheckDays(habit)
	fmt.Println(days)
	store.PerformHabit(&habit, days)
}
