package websocket_client

import (
	"sync"

	"github.com/gorilla/websocket"
)

type UserAndConn struct {
	Client map[string]*websocket.Conn
	mu     sync.RWMutex
}

func NewUserAndConn() *UserAndConn {
	return &UserAndConn{
		Client: make(map[string]*websocket.Conn),
	}
}

func (u *UserAndConn) AddUserAndConn(
	userID string,
	conn *websocket.Conn,
) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.Client[userID] = conn
}

func (u *UserAndConn) GetUserConnWithID(
	userID string,
) *websocket.Conn {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.Client[userID]
}

func (u *UserAndConn) DeleteUserAndConn(
	userID string,
) {
	u.mu.Lock()
	defer u.mu.Unlock()

	delete(u.Client, userID)
}