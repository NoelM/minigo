package databases

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cockroachdb/pebble"
)

var infoLog = log.New(os.Stdout, "[notel] info:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var warnLog = log.New(os.Stdout, "[notel] warn:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
var errorLog = log.New(os.Stdout, "[notel] error:", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

type User struct {
	Nick        string    `json:"nick"`
	PwdHash     string    `json:"pwd_hash"`
	Bio         string    `json:"bio"`
	Tel         string    `json:"tel"`
	Location    string    `json:"location"`
	LastConnect time.Time `json:"last_connect"`
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

func (u *UsersDatabase) loadUser(nick string) (user User, err error) {
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

func (u *UsersDatabase) setUser(user User) (err error) {
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
	if u.UserExists(nick) {
		return fmt.Errorf("user already exists")
	}

	usr := User{
		Nick:        nick,
		PwdHash:     u.getHashB64(pwd),
		LastConnect: time.Now(),
	}

	if err := u.setUser(usr); err != nil {
		return fmt.Errorf("unable to add user nick=%s: %s", nick, err.Error())
	}

	return nil
}

func (u *UsersDatabase) LogUser(nick, pwd string) bool {
	infoLog.Printf("attempt to log=%s\n", nick)

	user, err := u.loadUser(nick)
	if err != nil {
		errorLog.Printf("login error: nick=%s: %s\n", nick, err.Error())
		return false
	}

	lastAllowedConnection := time.Now().Add(-30 * 24 * time.Hour)
	if user.LastConnect.Before(lastAllowedConnection) {
		if err = u.DB.Delete([]byte(nick), pebble.Sync); err != nil {
			errorLog.Printf("login error: nick=%s: %s\n", nick, err.Error())
		}
		return false
	}

	return user.PwdHash == u.getHashB64(pwd)
}

func (u *UsersDatabase) ChangePassword(nick string, pwd string) bool {
	user, err := u.loadUser(nick)
	if err != nil {
		errorLog.Printf("change pwd error: nick=%s: %s\n", nick, err.Error())
		return false
	}

	user.PwdHash = u.getHashB64(pwd)

	if err = u.setUser(user); err != nil {
		errorLog.Printf("change pwd error: nick=%s: %s\n", nick, err.Error())
		return false
	}

	infoLog.Printf("password changed for: nick=%s\n", nick)
	return true
}

func (u *UsersDatabase) Quit() {
	u.DB.Close()
}
