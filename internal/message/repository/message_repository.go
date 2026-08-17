package repository

import (
	"gorm.io/gorm"

	"gin/internal/message/entity"
)

type MessageRepository interface {
	// Rooms & membership
	FindDirectRoom(userA, userB uint) (*entity.Room, error)
	GetOrCreateDirectRoom(userA, userB uint) (*entity.Room, error)
	AddMember(roomID, userID uint) error
	IsMember(roomID, userID uint) (bool, error)
	GetRoomMemberIDs(roomID uint) ([]uint, error)

	// Messages
	Store(msg *entity.Message) error
	GetRoomMessages(roomID uint, limit, offset int) ([]entity.Message, error)

	// Users
	UserExists(id uint) (bool, error)
}

type messageRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *messageRepository {
	return &messageRepository{db: db}
}

func findDirectRoom(tx *gorm.DB, userA, userB uint) (*entity.Room, error) {
	var rooms []entity.Room

	err := tx.
		Model(&entity.Room{}).
		Joins("JOIN room_members rm1 ON rm1.room_id = rooms.id AND rm1.user_id = ?", userA).
		Where("rooms.type = ?", entity.RoomTypeDirect).
		Find(&rooms).Error

	if err != nil {
		return nil, err
	}

	for _, room := range rooms {
		var count int64
		if err := tx.Model(&entity.RoomMember{}).
			Where("room_id = ?", room.ID).
			Count(&count).Error; err != nil {
			return nil, err
		}

		if count != 2 {
			continue
		}

		var bMember int64
		if err := tx.Model(&entity.RoomMember{}).
			Where("room_id = ? AND user_id = ?", room.ID, userB).
			Count(&bMember).Error; err != nil {
			return nil, err
		}

		if bMember > 0 {
			return &room, nil
		}
	}

	return nil, nil
}

func (r *messageRepository) FindDirectRoom(userA, userB uint) (*entity.Room, error) {
	return findDirectRoom(r.db, userA, userB)
}

func (r *messageRepository) GetOrCreateDirectRoom(userA, userB uint) (*entity.Room, error) {
	var room *entity.Room

	err := r.db.Transaction(func(tx *gorm.DB) error {
		existing, err := findDirectRoom(tx, userA, userB)
		if err != nil {
			return err
		}
		if existing != nil {
			room = existing
			return nil
		}

		newRoom := entity.Room{Type: entity.RoomTypeDirect}
		if err := tx.Create(&newRoom).Error; err != nil {
			return err
		}

		if err := tx.Create(&entity.RoomMember{RoomID: newRoom.ID, UserID: userA}).Error; err != nil {
			return err
		}
		if err := tx.Create(&entity.RoomMember{RoomID: newRoom.ID, UserID: userB}).Error; err != nil {
			return err
		}

		room = &newRoom
		return nil
	})

	return room, err
}

func (r *messageRepository) AddMember(roomID, userID uint) error {
	member := entity.RoomMember{RoomID: roomID, UserID: userID}
	return r.db.Create(&member).Error
}

func (r *messageRepository) IsMember(roomID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&entity.RoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *messageRepository) GetRoomMemberIDs(roomID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&entity.RoomMember{}).
		Where("room_id = ?", roomID).
		Pluck("user_id", &ids).Error
	return ids, err
}

func (r *messageRepository) Store(msg *entity.Message) error {
	return r.db.Create(msg).Error
}

func (r *messageRepository) GetRoomMessages(roomID uint, limit, offset int) ([]entity.Message, error) {
	var messages []entity.Message
	err := r.db.
		Where("room_id = ?", roomID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *messageRepository) UserExists(id uint) (bool, error) {
	var count int64
	err := r.db.Table("users").Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
