package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"shalom/internal/models"
)

type ChannelRepository struct {
	db *pgxpool.Pool
}

func NewChannelRepository(db *pgxpool.Pool) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) Create(ctx context.Context, ch *models.Channel) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO channels (id, name, description, created_by, is_public, member_count, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, 1, NOW(), NOW())
		 RETURNING id, created_at, updated_at`,
		ch.ID, ch.Name, ch.Description, ch.CreatedBy, ch.IsPublic,
	).Scan(&ch.ID, &ch.CreatedAt, &ch.UpdatedAt)
}

func (r *ChannelRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	ch := &models.Channel{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, created_by, is_public, member_count, created_at, updated_at
		 FROM channels WHERE id = $1`, id,
	).Scan(&ch.ID, &ch.Name, &ch.Description, &ch.CreatedBy, &ch.IsPublic, &ch.MemberCount, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (r *ChannelRepository) GetPublic(ctx context.Context, limit int) ([]models.Channel, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, description, created_by, is_public, member_count, created_at, updated_at
		 FROM channels WHERE is_public = TRUE ORDER BY member_count DESC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []models.Channel
	for rows.Next() {
		var ch models.Channel
		rows.Scan(&ch.ID, &ch.Name, &ch.Description, &ch.CreatedBy, &ch.IsPublic, &ch.MemberCount, &ch.CreatedAt, &ch.UpdatedAt)
		channels = append(channels, ch)
	}
	return channels, nil
}

func (r *ChannelRepository) AddMember(ctx context.Context, channelID, userID uuid.UUID, role string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO channel_members (id, channel_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (channel_id, user_id) DO NOTHING`,
		uuid.New(), channelID, userID, role,
	)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `UPDATE channels SET member_count = member_count + 1, updated_at = NOW() WHERE id = $1`, channelID)
	return err
}

func (r *ChannelRepository) RemoveMember(ctx context.Context, channelID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM channel_members WHERE channel_id = $1 AND user_id = $2`, channelID, userID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `UPDATE channels SET member_count = member_count - 1, updated_at = NOW() WHERE id = $1`, channelID)
	return err
}

func (r *ChannelRepository) GetMembers(ctx context.Context, channelID uuid.UUID) ([]models.ChannelMember, error) {
	rows, err := r.db.Query(ctx,
		`SELECT cm.id, cm.channel_id, cm.user_id, cm.role, cm.joined_at,
		        u.username, u.display_name, u.avatar_url
		 FROM channel_members cm JOIN users u ON cm.user_id = u.id
		 WHERE cm.channel_id = $1 ORDER BY cm.joined_at`, channelID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.ChannelMember
	for rows.Next() {
		var m models.ChannelMember
		rows.Scan(&m.ID, &m.ChannelID, &m.UserID, &m.Role, &m.JoinedAt,
			&m.Username, &m.DisplayName, &m.AvatarURL)
		members = append(members, m)
	}
	return members, nil
}

func (r *ChannelRepository) IsMember(ctx context.Context, channelID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM channel_members WHERE channel_id = $1 AND user_id = $2)`,
		channelID, userID,
	).Scan(&exists)
	return exists, err
}

func (r *ChannelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM channels WHERE id = $1`, id)
	return err
}
