package store

import (
	"io"
	"time"
)

// Habit struct has all habit attributes
type Habit struct {
	ID            int
	Name          string
	LastPerformed time.Time
	Streak        int
	Output        io.Writer
}

//func RunCLI() {
//	MySQLURL := os.Getenv("MYSQL_URL")
//	if MySQLURL == "" {
//		MySQLURL = "tester:secret@tcp(127.0.0.:3306)/habits?parseTime=true"
//	}
//	store, err := FromMySQL(MySQLURL)
//	if err != nil {
//		fmt.Fprintf(os.Stderr, "opening %q database: %v\n", err, MySQLURL)
//		os.Exit(1)
//	}
//	habitName := strings.Join(os.Args[1:], " ")
//	if len(os.Args) == 1 {
//		fmt.Println("You are tracking following habits: ")
//		for _, habit := range store.AllHabits() {
//			fmt.Println(habit.Name)
//		}
//		return
//	}
//	habit, err := store.GetHabit(habitName)
//	if err != nil {
//		store.Print("searching for record returned an error:%v", err)
//	}
//	if habit == nil {
//		store.Create(Habit{Name: habitName})
//		return
//	}
//	days := store.LastCheckDays(*habit)
//	fmt.Println(days)
//	store.PerformHabit(*habit, days)
//}
