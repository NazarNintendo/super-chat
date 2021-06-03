package database

import (
	"fmt"
	"github.com/joho/godotenv"
	log "gitlab.starlink.ua/high-school-prod/chat/logger"
	"math/rand"
	"testing"
	"time"
)

// Loads values from .env into the system
func init() {
	if err := godotenv.Load(); err != nil {
		log.Logger.Fatal(err)
	}
}

func TestEstablishConnection(t *testing.T) {
	got := establishConnection()

	// Check type of the returned value
	if fmt.Sprintf("%T", got) != "*sql.DB" {
		t.Errorf("Type is %T, want *sql.DB", got)
	}

	// Checking whether this is a working instance of the DB
	_, err := got.Query("SELECT username FROM users WHERE id = 1")
	if err != nil {
		t.Errorf("Error when querying %v, want none", err)
	}
}

func TestNew(t *testing.T) {
	got := New()

	// Check type of the returned value
	if fmt.Sprintf("%T", got) != "*database.Database" {
		t.Errorf("Type is %T, want *database.Database", got)
	}

	// Check type of the 'sessGuid' and 'MsgChan' fields
	if fmt.Sprintf("%T", got.psql) != "*sql.DB" {
		t.Errorf("Type is %T, want *sql.DB", got.psql)
	}
	if fmt.Sprintf("%T", got.MsgChan) != "chan model.Message" {
		t.Errorf("Type is %T, want chan model.Message", got.MsgChan)
	}
}

func TestReadRecentMessages(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(10)

	got := New().ReadRecentMessages("fake-guid", num, "")

	// Check number of messages is the same as passed 'num'
	if len(got.Messages) != 0 {
		t.Errorf("Read %v messages, want 0", len(got.Messages))
	}
}
