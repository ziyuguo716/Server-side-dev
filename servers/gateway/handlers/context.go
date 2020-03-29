package handlers

import (
	"sync"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/models/users"
	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/sessions"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

//TODO: define a handler context struct that
//will be a receiver on any of your HTTP
//handler functions that need access to
//globals, such as the key used for signing
//and verifying SessionIDs, the session store
//and the user store

//SessionContext captures the signing key, session and user info
type SessionContext struct {
	Key     string               `json:"-"`
	Session *sessions.RedisStore `json:"session"`
	User    users.Store          `json:"user"`
}

//WebsocketContext stores map of userid to websocket conn and a thread safe lock
type WebsocketContext struct {
	Context       *SessionContext         `json:"-"`
	Connections   map[int]*websocket.Conn `json:"-"`
	Lock          *sync.Mutex             `json:"-"`
	RabbitChannel *amqp.Channel           `json:"-"`
}
