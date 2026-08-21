package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"gin/internal/message/entity"
)

func TestGetRoomMessagesOrder(t *testing.T) {
	// Set up test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect to test database:", err)
	}

	// Migrate schema
	err = db.AutoMigrate(&entity.Room{}, &entity.RoomMember{}, &entity.Message{})
	if err != nil {
		t.Fatal("Failed to migrate schema:", err)
	}

	// Create repository
	repo := New(db)

	// Create test data
	room := entity.Room{Type: entity.RoomTypeDirect}
	if err := db.Create(&room).Error; err != nil {
		t.Fatal("Failed to create test room:", err)
	}

	// Create test messages with different timestamps
	now := time.Now()
	messages := []entity.Message{
		{RoomID: room.ID, SenderID: 1, Content: "First message", CreatedAt: now.Add(-2 * time.Hour)},
		{RoomID: room.ID, SenderID: 1, Content: "Second message", CreatedAt: now.Add(-1 * time.Hour)},
		{RoomID: room.ID, SenderID: 1, Content: "Third message", CreatedAt: now},
	}

	for _, msg := range messages {
		if err := db.Create(&msg).Error; err != nil {
			t.Fatal("Failed to create test message:", err)
		}
	}

	// Test GetRoomMessages with DESC order
	retrievedMessages, err := repo.GetRoomMessages(room.ID, 3, 0)
	if err != nil {
		t.Fatal("Failed to get room messages:", err)
	}

	// Verify we got 3 messages
	assert.Len(t, retrievedMessages, 3, "Should retrieve 3 messages")

	// Verify the order is DESC (newest first)
	assert.Equal(t, "Third message", retrievedMessages[0].Content, "First message should be the newest")
	assert.Equal(t, "Second message", retrievedMessages[1].Content, "Second message should be middle")
	assert.Equal(t, "First message", retrievedMessages[2].Content, "Last message should be the oldest")

	// Verify timestamps are in descending order
	assert.True(t, retrievedMessages[0].CreatedAt.After(retrievedMessages[1].CreatedAt), "First message should be newer than second")
	assert.True(t, retrievedMessages[1].CreatedAt.After(retrievedMessages[2].CreatedAt), "Second message should be newer than third")
}