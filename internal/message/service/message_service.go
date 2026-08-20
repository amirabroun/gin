package service

import (
	"errors"

	"gin/internal/message/entity"
	"gin/internal/message/repository"
	"gin/internal/ws"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrNotRoomMember   = errors.New("user is not a member of this room")
	ErrSelfDirectChat  = errors.New("cannot start a direct chat with yourself")
	ErrNoGroupMembers  = errors.New("group room requires at least one member")
	ErrInvalidRoomName = errors.New("room name is required")
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

	s.hub.Deliver(ws.MessageEnvelope(senderID, roomID, content).ToUsers(memberIDs))

	return nil
}

func (s *MessageService) PublishTyping(senderID, roomID uint, content string) {
	isMember, err := s.repo.IsMember(roomID, senderID)
	if err != nil || !isMember {
		return
	}

	memberIDs, err := s.repo.GetRoomMemberIDs(roomID)
	if err != nil {
		return
	}

	memberIDs = removeID(memberIDs, senderID)

	s.hub.Deliver(ws.TypingEnvelope(senderID, roomID, content).ToUsers(memberIDs))
}

func removeID(ids []uint, target uint) []uint {
	filtered := ids[:0]

	for _, id := range ids {
		if id != target {
			filtered = append(filtered, id)
		}
	}

	return filtered
}

func (s *MessageService) GetContactIDs(userID uint) ([]uint, error) {
	return s.repo.GetContactIDs(userID)
}

func (s *MessageService) GetUserRooms(userID uint) ([]entity.Room, error) {
	rooms, err := s.repo.GetUserRooms(userID)
	if err != nil {
		return nil, err
	}

	names, err := s.repo.GetUserNames(collectDirectPeerIDs(rooms, userID))
	if err != nil {
		return nil, err
	}

	for i := range rooms {
		if rooms[i].Type == entity.RoomTypeDirect {
			rooms[i].Name = peerDisplayName(rooms[i].Members, userID, names)
		}
	}

	return rooms, nil
}

func collectDirectPeerIDs(rooms []entity.Room, ownID uint) []uint {
	seen := make(map[uint]struct{}, len(rooms))
	ids := make([]uint, 0, len(rooms))
	for _, room := range rooms {
		if room.Type != entity.RoomTypeDirect {
			continue
		}
		for _, m := range room.Members {
			if m.UserID != ownID {
				if _, dup := seen[m.UserID]; !dup {
					seen[m.UserID] = struct{}{}
					ids = append(ids, m.UserID)
				}
				break
			}
		}
	}
	return ids
}

func peerDisplayName(members []entity.RoomMember, ownID uint, names map[uint]string) *string {
	for _, m := range members {
		if m.UserID != ownID {
			if name, ok := names[m.UserID]; ok {
				return &name
			}
			break
		}
	}
	return nil
}

func (s *MessageService) CreateGroupRoom(creatorID uint, memberIDs []uint, name string) (*entity.Room, error) {
	if name == "" {
		return nil, ErrInvalidRoomName
	}

	if len(memberIDs) == 0 {
		return nil, ErrNoGroupMembers
	}

	// Validate that every target user exists.
	for _, id := range memberIDs {
		if id == creatorID {
			continue
		}
		exists, err := s.repo.UserExists(id)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrUserNotFound
		}
	}

	return s.repo.CreateGroupRoom(creatorID, memberIDs, name)
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
