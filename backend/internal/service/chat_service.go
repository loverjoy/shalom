package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"shalom/internal/models"
)

type ChatService struct {
	messages *mongo.Collection
	rooms    *mongo.Collection
}

func NewChatService(db *mongo.Database) *ChatService {
	return &ChatService{
		messages: db.Collection("messages"),
		rooms:    db.Collection("chat_rooms"),
	}
}

func (s *ChatService) SendTextMessage(ctx context.Context, chatRoomID string, senderID uuid.UUID, senderName, content string, replyTo *uuid.UUID) (*models.Message, error) {
	msg := models.Message{
		ID:         uuid.New(),
		ChatRoomID: chatRoomID,
		SenderID:   senderID,
		SenderName: senderName,
		Content:    content,
		Type:       models.MessageTypeText,
		ReplyTo:    replyTo,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := s.messages.InsertOne(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Update room last message
	s.rooms.UpdateOne(ctx, bson.M{"_id": chatRoomID}, bson.M{
		"$set": bson.M{"last_message": msg, "updated_at": time.Now()},
		"$inc": bson.M{"message_count": 1},
	})

	return &msg, nil
}

func (s *ChatService) SendFileMessage(ctx context.Context, chatRoomID string, senderID uuid.UUID, senderName, fileURL, fileName string, fileSize int64, msgType models.MessageType) (*models.Message, error) {
	msg := models.Message{
		ID:         uuid.New(),
		ChatRoomID: chatRoomID,
		SenderID:   senderID,
		SenderName: senderName,
		Content:    fileName,
		Type:       msgType,
		FileURL:    fileURL,
		FileName:   fileName,
		FileSize:   fileSize,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := s.messages.InsertOne(ctx, msg)
	if err != nil {
		return nil, err
	}

	s.rooms.UpdateOne(ctx, bson.M{"_id": chatRoomID}, bson.M{
		"$set": bson.M{"last_message": msg, "updated_at": time.Now()},
		"$inc": bson.M{"message_count": 1},
	})

	return &msg, nil
}

func (s *ChatService) SendVoiceMessage(ctx context.Context, chatRoomID string, senderID uuid.UUID, senderName, fileURL string, duration int) (*models.Message, error) {
	msg := models.Message{
		ID:            uuid.New(),
		ChatRoomID:    chatRoomID,
		SenderID:      senderID,
		SenderName:    senderName,
		Content:       "Voice message",
		Type:          models.MessageTypeVoice,
		FileURL:       fileURL,
		VoiceDuration: duration,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err := s.messages.InsertOne(ctx, msg)
	if err != nil {
		return nil, err
	}

	s.rooms.UpdateOne(ctx, bson.M{"_id": chatRoomID}, bson.M{
		"$set": bson.M{"last_message": msg, "updated_at": time.Now()},
		"$inc": bson.M{"message_count": 1},
	})

	return &msg, nil
}

func (s *ChatService) SendPollMessage(ctx context.Context, chatRoomID string, senderID uuid.UUID, senderName, question string, options []string, isAnonymous bool) (*models.Message, error) {
	pollOptions := make([]models.PollOption, len(options))
	for i, opt := range options {
		pollOptions[i] = models.PollOption{Text: opt, VoteCount: 0}
	}

	msg := models.Message{
		ID:         uuid.New(),
		ChatRoomID: chatRoomID,
		SenderID:   senderID,
		SenderName: senderName,
		Content:    question,
		Type:       models.MessageTypePoll,
		Poll: &models.Poll{
			Question:    question,
			Options:     pollOptions,
			IsAnonymous: isAnonymous,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := s.messages.InsertOne(ctx, msg)
	if err != nil {
		return nil, err
	}

	s.rooms.UpdateOne(ctx, bson.M{"_id": chatRoomID}, bson.M{
		"$set": bson.M{"last_message": msg, "updated_at": time.Now()},
		"$inc": bson.M{"message_count": 1},
	})

	return &msg, nil
}

func (s *ChatService) VotePoll(ctx context.Context, messageID string, userID uuid.UUID, optionIndex int) error {
	filter := bson.M{"_id": messageID}
	update := bson.M{
		"$inc": bson.M{"poll.options." + string(rune('0'+optionIndex)) + ".vote_count": 1},
		"$addToSet": bson.M{"poll.options." + string(rune('0'+optionIndex)) + ".voters": userID},
	}
	_, err := s.messages.UpdateOne(ctx, filter, update)
	return err
}

func (s *ChatService) AddReaction(ctx context.Context, messageID string, userID uuid.UUID, emoji string) error {
	_, err := s.messages.UpdateOne(ctx,
		bson.M{"_id": messageID},
		bson.M{"$addToSet": bson.M{"reactions": models.Reaction{UserID: userID, Emoji: emoji}}},
	)
	return err
}

func (s *ChatService) RemoveReaction(ctx context.Context, messageID string, userID uuid.UUID, emoji string) error {
	_, err := s.messages.UpdateOne(ctx,
		bson.M{"_id": messageID},
		bson.M{"$pull": bson.M{"reactions": bson.M{"user_id": userID, "emoji": emoji}}},
	)
	return err
}

func (s *ChatService) EditMessage(ctx context.Context, messageID string, userID uuid.UUID, newContent string) error {
	_, err := s.messages.UpdateOne(ctx,
		bson.M{"_id": messageID, "sender_id": userID},
		bson.M{"$set": bson.M{"content": newContent, "is_edited": true, "updated_at": time.Now()}},
	)
	return err
}

func (s *ChatService) DeleteMessage(ctx context.Context, messageID string, userID uuid.UUID) error {
	_, err := s.messages.UpdateOne(ctx,
		bson.M{"_id": messageID, "sender_id": userID},
		bson.M{"$set": bson.M{"is_deleted": true, "content": "This message was deleted", "updated_at": time.Now()}},
	)
	return err
}

func (s *ChatService) PinMessage(ctx context.Context, messageID string, pin bool) error {
	_, err := s.messages.UpdateOne(ctx,
		bson.M{"_id": messageID},
		bson.M{"$set": bson.M{"is_pinned": pin}},
	)
	return err
}

func (s *ChatService) GetMessages(ctx context.Context, chatRoomID string, limit int64, before *time.Time) ([]models.Message, error) {
	filter := bson.M{"chat_room_id": chatRoomID, "is_deleted": false}
	if before != nil {
		filter["created_at"] = bson.M{"$lt": before}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit)

	cursor, err := s.messages.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *ChatService) CreateChatRoom(ctx context.Context, roomType, name string, meetingID string, createdBy uuid.UUID, members []string) (*models.ChatRoom, error) {
	room := models.ChatRoom{
		ID:           uuid.New().String(),
		Type:         roomType,
		Name:         name,
		MeetingID:    meetingID,
		Members:      members,
		CreatedBy:    createdBy,
		MessageCount: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := s.rooms.InsertOne(ctx, room)
	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (s *ChatService) GetChatRooms(ctx context.Context, userID uuid.UUID) ([]models.ChatRoom, error) {
	cursor, err := s.rooms.Find(ctx, bson.M{"members": userID.String()})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rooms []models.ChatRoom
	if err := cursor.All(ctx, &rooms); err != nil {
		return nil, err
	}
	return rooms, nil
}
