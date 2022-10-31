package main

import (
	"fmt"
	mgo "github.com/NoelM/minigo"
	"log"
	"net"
	"os"
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
		go handleRequest(conn)
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	tcpd := mgo.NewTCPDriver(conn)
	mntl := mgo.NewMinitel(tcpd)

	var pid uint
	var key uint
	var err error

	p := mgo.NewTestPage(mntl)
	p.Draw()
	for {
		log.Printf("PID=%d", pid)
		if key, err = mntl.RecvKey(); err != nil {
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
	}
}
