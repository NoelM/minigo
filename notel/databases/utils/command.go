package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/NoelM/minigo/notel/databases"
)

func main() {
	// os.Args is a slice of strings that contains all the command-line arguments.
	// The first element, os.Args[0], is the path to the executable.
	// The subsequent elements are the arguments passed to the program.

	if len(os.Args) < 2 {
		fmt.Println("Please provide some command-line arguments.")
		return
	}

	command := os.Args[1]

	switch command {
	case "pass":
		ChangePassword(os.Args[1:])

	case "users":
		ListUsers(os.Args[1:])

	case "help":
		fmt.Println("help")
		fmt.Println("./utils pass [user-db-path] [nickname] [password]")
		fmt.Println("./utils users [user-db-path]")

	default:
		fmt.Println("no match")
	}

}

func ListUsers(args []string) {
	dbPath := args[1]

	usersDb := databases.NewUsersDatabase()
	usersDb.LoadDatabase(dbPath)

	users, err := usersDb.ListAllowedUsers()
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		fmt.Printf("> %s\n", u.Nick)
		fmt.Printf("  %s\n", u.LastConnect.Format("02/01/2006 15:04:05"))
		fmt.Printf("  +30j: %t\n", time.Until(u.LastConnect) > 24*30*time.Hour)
		fmt.Printf("  %s\n", u.Bio)
		fmt.Printf("  %s\n", u.PwdHash)
		fmt.Printf("  %s\n", u.Location)
		fmt.Printf("  %s\n", u.Tel)
		fmt.Printf("  Rep: %t\n", u.AppAnnuaire)
	}

}

func ChangePassword(args []string) {
	// os.Args is a slice of strings that contains all the command-line arguments.
	// The first element, os.Args[0], is the path to the executable.
	// The subsequent elements are the arguments passed to the program.

	if len(os.Args) < 2 {
		fmt.Println("Please provide some command-line arguments.")
		return
	}

	dbPath := args[1]
	nick := args[2]
	pwd := args[3]

	usersDb := databases.NewUsersDatabase()
	usersDb.LoadDatabase(dbPath)

	if changed := usersDb.ChangePassword(nick, pwd); !changed {
		log.Fatal("unable to change pwd")
	}

	u, err := usersDb.LoadUser(nick)
	if err != nil {
		log.Fatal(err)
	}
	u.LastConnect = time.Now()

	if err = usersDb.SetUser(u); err != nil {
		log.Fatal(err)
	}
}
