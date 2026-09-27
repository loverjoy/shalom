package models

import (
	"time"

	"github.com/google/uuid"
)

// Chat message stored in MongoDB
type Message struct {
	ID            uuid.UUID     `bson:"_id,omitempty" json:"id"`
	ChatRoomID    string        `bson:"chat_room_id" json:"chat_room_id"`
	SenderID      uuid.UUID     `bson:"sender_id" json:"sender_id"`
	SenderName    string        `bson:"sender_name" json:"sender_name"`
	Content       string        `bson:"content" json:"content"`
	Type          MessageType   `bson:"type" json:"type"` // text, file, voice, image, video, poll, system
	FileURL       string        `bson:"file_url,omitempty" json:"file_url,omitempty"`
	FileName      string        `bson:"file_name,omitempty" json:"file_name,omitempty"`
	FileSize      int64         `bson:"file_size,omitempty" json:"file_size,omitempty"`
	VoiceDuration int           `bson:"voice_duration,omitempty" json:"voice_duration,omitempty"`
	ReplyTo       *uuid.UUID    `bson:"reply_to,omitempty" json:"reply_to,omitempty"`
	ForwardedFrom *uuid.UUID    `bson:"forwarded_from,omitempty" json:"forwarded_from,omitempty"`
	IsPinned      bool          `bson:"is_pinned" json:"is_pinned"`
	IsEdited      bool          `bson:"is_edited" json:"is_edited"`
	IsDeleted     bool          `bson:"is_deleted" json:"is_deleted"`
	Reactions     []Reaction    `bson:"reactions,omitempty" json:"reactions,omitempty"`
	Poll          *Poll         `bson:"poll,omitempty" json:"poll,omitempty"`
	CreatedAt     time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at" json:"updated_at"`
}

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeFile   MessageType = "file"
	MessageTypeVoice  MessageType = "voice"
	MessageTypeImage  MessageType = "image"
	MessageTypeVideo  MessageType = "video"
	MessageTypePoll   MessageType = "poll"
	MessageTypeSystem MessageType = "system"
)

type Reaction struct {
	UserID uuid.UUID `bson:"user_id" json:"user_id"`
	Emoji  string    `bson:"emoji" json:"emoji"`
}

type Poll struct {
	Question   string       `bson:"question" json:"question"`
	Options    []PollOption `bson:"options" json:"options"`
	IsAnonymous bool        `bson:"is_anonymous" json:"is_anonymous"`
	ExpiresAt  *time.Time   `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
}

type PollOption struct {
	Text      string       `bson:"text" json:"text"`
	VoteCount int          `bson:"vote_count" json:"vote_count"`
	Voters    []uuid.UUID  `bson:"voters,omitempty" json:"voters,omitempty"`
}

type ChatRoom struct {
	ID           string    `bson:"_id,omitempty" json:"id"`
	Type         string    `bson:"type" json:"type"` // public, private, meeting, group
	Name         string    `bson:"name" json:"name"`
	MeetingID    string    `bson:"meeting_id,omitempty" json:"meeting_id,omitempty"`
	Members      []string  `bson:"members" json:"members"`
	CreatedBy    uuid.UUID `bson:"created_by" json:"created_by"`
	LastMessage  *Message  `bson:"last_message,omitempty" json:"last_message,omitempty"`
	MessageCount int       `bson:"message_count" json:"message_count"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
}
