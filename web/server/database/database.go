// Package database provides an interface for database interaction, including
// opening a connection, reading and writing to the DB etc.
package database

import (
	"database/sql"
	_ "fmt"
	_ "github.com/lib/pq"
	log "gitlab.starlink.ua/high-school-prod/chat/logger"
	"gitlab.starlink.ua/high-school-prod/chat/server/model"
	"math"
	"os"
	"strconv"
)

// Database - an abstraction for DB interaction.
type Database struct {
	psql    *sql.DB
	MsgChan chan model.Message
}

// New - will construct and return a Database instance.
func New() *Database {
	db := &Database{
		psql:    establishConnection(),
		MsgChan: make(chan model.Message),
	}
	go db.databaseHandler()
	log.Logger.Infof("Created a new Database instance")
	return db
}

// establishConnection - will establish a connection with the DB.
func establishConnection() *sql.DB {
	dbSource, exists := os.LookupEnv("DB_SOURCE")
	if exists {
		psql, err := sql.Open("postgres", dbSource)
		if err != nil {
			log.Logger.Fatal(err)
		}
		log.Logger.Infof("Opened a new DB connection")
		return psql
	} else {
		log.Logger.Fatal("No DB_SOURCE in .env file")
		return nil
	}
}

// ReadRecentMessages - will read recent messages from the db and return as a slice.
func (db *Database) ReadRecentMessages(guid string, numMsgs int, pageToken string) (payload model.Payload) {
	var (
		msgs *sql.Rows
		err  error
	)

	// Read last numMsgs messages from the DB.
	stmt, err := db.psql.Prepare(`SELECT m.id, user_id, username, text, timestamp, chat_guid 
										 FROM messages m 
   										 	INNER JOIN users u ON u.id = m.user_id 
										 WHERE m.chat_guid=$1 
										 AND m.id < $3
										 ORDER BY m.timestamp DESC
										 LIMIT $2`)
	if err != nil {
		log.Logger.Fatalf("Error when preparing SQL statement - %s", err)
	}

	// If msgId is provided - select messages that are 'older' (their id is lesser).
	// If it's not - select messages, which ids are lesser that an arbitrary big number.
	msgID := decrypt(pageToken)
	log.Logger.Infof("Decrypted page token %s into message ID %s", pageToken, msgID)
	if msgID != "" {
		msgs, err = stmt.Query(guid, numMsgs, msgID)
	} else {
		msgs, err = stmt.Query(guid, numMsgs, strconv.Itoa(math.MaxInt32))
	}
	if err != nil {
		log.Logger.Fatalf("Error when querying SQL statement - %s", err)
	}

	// Close the connection and return it to the pool.
	defer func() {
		err := msgs.Close()
		if err != nil {
			log.Logger.Fatal(err)
		}
	}()

	var (
		lastMsgs  = make([]model.Message, 0)
		lastMsgID string
	)

	// Iterate over queried data, scan it into the variables and append to the destination string.
	for msgs.Next() {
		// Declaring variables to describe types of queried data.
		var (
			msgID     string
			userID    string
			userName  string
			timestamp string
			text      string
			chatGUID  string
		)

		err := msgs.Scan(&msgID, &userID, &userName, &text, &timestamp, &chatGUID)
		if err != nil {
			log.Logger.Fatal(err)
		}
		lastMsgs = append(lastMsgs, model.Message{userID, userName, timestamp, text, chatGUID})
		lastMsgID = msgID
	}
	log.Logger.Infof("Fetched %v messages", len(lastMsgs))

	// Check for errors after we’re done iterating over the rows.
	if err = msgs.Err(); err != nil {
		log.Logger.Fatal(err)
	}

	// Encrypt the last message's ID as a new page token
	var newPageToken = encrypt(lastMsgID)
	if lastMsgID != "" {
		log.Logger.Infof("Encrypted message ID %s into page token %s", lastMsgID, newPageToken)
	} else {
		log.Logger.Infof("Returned empty page token")
	}

	payload = model.Payload{
		Messages:  lastMsgs,
		PageToken: newPageToken,
	}

	return payload
}

// ValidateUserChat - will validate that the chatGUID exists in the db and that the userID has access to the guid.
func (db *Database) ValidateUserChat(userID, chatGUID string) bool {
	stmt, err := db.psql.Prepare("SELECT COUNT(*) FROM chats_users WHERE user_id=$1 AND chat_guid=$2")
	if err != nil {
		log.Logger.Fatal(err)
	}
	rows, err := stmt.Query(userID, chatGUID)
	if err != nil {
		log.Logger.Fatal(err)
	}

	// Close the connection and return it to the pool.
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Logger.Fatal(err)
		}
	}()
	matches := 0
	rows.Next()
	if err := rows.Scan(&matches); err != nil {
		log.Logger.Fatal(err)
	}

	return matches > 0
}

// A go routine that monitors message channel and updates the database with new messages.
func (db *Database) databaseHandler() {
	for {
		msg := <-db.MsgChan

		// Insert the message into the DB.
		stmt, err := db.psql.Prepare("INSERT INTO messages(user_id, text, timestamp, chat_guid) VALUES($1, $2, $3, $4)")
		if err != nil {
			log.Logger.Fatal(err)
		}
		_, err = stmt.Exec(msg.UserID, msg.Text, msg.Timestamp, msg.ChatGUID)
		if err != nil {
			log.Logger.Fatal(err)
		}

		log.Logger.Infof("Successfully saved the message to the DB")
	}
}
