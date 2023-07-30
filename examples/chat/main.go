package main

import (
	"fmt"
	mgo "github.com/NoelM/minigo"
	"github.com/gobwas/ws"
	"log"
	"math/rand"
	"net/http"
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

/*const (
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
*/

func main() {
	page := &ChatPage{
		users:    []User{},
		messages: []Message{},
	}

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		wsd := mgo.NewWebSocketDriver(conn)
		mini := mgo.NewMinitel(wsd)

		username := Pseudos[rand.Intn(4)]

		page.usersMtx.Lock()
		page.users = append(page.users, User{
			logTime: time.Now(),
			logged:  true,
			pseudo:  username,
		})
		page.usersMtx.Unlock()

		handleRequest(mini, page, username)
	})

	err := http.ListenAndServe("192.168.1.27:3615", fn)
	log.Fatal(err)
}

// Handles incoming requests.
func handleRequest(mntl *mgo.Minitel, page *ChatPage, username string) {
	var buffer []byte

	recvChan := make(chan uint)

	lastMessageId := -1
	lastUserId := -1

	go func() {
		for {
			if mntl.IsClosed() {
				return
			}

			key, err := mntl.ReadKey()
			if err != nil {
				fmt.Printf("unable to read key: %s", err)
			}
			recvChan <- key
		}
	}()

	mntl.CursorOn()
	//drawPage(mntl, page, username)
	for {
		if mntl.IsClosed() {
			return
		}

		var key uint

		select {
		case key = <-recvChan:
		default:
			updated := false
			if lastUserId != len(page.users) {
				updated = true
				lastUserId = len(page.users)
				//updateConnected(mntl, page, username)
			}

			if lastMessageId != len(page.messages) {
				updated = true
				lastMessageId = len(page.messages)
				//updateMessages(mntl, page)
			}

			if updated {
				//moveCursorToText(mntl, len(buffer))
			}

			time.Sleep(10 * time.Millisecond)
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

func drawPage(minitel *mgo.Minitel, page *ChatPage, username string) {
	drawHeader(minitel)
	updateConnected(minitel, page, username)
	updateMessages(minitel, page)
	updateTextZone(minitel)
}

func moveCursorToText(minitel *mgo.Minitel, msgLen int) {
	var buf []byte
	row := msgLen / mgo.ColonnesSimple
	col := msgLen - mgo.ColonnesSimple*row

	buf = mgo.GetMoveCursorXY(buf, col+1, 20+row)

	minitel.WriteBytes(buf)
}

func updateTextZone(minitel *mgo.Minitel) {
	var buf []byte

	buf = mgo.GetMoveCursorXY(buf, 1, 20)
	buf = mgo.GetCleanScreenFromCursor(buf)

	buf = mgo.GetMoveCursorXY(buf, 25, 25)
	buf = mgo.GetMessage(buf, "MESSAGE + ENVOI")
	buf = mgo.GetMoveCursorXY(buf, 1, 20)
	buf = append(buf, mgo.GetByteWithParity(mgo.CursorOn))

	minitel.WriteBytes(buf)
}

func updateMessages(minitel *mgo.Minitel, page *ChatPage) {
	var buf []byte
	var countNbLines int

	buf = mgo.GetMoveCursorXY(buf, 1, 4)
	page.messagesMtx.RLock()
	for mid := len(page.messages) - 1; mid >= 0; mid-- {
		msg := page.messages[mid]

		formatted := fmt.Sprintf("(%s) %s > %s", msg.timestamp.Format("15:04"), msg.user, msg.message)
		countNbLines += 1 + len(formatted)/mgo.ColonnesSimple
		if countNbLines+4 > 18 {
			break
		}

		buf = mgo.GetMessage(buf, formatted)

		buf = mgo.GetMoveCursorReturn(buf, 1)
	}
	page.messagesMtx.RUnlock()

	minitel.WriteBytes(buf)
}

func drawHeader(minitel *mgo.Minitel) {
	var buf []byte

	buf = append(buf, mgo.Ff) // Screen cleanup, cursor at 1,1, article separator
	buf = mgo.GetMessage(buf, fmt.Sprintf("=== MESSAGERIE ==="))

	minitel.WriteBytes(buf)
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

	minitel.WriteBytes(buf)
}
