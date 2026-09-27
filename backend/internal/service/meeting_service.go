package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/livekit/protocol/auth"
	lksdk "github.com/livekit/server-sdk-go"

	"shalom/internal/config"
	"shalom/internal/models"
)

type MeetingService struct {
	db      *pgxpool.Pool
	config  *config.Config
}

func NewMeetingService(db *pgxpool.Pool, cfg *config.Config) *MeetingService {
	return &MeetingService{db: db, config: cfg}
}

type CreateMeetingRequest struct {
	Title           string     `json:"title" binding:"required"`
	Description     string     `json:"description"`
	MaxParticipants int        `json:"max_participants"`
	ScheduledAt     *time.Time `json:"scheduled_at"`
}

type JoinMeetingRequest struct {
	MeetingCode string `json:"meeting_code" binding:"required"`
}

type MeetingResponse struct {
	Meeting      models.Meeting              `json:"meeting"`
	Token        string                      `json:"token,omitempty"`
	RoomName     string                      `json:"room_name,omitempty"`
	Participants []models.MeetingParticipant `json:"participants,omitempty"`
}

func (s *MeetingService) CreateMeeting(ctx context.Context, hostID uuid.UUID, req CreateMeetingRequest) (*models.Meeting, error) {
	code := generateMeetingCode()
	maxParticipants := req.MaxParticipants
	if maxParticipants == 0 {
		maxParticipants = 500
	}

	meeting := models.Meeting{}
	err := s.db.QueryRow(ctx,
		`INSERT INTO meetings (id, title, description, host_id, meeting_code, join_link, status, max_participants, scheduled_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'scheduled', $7, $8, NOW(), NOW())
		 RETURNING id, title, description, host_id, meeting_code, join_link, status, max_participants, scheduled_at, created_at, updated_at`,
		uuid.New(), req.Title, req.Description, hostID, code, "/meet/"+code, maxParticipants, req.ScheduledAt,
	).Scan(&meeting.ID, &meeting.Title, &meeting.Description, &meeting.HostID, &meeting.MeetingCode, &meeting.JoinLink, &meeting.Status, &meeting.MaxParticipants, &meeting.ScheduledAt, &meeting.CreatedAt, &meeting.UpdatedAt)
	if err != nil {
		return nil, err
	}

	s.db.Exec(ctx,
		`INSERT INTO meeting_participants (id, meeting_id, user_id, role, joined_at) VALUES ($1, $2, $3, 'host', NOW())`,
		uuid.New(), meeting.ID, hostID,
	)

	return &meeting, nil
}

func (s *MeetingService) StartMeeting(ctx context.Context, meetingID, userID uuid.UUID) (*MeetingResponse, error) {
	var meeting models.Meeting
	err := s.db.QueryRow(ctx,
		`SELECT id, title, description, host_id, meeting_code, join_link, status, max_participants, scheduled_at, created_at, updated_at
		 FROM meetings WHERE id = $1`, meetingID,
	).Scan(&meeting.ID, &meeting.Title, &meeting.Description, &meeting.HostID, &meeting.MeetingCode, &meeting.JoinLink, &meeting.Status, &meeting.MaxParticipants, &meeting.ScheduledAt, &meeting.CreatedAt, &meeting.UpdatedAt)
	if err != nil {
		return nil, errors.New("meeting not found")
	}

	if meeting.HostID != userID {
		return nil, errors.New("only host can start meeting")
	}

	s.db.Exec(ctx, `UPDATE meetings SET status = 'active', started_at = NOW(), updated_at = NOW() WHERE id = $1`, meetingID)
	meeting.Status = "active"

	roomName := "meeting-" + meeting.MeetingCode
	token, err := s.generateLiveKitToken(userID, roomName, true)
	if err != nil {
		return nil, err
	}

	return &MeetingResponse{Meeting: meeting, Token: token, RoomName: roomName}, nil
}

func (s *MeetingService) JoinMeeting(ctx context.Context, userID uuid.UUID, req JoinMeetingRequest) (*MeetingResponse, error) {
	var meeting models.Meeting
	err := s.db.QueryRow(ctx,
		`SELECT id, title, description, host_id, meeting_code, join_link, status, max_participants, scheduled_at, created_at, updated_at
		 FROM meetings WHERE meeting_code = $1`, req.MeetingCode,
	).Scan(&meeting.ID, &meeting.Title, &meeting.Description, &meeting.HostID, &meeting.MeetingCode, &meeting.JoinLink, &meeting.Status, &meeting.MaxParticipants, &meeting.ScheduledAt, &meeting.CreatedAt, &meeting.UpdatedAt)
	if err != nil {
		return nil, errors.New("meeting not found")
	}

	if meeting.Status == "ended" {
		return nil, errors.New("meeting has ended")
	}

	var count int
	s.db.QueryRow(ctx, `SELECT COUNT(*) FROM meeting_participants WHERE meeting_id = $1 AND left_at IS NULL`, meeting.ID).Scan(&count)
	if count >= meeting.MaxParticipants {
		return nil, errors.New("meeting is full")
	}

	role := "listener"
	if meeting.HostID == userID {
		role = "host"
	}

	s.db.Exec(ctx,
		`INSERT INTO meeting_participants (id, meeting_id, user_id, role, joined_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (meeting_id, user_id) DO UPDATE SET left_at = NULL`,
		uuid.New(), meeting.ID, userID, role,
	)

	roomName := "meeting-" + meeting.MeetingCode
	token, err := s.generateLiveKitToken(userID, roomName, role == "host" || role == "cohost")
	if err != nil {
		return nil, err
	}

	return &MeetingResponse{Meeting: meeting, Token: token, RoomName: roomName}, nil
}

func (s *MeetingService) EndMeeting(ctx context.Context, meetingID, userID uuid.UUID) error {
	var hostID uuid.UUID
	err := s.db.QueryRow(ctx, `SELECT host_id FROM meetings WHERE id = $1`, meetingID).Scan(&hostID)
	if err != nil {
		return errors.New("meeting not found")
	}
	if hostID != userID {
		return errors.New("only host can end meeting")
	}
	s.db.Exec(ctx, `UPDATE meetings SET status = 'ended', ended_at = NOW(), updated_at = NOW() WHERE id = $1`, meetingID)
	s.db.Exec(ctx, `UPDATE meeting_participants SET left_at = NOW() WHERE meeting_id = $1 AND left_at IS NULL`, meetingID)
	return nil
}

func (s *MeetingService) GetMeeting(ctx context.Context, meetingID uuid.UUID) (*models.Meeting, error) {
	var meeting models.Meeting
	err := s.db.QueryRow(ctx,
		`SELECT id, title, description, host_id, meeting_code, join_link, status, max_participants, scheduled_at, started_at, ended_at, created_at, updated_at
		 FROM meetings WHERE id = $1`, meetingID,
	).Scan(&meeting.ID, &meeting.Title, &meeting.Description, &meeting.HostID, &meeting.MeetingCode, &meeting.JoinLink, &meeting.Status, &meeting.MaxParticipants, &meeting.ScheduledAt, &meeting.StartedAt, &meeting.EndedAt, &meeting.CreatedAt, &meeting.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &meeting, nil
}

func (s *MeetingService) GetParticipants(ctx context.Context, meetingID uuid.UUID) ([]models.MeetingParticipant, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, meeting_id, user_id, role, is_muted, is_video_on, is_screen_share, hand_raised, joined_at, left_at
		 FROM meeting_participants WHERE meeting_id = $1 ORDER BY joined_at`, meetingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []models.MeetingParticipant
	for rows.Next() {
		var p models.MeetingParticipant
		rows.Scan(&p.ID, &p.MeetingID, &p.UserID, &p.Role, &p.IsMuted, &p.IsVideoOn, &p.IsScreenShare, &p.HandRaised, &p.JoinedAt, &p.LeftAt)
		participants = append(participants, p)
	}
	return participants, nil
}

func (s *MeetingService) UpdateParticipant(ctx context.Context, meetingID, userID uuid.UUID, isMuted, isVideoOn, isScreenShare, handRaised *bool) error {
	if isMuted != nil {
		s.db.Exec(ctx, `UPDATE meeting_participants SET is_muted = $1 WHERE meeting_id = $2 AND user_id = $3`, *isMuted, meetingID, userID)
	}
	if isVideoOn != nil {
		s.db.Exec(ctx, `UPDATE meeting_participants SET is_video_on = $1 WHERE meeting_id = $2 AND user_id = $3`, *isVideoOn, meetingID, userID)
	}
	if isScreenShare != nil {
		s.db.Exec(ctx, `UPDATE meeting_participants SET is_screen_share = $1 WHERE meeting_id = $2 AND user_id = $3`, *isScreenShare, meetingID, userID)
	}
	if handRaised != nil {
		s.db.Exec(ctx, `UPDATE meeting_participants SET hand_raised = $1 WHERE meeting_id = $2 AND user_id = $3`, *handRaised, meetingID, userID)
	}
	return nil
}

func (s *MeetingService) KickParticipant(ctx context.Context, meetingID, hostID, targetID uuid.UUID) error {
	var hostRole string
	err := s.db.QueryRow(ctx,
		`SELECT role FROM meeting_participants WHERE meeting_id = $1 AND user_id = $2`, meetingID, hostID,
	).Scan(&hostRole)
	if err != nil || (hostRole != "host" && hostRole != "cohost") {
		return errors.New("insufficient permissions")
	}
	s.db.Exec(ctx,
		`UPDATE meeting_participants SET left_at = NOW() WHERE meeting_id = $1 AND user_id = $2`, meetingID, targetID,
	)
	return nil
}

func (s *MeetingService) generateLiveKitToken(userID uuid.UUID, roomName string, canPublish bool) (string, error) {
	grant := &lksdk.VideoGrants{
		RoomJoin:       true,
		Room:           roomName,
		CanPublish:     canPublish,
		CanSubscribe:   true,
		CanPublishData: true,
	}
	token, err := auth.CreateAPIToken(s.config.LiveKit.APIKey, s.config.LiveKit.APISecret, userID.String(), "", grant, nil)
	if err != nil {
		return "", err
	}
	return token, nil
}
