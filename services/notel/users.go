package main

import (
	"bufio"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type User struct {
	Nick      string            `json:"nick"`
	PwdHash   [sha512.Size]byte `json:"pwd_hash"`
	LastLogin time.Time         `json:"last_login"`
}

type UsersDatabase struct {
	filePath string
	file     *os.File
	users    map[string]User
	mutex    sync.RWMutex
}

func NewUsersDatabase() *UsersDatabase {
	return &UsersDatabase{}
}

func (u *UsersDatabase) LoadDatabase(filePath string) error {
	u.filePath = filePath
	u.users = make(map[string]User)

	if err := u.readDatabase(); err != nil {
		errorLog.Printf("unable to load users database: %s\n", err.Error())
		return err
	}

	minLastLogin := time.Now().Add(-30 * 24 * time.Hour)
	for nick, usr := range u.users {
		if usr.LastLogin.Before(minLastLogin) {
			warnLog.Printf("removed user=%s, last login=%s\n", nick, usr.LastLogin.Format(time.RFC3339))
			delete(u.users, nick)
		}
	}
	return nil
}

func (u *UsersDatabase) readDatabase() error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	filedb, err := os.OpenFile(u.filePath, os.O_RDONLY|os.O_CREATE, 0755)
	if err != nil {
		errorLog.Printf("unable to get database: %s\n", err.Error())
		return err
	}
	infoLog.Printf("opened database: %s\n", u.filePath)

	scanner := bufio.NewScanner(filedb)
	scanner.Split(bufio.ScanLines)

	line := 0
	for scanner.Scan() {
		var usr User
		if err := json.Unmarshal([]byte(scanner.Text()), &usr); err != nil {
			errorLog.Printf("unable to marshal line %d: %s\n", line, err.Error())
			continue
		}

		u.users[usr.Nick] = usr
	}
	filedb.Close()

	infoLog.Printf("loaded %d users from database\n", len(u.users))

	return nil
}

func (u *UsersDatabase) updateDatabase() error {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	filedb, err := os.OpenFile(u.filePath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		errorLog.Printf("unable to get database: %s\n", err.Error())
		return err
	}
	infoLog.Printf("opened database: %s\n", u.filePath)

	for nick, usr := range u.users {
		if b, err := json.Marshal(usr); err != nil {
			errorLog.Printf("unable to marshal user=%s: %s\n", nick, err.Error())
		} else {
			b = append(b, '\n')
			if _, err := filedb.Write(b); err != nil {
				errorLog.Printf("unable to write user=%s: %s\n", nick, err.Error())
			}
		}
	}
	return filedb.Close()
}

func (u *UsersDatabase) UserExists(nick string) bool {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	_, ok := u.users[nick]
	return ok
}

func (u *UsersDatabase) AddUser(nick, pwd string) error {
	if u.UserExists(nick) {
		return fmt.Errorf("user already exists")
	}

	u.mutex.Lock()
	u.users[nick] = User{
		Nick:      nick,
		PwdHash:   sha512.Sum512([]byte(pwd)),
		LastLogin: time.Now(),
	}
	u.mutex.Unlock()

	if err := u.updateDatabase(); err != nil {
		errorLog.Printf("unable to update the database on disk: %s\n", err.Error())
		return err
	}

	return nil
}

func (u *UsersDatabase) LogUser(nick, pwd string) bool {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	usr, ok := u.users[nick]
	if !ok {
		return false
	}

	if usr.PwdHash != sha512.Sum512([]byte(pwd)) {
		return false
	}

	return true
}

func (u *UsersDatabase) Quit() {
	u.file.Close()
}
