package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"shalom/internal/models"
)

// BandwidthOptimizer handles adaptive bitrate and bandwidth modes
type BandwidthOptimizer struct {
	rdb *redis.Client
}

func NewBandwidthOptimizer(rdb *redis.Client) *BandwidthOptimizer {
	return &BandwidthOptimizer{rdb: rdb}
}

type UserBandwidthProfile struct {
	UserID          uuid.UUID             `json:"user_id"`
	CurrentMode     models.BandwidthMode  `json:"current_mode"`
	NetworkCondition models.NetworkCondition `json:"network_condition"`
	VideoQuality    models.VideoQuality   `json:"video_quality"`
	IsAudioOnly     bool                  `json:"is_audio_only"`
	LastUpdated     time.Time             `json:"last_updated"`
	BandwidthKbps   int                   `json:"bandwidth_kbps"`
}

type QualityPreset struct {
	MaxBitrate  int    `json:"max_bitrate"`
	Resolution  string `json:"resolution"`
	FPS         int    `json:"fps"`
	AudioOnly   bool   `json:"audio_only"`
	DataPerHour string `json:"data_per_hour"`
}

var QualityPresets = map[models.BandwidthMode]QualityPreset{
	models.ModeUltraS: {
		MaxBitrate:  32,
		Resolution:  "audio_only",
		FPS:         0,
		AudioOnly:   true,
		DataPerHour: "8-15 MB",
	},
	models.ModeEconomy: {
		MaxBitrate:  100,
		Resolution:  "144p",
		FPS:         15,
		AudioOnly:   false,
		DataPerHour: "50-100 MB",
	},
	models.ModeStandard: {
		MaxBitrate:  500,
		Resolution:  "360p",
		FPS:         24,
		AudioOnly:   false,
		DataPerHour: "150-250 MB",
	},
	models.ModeHD: {
		MaxBitrate:  2000,
		Resolution:  "720p",
		FPS:         30,
		AudioOnly:   false,
		DataPerHour: "500-800 MB",
	},
}

// DetectNetworkCondition analyzes user's network metrics
func (b *BandwidthOptimizer) DetectNetworkCondition(ctx context.Context, userID uuid.UUID, bandwidthKbps int) models.NetworkCondition {
	var condition models.NetworkCondition

	switch {
	case bandwidthKbps < 50:
		condition = models.NetworkVeryPoor
	case bandwidthKbps < 150:
		condition = models.NetworkPoor
	case bandwidthKbps < 500:
		condition = models.NetworkAverage
	case bandwidthKbps < 2000:
		condition = models.NetworkGood
	default:
		condition = models.NetworkExcellent
	}

	// Auto-switch mode based on condition
	mode := b.mapConditionToMode(condition)
	b.UpdateProfile(ctx, userID, mode, condition, bandwidthKbps)

	return condition
}

func (b *BandwidthOptimizer) mapConditionToMode(condition models.NetworkCondition) models.BandwidthMode {
	switch condition {
	case models.NetworkVeryPoor:
		return models.ModeUltraS
	case models.NetworkPoor:
		return models.ModeEconomy
	case models.NetworkAverage:
		return models.ModeStandard
	case models.NetworkGood, models.NetworkExcellent:
		return models.ModeHD
	default:
		return models.ModeStandard
	}
}

func (b *BandwidthOptimizer) UpdateProfile(ctx context.Context, userID uuid.UUID, mode models.BandwidthMode, condition models.NetworkCondition, bandwidthKbps int) {
	profile := UserBandwidthProfile{
		UserID:           userID,
		CurrentMode:      mode,
		NetworkCondition: condition,
		VideoQuality:     b.modeToQuality(mode),
		IsAudioOnly:      mode == models.ModeUltraS,
		LastUpdated:      time.Now(),
		BandwidthKbps:    bandwidthKbps,
	}

	data, _ := json.Marshal(profile)
	b.rdb.Set(ctx, "bandwidth:"+userID.String(), data, time.Hour)
}

func (b *BandwidthOptimizer) modeToQuality(mode models.BandwidthMode) models.VideoQuality {
	switch mode {
	case models.ModeUltraS:
		return ""
	case models.ModeEconomy:
		return models.Quality144p
	case models.ModeStandard:
		return models.Quality360p
	case models.ModeHD:
		return models.Quality720p
	default:
		return models.Quality360p
	}
}

func (b *BandwidthOptimizer) GetProfile(ctx context.Context, userID uuid.UUID) (*UserBandwidthProfile, error) {
	data, err := b.rdb.Get(ctx, "bandwidth:"+userID.String()).Bytes()
	if err != nil {
		// Default to standard mode
		defaultProfile := &UserBandwidthProfile{
			UserID:       userID,
			CurrentMode:  models.ModeStandard,
			VideoQuality: models.Quality360p,
			LastUpdated:  time.Now(),
		}
		return defaultProfile, nil
	}

	var profile UserBandwidthProfile
	json.Unmarshal(data, &profile)
	return &profile, nil
}

func (b *BandwidthOptimizer) GetPreset(mode models.BandwidthMode) QualityPreset {
	return QualityPresets[mode]
}

// SwitchMode allows manual mode switching
func (b *BandwidthOptimizer) SwitchMode(ctx context.Context, userID uuid.UUID, mode models.BandwidthMode) *UserBandwidthProfile {
	condition := models.NetworkAverage
	bandwidthKbps := 500

	switch mode {
	case models.ModeUltraS:
		condition = models.NetworkVeryPoor
		bandwidthKbps = 32
	case models.ModeEconomy:
		condition = models.NetworkPoor
		bandwidthKbps = 100
	case models.ModeStandard:
		condition = models.NetworkAverage
		bandwidthKbps = 500
	case models.ModeHD:
		condition = models.NetworkGood
		bandwidthKbps = 2000
	}

	profile := UserBandwidthProfile{
		UserID:           userID,
		CurrentMode:      mode,
		NetworkCondition: condition,
		VideoQuality:     b.modeToQuality(mode),
		IsAudioOnly:      mode == models.ModeUltraS,
		LastUpdated:      time.Now(),
		BandwidthKbps:    bandwidthKbps,
	}

	data, _ := json.Marshal(profile)
	b.rdb.Set(ctx, "bandwidth:"+userID.String(), data, time.Hour)

	return &profile
}

// ConnectionRecovery handles automatic reconnection
type ConnectionRecovery struct {
	rdb *redis.Client
}

func NewConnectionRecovery(rdb *redis.Client) *ConnectionRecovery {
	return &ConnectionRecovery{rdb: rdb}
}

type ReconnectionState struct {
	UserID      uuid.UUID `json:"user_id"`
	MeetingID   string    `json:"meeting_id"`
	Attempts    int       `json:"attempts"`
	MaxAttempts int       `json:"max_attempts"`
	LastAttempt time.Time `json:"last_attempt"`
	Offline     bool      `json:"offline"`
}

func (cr *ConnectionRecovery) OnDisconnect(ctx context.Context, userID uuid.UUID, meetingID string) {
	state := ReconnectionState{
		UserID:      userID,
		MeetingID:   meetingID,
		Attempts:    0,
		MaxAttempts: 10,
		LastAttempt: time.Now(),
		Offline:     true,
	}
	data, _ := json.Marshal(state)
	cr.rdb.Set(ctx, fmt.Sprintf("reconnect:%s", userID.String()), data, 10*time.Minute)

	// Store user as away
	cr.rdb.Set(ctx, "presence:"+userID.String(), "away", 10*time.Minute)
}

func (cr *ConnectionRecovery) OnReconnect(ctx context.Context, userID uuid.UUID) {
	cr.rdb.Del(ctx, fmt.Sprintf("reconnect:%s", userID.String()))
	cr.rdb.Set(ctx, "presence:"+userID.String(), "online", 72*time.Hour)
}

func (cr *ConnectionRecovery) GetPendingMessages(ctx context.Context, userID uuid.UUID) ([]interface{}, error) {
	key := fmt.Sprintf("offline_queue:%s", userID.String())
	data, err := cr.rdb.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var messages []interface{}
	for _, d := range data {
		var msg interface{}
		json.Unmarshal([]byte(d), &msg)
		messages = append(messages, msg)
	}
	return messages, nil
}

func (cr *ConnectionRecovery) QueueOfflineMessage(ctx context.Context, userID uuid.UUID, message interface{}) {
	data, _ := json.Marshal(message)
	cr.rdb.RPush(ctx, fmt.Sprintf("offline_queue:%s", userID.String()), data)
	cr.rdb.Expire(ctx, fmt.Sprintf("offline_queue:%s", userID.String()), 24*time.Hour)
}

// PresenceService manages user online/offline status
type PresenceService struct {
	rdb *redis.Client
}

func NewPresenceService(rdb *redis.Client) *PresenceService {
	return &PresenceService{rdb: rdb}
}

func (ps *PresenceService) SetOnline(ctx context.Context, userID uuid.UUID) {
	ps.rdb.Set(ctx, "presence:"+userID.String(), "online", 72*time.Hour)
}

func (ps *PresenceService) SetOffline(ctx context.Context, userID uuid.UUID) {
	ps.rdb.Set(ctx, "presence:"+userID.String(), "offline", 24*time.Hour)
}

func (ps *PresenceService) SetAway(ctx context.Context, userID uuid.UUID) {
	ps.rdb.Set(ctx, "presence:"+userID.String(), "away", 10*time.Minute)
}

func (ps *PresenceService) GetStatus(ctx context.Context, userID uuid.UUID) string {
	status, err := ps.rdb.Get(ctx, "presence:"+userID.String()).Result()
	if err != nil {
		return "offline"
	}
	return status
}
