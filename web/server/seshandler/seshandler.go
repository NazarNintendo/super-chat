// Package seshandler provides an implementation of the session handler.
// Session handler is consulted when user visits the chat URL.
// It will validate the request and distribute clients across existing sessions or create new one.
package seshandler

import (
	"github.com/gorilla/websocket"
	log "gitlab.starlink.ua/high-school-prod/chat/logger"
	"gitlab.starlink.ua/high-school-prod/chat/server/database"
	"gitlab.starlink.ua/high-school-prod/chat/server/model"
	"gitlab.starlink.ua/high-school-prod/chat/server/session"
	"net/http"
	"strconv"
)

// SessionHandler - contains a map of currently opened sessions and a pointer to the DB
type SessionHandler struct {
	sessions map[string]*session.Session
	db       *database.Database
}

// New - will create a new session handler
func New() *SessionHandler {
	log.Logger.Infof("Started Session Handler")
	handler := &SessionHandler{
		sessions: make(map[string]*session.Session),
		db:       database.New(),
	}
	return handler
}

// Handle - will validate and distribute incoming requests over the right sessions
func (sh *SessionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	guid := r.URL.Query().Get("guid")
	token := r.URL.Query().Get("token")

	log.Logger.Infof("Received upgrade request from token:%s to guid:%s", token, guid)

	// Check the request's origin
	if !checkOrigin(r) {
		log.Logger.Warnf("Couldn't verify origin [%s], dropping connection...", r.Header.Get("Origin"))
		http.Error(w, "Bad Origin", http.StatusForbidden)
		return
	}

	// "Upgrading" HTTP request to the WebSocket protocol
	if upgrade := websocket.IsWebSocketUpgrade(r); upgrade == false {
		log.Logger.Warnf("WebSocket upgrade not present in the request, dropping connection...")
		http.Error(w, "Method not allowed. Need upgrade to websockets", http.StatusMethodNotAllowed)
		return
	}

	resp, err := fetchUser(token)

	// Check that fetching user didn't yield an error
	if err != nil {
		log.Logger.Warnf("Couldn't verify token: err [%s], dropping connection...", err)
		http.Error(w, "Bad token", http.StatusForbidden)
		return
	}

	userID, username := strconv.Itoa(resp.Data.User.ID), resp.Data.User.Username

	// Check that user has access to the given guid
	if !sh.db.ValidateUserChat(userID, guid) {
		log.Logger.Warnf("Bad guid [%s] : userID [%s], dropping connection...\n", guid, userID)
		http.Error(w, "Bad GUID", http.StatusForbidden)
		return
	}

	client := &model.Client{
		UserID:   userID,
		Username: username,
	}

	// Add user to the session when successfully validated
	sh.addToSession(w, r, guid, client)
}

// handleSession - method to add a user to an existing or new session, deletes the session after use
func (sh *SessionHandler) addToSession(w http.ResponseWriter, r *http.Request, guid string, client *model.Client) {
	sess, ok := sh.sessions[guid]

	// Create a new session or use the existing one
	if !ok {
		sess = session.New(guid, sh.db)
		sh.sessions[guid] = sess
	}

	deleteSession := sess.UpgradeAndHandle(w, r, client)

	// Delete current session if deleteSession is true
	if deleteSession {
		delete(sh.sessions, sess.GUID)
		log.Logger.Infof("Deleting session with GUID %v", sess.GUID)
	}
}
