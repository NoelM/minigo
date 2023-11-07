package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
)

type User struct {
	Nick        string    `json:"nick"`
	PwdHash     string    `json:"pwd_hash"`
	Bio         string    `json:"bio"`
	Tel         string    `json:"tel"`
	Location    string    `json:"location"`
	LastConnect time.Time `json:"last_connect"`
}

type UsersDatabase struct {
	DB    *pebble.DB
	mutex sync.RWMutex
}

func NewUsersDatabase() *UsersDatabase {
	return &UsersDatabase{}
}

func (u *UsersDatabase) LoadDatabase(dir string) error {
	db, err := pebble.Open(dir, &pebble.Options{})
	if err != nil {
		return err
	}
	u.DB = db

	return nil
}

func (u *UsersDatabase) UserExists(nick string) bool {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	_, _, err := u.DB.Get([]byte(nick))
	if err != nil || err == pebble.ErrNotFound {
		return false
	}

	return true
}

func (u *UsersDatabase) AddUser(nick, pwd string) error {
	if u.UserExists(nick) {
		return fmt.Errorf("user already exists")
	}

	hash := sha512.Sum512([]byte(pwd))
	hashB64 := base64.StdEncoding.EncodeToString(hash[:])

	usr := &User{
		Nick:        nick,
		PwdHash:     hashB64,
		LastConnect: time.Now(),
	}

	val, err := json.Marshal(usr)
	if err != nil {
		return fmt.Errorf("unable to marshall user nick=%s: %s", nick, err.Error())
	}

	u.mutex.Lock()
	if err = u.DB.Set([]byte(nick), val, pebble.Sync); err != nil {
		return fmt.Errorf("unable to add user nick=%s: %s", nick, err.Error())
	}
	u.mutex.Unlock()

	return nil
}

func (u *UsersDatabase) LogUser(nick, pwd string) bool {
	u.mutex.RLock()
	defer u.mutex.RUnlock()

	val, closer, err := u.DB.Get([]byte(nick))
	if err != nil {
		errorLog.Printf("login error: nick=%s: %s\n", nick, err.Error())
		return false
	}

	var user *User
	if err = json.Unmarshal(val, user); err != nil {
		errorLog.Printf("login error: nick=%s: %s\n", nick, err.Error())
		return false
	}
	closer.Close()

	hash := sha512.Sum512([]byte(pwd))
	hashB64 := base64.StdEncoding.EncodeToString(hash[:])

	return user.PwdHash == hashB64
}

func (u *UsersDatabase) Quit() {
	u.DB.Close()
}
