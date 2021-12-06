package habits

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const DBFile = "./data.db"

func RunCli() {

	habitName := strings.Join(os.Args[1:], " ")
	store := FromSQLite(dbFIle)
	if len(os.Args) == 1 {
		fmt.Println("You are tracking following habits: ")
		habits := store.GetAll()
		for _, habit := range store.AllHabits() {
			if habit.Done {
				continue
			}
			fmt.Println(habit.Name)
		}
		return
	}
	habit, found := store.GetOne(habitName)
	if !found {
		store.Add(habitName)
	}
		days := habit.LastCheckDays(Now())
		fmt.Println(days)
		store.DecisionsHandler(&habit, days, Now())
	}

}
