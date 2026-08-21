package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gin/internal/message/entity"
)

type MessageRepository interface {
	// Rooms & membership
	GetOrCreateDirectRoom(userA, userB uint) (*entity.Room, error)
	IsMember(roomID, userID uint) (bool, error)
	GetRoomMemberIDs(roomID uint) ([]uint, error)
	GetUserRooms(userID uint) ([]entity.Room, error)
	GetUserNames(ids []uint) (map[uint]string, error)
	CreateGroupRoom(creatorID uint, memberIDs []uint, name string) (*entity.Room, error)

	// Messages
	Store(msg *entity.Message) error
	GetRoomMessages(roomID uint, limit, offset int) ([]entity.Message, error)

	// Read state
	MarkReadIfValid(userID, roomID, messageID uint) error
	MarkAllRead(userID, roomID, messageID uint) error
	GetReadStates(userID uint) ([]entity.RoomState, error)
	GetRoomReadStates(roomID uint) ([]entity.RoomState, error)

	// Users
	UserExists(id uint) (bool, error)
	GetContactIDs(userID uint) ([]uint, error)
}

type messageRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *messageRepository {
	return &messageRepository{db: db}
}


func (r *messageRepository) GetOrCreateDirectRoom(userA, userB uint) (*entity.Room, error) {
	var room *entity.Room

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Find existing direct room between userA and userB
		var existingRooms []entity.Room
		err := tx.Model(&entity.Room{}).
			Joins("JOIN room_members rm1 ON rm1.room_id = rooms.id AND rm1.user_id = ?", userA).
			Where("rooms.type = ?", entity.RoomTypeDirect).
			Find(&existingRooms).Error
		if err != nil {
			return err
		}

		for _, r := range existingRooms {
			var count int64
			if err := tx.Model(&entity.RoomMember{}).
				Where("room_id = ?", r.ID).
				Count(&count).Error; err != nil {
				return err
			}
			if count != 2 {
				continue
			}

			var bMember int64
			if err := tx.Model(&entity.RoomMember{}).
				Where("room_id = ? AND user_id = ?", r.ID, userB).
				Count(&bMember).Error; err != nil {
				return err
			}

			if bMember > 0 {
				room = &r
				return nil
			}
		}

		if room != nil {
			return nil
		}

		// Create new direct room
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

func (r *messageRepository) GetUserRooms(userID uint) ([]entity.Room, error) {
	var rooms []entity.Room

	err := r.db.
		Joins("JOIN room_members rm ON rm.room_id = rooms.id").
		Where("rm.user_id = ?", userID).
		Preload("Members").
		Preload("LastMessage", func(db *gorm.DB) *gorm.DB {
			return db.Joins(
				"JOIN (SELECT MAX(id) AS max_id FROM messages WHERE room_id IN (SELECT room_id FROM room_members WHERE user_id = ?) GROUP BY room_id) lm ON lm.max_id = messages.id",
				userID,
			)
		}).
		Order("rooms.created_at DESC").
		Find(&rooms).Error

	if err != nil {
		return nil, err
	}

	// Compute unread_count = number of messages in the room after the user's
	// last-read message_id. Users with no read state (last_read_message_id is
	// NULL) are treated as unread from the first message.
	type unreadRow struct {
		RoomID uint
		Count  int64
	}
	var unread []unreadRow
	err = r.db.Table("messages m").
		Select("m.room_id, COUNT(*) AS count").
		Joins("LEFT JOIN room_states rs ON rs.room_id = m.room_id AND rs.user_id = ?", userID).
		Where("m.room_id IN (?)",
			r.db.Model(&entity.RoomMember{}).Select("room_id").Where("user_id = ?", userID),
		).
		Where("COALESCE(rs.last_read_message_id, 0) < m.id").
		Group("m.room_id").
		Scan(&unread).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[uint]int64, len(unread))
	for _, u := range unread {
		counts[u.RoomID] = u.Count
	}

	for i := range rooms {
		rooms[i].UnreadCount = counts[rooms[i].ID]
	}

	return rooms, nil
}

func (r *messageRepository) GetUserNames(ids []uint) (map[uint]string, error) {
	names := map[uint]string{}

	if len(ids) == 0 {
		return names, nil
	}

	type userRow struct {
		ID   uint
		Name string
	}

	var rows []userRow
	err := r.db.Table("users").
		Select("id, name").
		Where("id IN (?)", ids).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		names[row.ID] = row.Name
	}

	return names, nil
}

func (r *messageRepository) CreateGroupRoom(creatorID uint, memberIDs []uint, name string) (*entity.Room, error) {
	var room *entity.Room

	err := r.db.Transaction(func(tx *gorm.DB) error {
		newRoom := entity.Room{Type: entity.RoomTypeGroup, Name: &name}
		if err := tx.Create(&newRoom).Error; err != nil {
			return err
		}

		members := []entity.RoomMember{
			{RoomID: newRoom.ID, UserID: creatorID},
		}
		for _, id := range memberIDs {
			if id == creatorID {
				continue
			}
			members = append(members, entity.RoomMember{RoomID: newRoom.ID, UserID: id})
		}

		if err := tx.Create(&members).Error; err != nil {
			return err
		}

		room = &newRoom
		return nil
	})

	return room, err
}

func (r *messageRepository) Store(msg *entity.Message) error {
	return r.db.Create(msg).Error
}

func (r *messageRepository) GetRoomMessages(roomID uint, limit, offset int) ([]entity.Message, error) {
	var messages []entity.Message

	// For offset-based pagination on time-ordered data, we need to be more careful
	// to avoid inconsistencies when new messages are added during pagination.
	// We'll use a subquery to get the IDs first, then join to get the full messages.
	// This approach is more reliable than simple OFFSET with ORDER BY.

	var messageIDs []uint
	err := r.db.Model(&entity.Message{}).
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Pluck("id", &messageIDs).Error
	if err != nil {
		return nil, err
	}

	if len(messageIDs) == 0 {
		return []entity.Message{}, nil
	}

	// Now get the full messages in the correct order
	err = r.db.
		Where("id IN (?)", messageIDs).
		Order("created_at DESC").
		Find(&messages).Error

	return messages, err
}


func (r *messageRepository) maxExpr(column string, _ uint) string {
	if r.db.Dialector.Name() == "sqlite" {
		return fmt.Sprintf("MAX(%s, ?)", column)
	}
	return fmt.Sprintf("GREATEST(%s, ?)", column)
}

// MarkReadIfValid marks a message as read only if:
// 1. The user is a member of the room
// 2. The message belongs to the room
// This is done in a single query using JOINs for better performance.
func (r *messageRepository) MarkReadIfValid(userID, roomID, messageID uint) error {
	expr := r.maxExpr("last_read_message_id", messageID)

	// First check if the conditions are met (user is in room AND message belongs to room)
	var count int64
	if err := r.db.Table("room_members rm").
		Joins("JOIN messages m ON m.room_id = rm.room_id").
		Where("rm.user_id = ? AND rm.room_id = ? AND m.id = ?", userID, roomID, messageID).
		Count(&count).Error; err != nil {
		return err
	}

	// If no matching rows, don't insert/update (user not in room or message doesn't belong)
	if count == 0 {
		return nil
	}

	// Insert or update the read state
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "room_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_read_message_id": gorm.Expr(expr, messageID),
			"updated_at":           time.Now(),
		}),
	}).Create(&entity.RoomState{
		UserID:            userID,
		RoomID:            roomID,
		LastReadMessageID: &messageID,
	}).Error
}

func (r *messageRepository) GetReadStates(userID uint) ([]entity.RoomState, error) {
	var states []entity.RoomState
	err := r.db.
		Preload("LastReadMessage").
		Where("user_id = ?", userID).
		Find(&states).Error
	return states, err
}

// GetRoomReadStates retrieves read states for all users in a specific room.
// This returns information about which messages each user has read in the room.
func (r *messageRepository) GetRoomReadStates(roomID uint) ([]entity.RoomState, error) {
	if roomID == 0 {
		return nil, nil
	}

	var states []entity.RoomState
	err := r.db.
		Where("room_id = ?", roomID).
		Find(&states).Error
	return states, err
}

func (r *messageRepository) UserExists(id uint) (bool, error) {
	var count int64
	err := r.db.Table("users").Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *messageRepository) GetContactIDs(userID uint) ([]uint, error) {
	var ids []uint

	err := r.db.Model(&entity.RoomMember{}).
		Where("room_id IN (?)",
			r.db.Model(&entity.RoomMember{}).
				Select("room_id").
				Where("user_id = ?", userID),
		).
		Where("user_id <> ?", userID).
		Distinct().
		Pluck("user_id", &ids).Error

	return ids, err
}

// MarkAllRead marks all messages in a room as read for the current user.
// This is used when the user sends a message - all messages in that chat are considered read.
// It marks all messages up to and including the specified messageID as read.
func (r *messageRepository) MarkAllRead(userID, roomID, messageID uint) error {
	// Use the same logic as MarkReadIfValid but without the validation
	// (since we're the sender, we know we're in the room and the message belongs)
	expr := r.maxExpr("last_read_message_id", messageID)

	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "room_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_read_message_id": gorm.Expr(expr, messageID),
			"updated_at":           time.Now(),
		}),
	}).Create(&entity.RoomState{
		UserID:            userID,
		RoomID:            roomID,
		LastReadMessageID: &messageID,
	}).Error
}
