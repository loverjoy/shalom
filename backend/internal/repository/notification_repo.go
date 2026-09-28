package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"shalom/internal/models"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO notifications (id, user_id, type, title, body, data, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())
		 RETURNING id, created_at`,
		n.ID, n.UserID, n.Type, n.Title, n.Body, n.Data,
	).Scan(&n.ID, &n.CreatedAt)
}

func (r *NotificationRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, type, title, body, data, is_read, created_at
		 FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []models.Notification
	for rows.Next() {
		var n models.Notification
		rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Data, &n.IsRead, &n.CreatedAt)
		notifs = append(notifs, n)
	}
	return notifs, nil
}

func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`, userID,
	).Scan(&count)
	return count, err
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET is_read = TRUE WHERE id = $1`, id)
	return err
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`, userID)
	return err
}

func (r *NotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM notifications WHERE id = $1`, id)
	return err
}

func (r *NotificationRepository) DeleteOlderThan(ctx context.Context, userID uuid.UUID, before time.Time) (int64, error) {
	result, err := r.db.Exec(ctx,
		`DELETE FROM notifications WHERE user_id = $1 AND created_at < $2`, userID, before,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
