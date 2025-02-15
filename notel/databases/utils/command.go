package main

import (
	"fmt"
	"os"

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

	dbPath := os.Args[1]
	nick := os.Args[2]
	pwd := os.Args[3]

	usersDb := databases.NewUsersDatabase()
	usersDb.LoadDatabase(dbPath)

	usersDb.ChangePassword(nick, pwd)
}
