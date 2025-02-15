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

	dbPath := os.Args[1]
	nick := os.Args[2]
	pwd := os.Args[3]

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
