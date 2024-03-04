package minigo

import (
	"context"

	"nhooyr.io/websocket"
)

type Websocket struct {
	conn *websocket.Conn
	ctx  context.Context

	connected bool
}

func NewWebsocket(conn *websocket.Conn, ctx context.Context) (*Websocket, error) {
	return &Websocket{
		conn: conn,
		ctx:  ctx,
	}, nil
}

func (ws *Websocket) Init() error {
	ws.connected = true
	return nil
}

func (ws *Websocket) Write(b []byte) error {
	//err := ws.conn.Write(ws.ctx, websocket.MessageBinary, b)
	err := ws.conn.Write(ws.ctx, websocket.MessageText, b)

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
	_, msg, err := ws.conn.Read(ws.ctx)

	if err != nil {
		if websocket.CloseStatus(err) == websocket.StatusAbnormalClosure ||
			websocket.CloseStatus(err) == websocket.StatusNormalClosure {

			ws.connected = false

			return nil, &ConnectorError{code: ClosedConnection, raw: err}
		} else {
			return nil, &ConnectorError{code: Unsupported, raw: err}
		}
	}

	return msg, nil
}

func (ws *Websocket) Connected() bool {
	return ws.connected
}

func (ws *Websocket) Disconnect() {
	ws.connected = false
}
