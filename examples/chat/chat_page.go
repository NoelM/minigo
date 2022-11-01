package main

import (
	mgo "github.com/NoelM/minigo"
	"log"
	"sync"
	"time"
)

type User struct {
	logTime time.Time
	logged  bool
	pseudo  string
}

type Message struct {
	timestamp time.Time
	user      string
	message   []byte
}

type ChatPage struct {
	usersMtx    sync.RWMutex
	users       []User
	messagesMtx sync.RWMutex
	messages    []Message
}

func NewChatPage() ChatPage {
	return ChatPage{
		users:    []User{},
		messages: []Message{},
	}
}

var Pseudos = []string{
	"antarian",
	"balorian",
	"quark",
	"kolos",
}

func (p *ChatPage) registerNewUser() string {
	username := "foo" //Pseudos[p.uid]
	p.users = append(p.users,
		User{
			logTime: time.Now(),
			logged:  true,
			pseudo:  username,
		})

	return username
}

func (p *ChatPage) NewSession(driver mgo.Driver) uint {
	minitel := mgo.NewMinitel(driver)
	var pid uint
	var key uint
	var err error

	p.Draw()
	for {
		if key, err = minitel.RecvKey(); err != nil {
			log.Printf("unable to receive key: %s", err.Error())
		}

		switch key {
		case mgo.Envoi:
			pid = p.Envoi()
		case mgo.Retour:
			pid = p.Retour()
		case mgo.Repetition:
			pid = p.Repetition()
		case mgo.Guide:
			pid = p.Guide()
		case mgo.Annulation:
			pid = p.Annulation()
		case mgo.Sommaire:
			pid = p.Sommaire()
		case mgo.Correction:
			pid = p.Correction()
		case mgo.Suite:
			pid = p.Suite()
		case mgo.ConnexionFin:
			pid = p.ConnexionFin()
		default:
			p.NewKey(key)
		}

		if pid != NullPid {
			return pid
		}
	}
}

func (p *ChatPage) Draw() {
	var buf []byte

	buf = append(buf, mgo.Ff) // Screen cleanup, cursor at 1,1, article separator
	buf = mgo.GetMessage(buf, "Connecte:")
	buf = mgo.GetMoveCursorReturn(buf, 1)
	buf = mgo.GetMessage(buf, "Telematique 2000")
	buf = mgo.GetMoveCursorReturn(buf, 1)
	buf = mgo.GetMessage(buf, " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_abcdefghijklmnopqrstuvwxyz")
}

func (p *ChatPage) Envoi() uint {
	return NullPid
}

func (p *ChatPage) Retour() uint {
	return NullPid
}

func (p *ChatPage) Repetition() uint {
	return NullPid
}

func (p *ChatPage) Guide() uint {
	return NullPid
}

func (p *ChatPage) Annulation() uint {
	return NullPid
}

func (p *ChatPage) Sommaire() uint {
	return NullPid
}

func (p *ChatPage) Correction() uint {
	return NullPid
}

func (p *ChatPage) Suite() uint {
	return NullPid
}

func (p *ChatPage) ConnexionFin() uint {
	return NullPid
}

func (p *ChatPage) NewKey(k uint) {
	log.Printf("got key: %d", k)
}
