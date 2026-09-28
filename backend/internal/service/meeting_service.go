package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"

	"shalom/internal/config"
	"shalom/internal/models"
	"shalom/internal/repository"
)

type MeetingService struct {
	config      *config.Config
	meetingRepo *repository.MeetingRepository
}

func NewMeetingService(cfg *config.Config, meetingRepo *repository.MeetingRepository) *MeetingService {
	return &MeetingService{config: cfg, meetingRepo: meetingRepo}
}

type CreateMeetingRequest struct {
	Title           string     `json:"title" binding:"required"`
	Description     string     `json:"description"`
	MaxParticipants int        `json:"max_participants"`
	IsWaitingRoom   bool       `json:"is_waiting_room"`
	Password        string     `json:"password"`
	ScheduledAt     *time.Time `json:"scheduled_at"`
}

type JoinMeetingRequest struct {
	MeetingCode string `json:"meeting_code" binding:"required"`
	Password    string `json:"password"`
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

	meeting := &models.Meeting{
		ID:              uuid.New(),
		Title:           req.Title,
		Description:     req.Description,
		HostID:          hostID,
		MeetingCode:     code,
		JoinLink:        "/meet/" + code,
		Status:          models.MeetingScheduled,
		MaxParticipants: maxParticipants,
		IsWaitingRoom:   req.IsWaitingRoom,
		Password:        req.Password,
		ScheduledAt:     req.ScheduledAt,
	}

	if err := s.meetingRepo.Create(ctx, meeting); err != nil {
		return nil, err
	}

	// Add host as participant
	s.meetingRepo.AddParticipant(ctx, &models.MeetingParticipant{
		ID:        uuid.New(),
		MeetingID: meeting.ID,
		UserID:    hostID,
		Role:      models.RoleHost,
	})

	return meeting, nil
}

func (s *MeetingService) StartMeeting(ctx context.Context, meetingID, userID uuid.UUID) (*MeetingResponse, error) {
	meeting, err := s.meetingRepo.GetByID(ctx, meetingID)
	if err != nil {
		return nil, errors.New("meeting not found")
	}

	if meeting.HostID != userID {
		return nil, errors.New("only host can start meeting")
	}

	s.meetingRepo.UpdateStatus(ctx, meetingID, models.MeetingActive)
	meeting.Status = models.MeetingActive

	roomName := "meeting-" + meeting.MeetingCode
	token, err := s.generateLiveKitToken(userID, roomName, true)
	if err != nil {
		return nil, err
	}

	return &MeetingResponse{Meeting: *meeting, Token: token, RoomName: roomName}, nil
}

func (s *MeetingService) JoinMeeting(ctx context.Context, userID uuid.UUID, req JoinMeetingRequest) (*MeetingResponse, error) {
	meeting, err := s.meetingRepo.GetByCode(ctx, req.MeetingCode)
	if err != nil {
		return nil, errors.New("meeting not found")
	}

	if meeting.Status == models.MeetingEnded {
		return nil, errors.New("meeting has ended")
	}

	// Check password
	if meeting.Password != "" && meeting.Password != req.Password {
		return nil, errors.New("invalid meeting password")
	}

	// Check capacity
	count, err := s.meetingRepo.GetActiveCount(ctx, meeting.ID)
	if err != nil {
		return nil, err
	}
	if count >= meeting.MaxParticipants {
		return nil, errors.New("meeting is full")
	}

	// Determine role
	role := models.RoleListener
	if meeting.HostID == userID {
		role = models.RoleHost
	}

	// Handle waiting room
	if meeting.IsWaitingRoom && role != models.RoleHost {
		s.meetingRepo.AddToWaitingRoom(ctx, meeting.ID, userID)
		return &MeetingResponse{Meeting: *meeting, Token: "", RoomName: ""}, nil
	}

	s.meetingRepo.AddParticipant(ctx, &models.MeetingParticipant{
		ID:        uuid.New(),
		MeetingID: meeting.ID,
		UserID:    userID,
		Role:      role,
	})

	// Log attendance
	s.meetingRepo.LogAttendance(ctx, meeting.ID, userID)

	roomName := "meeting-" + meeting.MeetingCode
	token, err := s.generateLiveKitToken(userID, roomName, role == models.RoleHost || role == models.RoleCohost)
	if err != nil {
		return nil, err
	}

	return &MeetingResponse{Meeting: *meeting, Token: token, RoomName: roomName}, nil
}

func (s *MeetingService) EndMeeting(ctx context.Context, meetingID, userID uuid.UUID) error {
	meeting, err := s.meetingRepo.GetByID(ctx, meetingID)
	if err != nil {
		return errors.New("meeting not found")
	}

	if meeting.HostID != userID {
		return errors.New("only host can end meeting")
	}

	return s.meetingRepo.UpdateStatus(ctx, meetingID, models.MeetingEnded)
}

func (s *MeetingService) GetMeeting(ctx context.Context, meetingID uuid.UUID) (*models.Meeting, error) {
	return s.meetingRepo.GetByID(ctx, meetingID)
}

func (s *MeetingService) GetParticipants(ctx context.Context, meetingID uuid.UUID) ([]models.MeetingParticipant, error) {
	return s.meetingRepo.GetParticipants(ctx, meetingID)
}

func (s *MeetingService) UpdateParticipant(ctx context.Context, meetingID, userID uuid.UUID, isMuted, isVideoOn, isScreenShare, handRaised *bool) error {
	fields := make(map[string]interface{})
	if isMuted != nil {
		fields["is_muted"] = *isMuted
	}
	if isVideoOn != nil {
		fields["is_video_on"] = *isVideoOn
	}
	if isScreenShare != nil {
		fields["is_screen_share"] = *isScreenShare
	}
	if handRaised != nil {
		fields["hand_raised"] = *handRaised
	}
	if len(fields) == 0 {
		return nil
	}
	return s.meetingRepo.UpdateParticipant(ctx, meetingID, userID, fields)
}

func (s *MeetingService) KickParticipant(ctx context.Context, meetingID, hostID, targetID uuid.UUID) error {
	role, err := s.meetingRepo.GetRole(ctx, meetingID, hostID)
	if err != nil || (role != string(models.RoleHost) && role != string(models.RoleCohost)) {
		return errors.New("insufficient permissions")
	}
	return s.meetingRepo.BanParticipant(ctx, meetingID, targetID)
}

func (s *MeetingService) AdmitFromWaitingRoom(ctx context.Context, meetingID, hostID, targetID uuid.UUID) error {
	role, err := s.meetingRepo.GetRole(ctx, meetingID, hostID)
	if err != nil || (role != string(models.RoleHost) && role != string(models.RoleCohost)) {
		return errors.New("insufficient permissions")
	}

	if err := s.meetingRepo.AdmitFromWaitingRoom(ctx, meetingID, targetID, hostID); err != nil {
		return err
	}

	return s.meetingRepo.AddParticipant(ctx, &models.MeetingParticipant{
		ID:        uuid.New(),
		MeetingID: meetingID,
		UserID:    targetID,
		Role:      models.RoleListener,
	})
}

func (s *MeetingService) GetWaitingRoom(ctx context.Context, meetingID uuid.UUID) ([]models.WaitingRoomEntry, error) {
	return s.meetingRepo.GetWaitingRoom(ctx, meetingID)
}

func (s *MeetingService) GetRecordings(ctx context.Context, meetingID uuid.UUID) ([]models.Recording, error) {
	return s.meetingRepo.GetRecordings(ctx, meetingID)
}

func (s *MeetingService) GetAttendance(ctx context.Context, meetingID uuid.UUID) ([]models.Attendance, error) {
	return s.meetingRepo.GetAttendance(ctx, meetingID)
}

func generateMeetingCode() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)[:8]
}

func (s *MeetingService) generateLiveKitToken(userID uuid.UUID, roomName string, canPublish bool) (string, error) {
	grant := &auth.VideoGrants{
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
