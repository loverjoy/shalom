package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"shalom/internal/models"
)

type MeetingRepository struct {
	db *pgxpool.Pool
}

func NewMeetingRepository(db *pgxpool.Pool) *MeetingRepository {
	return &MeetingRepository{db: db}
}

func (r *MeetingRepository) Create(ctx context.Context, m *models.Meeting) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO meetings (id, title, description, host_id, meeting_code, join_link, status, max_participants,
		        is_waiting_room, password, scheduled_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		 RETURNING id, title, description, host_id, meeting_code, join_link, status, max_participants,
		        is_waiting_room, scheduled_at, created_at, updated_at`,
		m.ID, m.Title, m.Description, m.HostID, m.MeetingCode, m.JoinLink, m.Status, m.MaxParticipants,
		m.IsWaitingRoom, m.Password, m.ScheduledAt,
	).Scan(&m.ID, &m.Title, &m.Description, &m.HostID, &m.MeetingCode, &m.JoinLink, &m.Status, &m.MaxParticipants,
		&m.IsWaitingRoom, &m.ScheduledAt, &m.CreatedAt, &m.UpdatedAt)
}

func (r *MeetingRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Meeting, error) {
	m := &models.Meeting{}
	err := r.db.QueryRow(ctx,
		`SELECT id, title, description, host_id, meeting_code, join_link, status, max_participants,
		        is_recording, is_waiting_room, is_breakout, parent_meeting_id, scheduled_at, started_at, ended_at,
		        duration_seconds, created_at, updated_at
		 FROM meetings WHERE id = $1`, id,
	).Scan(&m.ID, &m.Title, &m.Description, &m.HostID, &m.MeetingCode, &m.JoinLink, &m.Status, &m.MaxParticipants,
		&m.IsRecording, &m.IsWaitingRoom, &m.IsBreakout, &m.ParentMeetingID, &m.ScheduledAt, &m.StartedAt, &m.EndedAt,
		&m.DurationSeconds, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MeetingRepository) GetByCode(ctx context.Context, code string) (*models.Meeting, error) {
	m := &models.Meeting{}
	err := r.db.QueryRow(ctx,
		`SELECT id, title, description, host_id, meeting_code, join_link, status, max_participants,
		        is_recording, is_waiting_room, is_breakout, parent_meeting_id, scheduled_at, started_at, ended_at,
		        duration_seconds, created_at, updated_at
		 FROM meetings WHERE meeting_code = $1`, code,
	).Scan(&m.ID, &m.Title, &m.Description, &m.HostID, &m.MeetingCode, &m.JoinLink, &m.Status, &m.MaxParticipants,
		&m.IsRecording, &m.IsWaitingRoom, &m.IsBreakout, &m.ParentMeetingID, &m.ScheduledAt, &m.StartedAt, &m.EndedAt,
		&m.DurationSeconds, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MeetingRepository) GetByHost(ctx context.Context, hostID uuid.UUID, limit int) ([]models.Meeting, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, description, host_id, meeting_code, join_link, status, max_participants,
		        is_recording, is_waiting_room, is_breakout, parent_meeting_id, scheduled_at, started_at, ended_at,
		        duration_seconds, created_at, updated_at
		 FROM meetings WHERE host_id = $1 ORDER BY created_at DESC LIMIT $2`, hostID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var meetings []models.Meeting
	for rows.Next() {
		var m models.Meeting
		rows.Scan(&m.ID, &m.Title, &m.Description, &m.HostID, &m.MeetingCode, &m.JoinLink, &m.Status, &m.MaxParticipants,
			&m.IsRecording, &m.IsWaitingRoom, &m.IsBreakout, &m.ParentMeetingID, &m.ScheduledAt, &m.StartedAt, &m.EndedAt,
			&m.DurationSeconds, &m.CreatedAt, &m.UpdatedAt)
		meetings = append(meetings, m)
	}
	return meetings, nil
}

func (r *MeetingRepository) GetUpcoming(ctx context.Context, userID uuid.UUID, limit int) ([]models.Meeting, error) {
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT m.id, m.title, m.description, m.host_id, m.meeting_code, m.join_link, m.status, m.max_participants,
		        m.is_recording, m.is_waiting_room, m.is_breakout, m.parent_meeting_id, m.scheduled_at, m.started_at, m.ended_at,
		        m.duration_seconds, m.created_at, m.updated_at
		 FROM meetings m
		 LEFT JOIN meeting_participants mp ON m.id = mp.meeting_id
		 WHERE (m.host_id = $1 OR mp.user_id = $1) AND m.status IN ('scheduled', 'active')
		 ORDER BY COALESCE(m.scheduled_at, m.created_at) ASC LIMIT $2`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var meetings []models.Meeting
	for rows.Next() {
		var m models.Meeting
		rows.Scan(&m.ID, &m.Title, &m.Description, &m.HostID, &m.MeetingCode, &m.JoinLink, &m.Status, &m.MaxParticipants,
			&m.IsRecording, &m.IsWaitingRoom, &m.IsBreakout, &m.ParentMeetingID, &m.ScheduledAt, &m.StartedAt, &m.EndedAt,
			&m.DurationSeconds, &m.CreatedAt, &m.UpdatedAt)
		meetings = append(meetings, m)
	}
	return meetings, nil
}

func (r *MeetingRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.MeetingStatus) error {
	var extra string
	switch status {
	case models.MeetingActive:
		extra = ", started_at = NOW()"
	case models.MeetingEnded:
		extra = ", ended_at = NOW()"
	}
	query := `UPDATE meetings SET status = $1, updated_at = NOW()` + extra + ` WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *MeetingRepository) SetRecording(ctx context.Context, id uuid.UUID, isRecording bool) error {
	_, err := r.db.Exec(ctx, `UPDATE meetings SET is_recording = $1, updated_at = NOW() WHERE id = $2`, isRecording, id)
	return err
}

func (r *MeetingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM meetings WHERE id = $1`, id)
	return err
}

func (r *MeetingRepository) GetActiveCount(ctx context.Context, id uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM meeting_participants WHERE meeting_id = $1 AND left_at IS NULL`, id,
	).Scan(&count)
	return count, err
}

// Participants

func (r *MeetingRepository) AddParticipant(ctx context.Context, p *models.MeetingParticipant) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO meeting_participants (id, meeting_id, user_id, role, is_muted, bandwidth_mode, joined_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())
		 ON CONFLICT (meeting_id, user_id) DO UPDATE SET left_at = NULL, role = $4
		 RETURNING id, joined_at`,
		p.ID, p.MeetingID, p.UserID, p.Role, p.IsMuted, p.BandwidthMode,
	).Scan(&p.ID, &p.JoinedAt)
}

func (r *MeetingRepository) RemoveParticipant(ctx context.Context, meetingID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE meeting_participants SET left_at = NOW(), duration_seconds = EXTRACT(EPOCH FROM (NOW() - joined_at))::INT
		 WHERE meeting_id = $1 AND user_id = $2 AND left_at IS NULL`,
		meetingID, userID,
	)
	return err
}

func (r *MeetingRepository) GetParticipants(ctx context.Context, meetingID uuid.UUID) ([]models.MeetingParticipant, error) {
	rows, err := r.db.Query(ctx,
		`SELECT mp.id, mp.meeting_id, mp.user_id, mp.role, mp.is_muted, mp.is_video_on, mp.is_screen_share,
		        mp.hand_raised, mp.is_banned, mp.bandwidth_mode, mp.joined_at, mp.left_at, mp.duration_seconds,
		        u.username, u.display_name, u.avatar_url
		 FROM meeting_participants mp JOIN users u ON mp.user_id = u.id
		 WHERE mp.meeting_id = $1 ORDER BY mp.joined_at`, meetingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []models.MeetingParticipant
	for rows.Next() {
		var p models.MeetingParticipant
		rows.Scan(&p.ID, &p.MeetingID, &p.UserID, &p.Role, &p.IsMuted, &p.IsVideoOn, &p.IsScreenShare,
			&p.HandRaised, &p.IsBanned, &p.BandwidthMode, &p.JoinedAt, &p.LeftAt, &p.DurationSeconds,
			&p.Username, &p.DisplayName, &p.AvatarURL)
		participants = append(participants, p)
	}
	return participants, nil
}

func (r *MeetingRepository) GetParticipant(ctx context.Context, meetingID, userID uuid.UUID) (*models.MeetingParticipant, error) {
	p := &models.MeetingParticipant{}
	err := r.db.QueryRow(ctx,
		`SELECT id, meeting_id, user_id, role, is_muted, is_video_on, is_screen_share, hand_raised,
		        is_banned, bandwidth_mode, joined_at, left_at, duration_seconds
		 FROM meeting_participants WHERE meeting_id = $1 AND user_id = $2`, meetingID, userID,
	).Scan(&p.ID, &p.MeetingID, &p.UserID, &p.Role, &p.IsMuted, &p.IsVideoOn, &p.IsScreenShare,
		&p.HandRaised, &p.IsBanned, &p.BandwidthMode, &p.JoinedAt, &p.LeftAt, &p.DurationSeconds)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *MeetingRepository) UpdateParticipant(ctx context.Context, meetingID, userID uuid.UUID, fields map[string]interface{}) error {
	for k, v := range fields {
		_, err := r.db.Exec(ctx,
			`UPDATE meeting_participants SET `+k+` = $1 WHERE meeting_id = $2 AND user_id = $3`,
			v, meetingID, userID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MeetingRepository) BanParticipant(ctx context.Context, meetingID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE meeting_participants SET is_banned = TRUE, left_at = NOW() WHERE meeting_id = $1 AND user_id = $2`,
		meetingID, userID,
	)
	return err
}

func (r *MeetingRepository) GetRole(ctx context.Context, meetingID, userID uuid.UUID) (string, error) {
	var role string
	err := r.db.QueryRow(ctx,
		`SELECT role FROM meeting_participants WHERE meeting_id = $1 AND user_id = $2`, meetingID, userID,
	).Scan(&role)
	return role, err
}

// Waiting Room

func (r *MeetingRepository) AddToWaitingRoom(ctx context.Context, meetingID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO waiting_room (id, meeting_id, user_id, created_at) VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (meeting_id, user_id) DO NOTHING`,
		uuid.New(), meetingID, userID,
	)
	return err
}

func (r *MeetingRepository) AdmitFromWaitingRoom(ctx context.Context, meetingID, userID, admittedBy uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE waiting_room SET admitted = TRUE, admitted_by = $1, admitted_at = NOW()
		 WHERE meeting_id = $2 AND user_id = $3`,
		admittedBy, meetingID, userID,
	)
	return err
}

func (r *MeetingRepository) GetWaitingRoom(ctx context.Context, meetingID uuid.UUID) ([]models.WaitingRoomEntry, error) {
	rows, err := r.db.Query(ctx,
		`SELECT wr.id, wr.meeting_id, wr.user_id, wr.admitted, wr.admitted_by, wr.created_at, wr.admitted_at,
		        u.username, u.display_name, u.avatar_url
		 FROM waiting_room wr JOIN users u ON wr.user_id = u.id
		 WHERE wr.meeting_id = $1 AND wr.admitted = FALSE ORDER BY wr.created_at`, meetingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.WaitingRoomEntry
	for rows.Next() {
		var e models.WaitingRoomEntry
		rows.Scan(&e.ID, &e.MeetingID, &e.UserID, &e.Admitted, &e.AdmittedBy, &e.CreatedAt, &e.AdmittedAt,
			&e.Username, &e.DisplayName, &e.AvatarURL)
		entries = append(entries, e)
	}
	return entries, nil
}

// Recordings

func (r *MeetingRepository) CreateRecording(ctx context.Context, rec *models.Recording) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO recordings (id, meeting_id, started_by, status, started_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, NOW(), NOW(), NOW())
		 RETURNING id, started_at, created_at`,
		rec.ID, rec.MeetingID, rec.StartedBy, rec.Status,
	).Scan(&rec.ID, &rec.StartedAt, &rec.CreatedAt)
}

func (r *MeetingRepository) UpdateRecording(ctx context.Context, rec *models.Recording) error {
	_, err := r.db.Exec(ctx,
		`UPDATE recordings SET status = $1, file_url = $2, file_size = $3, duration_seconds = $4,
		        ended_at = $5, processed_at = $6, updated_at = NOW() WHERE id = $7`,
		rec.Status, rec.FileURL, rec.FileSize, rec.DurationSeconds, rec.EndedAt, rec.ProcessedAt, rec.ID,
	)
	return err
}

func (r *MeetingRepository) GetRecordings(ctx context.Context, meetingID uuid.UUID) ([]models.Recording, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, meeting_id, started_by, status, file_url, file_size, duration_seconds, started_at, ended_at, processed_at, created_at
		 FROM recordings WHERE meeting_id = $1 ORDER BY created_at DESC`, meetingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recordings []models.Recording
	for rows.Next() {
		var rec models.Recording
		rows.Scan(&rec.ID, &rec.MeetingID, &rec.StartedBy, &rec.Status, &rec.FileURL, &rec.FileSize,
			&rec.DurationSeconds, &rec.StartedAt, &rec.EndedAt, &rec.ProcessedAt, &rec.CreatedAt)
		recordings = append(recordings, rec)
	}
	return recordings, nil
}

// Attendance

func (r *MeetingRepository) LogAttendance(ctx context.Context, meetingID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO attendance (id, meeting_id, user_id, joined_at) VALUES ($1, $2, $3, NOW())`,
		uuid.New(), meetingID, userID,
	)
	return err
}

func (r *MeetingRepository) GetAttendance(ctx context.Context, meetingID uuid.UUID) ([]models.Attendance, error) {
	rows, err := r.db.Query(ctx,
		`SELECT a.id, a.meeting_id, a.user_id, a.joined_at, a.left_at, a.duration,
		        u.username, u.display_name, u.avatar_url
		 FROM attendance a JOIN users u ON a.user_id = u.id
		 WHERE a.meeting_id = $1 ORDER BY a.joined_at`, meetingID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.Attendance
	for rows.Next() {
		var a models.Attendance
		rows.Scan(&a.ID, &a.MeetingID, &a.UserID, &a.JoinedAt, &a.LeftAt, &a.Duration,
			&a.Username, &a.DisplayName, &a.AvatarURL)
		records = append(records, a)
	}
	return records, nil
}
