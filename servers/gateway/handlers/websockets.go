package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/models/users"
	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/sessions"
	"github.com/gorilla/websocket"
)

//TODO: add a handler that upgrades clients to a WebSocket connection
//and adds that to a list of WebSockets to notify when events are
//read from the RabbitMQ server. Remember to synchronize changes
//to this list, as handlers are called concurrently from multiple
//goroutines.

//TODO: start a goroutine that connects to the RabbitMQ server,
//reads events off the queue, and broadcasts them to all of
//the existing WebSocket connections that should hear about
//that event. If you get an error writing to the WebSocket,
//just close it and remove it from the list
//(client went away without closing from
//their end). Also make sure you start a read pump that
//reads incoming control messages, as described in the
//Gorilla WebSocket API documentation:
//http://godoc.org/github.com/gorilla/websocket

//RabbitMessage stores object from RMQ
type RabbitMessage struct {
	Type      string  `json:"type"`
	Message   Message `json:"message"`
	MessageID string  `json:"messageID"`
	ChannelID string  `json:"channelID"`
	Channel   Channel `json:"channel"`
	UserIDs   []int   `json:"userIDs"`
}

//Channel stores channel info
type Channel struct {
	ID          string       `json:"_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Private     bool         `json:"private"`
	Members     []users.User `json:"members"`
	CreatedAt   string       `json:"createdAt"`
	Creator     users.User   `json:"creator"`
	EditedAt    string       `json:"editedAt"`
}

//Message stores message info
type Message struct {
	ChannelID string     `json:"channelID"`
	Body      string     `json:"body"`
	CreatedAt string     `json:"createdAt"`
	Creator   users.User `json:"creator"`
	EditedAt  string     `json:"editedAt"`
}

// Control messages for websocket
const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return r.Header.Get("Origin") == "https://ziyuguo.me"
	},
}

//WebSocketHandler handles all requests for general user actions
func (wsc *WebsocketContext) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// handle the websocket handshake
	if r.Header.Get("Origin") != "https://ziyuguo.me" {
		http.Error(w, "Websocket Connection Refused", 403)
		return
	}
	//Authenticate
	sessionState := &SessionState{}
	_, err := sessions.GetState(r, wsc.Context.Key, wsc.Context.Session, sessionState)
	if err != nil {
		http.Error(w, "Unauthorized", 401)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to open websocket connection", 401)
	}

	user := sessionState.User
	userID := int(user.ID)
	//TODO: add conn to connections list
	wsc.InsertConnection(conn, userID)

	go echo(conn, wsc, userID)
}

//echo reads incoming messages
func echo(conn *websocket.Conn, wsc *WebsocketContext, userID int) {
	defer conn.Close()
	defer wsc.RemoveConnection(userID)

	for { // infinite loop
		messageType, _, err := conn.ReadMessage()
		if messageType == websocket.CloseMessage {
			fmt.Println("Close message received.")
			break
		} else if err != nil {
			fmt.Println("Error reading message.")
			break
		}
	}
}

//StartRabbitConsumer connects to MQ and consumes events from queue
func (wsc *WebsocketContext) StartRabbitConsumer() {
	rabbitQueue, err := wsc.RabbitChannel.QueueDeclare(
		"info441", // name
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := wsc.RabbitChannel.Consume(
		rabbitQueue.Name, // queue
		"",               // consumer
		true,             // auto-ack
		false,            // exclusive
		false,            // no-local
		false,            // no-wait
		nil,              // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		for {
			for d := range msgs {
				m := RabbitMessage{}
				err := json.Unmarshal([]byte(d.Body), &m)
				if err != nil {
					fmt.Println("Error reading json", err)
					break
				}
				//check if channel is private
				//If yes, broadcast only to users in userID array
				//If no, broadcat to all
				if len(m.UserIDs) != 0 {
					log.Printf("write to private users: " + m.Type)
					wsc.WriteToPrivateConnections(m, m.UserIDs)
				} else {
					log.Printf("write to ALL: " + m.Type)
					wsc.WriteToAllConnections(m)
				}
			}
		}

	}()
}

//InsertConnection Thread-safe method for inserting a connection
func (wsc *WebsocketContext) InsertConnection(conn *websocket.Conn, userID int) {
	wsc.Lock.Lock()
	// insert socket connection
	wsc.Connections[userID] = conn
	wsc.Lock.Unlock()
}

//RemoveConnection Thread-safe method for inserting a connection
func (wsc *WebsocketContext) RemoveConnection(userID int) {
	wsc.Lock.Lock()
	// insert socket connection
	delete(wsc.Connections, userID)
	wsc.Lock.Unlock()
}

//failOnError send error message, if any
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

//WriteToAllConnections broadcast to all websocket conns
func (wsc *WebsocketContext) WriteToAllConnections(msg interface{}) {
	for k, conn := range wsc.Connections {
		writeError := conn.WriteJSON(msg)
		//if error writing, close connection
		if writeError != nil {
			conn.Close()
			wsc.RemoveConnection(k)
		}
	}
}

//WriteToPrivateConnections broadcast to listed websocket conns
func (wsc *WebsocketContext) WriteToPrivateConnections(msg interface{}, userIDs []int) {
	for _, id := range userIDs {
		conn := wsc.Connections[id]
		writeError := conn.WriteJSON(msg)
		//if error writing, close connection
		if writeError != nil {
			conn.Close()
			wsc.RemoveConnection(id)
		}
	}
}
