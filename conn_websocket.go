package minigo

import (
	"context"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

type Websocket struct {
	write   http.ResponseWriter
	request *http.Request

	conn *websocket.Conn
	ctx  context.Context

	connected bool
}

func NewWebsocket(write http.ResponseWriter, request *http.Request) (*Websocket, error) {
	return &Websocket{
		write:   write,
		request: request,
	}, nil
}

func (ws *Websocket) Init() error {
	var err error

	ws.conn, err = websocket.Accept(ws.write, ws.request, &websocket.AcceptOptions{OriginPatterns: []string{"*"}})
	if err != nil {
		errorLog.Printf("unable to open websocket connection: %s\n", err.Error())
		return &ConnectorError{code: InvalidInit, raw: err}
	}

	defer ws.conn.Close(websocket.StatusInternalError, "websocket internal error, quitting")
	infoLog.Printf("new connection from IP=%s\n", ws.request.RemoteAddr)

	ws.conn.SetReadLimit(1024)

	var cancel context.CancelFunc
	ws.ctx, cancel = context.WithTimeout(ws.request.Context(), time.Minute*10)
	defer cancel()

	ws.connected = true
	return nil
}

func (ws *Websocket) Write(b []byte) error {
	err := ws.conn.Write(ws.ctx, websocket.MessageBinary, b)

	if err != nil {
		if websocket.CloseStatus(err) == websocket.StatusAbnormalClosure ||
			websocket.CloseStatus(err) == websocket.StatusNormalClosure {

			ws.connected = false

			return &ConnectorError{code: ClosedConnection, raw: err}
		} else {
			return &ConnectorError{code: Unsupported, raw: err}
		}
	}

	return nil
}

func (ws *Websocket) Read() ([]byte, error) {
	msgType, msg, err := ws.conn.Read(ws.ctx)

	if err != nil {
		if websocket.CloseStatus(err) == websocket.StatusAbnormalClosure ||
			websocket.CloseStatus(err) == websocket.StatusNormalClosure {

			ws.connected = false

			return nil, &ConnectorError{code: ClosedConnection, raw: err}
		} else {
			return nil, &ConnectorError{code: Unsupported, raw: err}
		}
	}

	if msgType != websocket.MessageBinary {
		return nil, &ConnectorError{code: InvalidData, raw: err}
	}

	return msg, nil
}

func (ws *Websocket) Connected() bool {
	return ws.connected
}
