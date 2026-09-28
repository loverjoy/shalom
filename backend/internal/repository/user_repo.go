package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"shalom/internal/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO users (id, email, username, display_name, password_hash, avatar_url, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		 RETURNING id, email, username, display_name, avatar_url, status, last_seen, created_at, updated_at`,
		user.ID, user.Email, user.Username, user.DisplayName, user.PasswordHash, user.AvatarURL, user.Status,
	).Scan(&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.AvatarURL, &user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, username, display_name, avatar_url, status, last_seen, created_at, updated_at
		 FROM users WHERE id = $1 AND is_active = TRUE`, id,
	).Scan(&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.AvatarURL, &user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, username, display_name, password_hash, avatar_url, status, last_seen, created_at, updated_at
		 FROM users WHERE email = $1 AND is_active = TRUE`, email,
	).Scan(&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.PasswordHash, &user.AvatarURL, &user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, username, display_name, password_hash, avatar_url, status, last_seen, created_at, updated_at
		 FROM users WHERE username = $1 AND is_active = TRUE`, username,
	).Scan(&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.PasswordHash, &user.AvatarURL, &user.Status, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET display_name = $1, avatar_url = $2, bio = $3, phone = $4, updated_at = NOW()
		 WHERE id = $5`,
		user.DisplayName, user.AvatarURL, user.Bio, user.Phone, user.ID,
	)
	return err
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.UserStatus) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET status = $1, last_seen = NOW(), updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	return err
}

func (r *UserRepository) SetLastSeen(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET last_seen = NOW() WHERE id = $1`, id)
	return err
}

func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *UserRepository) Search(ctx context.Context, query string, limit int) ([]models.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, email, username, display_name, avatar_url, status, last_seen, created_at, updated_at
		 FROM users WHERE is_active = TRUE
		 AND (username ILIKE $1 OR display_name ILIKE $1 OR email ILIKE $1)
		 ORDER BY last_seen DESC LIMIT $2`,
		"%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Email, &u.Username, &u.DisplayName, &u.AvatarURL, &u.Status, &u.LastSeen, &u.CreatedAt, &u.UpdatedAt)
		users = append(users, u)
	}
	return users, nil
}

// Settings

func (r *UserRepository) GetSettings(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	s := &models.UserSettings{}
	err := r.db.QueryRow(ctx,
		`SELECT user_id, notifications_enabled, sound_enabled, default_bandwidth, language, theme,
		        auto_reconnect, show_online_status, allow_direct_messages, camera_default_on, mic_default_muted,
		        created_at, updated_at
		 FROM user_settings WHERE user_id = $1`, userID,
	).Scan(&s.UserID, &s.NotificationsEnabled, &s.SoundEnabled, &s.DefaultBandwidth, &s.Language, &s.Theme,
		&s.AutoReconnect, &s.ShowOnlineStatus, &s.AllowDirectMessages, &s.CameraDefaultOn, &s.MicDefaultMuted,
		&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *UserRepository) UpsertSettings(ctx context.Context, s *models.UserSettings) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO user_settings (user_id, notifications_enabled, sound_enabled, default_bandwidth, language, theme,
		        auto_reconnect, show_online_status, allow_direct_messages, camera_default_on, mic_default_muted, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		 ON CONFLICT (user_id) DO UPDATE SET
		        notifications_enabled = $2, sound_enabled = $3, default_bandwidth = $4, language = $5, theme = $6,
		        auto_reconnect = $7, show_online_status = $8, allow_direct_messages = $9, camera_default_on = $10, mic_default_muted = $11,
		        updated_at = NOW()`,
		s.UserID, s.NotificationsEnabled, s.SoundEnabled, s.DefaultBandwidth, s.Language, s.Theme,
		s.AutoReconnect, s.ShowOnlineStatus, s.AllowDirectMessages, s.CameraDefaultOn, s.MicDefaultMuted,
	)
	return err
}

// Contacts

func (r *UserRepository) AddContact(ctx context.Context, userID, contactID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO contacts (id, user_id, contact_id, created_at) VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (user_id, contact_id) DO NOTHING`,
		uuid.New(), userID, contactID,
	)
	return err
}

func (r *UserRepository) RemoveContact(ctx context.Context, userID, contactID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM contacts WHERE user_id = $1 AND contact_id = $2`, userID, contactID)
	return err
}

func (r *UserRepository) BlockContact(ctx context.Context, userID, contactID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE contacts SET is_blocked = TRUE WHERE user_id = $1 AND contact_id = $2`, userID, contactID)
	return err
}

func (r *UserRepository) GetContacts(ctx context.Context, userID uuid.UUID) ([]models.Contact, error) {
	rows, err := r.db.Query(ctx,
		`SELECT c.id, c.user_id, c.contact_id, c.nickname, c.is_blocked, c.created_at,
		        u.username, u.display_name, u.avatar_url, u.status
		 FROM contacts c JOIN users u ON c.contact_id = u.id
		 WHERE c.user_id = $1 ORDER BY c.created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var c models.Contact
		rows.Scan(&c.ID, &c.UserID, &c.ContactID, &c.Nickname, &c.IsBlocked, &c.CreatedAt,
			&c.ContactUsername, &c.ContactDisplayName, &c.ContactAvatar, &c.ContactStatus)
		contacts = append(contacts, c)
	}
	return contacts, nil
}

func (r *UserRepository) IsContact(ctx context.Context, userID, contactID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM contacts WHERE user_id = $1 AND contact_id = $2)`,
		userID, contactID,
	).Scan(&exists)
	return exists, err
}
