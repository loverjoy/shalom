package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email" unique:"true"`
	Username     string     `json:"username" db:"username" unique:"true"`
	DisplayName  string     `json:"display_name" db:"display_name"`
	PasswordHash string     `json:"-" db:"password_hash"`
	AvatarURL    string     `json:"avatar_url" db:"avatar_url"`
	Status       string     `json:"status" db:"status"` // online, offline, away
	LastSeen     time.Time  `json:"last_seen" db:"last_seen"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type Meeting struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Title         string     `json:"title" db:"title"`
	Description   string     `json:"description" db:"description"`
	HostID        uuid.UUID  `json:"host_id" db:"host_id"`
	MeetingCode   string     `json:"meeting_code" db:"meeting_code" unique:"true"`
	JoinLink      string     `json:"join_link" db:"join_link"`
	Status        string     `json:"status" db:"status"` // scheduled, active, ended
	MaxParticipants int      `json:"max_participants" db:"max_participants"`
	ScheduledAt   *time.Time `json:"scheduled_at" db:"scheduled_at"`
	StartedAt     *time.Time `json:"started_at" db:"started_at"`
	EndedAt       *time.Time `json:"ended_at" db:"ended_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type MeetingParticipant struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	MeetingID    uuid.UUID  `json:"meeting_id" db:"meeting_id"`
	UserID       uuid.UUID  `json:"user_id" db:"user_id"`
	Role         string     `json:"role" db:"role"` // host, cohost, speaker, listener, moderator, guest
	IsMuted      bool       `json:"is_muted" db:"is_muted"`
	IsVideoOn    bool       `json:"is_video_on" db:"is_video_on"`
	IsScreenShare bool     `json:"is_screen_share" db:"is_screen_share"`
	HandRaised   bool       `json:"hand_raised" db:"hand_raised"`
	JoinedAt     time.Time  `json:"joined_at" db:"joined_at"`
	LeftAt       *time.Time `json:"left_at" db:"left_at"`
}

type BandwidthMode string

const (
	ModeUltraS BandwidthMode = "ultra_saving" // Audio only: 8-15 MB/hr
	ModeEconomy BandwidthMode = "economy"      // 144P: 50-100 MB/hr
	ModeStandard BandwidthMode = "standard"    // 360P: 150-250 MB/hr
	ModeHD      BandwidthMode = "hd"           // 720P: 500-800 MB/hr
)

type VideoQuality string

const (
	Quality144p VideoQuality = "144p"
	Quality360p VideoQuality = "360p"
	Quality720p VideoQuality = "720p"
	Quality1080p VideoQuality = "1080p"
)

type NetworkCondition string

const (
	NetworkVeryPoor NetworkCondition = "very_poor"
	NetworkPoor     NetworkCondition = "poor"
	NetworkAverage  NetworkCondition = "average"
	NetworkGood     NetworkCondition = "good"
	NetworkExcellent NetworkCondition = "excellent"
)
