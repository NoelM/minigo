package main

import (
	"fmt"
	mgo "github.com/NoelM/minigo"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

const (
	NullPid = iota
	ChatPid
)

/*
var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }} // use default options

func server(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		msg, _ := GetMessage("BONJOUR! ")
		err = c.WriteMessage(websocket.BinaryMessage, msg)
		if err != nil {
			log.Println("write:", err)
			break
		}
		msg = GetMoveCursorDown(1)
		err = c.WriteMessage(websocket.BinaryMessage, msg)
		time.Sleep(time.Second)
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", server)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
*/

const (
	ConnHost = "192.168.1.10"
	ConnPort = "3615"
	ConnType = "tcp"
)

func main() {
	page := &ChatPage{
		users:    []User{},
		messages: []Message{},
	}

	// Listen for incoming connections.
	l, err := net.Listen(ConnType, ConnHost+":"+ConnPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + ConnHost + ":" + ConnPort)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn, page)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn, page *ChatPage) {
	tcpd := mgo.NewTCPDriver(conn, time.Second)
	mntl := mgo.NewMinitel(tcpd)

	var buffer []byte
	var key uint
	var err error

	username := Pseudos[rand.Intn(5)]
	page.usersMtx.Lock()
	page.users = append(page.users, User{
		logTime: time.Now(),
		logged:  true,
		pseudo:  username,
	})
	page.usersMtx.Unlock()

	lastMessageId := -1
	lastUserId := -1

	drawPage(mntl, page, username, lastUserId)
	for {
		key, err = mntl.RecvKey()
		if err != nil {
			updated := false
			if lastUserId != len(page.users) {
				updated = true
				lastUserId = len(page.users)
				updateConnected(mntl, page, username)
			}

			if lastMessageId != len(page.messages) {
				updated = true
				lastMessageId = len(page.messages)
				updateMessages(mntl, page, lastMessageId)
			}

			if updated {
				moveCursorToText(mntl, len(buffer))
			}

			log.Printf("unable to receive key: %s", err.Error())
			continue
		}

		switch key {
		case mgo.Envoi:
			page.messagesMtx.Lock()
			page.messages = append(page.messages, Message{
				timestamp: time.Now(),
				user:      username,
				message:   buffer,
			})
			page.messagesMtx.Unlock()
			buffer = nil
			updateTextZone(mntl)
			//pid = p.Envoi()
		case mgo.Retour:
			continue
			//pid = p.Retour()
		case mgo.Repetition:
			continue
			//pid = p.Repetition()
		case mgo.Guide:
			continue
			//pid = p.Guide()
		case mgo.Annulation:
			continue
			//pid = p.Annulation()
		case mgo.Sommaire:
			continue
			//pid = p.Sommaire()
		case mgo.Correction:
			continue
			//pid = p.Correction()
		case mgo.Suite:
			continue
			//pid = p.Suite()
		case mgo.ConnexionFin:
			continue
			//pid = p.ConnexionFin()
		default:
			buffer = append(buffer, byte(key))
		}
	}
}

func drawPage(minitel *mgo.Minitel, page *ChatPage, username string, lastMessageId int) {
	drawHeader(minitel)
	updateConnected(minitel, page, username)
	updateMessages(minitel, page, lastMessageId)
	updateTextZone(minitel)
}

func moveCursorToText(minitel *mgo.Minitel, msgLen int) {
	var buf []byte
	row := msgLen / mgo.ColonnesSimple
	col := msgLen - mgo.ColonnesSimple*row

	buf = mgo.GetMoveCursorXY(buf, col+1, 20+row)

	minitel.SendBytes(buf)
}

func updateTextZone(minitel *mgo.Minitel) {
	var buf []byte

	buf = mgo.GetMoveCursorXY(buf, 1, 20)
	buf = mgo.GetCleanScreenFromCursor(buf)

	buf = mgo.GetMoveCursorXY(buf, 25, 25)
	buf = mgo.GetMessage(buf, "MESSAGE + ENVOI")
	buf = mgo.GetMoveCursorXY(buf, 1, 20)
	buf = append(buf, mgo.GetByteWithParity(mgo.Con))

	minitel.SendBytes(buf)
}

func updateMessages(minitel *mgo.Minitel, page *ChatPage, lastMessageId int) {
	var buf []byte

	buf = mgo.GetMoveCursorXY(buf, 1, 4)
	page.messagesMtx.RLock()
	for mid := len(page.messages) - 1; mid >= 0; mid-- {
		msg := page.messages[mid]

		buf = mgo.GetMessage(buf, fmt.Sprintf("(%s) %s > ", msg.timestamp.Format("15:04"), msg.user))
		buf = mgo.GetMessage(buf, fmt.Sprintf("%s", msg.message))
		buf = mgo.GetMoveCursorReturn(buf, 1)
	}
	page.messagesMtx.RUnlock()

	minitel.SendBytes(buf)
}

func drawHeader(minitel *mgo.Minitel) {
	var buf []byte

	buf = append(buf, mgo.Ff) // Screen cleanup, cursor at 1,1, article separator
	buf = mgo.GetMessage(buf, fmt.Sprintf("=== MESSAGERIE ==="))

	minitel.SendBytes(buf)
}

func updateConnected(minitel *mgo.Minitel, page *ChatPage, username string) {
	var buf []byte

	buf = mgo.GetMoveCursorXY(buf, 1, 2)
	buf = mgo.GetMessage(buf, "Connectes : ")
	page.usersMtx.RLock()
	for i, u := range page.users {
		if u.pseudo == username {
			buf = mgo.GetMessage(buf, fmt.Sprintf("*%s", u.pseudo))
		} else {
			buf = mgo.GetMessage(buf, fmt.Sprintf("%s", u.pseudo))
		}

		if i < len(page.users)-1 {
			buf = mgo.GetMessage(buf, ", ")
		}
	}
	page.usersMtx.RUnlock()

	minitel.SendBytes(buf)
}
