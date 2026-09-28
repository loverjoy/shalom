package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ============================================================
// ENUM TYPES
// ============================================================

type UserStatus string

const (
	UserOnline  UserStatus = "online"
	UserOffline UserStatus = "offline"
	UserAway    UserStatus = "away"
	UserDND     UserStatus = "dnd"
)

type MeetingStatus string

const (
	MeetingScheduled MeetingStatus = "scheduled"
	MeetingActive    MeetingStatus = "active"
	MeetingEnded     MeetingStatus = "ended"
	MeetingCancelled MeetingStatus = "cancelled"
)

type ParticipantRole string

const (
	RoleHost     ParticipantRole = "host"
	RoleCohost   ParticipantRole = "cohost"
	RoleSpeaker  ParticipantRole = "speaker"
	RoleListener ParticipantRole = "listener"
	RoleModerator ParticipantRole = "moderator"
	RoleGuest    ParticipantRole = "guest"
)

type RoomType string

const (
	RoomPublic   RoomType = "public"
	RoomPrivate  RoomType = "private"
	RoomMeeting  RoomType = "meeting"
	RoomGroup    RoomType = "group"
	RoomChannel  RoomType = "channel"
)

type NotificationType string

const (
	NotifMeetingInvite  NotificationType = "meeting_invite"
	NotifMeetingStart   NotificationType = "meeting_start"
	NotifMessage        NotificationType = "message"
	NotifMention        NotificationType = "mention"
	NotifReaction       NotificationType = "reaction"
	NotifSystem         NotificationType = "system"
	NotifRecordingReady NotificationType = "recording_ready"
)

type RecordingStatus string

const (
	RecordingRecording  RecordingStatus = "recording"
	RecordingProcessing RecordingStatus = "processing"
	RecordingReady      RecordingStatus = "ready"
	RecordingFailed     RecordingStatus = "failed"
	RecordingDeleted    RecordingStatus = "deleted"
)

type BandwidthMode string

const (
	ModeUltraS  BandwidthMode = "ultra_saving"
	ModeEconomy BandwidthMode = "economy"
	ModeStandard BandwidthMode = "standard"
	ModeHD      BandwidthMode = "hd"
)

type VideoQuality string

const (
	Quality144p  VideoQuality = "144p"
	Quality360p  VideoQuality = "360p"
	Quality720p  VideoQuality = "720p"
	Quality1080p VideoQuality = "1080p"
)

type NetworkCondition string

const (
	NetworkVeryPoor  NetworkCondition = "very_poor"
	NetworkPoor      NetworkCondition = "poor"
	NetworkAverage   NetworkCondition = "average"
	NetworkGood      NetworkCondition = "good"
	NetworkExcellent NetworkCondition = "excellent"
)

type FileType string

const (
	FileTypeImage     FileType = "image"
	FileTypeVideo     FileType = "video"
	FileTypeDocument  FileType = "document"
	FileTypeAudio     FileType = "audio"
	FileTypeOther     FileType = "other"
)

// ============================================================
// USER
// ============================================================

type User struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Email         string     `json:"email" db:"email"`
	Username      string     `json:"username" db:"username"`
	DisplayName   string     `json:"display_name" db:"display_name"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	AvatarURL     string     `json:"avatar_url" db:"avatar_url"`
	Bio           string     `json:"bio" db:"bio"`
	Phone         string     `json:"phone" db:"phone"`
	Status        UserStatus `json:"status" db:"status"`
	LastSeen      time.Time  `json:"last_seen" db:"last_seen"`
	EmailVerified bool       `json:"email_verified" db:"email_verified"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type UserSettings struct {
	UserID              uuid.UUID    `json:"user_id" db:"user_id"`
	NotificationsEnabled bool       `json:"notifications_enabled" db:"notifications_enabled"`
	SoundEnabled         bool       `json:"sound_enabled" db:"sound_enabled"`
	DefaultBandwidth     BandwidthMode `json:"default_bandwidth" db:"default_bandwidth"`
	Language             string     `json:"language" db:"language"`
	Theme                string     `json:"theme" db:"theme"`
	AutoReconnect        bool       `json:"auto_reconnect" db:"auto_reconnect"`
	ShowOnlineStatus     bool       `json:"show_online_status" db:"show_online_status"`
	AllowDirectMessages  bool       `json:"allow_direct_messages" db:"allow_direct_messages"`
	CameraDefaultOn      bool       `json:"camera_default_on" db:"camera_default_on"`
	MicDefaultMuted      bool       `json:"mic_default_muted" db:"mic_default_muted"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

type Contact struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	UserID            uuid.UUID  `json:"user_id" db:"user_id"`
	ContactID         uuid.UUID  `json:"contact_id" db:"contact_id"`
	Nickname          string     `json:"nickname" db:"nickname"`
	IsBlocked         bool       `json:"is_blocked" db:"is_blocked"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	ContactUsername   string     `json:"contact_username" db:"contact_username"`
	ContactDisplayName string    `json:"contact_display_name" db:"contact_display_name"`
	ContactAvatar     string    `json:"contact_avatar" db:"contact_avatar"`
	ContactStatus     string    `json:"contact_status" db:"contact_status"`
}

// ============================================================
// MEETING
// ============================================================

type Meeting struct {
	ID               uuid.UUID     `json:"id" db:"id"`
	Title            string        `json:"title" db:"title"`
	Description      string        `json:"description" db:"description"`
	HostID           uuid.UUID     `json:"host_id" db:"host_id"`
	MeetingCode      string        `json:"meeting_code" db:"meeting_code"`
	JoinLink         string        `json:"join_link" db:"join_link"`
	Status           MeetingStatus `json:"status" db:"status"`
	MaxParticipants  int           `json:"max_participants" db:"max_participants"`
	IsRecording      bool          `json:"is_recording" db:"is_recording"`
	IsWaitingRoom    bool          `json:"is_waiting_room" db:"is_waiting_room"`
	IsBreakout       bool          `json:"is_breakout" db:"is_breakout"`
	ParentMeetingID  *uuid.UUID    `json:"parent_meeting_id,omitempty" db:"parent_meeting_id"`
	Password         string        `json:"-" db:"password"`
	ScheduledAt      *time.Time    `json:"scheduled_at" db:"scheduled_at"`
	StartedAt        *time.Time    `json:"started_at" db:"started_at"`
	EndedAt          *time.Time    `json:"ended_at" db:"ended_at"`
	DurationSeconds  int           `json:"duration_seconds" db:"duration_seconds"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" db:"updated_at"`
}

type MeetingParticipant struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	MeetingID       uuid.UUID     `json:"meeting_id" db:"meeting_id"`
	UserID          uuid.UUID     `json:"user_id" db:"user_id"`
	Role            ParticipantRole `json:"role" db:"role"`
	IsMuted         bool          `json:"is_muted" db:"is_muted"`
	IsVideoOn       bool          `json:"is_video_on" db:"is_video_on"`
	IsScreenShare   bool          `json:"is_screen_share" db:"is_screen_share"`
	HandRaised      bool          `json:"hand_raised" db:"hand_raised"`
	IsBanned        bool          `json:"is_banned" db:"is_banned"`
	BandwidthMode   BandwidthMode `json:"bandwidth_mode" db:"bandwidth_mode"`
	JoinedAt        time.Time     `json:"joined_at" db:"joined_at"`
	LeftAt          *time.Time    `json:"left_at" db:"left_at"`
	DurationSeconds int           `json:"duration_seconds" db:"duration_seconds"`
	// Joined from users table
	Username    string `json:"username" db:"username"`
	DisplayName string `json:"display_name" db:"display_name"`
	AvatarURL   string `json:"avatar_url" db:"avatar_url"`
}

type WaitingRoomEntry struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	MeetingID   uuid.UUID  `json:"meeting_id" db:"meeting_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	Admitted    bool       `json:"admitted" db:"admitted"`
	AdmittedBy  *uuid.UUID `json:"admitted_by" db:"admitted_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	AdmittedAt  *time.Time `json:"admitted_at" db:"admitted_at"`
	Username    string     `json:"username" db:"username"`
	DisplayName string     `json:"display_name" db:"display_name"`
	AvatarURL   string     `json:"avatar_url" db:"avatar_url"`
}

type Recording struct {
	ID              uuid.UUID        `json:"id" db:"id"`
	MeetingID       uuid.UUID        `json:"meeting_id" db:"meeting_id"`
	StartedBy       uuid.UUID        `json:"started_by" db:"started_by"`
	Status          RecordingStatus  `json:"status" db:"status"`
	FileURL         string           `json:"file_url" db:"file_url"`
	FileSize        int64            `json:"file_size" db:"file_size"`
	DurationSeconds int              `json:"duration_seconds" db:"duration_seconds"`
	StartedAt       time.Time        `json:"started_at" db:"started_at"`
	EndedAt         *time.Time       `json:"ended_at" db:"ended_at"`
	ProcessedAt     *time.Time       `json:"processed_at" db:"processed_at"`
	CreatedAt       time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at" db:"updated_at"`
}

type Attendance struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	MeetingID   uuid.UUID  `json:"meeting_id" db:"meeting_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	JoinedAt    time.Time  `json:"joined_at" db:"joined_at"`
	LeftAt      *time.Time `json:"left_at" db:"left_at"`
	Duration    int        `json:"duration" db:"duration"`
	Username    string     `json:"username" db:"username"`
	DisplayName string     `json:"display_name" db:"display_name"`
	AvatarURL   string     `json:"avatar_url" db:"avatar_url"`
}

type Notification struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	UserID    uuid.UUID        `json:"user_id" db:"user_id"`
	Type      NotificationType `json:"type" db:"type"`
	Title     string           `json:"title" db:"title"`
	Body      string           `json:"body" db:"body"`
	Data      json.RawMessage  `json:"data" db:"data"`
	IsRead    bool             `json:"is_read" db:"is_read"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}

type FileUpload struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UploaderID uuid.UUID `json:"uploader_id" db:"uploader_id"`
	FileName   string    `json:"file_name" db:"file_name"`
	FileType   FileType  `json:"file_type" db:"file_type"`
	FileSize   int64     `json:"file_size" db:"file_size"`
	MimeType   string    `json:"mime_type" db:"mime_type"`
	StorageKey string    `json:"storage_key" db:"storage_key"`
	URL        string    `json:"url" db:"url"`
	IsPublic   bool      `json:"is_public" db:"is_public"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type Channel struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	MemberCount int       `json:"member_count" db:"member_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ChannelMember struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ChannelID   uuid.UUID `json:"channel_id" db:"channel_id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Role        string    `json:"role" db:"role"`
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
	Username    string    `json:"username" db:"username"`
	DisplayName string    `json:"display_name" db:"display_name"`
	AvatarURL   string    `json:"avatar_url" db:"avatar_url"`
}
