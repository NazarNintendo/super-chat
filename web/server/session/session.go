// Package session manages the full lifecycle of a session.
// It sends and receives messages to and from the clients, adds and deletes clients etc.
package session

import (
	"github.com/gorilla/websocket"
	log "gitlab.starlink.ua/high-school-prod/chat/logger"
	"gitlab.starlink.ua/high-school-prod/chat/server/database"
	"gitlab.starlink.ua/high-school-prod/chat/server/model"
	"net/http"
	"time"
)

// Session - handles a single chat session for a set of clients.
type Session struct {
	GUID      string
	db        *database.Database
	clients   map[*websocket.Conn]*model.Client
	broadcast chan model.Message
}

// New will construct and return a new session.
func New(GUID string, dbP *database.Database) *Session {
	session := &Session{
		GUID:      GUID,
		db:        dbP,
		clients:   make(map[*websocket.Conn]*model.Client),
		broadcast: make(chan model.Message),
	}

	log.Logger.Infof("Opened a new chat session with id %v", session.GUID)

	go session.handleMessages() // Listens to the incoming messages.

	return session
}

// A go routine that monitors broadcast channel and populates clients' feed.
func (session *Session) handleMessages() {
	for {
		msg := <-session.broadcast
		log.Logger.Infof("Transmitting to all clients: %s", msg)
		for client := range session.clients {
			msgs := make([]model.Message, 0)
			msgs = append(msgs, msg)
			payload := model.Payload{
				Messages:  msgs,
				PageToken: "",
			}
			err := client.WriteJSON(&payload)
			if err != nil {
				log.Logger.Error(err)
				return
			}
		}
	}
}

// Declaring an upgrader in order to establish the WebSocket connection.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// UpgradeAndHandle upgrades an incoming request to a WebSocket and handles messages.
// Returns true if session needs to be deleted, otherwise - false.
func (session *Session) UpgradeAndHandle(w http.ResponseWriter, r *http.Request, client *model.Client) (deleteSession bool) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Logger.Error(err)
	}

	session.clients[conn] = client
	log.Logger.Infof("Adding client with id %s to the session with GUID %s.", client.UserID, session.GUID)

	// Notify other clients that user has gone online.
	session.sendOnlineNotification(*client, true)

	// Notify this user about other clients' statuses.
	session.sendStatuses(conn)

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Logger.Error(err)
		}

		log.Logger.Infof("Deleting client with id %s from the session with GUID %s.", client.UserID, session.GUID)
		deleteSession = session.deleteClient(conn)

		// Notify other clients that user has gone offline.
		session.sendOnlineNotification(*client, false)

		return
	}()

	// Send the recent messages to the new client.
	payload := session.db.ReadRecentMessages(session.GUID, 25, "")
	log.Logger.Infof("Sending \n%s", payload.Messages)
	err = conn.WriteJSON(&payload)
	if err != nil {
		log.Logger.Error(err)
		return
	}

	for {
		// Reading the received payload as a JSON.
		payload := model.Payload{}
		err := conn.ReadJSON(&payload)
		if err != nil {
			log.Logger.Error(err)
			return
		}

		// If payload has a page token, respond with the corresponding messages, and do not broadcast.
		if payload.PageToken != "" {
			log.Logger.Infof("Received page token %s", payload.PageToken)
			payload := session.db.ReadRecentMessages(session.GUID, 25, payload.PageToken)
			err := conn.WriteJSON(&payload)
			if err != nil {
				log.Logger.Error(err)
				return
			}
			continue
		}

		receivedMsg := payload.Messages[0]

		// If message length is more than 8K chars - do not broadcast and do not save to the DB.
		if len(receivedMsg.Text) > 8000 {
			log.Logger.Errorf("Message too long - length [%v], from client [%s]", len(receivedMsg.Text), client)
			continue
		}

		timestamp := time.Now().UTC().Format("01-02-2006 15:04:05.000000 UTC")
		receivedMsg.Timestamp = timestamp
		receivedMsg.ChatGUID = session.GUID
		receivedMsg.UserID = session.clients[conn].UserID
		receivedMsg.Username = session.clients[conn].Username

		log.Logger.Infof("Message received: %s", receivedMsg)

		session.db.MsgChan <- receivedMsg
		session.broadcast <- receivedMsg
	}
}

// deleteClient deletes a client from session clients.
// Returns true if the session needs to be deleted, otherwise - false.
func (session *Session) deleteClient(conn *websocket.Conn) (deleteSession bool) {
	delete(session.clients, conn)

	// If this client was the last one in the session - delete the session.
	if len(session.clients) == 0 {
		deleteSession = true
	} else {
		deleteSession = false
	}
	return deleteSession
}

// sendOnlineNotification will send a notification to all clients when user goes joins or leaves the session.
func (session *Session) sendOnlineNotification(user model.Client, isOnline bool) {
	log.Logger.Infof("Sending notification to all clients: [%s] is online [%v]", user, isOnline)
	for conn := range session.clients {
		payload := model.Payload{
			Notification: &model.Notification{
				Client:   &user,
				IsOnline: isOnline,
			},
		}

		// Avoid sending notifications to ourselves.
		if session.clients[conn].UserID != user.UserID {
			err := conn.WriteJSON(&payload)
			if err != nil {
				log.Logger.Error(err)
				return
			}
		}
	}
}

// sendStatuses will send all clients' statuses to the user when he joins the session.
func (session *Session) sendStatuses(conn *websocket.Conn) {
	log.Logger.Infof("Sending all statuses to client [%s]", session.clients[conn])
	for _, client := range session.clients {
		payload := model.Payload{
			Notification: &model.Notification{
				Client:   client,
				IsOnline: true,
			},
		}

		// Avoid sending notifications to ourselves.
		if client.UserID != session.clients[conn].UserID {
			err := conn.WriteJSON(&payload)
			if err != nil {
				log.Logger.Error(err)
				return
			}
		}
	}
}
