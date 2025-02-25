package databases

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NoelM/minigo/notel/logs"
	"github.com/cockroachdb/pebble"
)

type User struct {
	Nick        string    `json:"nick"`
	PwdHash     string    `json:"pwd_hash"`
	Bio         string    `json:"bio"`
	Tel         string    `json:"tel"`
	Location    string    `json:"location"`
	LastConnect time.Time `json:"last_connect"`
	AppAnnuaire bool      `json:"annu,omitempty"`
}

type UsersDatabase struct {
	DB *pebble.DB
}

func NewUsersDatabase() *UsersDatabase {
	return &UsersDatabase{}
}

func (u *UsersDatabase) getHashB64(pwd string) (hashB64 string) {
	hash := sha512.Sum512([]byte(pwd))
	hashB64 = base64.StdEncoding.EncodeToString(hash[:])
	return
}

func (u *UsersDatabase) LoadUser(nick string) (user User, err error) {
	val, closer, err := u.DB.Get([]byte(nick))
	if err != nil {
		return User{}, fmt.Errorf("login error: nick=%s: %s", nick, err.Error())
	}

	if err = json.Unmarshal(val, &user); err != nil {
		return User{}, fmt.Errorf("login error: nick=%s: %s", nick, err.Error())
	}
	closer.Close()

	return
}

func (u *UsersDatabase) ListUsers() (users []User, err error) {
	iter, _ := u.DB.NewIter(nil)
	for iter.First(); iter.Valid(); iter.Next() {
		var user User

		if err = json.Unmarshal(iter.Value(), &user); err != nil {
			logs.ErrorLog("login error: nick=%s: %s\n", iter.Key(), err.Error())
			continue
		}
		users = append(users, user)
	}

	return users, err
}

func (u *UsersDatabase) ListAllowedUsers() (users []User, err error) {
	iter, _ := u.DB.NewIter(nil)
	for iter.First(); iter.Valid(); iter.Next() {
		var user User

		if err = json.Unmarshal(iter.Value(), &user); err != nil {
			logs.ErrorLog("login error: nick=%s: %s\n", iter.Key(), err.Error())
			continue
		}

		if !user.AppAnnuaire {
			continue
		}

		user.PwdHash = ""
		users = append(users, user)
	}

	return users, err
}

func (u *UsersDatabase) SetUser(user User) (err error) {
	val, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("unable to marshall user nick=%s: %s", user.Nick, err.Error())
	}

	if err = u.DB.Set([]byte(user.Nick), val, pebble.Sync); err != nil {
		return fmt.Errorf("unable to add user nick=%s: %s", user.Nick, err.Error())
	}

	return
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
	_, _, err := u.DB.Get([]byte(nick))
	if err != nil || err == pebble.ErrNotFound {
		return false
	}

	return true
}

func (u *UsersDatabase) AddUser(nick, pwd string) error {
	// Check if user exists
	existingUser, err := u.LoadUser(nick)
	if err == nil {
		// User exists, check if inactive for 30+ days
		lastAllowedConnection := time.Now().Add(-30 * 24 * time.Hour)
		if existingUser.LastConnect.Before(lastAllowedConnection) {
			// Delete old user data
			if err = u.DB.Delete([]byte(nick), pebble.Sync); err != nil {
				logs.ErrorLog("unable to delete inactive user nick=%s: %s\n", nick, err.Error())
				return fmt.Errorf("error cleaning up old user data")
			}
		} else {
			return fmt.Errorf("user already exists")
		}
	}

	// Create new user
	usr := User{
		Nick:        nick,
		PwdHash:     u.getHashB64(pwd),
		LastConnect: time.Now(),
	}

	if err := u.SetUser(usr); err != nil {
		return fmt.Errorf("unable to add user nick=%s: %s", nick, err.Error())
	}

	return nil
}

func (u *UsersDatabase) LogUser(nick, pwd string) bool {
	logs.InfoLog("attempt to log=%s\n", nick)

	user, err := u.LoadUser(nick)
	if err != nil {
		logs.ErrorLog("login error: nick=%s: %s\n", nick, err.Error())
		return false
	}

	if user.PwdHash == u.getHashB64(pwd) {
		user.LastConnect = time.Now()
		u.SetUser(user)
		return true
	} else {
		return false
	}
}

func (u *UsersDatabase) ChangePassword(nick string, pwd string) bool {
	user, err := u.LoadUser(nick)
	if err != nil {
		logs.ErrorLog("change pwd error: nick=%s: %s\n", nick, err.Error())
		return false
	}

	user.PwdHash = u.getHashB64(pwd)

	if err = u.SetUser(user); err != nil {
		logs.ErrorLog("change pwd error: nick=%s: %s\n", nick, err.Error())
		return false
	}

	logs.InfoLog("password changed for: nick=%s\n", nick)
	return true
}

func (u *UsersDatabase) DeleteUser(pseudo string) (err error) {
	if err = u.DB.Delete([]byte(pseudo), pebble.Sync); err != nil {
		logs.ErrorLog("cannot delete user=%s: %s\n", pseudo, err.Error())
	}

	logs.ErrorLog("deleted user=%s\n", pseudo)
	return
}

func (u *UsersDatabase) Quit() {
	u.DB.Close()
}
