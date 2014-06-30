package daemon

import (
	"fmt"
	"code.google.com/p/go.net/websocket"
	"time"
)

type Client struct {
	ws *websocket.Conn
	start time.Time
}


func (client *Client) String() string {
	uptime := time.Now().UTC().Sub(client.start)
	var addr string
	if client.ws.Request().Header.Get("X-Real-Ip") != "" {
		addr = client.ws.Request().Header.Get("X-Real-Ip")
	} else {
		addr = client.ws.Request().RemoteAddr
	}
	return fmt.Sprintf("%s is up from %s", addr, uptime)
}