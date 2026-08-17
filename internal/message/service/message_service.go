package service

import (
	"errors"

	"gin/internal/message/entity"
	"gin/internal/message/repository"
	"gin/internal/ws"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrNotRoomMember  = errors.New("user is not a member of this room")
	ErrSelfDirectChat = errors.New("cannot start a direct chat with yourself")
)

type MessageService struct {
	repo repository.MessageRepository
	hub  *ws.Hub
}

func New(repo repository.MessageRepository, hub *ws.Hub) *MessageService {
	return &MessageService{repo: repo, hub: hub}
}

func (s *MessageService) GetOrCreateDirectRoom(userA, userB uint) (*entity.Room, error) {
	if userA == userB {
		return nil, ErrSelfDirectChat
	}

	exists, err := s.repo.UserExists(userB)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrUserNotFound
	}

	return s.repo.GetOrCreateDirectRoom(userA, userB)
}

func (s *MessageService) Send(senderID, roomID uint, content string) error {
	isMember, err := s.repo.IsMember(roomID, senderID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotRoomMember
	}

	msg := entity.Message{
		RoomID:   roomID,
		SenderID: senderID,
		Content:  content,
	}
	if err := s.repo.Store(&msg); err != nil {
		return err
	}

	memberIDs, err := s.repo.GetRoomMemberIDs(roomID)
	if err != nil {
		return err
	}

	s.hub.PublishToRoom(&ws.RoomMessage{
		MemberIDs: memberIDs,
		Payload:   ws.MustEnvelope(ws.MessageTypeMessage, msg),
	})

	return nil
}

func (s *MessageService) GetContactIDs(userID uint) ([]uint, error) {
	return s.repo.GetContactIDs(userID)
}

func (s *MessageService) GetRoomMessages(userID, roomID uint, limit, offset int) ([]entity.Message, error) {
	isMember, err := s.repo.IsMember(roomID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotRoomMember
	}

	return s.repo.GetRoomMessages(roomID, limit, offset)
}
