package controller

import (
	"backend/config"
	"backend/database"
	"backend/model"
	"backend/utils"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

func GetAgentConversations(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Unauthorized - User ID not found",
		})
	}

	role, ok := c.Locals("role").(string)
	if !ok || role != "agent" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied. Agent only.",
		})
	}

	var agent model.User
	if err := database.DB.First(&agent, userID).Error; err != nil || agent.Role != model.RoleAgent {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "User is not an agent",
		})
	}

	status := c.Query("status", "all")
	limit := c.Query("limit", "10")
	offset := c.Query("offset", "0")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	cacheKey := fmt.Sprintf("agent:conversations:%d:%s:%s:%s", userID, status, limit, offset)

	var cachedResponse fiber.Map
	err := utils.GetCache(cacheKey, &cachedResponse)
	if err == nil && cachedResponse != nil {
		return c.JSON(cachedResponse)
	}

	var channels []model.Channel
	query := database.DB.Model(&model.Channel{}).
		Where("assigned_agent_id = ?", userID)

	if status != "all" {
		query = query.Where("status = ?", status)
	}

	query = query.Limit(limitInt).Offset(offsetInt).
		Order("id DESC")

	if err := query.Find(&channels).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch channels",
		})
	}

	var total int64
	countQuery := database.DB.Model(&model.Channel{}).Where("assigned_agent_id = ?", userID)
	if status != "all" {
		countQuery = countQuery.Where("status = ?", status)
	}
	countQuery.Count(&total)

	var responseData []fiber.Map
	for _, channel := range channels {
		var customer model.User
		if err := database.DB.Select("id", "email", "full_name").First(&customer, channel.CustomerID).Error; err != nil {
			customer = model.User{
				ID:       channel.CustomerID,
				FullName: "Unknown",
				Email:    "unknown@email.com",
			}
		}

		lastMessage := getLastMessageFromCacheOrDB(channel.ID)
		unreadCount := getUnreadCountFromCacheOrDB(channel.ID, "customer")

		responseData = append(responseData, fiber.Map{
			"id":                channel.ID,
			"tenant_id":         channel.TenantID,
			"customer_id":       channel.CustomerID,
			"customer_name":     customer.FullName,
			"customer_email":    customer.Email,
			"status":            channel.Status,
			"assigned_agent_id": channel.AssignedAgentID,
			"last_message": fiber.Map{
				"id":          lastMessage.ID,
				"message":     lastMessage.Message,
				"sender_type": lastMessage.SenderType,
				"created_at":  lastMessage.CreatedAt,
			},
			"unread_count": unreadCount,
		})
	}

	response := fiber.Map{
		"success": true,
		"data":    responseData,
		"pagination": fiber.Map{
			"total":  total,
			"limit":  limitInt,
			"offset": offsetInt,
		},
	}

	utils.SetCache(cacheKey, response, 30*time.Second)

	return c.JSON(response)
}

func GetChannelByID(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(uint)
	role, _ := c.Locals("role").(string)

	channelID := c.Params("id")
	if channelID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Channel ID is required",
		})
	}

	cacheKey := fmt.Sprintf("channel:%s:role:%s:user:%d", channelID, role, userID)

	var cachedResponse fiber.Map
	err := utils.GetCache(cacheKey, &cachedResponse)
	if err == nil && cachedResponse != nil {
		return c.JSON(cachedResponse)
	}

	var channel model.Channel
	query := database.DB

	if role == "agent" {
		query = query.Where("assigned_agent_id = ?", userID)
	}

	if err := query.First(&channel, channelID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Channel not found",
		})
	}

	var messages []model.Message
	database.DB.Where("conversation_id = ?", channel.ID).
		Order("id ASC").
		Find(&messages)

	var customer model.User
	database.DB.Select("id", "email", "full_name").First(&customer, channel.CustomerID)

	if role == "agent" {
		database.DB.Model(&model.Message{}).
			Where("conversation_id = ? AND sender_type = ? AND is_read = ?",
				channel.ID, "customer", false).
			Update("is_read", true)

		utils.DeleteCache(fmt.Sprintf("unread:channel:%s:agent", channelID))
	} else if role == "user" {
		database.DB.Model(&model.Message{}).
			Where("conversation_id = ? AND sender_type = ? AND is_read = ?",
				channel.ID, "agent", false).
			Update("is_read", true)

		utils.DeleteCache(fmt.Sprintf("unread:channel:%s:user", channelID))
	}

	response := fiber.Map{
		"success": true,
		"data": fiber.Map{
			"channel": fiber.Map{
				"id":                channel.ID,
				"tenant_id":         channel.TenantID,
				"customer_id":       channel.CustomerID,
				"customer_name":     customer.FullName,
				"customer_email":    customer.Email,
				"status":            channel.Status,
				"assigned_agent_id": channel.AssignedAgentID,
				"created_at":        channel.CreatedAt,
				"updated_at":        channel.UpdatedAt,
			},
			"messages": messages,
		},
	}

	utils.SetCache(cacheKey, response, 10*time.Second)

	return c.JSON(response)
}

func AssignChannel(c fiber.Ctx) error {
	agentID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Unauthorized",
		})
	}

	role, _ := c.Locals("role").(string)
	if role != "agent" && role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied. Agent or admin only.",
		})
	}

	channelID := c.Params("id")
	if channelID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Channel ID is required",
		})
	}

	var channel model.Channel
	if err := database.DB.First(&channel, channelID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Channel not found",
		})
	}

	updates := map[string]interface{}{
		"assigned_agent_id": agentID,
		"status":            "assigned",
		"updated_at":        time.Now(),
	}

	if err := database.DB.Model(&channel).Updates(updates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to assign channel",
		})
	}

	invalidateChannelCache(channel.ID)
	invalidateAgentConversationsCache(agentID)

	if role == "admin" && channel.AssignedAgentID != agentID {
		invalidateAgentConversationsCache(channel.AssignedAgentID)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Channel assigned successfully",
		"data": fiber.Map{
			"channel_id":        channel.ID,
			"assigned_agent_id": agentID,
			"status":            "assigned",
		},
	})
}

func CloseChannel(c fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(uint)
	role, _ := c.Locals("role").(string)

	channelID := c.Params("id")
	if channelID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Channel ID is required",
		})
	}

	var channel model.Channel
	query := database.DB

	if role == "agent" {
		query = query.Where("assigned_agent_id = ?", userID)
	}

	if err := query.First(&channel, channelID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Channel not found or not assigned to you",
		})
	}

	if err := database.DB.Model(&channel).Update("status", "closed").Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to close channel",
		})
	}

	invalidateChannelCache(channel.ID)
	invalidateAgentConversationsCache(channel.AssignedAgentID)
	if channel.CustomerID > 0 {
		invalidateUserConversationsCache(channel.CustomerID)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Channel closed successfully",
	})
}

func SendMessage(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Unauthorized",
		})
	}

	role, _ := c.Locals("role").(string)
	channelID := c.Params("id")

	if channelID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Channel ID is required",
		})
	}

	var req struct {
		Message string `json:"message"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	if req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Message is required",
		})
	}

	var channel model.Channel
	if err := database.DB.First(&channel, channelID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Channel not found",
		})
	}

	if role == "agent" {
		if channel.AssignedAgentID != userID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   true,
				"message": "Access denied. This channel is not assigned to you.",
			})
		}
	} else if role == "user" {
		if channel.CustomerID != userID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   true,
				"message": "Access denied. This is not your channel.",
			})
		}
	}

	senderType := "agent"
	if role == "user" {
		senderType = "customer"
	}

	message := model.Message{
		ConversationID: channel.ID,
		SenderType:     senderType,
		Message:        req.Message,
		IsRead:         false,
	}

	if err := database.DB.Create(&message).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to send message",
		})
	}

	database.DB.Model(&channel).Update("updated_at", time.Now())

	invalidateChannelCache(channel.ID)
	invalidateLastMessageCache(channel.ID)
	invalidateAgentConversationsCache(channel.AssignedAgentID)
	if channel.CustomerID > 0 {
		invalidateUserConversationsCache(channel.CustomerID)
	}

	publishMessageToRedis(channel.ID, message)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Message sent successfully",
		"data":    message,
	})
}

func GetAvailableChannels(c fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	if role != "agent" && role != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": "Access denied",
		})
	}

	cacheKey := "channels:available"

	var cachedResponse fiber.Map
	err := utils.GetCache(cacheKey, &cachedResponse)
	if err == nil && cachedResponse != nil {
		return c.JSON(cachedResponse)
	}

	var channels []model.Channel
	if err := database.DB.Where("status = ? AND assigned_agent_id = ?", "open", 0).
		Order("id ASC").
		Find(&channels).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to fetch available channels",
		})
	}

	var responseData []fiber.Map
	for _, channel := range channels {
		var customer model.User
		database.DB.Select("id", "email", "full_name").First(&customer, channel.CustomerID)

		responseData = append(responseData, fiber.Map{
			"id":             channel.ID,
			"tenant_id":      channel.TenantID,
			"customer_id":    channel.CustomerID,
			"customer_name":  customer.FullName,
			"customer_email": customer.Email,
			"status":         channel.Status,
		})
	}

	response := fiber.Map{
		"success": true,
		"data":    responseData,
	}

	utils.SetCache(cacheKey, response, 15*time.Second)

	return c.JSON(response)
}

func GetChannelStats(c fiber.Ctx) error {
	agentID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Unauthorized",
		})
	}

	cacheKey := fmt.Sprintf("agent:stats:%d", agentID)

	var cachedResponse fiber.Map
	err := utils.GetCache(cacheKey, &cachedResponse)
	if err == nil && cachedResponse != nil {
		return c.JSON(cachedResponse)
	}

	var totalOpen, totalAssigned, totalClosed int64

	database.DB.Model(&model.Channel{}).Where("assigned_agent_id = ? AND status = ?", agentID, "open").Count(&totalOpen)
	database.DB.Model(&model.Channel{}).Where("assigned_agent_id = ? AND status = ?", agentID, "assigned").Count(&totalAssigned)
	database.DB.Model(&model.Channel{}).Where("assigned_agent_id = ? AND status = ?", agentID, "closed").Count(&totalClosed)

	var totalUnread int64
	database.DB.Model(&model.Message{}).
		Joins("JOIN channels ON messages.conversation_id = channels.id").
		Where("channels.assigned_agent_id = ? AND messages.sender_type = ?", agentID, "customer").
		Count(&totalUnread)

	response := fiber.Map{
		"success": true,
		"data": fiber.Map{
			"open":     totalOpen,
			"assigned": totalAssigned,
			"closed":   totalClosed,
			"total":    totalOpen + totalAssigned + totalClosed,
			"unread":   totalUnread,
		},
	}

	utils.SetCache(cacheKey, response, 20*time.Second)

	return c.JSON(response)
}

func CreateChannel(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Unauthorized",
		})
	}

	var req struct {
		TenantID uint   `json:"tenant_id"`
		Message  string `json:"message"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}

	if req.TenantID == 0 {
		req.TenantID = 1
	}

	if req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Message is required",
		})
	}

	tx := database.DB.Begin()

	channel := model.Channel{
		TenantID:        req.TenantID,
		CustomerID:      userID,
		Status:          "open",
		AssignedAgentID: 0,
	}

	if err := tx.Create(&channel).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to create channel",
		})
	}

	message := model.Message{
		ConversationID: channel.ID,
		SenderType:     "customer",
		Message:        req.Message,
		IsRead:         false,
	}

	if err := tx.Create(&message).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to send message",
		})
	}

	tx.Commit()

	utils.DeleteCache("channels:available")
	invalidateUserConversationsCache(userID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Channel created successfully",
		"data": fiber.Map{
			"channel_id": channel.ID,
			"message_id": message.ID,
		},
	})
}

func getLastMessageFromCacheOrDB(channelID uint) model.Message {
	cacheKey := fmt.Sprintf("channel:lastmessage:%d", channelID)

	var lastMessage model.Message
	err := utils.GetCache(cacheKey, &lastMessage)
	if err == nil && lastMessage.ID > 0 {
		return lastMessage
	}

	database.DB.Where("conversation_id = ?", channelID).
		Order("id DESC").
		First(&lastMessage)

	if lastMessage.ID > 0 {
		utils.SetCache(cacheKey, lastMessage, 10*time.Second)
	}

	return lastMessage
}

func getUnreadCountFromCacheOrDB(channelID uint, senderType string) int64 {
	cacheKey := fmt.Sprintf("unread:channel:%d:%s", channelID, senderType)

	var count int64
	err := utils.GetCache(cacheKey, &count)
	if err == nil {
		return count
	}

	database.DB.Model(&model.Message{}).
		Where("conversation_id = ? AND sender_type = ? AND is_read = ?",
			channelID, senderType, false).
		Count(&count)

	utils.SetCache(cacheKey, count, 5*time.Second)

	return count
}

func invalidateChannelCache(channelID uint) {
	pattern := fmt.Sprintf("channel:%d:*", channelID)
	keys, _ := config.RedisClient.Keys(config.Ctx, pattern).Result()
	for _, key := range keys {
		config.RedisClient.Del(config.Ctx, key)
	}

	utils.DeleteCache(fmt.Sprintf("channel:lastmessage:%d", channelID))
	utils.DeleteCache(fmt.Sprintf("unread:channel:%d:customer", channelID))
	utils.DeleteCache(fmt.Sprintf("unread:channel:%d:agent", channelID))
}

func invalidateAgentConversationsCache(agentID uint) {
	if agentID == 0 {
		return
	}
	pattern := fmt.Sprintf("agent:conversations:%d:*", agentID)
	keys, _ := config.RedisClient.Keys(config.Ctx, pattern).Result()
	for _, key := range keys {
		config.RedisClient.Del(config.Ctx, key)
	}

	utils.DeleteCache(fmt.Sprintf("agent:stats:%d", agentID))
}

func invalidateUserConversationsCache(userID uint) {
	if userID == 0 {
		return
	}
	pattern := fmt.Sprintf("user:conversations:%d:*", userID)
	keys, _ := config.RedisClient.Keys(config.Ctx, pattern).Result()
	for _, key := range keys {
		config.RedisClient.Del(config.Ctx, key)
	}
}

func invalidateLastMessageCache(channelID uint) {
	utils.DeleteCache(fmt.Sprintf("channel:lastmessage:%d", channelID))
}

func publishMessageToRedis(channelID uint, message model.Message) {
	messageJSON, _ := json.Marshal(message)
	config.RedisClient.Publish(config.Ctx, fmt.Sprintf("channel:%d", channelID), messageJSON)
}
