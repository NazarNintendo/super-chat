// Package model contains a few general structs used throughout the code.
package model

import (
	"fmt"
)

// Message - a message entity.
type Message struct {
	UserID    string `json:"userId,omitempty"`
	Username  string `json:"username,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Text      string `json:"text,omitempty"`
	ChatGUID  string `json:"chatGuid,omitempty"`
}

// Payload - an entity of WS exchange body.
type Payload struct {
	Messages     []Message     `json:"messages,omitempty"`
	PageToken    string        `json:"pageToken,omitempty"`
	Notification *Notification `json:"notification,omitempty"`
}

// Notification - a message that notifies that a user is online / offline.
type Notification struct {
	Client   *Client `json:"client,omitempty"`
	IsOnline bool    `json:"isOnline,omitempty"`
}

/*
	Format:
	UserID: int; Timestamp: string; ChatGUID: string; Username: string; Text: string;
*/
func (m Message) String() string {
	return fmt.Sprintf("UserID: %v; Timestamp: %v; ChatGUID: %v; Username: %v; Text: %v", m.UserID, m.Timestamp, m.ChatGUID, m.Username, m.Text)
}

// Client - a struct for ws client information.
type Client struct {
	UserID   string `json:"userId,omitempty"`
	Username string `json:"username,omitempty"`
}

/*
	Format:
	UserID: int; Username: string;
*/
func (c Client) String() string {
	return fmt.Sprintf("UserID: %v; Username: %v", c.UserID, c.Username)
}
