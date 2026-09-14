package websocket

import (
	"fmt"
	"net/http"

	websocket_client "bombe_main_server/internal/websocket/client"

	"github.com/gorilla/websocket"
)

type WSHandlerS struct {
	client   *websocket_client.UserAndConn
	MsgCh    chan []byte
	upgrader websocket.Upgrader
}

func NewWSHandler() *WSHandlerS {
	msgCh := make(chan []byte)
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,

		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return &WSHandlerS{
		client:   websocket_client.NewUserAndConn(),
		MsgCh:    msgCh,
		upgrader: upgrader,
	}
}

func (ws *WSHandlerS) WSHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}

	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		conn.Close()
		return
	}

	ws.client.AddUserAndConn(userID, conn)


	defer func() {
		ws.client.DeleteUserAndConn(userID)
		conn.Close()
	}()

	ws.ReadIncomingMessages(conn)
}

func (ws *WSHandlerS) ReadIncomingMessages(
	conn *websocket.Conn,
) {
	for {
		_, msg, err := conn.ReadMessage()
		
		if err != nil {
			fmt.Println("WebSocket read error:", err)
			return
		}

		ws.MsgCh <- msg
	}
}

func (ws *WSHandlerS) SendToTargetUser(
	targetUserID string,
	data any,
) error {

	conn := ws.client.GetUserConnWithID(targetUserID)

	if conn == nil {
		fmt.Println(conn)
		return fmt.Errorf(
			"user %s is not connected",
			targetUserID,
		)
	}

	if err := conn.WriteJSON(data); err != nil {
		fmt.Println("send to target user error",err)
		return fmt.Errorf(
			"failed to send message to user %s: %w",
			targetUserID,
			err,
		)
	}

	return nil
}
