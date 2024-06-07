package websocket

import (
	"github.com/gorilla/websocket"
	"github.com/injoyai/io"
	"net/http"
)

var _ io.ReadWriteCloser = (*Websocket)(nil)

type Websocket struct {
	*websocket.Conn
	io.Reader
}

func (this *Websocket) Write(p []byte) (int, error) {
	err := this.Conn.WriteMessage(websocket.BinaryMessage, p)
	return len(p), err
}

func (this *Websocket) ReadMessage() ([]byte, error) {
	_, bs, err := this.Conn.ReadMessage()
	return bs, err
}

func (this *Websocket) Close() error {
	return this.Conn.Close()
}

func NewEasy(url string) (*Websocket, error) {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	w := &Websocket{Conn: c}
	w.Reader = io.MReaderToReader(w)
	return w, nil
}

func New(dial *websocket.Dialer, url string, header http.Header) (*Websocket, error) {
	c, _, err := dial.Dial(url, header)
	if err != nil {
		return nil, err
	}
	w := &Websocket{Conn: c}
	w.Reader = io.MReaderToReader(w)
	return w, nil
}
