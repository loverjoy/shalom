package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	User         *UserRepository
	Meeting      *MeetingRepository
	Notification *NotificationRepository
	File         *FileRepository
	Channel      *ChannelRepository
}

func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		User:         NewUserRepository(db),
		Meeting:      NewMeetingRepository(db),
		Notification: NewNotificationRepository(db),
		File:         NewFileRepository(db),
		Channel:      NewChannelRepository(db),
	}
}
