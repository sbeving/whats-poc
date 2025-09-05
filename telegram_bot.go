package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// TelegramBot represents the Telegram bot with WhatsApp integration
type TelegramBot struct {
	bot             *tgbotapi.BotAPI
	userSessions    map[int64]*UserSession
	sessionMutex    sync.RWMutex
	storeContainer  *sqlstore.Container
}

// UserSession represents a user's WhatsApp session
type UserSession struct {
	TelegramUserID int64
	Client         *whatsmeow.Client
	Device         *store.Device
	IsConnected    bool
	IsLoggedIn     bool
}

// NewTelegramBot creates a new Telegram bot instance
func NewTelegramBot(token string, storeContainer *sqlstore.Container) (*TelegramBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %v", err)
	}

	return &TelegramBot{
		bot:            bot,
		userSessions:   make(map[int64]*UserSession),
		storeContainer: storeContainer,
	}, nil
}

// Start starts the Telegram bot
func (tb *TelegramBot) Start() error {
	tb.bot.Debug = false
	log.Infof("Authorized on account %s", tb.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tb.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			go tb.handleMessage(update.Message)
		}
	}

	return nil
}

// handleMessage processes incoming Telegram messages
func (tb *TelegramBot) handleMessage(message *tgbotapi.Message) {
	userID := message.From.ID
	
	// Handle commands
	if message.IsCommand() {
		tb.handleCommand(message)
		return
	}

	// Handle regular text messages
	tb.sendMessage(userID, "Please use commands to interact with the bot. Type /help for available commands.")
}

// handleCommand processes Telegram bot commands
func (tb *TelegramBot) handleCommand(message *tgbotapi.Message) {
	userID := message.From.ID
	command := message.Command()
	args := strings.Fields(message.CommandArguments())

	switch command {
	case "start":
		tb.handleStart(userID)
	case "help":
		tb.handleHelp(userID)
	case "login":
		tb.handleLogin(userID)
	case "status":
		tb.handleStatus(userID)
	case "listgroups":
		tb.handleListGroups(userID)
	case "getgroup":
		tb.handleGetGroup(userID, args)
	case "spoofed_reply":
		tb.handleSpoofedReply(userID, args)
	case "spoofed_img_reply":
		tb.handleSpoofedImgReply(userID, args)
	case "spoofed_demo":
		tb.handleSpoofedDemo(userID, args)
	default:
		tb.sendMessage(userID, "Unknown command. Type /help for available commands.")
	}
}

// handleStart handles the /start command
func (tb *TelegramBot) handleStart(userID int64) {
	welcome := `🤖 Welcome to WhatsApp Spoofing Bot!

This bot allows you to interact with WhatsApp spoofing features through Telegram.

⚠️ LEGAL DISCLAIMER: This tool is for educational and testing purposes only. Use responsibly and in accordance with WhatsApp's terms of service.

Available commands:
/help - Show help message
/login - Login to WhatsApp
/status - Check your WhatsApp connection status
/listgroups - List your WhatsApp groups
/getgroup <jid> - Get group information
/spoofed_reply <chat_jid> <msg_id> <spoofed_jid> <spoofed_text>|<text> - Send spoofed reply
/spoofed_img_reply <chat_jid> <msg_id> <spoofed_jid> <file_path> <spoofed_text>|<text> - Send spoofed image reply
/spoofed_demo <gender> <language> <chat_jid> <spoofed_jid> - Send spoofed demo message

To get started, use /login to connect your WhatsApp account.`

	tb.sendMessage(userID, welcome)
}

// handleHelp handles the /help command
func (tb *TelegramBot) handleHelp(userID int64) {
	help := `📋 Available Commands:

🔐 Authentication:
/login - Login to WhatsApp (shows QR code)
/status - Check WhatsApp connection status

📱 WhatsApp Operations:
/listgroups - List your WhatsApp groups
/getgroup <jid> - Get information about a specific group

🎭 Spoofing Operations:
/spoofed_reply <chat_jid> <msg_id> <spoofed_jid> <spoofed_text>|<text>
   - Send a spoofed reply message
   
/spoofed_img_reply <chat_jid> <msg_id> <spoofed_jid> <file_path> <spoofed_text>|<text>
   - Send a spoofed reply with image
   
/spoofed_demo <gender:boy|girl> <language:br|en> <chat_jid> <spoofed_jid>
   - Send a demo spoofed conversation

📝 Parameter Examples:
- chat_jid: 1234567890@s.whatsapp.net (for individual) or 1234567890-1234567890@g.us (for group)
- msg_id: Use ! for auto-generated ID
- spoofed_jid: 1234567890@s.whatsapp.net (user to impersonate)

⚠️ Use responsibly and ethically!`

	tb.sendMessage(userID, help)
}

// handleLogin handles WhatsApp login for a user
func (tb *TelegramBot) handleLogin(userID int64) {
	tb.sessionMutex.Lock()
	defer tb.sessionMutex.Unlock()

	// Check if user already has a session
	if session, exists := tb.userSessions[userID]; exists && session.IsLoggedIn {
		tb.sendMessage(userID, "✅ You are already logged in to WhatsApp!")
		return
	}

	// Create new device for this user
	device := tb.storeContainer.NewDevice()

	// Create WhatsApp client for this user
	client := whatsmeow.NewClient(device, waLog.Stdout("Client", "INFO", true))
	
	session := &UserSession{
		TelegramUserID: userID,
		Client:         client,
		Device:         device,
		IsConnected:    false,
		IsLoggedIn:     false,
	}
	
	tb.userSessions[userID] = session

	// Get QR code
	qrChan, err := client.GetQRChannel(context.Background())
	if err != nil {
		tb.sendMessage(userID, fmt.Sprintf("❌ Failed to get QR channel: %v", err))
		return
	}

	tb.sendMessage(userID, "📱 Generating QR code for WhatsApp login...")

	go func() {
		for evt := range qrChan {
			if evt.Event == "code" {
				// Send QR code to user
				tb.sendQRCode(userID, evt.Code)
			} else if evt.Event == "success" {
				session.IsLoggedIn = true
				session.IsConnected = true
				tb.sendMessage(userID, "✅ Successfully logged in to WhatsApp!")
				break
			} else if evt.Event == "timeout" {
				tb.sendMessage(userID, "⏰ QR code expired. Please try /login again.")
				break
			}
		}
	}()

	// Connect the client
	err = client.Connect()
	if err != nil {
		tb.sendMessage(userID, fmt.Sprintf("❌ Failed to connect: %v", err))
		return
	}

	session.IsConnected = true
}

// handleStatus checks the user's WhatsApp connection status
func (tb *TelegramBot) handleStatus(userID int64) {
	tb.sessionMutex.RLock()
	defer tb.sessionMutex.RUnlock()

	session, exists := tb.userSessions[userID]
	if !exists {
		tb.sendMessage(userID, "❌ No WhatsApp session found. Please use /login first.")
		return
	}

	status := "📊 WhatsApp Status:\n\n"
	if session.IsConnected {
		status += "🟢 Connection: Connected\n"
	} else {
		status += "🔴 Connection: Disconnected\n"
	}

	if session.IsLoggedIn {
		status += "✅ Login: Logged in\n"
		if session.Client.Store.ID != nil {
			status += fmt.Sprintf("📱 Phone: %s\n", session.Client.Store.ID.User)
		}
	} else {
		status += "❌ Login: Not logged in\n"
	}

	tb.sendMessage(userID, status)
}

// handleListGroups lists the user's WhatsApp groups
func (tb *TelegramBot) handleListGroups(userID int64) {
	session := tb.getUserSession(userID)
	if session == nil || !session.IsLoggedIn {
		tb.sendMessage(userID, "❌ Please login to WhatsApp first using /login")
		return
	}

	groups, err := session.Client.GetJoinedGroups()
	if err != nil {
		tb.sendMessage(userID, fmt.Sprintf("❌ Failed to get groups: %v", err))
		return
	}

	if len(groups) == 0 {
		tb.sendMessage(userID, "📝 You are not in any WhatsApp groups.")
		return
	}

	response := "📋 Your WhatsApp Groups:\n\n"
	for i, group := range groups {
		response += fmt.Sprintf("%d. %s\n   ID: `%s`\n\n", i+1, group.GroupName.Name, group.JID.String())
	}

	tb.sendMessage(userID, response)
}

// handleGetGroup gets information about a specific group
func (tb *TelegramBot) handleGetGroup(userID int64, args []string) {
	if len(args) < 1 {
		tb.sendMessage(userID, "❌ Usage: /getgroup <group_jid>")
		return
	}

	session := tb.getUserSession(userID)
	if session == nil || !session.IsLoggedIn {
		tb.sendMessage(userID, "❌ Please login to WhatsApp first using /login")
		return
	}

	groupJID, ok := parseJID(args[0])
	if !ok {
		tb.sendMessage(userID, "❌ Invalid group JID format")
		return
	}

	if groupJID.Server != types.GroupServer {
		tb.sendMessage(userID, fmt.Sprintf("❌ Input must be a group JID (@%s)", types.GroupServer))
		return
	}

	groupInfo, err := session.Client.GetGroupInfo(groupJID)
	if err != nil {
		tb.sendMessage(userID, fmt.Sprintf("❌ Failed to get group info: %v", err))
		return
	}

	response := fmt.Sprintf("📋 Group Information:\n\n")
	response += fmt.Sprintf("📝 Name: %s\n", groupInfo.Name)
	response += fmt.Sprintf("🆔 JID: `%s`\n", groupInfo.JID.String())
	response += fmt.Sprintf("👥 Participants: %d\n", len(groupInfo.Participants))
	response += fmt.Sprintf("👑 Owner: %s\n", groupInfo.OwnerJID.String())
	response += fmt.Sprintf("📅 Created: %s\n", groupInfo.GroupCreated.Format("2006-01-02 15:04:05"))

	tb.sendMessage(userID, response)
}

// handleSpoofedReply handles spoofed reply command
func (tb *TelegramBot) handleSpoofedReply(userID int64, args []string) {
	if len(args) < 4 {
		tb.sendMessage(userID, "❌ Usage: /spoofed_reply <chat_jid> <msg_id> <spoofed_jid> <spoofed_text>|<text>")
		return
	}

	session := tb.getUserSession(userID)
	if session == nil || !session.IsLoggedIn {
		tb.sendMessage(userID, "❌ Please login to WhatsApp first using /login")
		return
	}

	// Temporarily set global cli to user's client for compatibility
	originalCli := cli
	cli = session.Client
	defer func() { cli = originalCli }()

	result := cmdSendSpoofedReply(args)
	tb.sendMessage(userID, result)
}

// handleSpoofedImgReply handles spoofed image reply command
func (tb *TelegramBot) handleSpoofedImgReply(userID int64, args []string) {
	if len(args) < 5 {
		tb.sendMessage(userID, "❌ Usage: /spoofed_img_reply <chat_jid> <msg_id> <spoofed_jid> <file_path> <spoofed_text>|<text>")
		return
	}

	session := tb.getUserSession(userID)
	if session == nil || !session.IsLoggedIn {
		tb.sendMessage(userID, "❌ Please login to WhatsApp first using /login")
		return
	}

	// Temporarily set global cli to user's client for compatibility
	originalCli := cli
	cli = session.Client
	defer func() { cli = originalCli }()

	result := cmdSendSpoofedImgReply(args)
	tb.sendMessage(userID, result)
}

// handleSpoofedDemo handles spoofed demo command
func (tb *TelegramBot) handleSpoofedDemo(userID int64, args []string) {
	if len(args) < 4 {
		tb.sendMessage(userID, "❌ Usage: /spoofed_demo <gender:boy|girl> <language:br|en> <chat_jid> <spoofed_jid>")
		return
	}

	session := tb.getUserSession(userID)
	if session == nil || !session.IsLoggedIn {
		tb.sendMessage(userID, "❌ Please login to WhatsApp first using /login")
		return
	}

	// Temporarily set global cli to user's client for compatibility
	originalCli := cli
	cli = session.Client
	defer func() { cli = originalCli }()

	result := cmdSendSpoofedDemo(args)
	tb.sendMessage(userID, result)
}

// getUserSession gets a user's WhatsApp session
func (tb *TelegramBot) getUserSession(userID int64) *UserSession {
	tb.sessionMutex.RLock()
	defer tb.sessionMutex.RUnlock()
	return tb.userSessions[userID]
}

// sendMessage sends a message to a Telegram user
func (tb *TelegramBot) sendMessage(userID int64, text string) {
	msg := tgbotapi.NewMessage(userID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := tb.bot.Send(msg)
	if err != nil {
		log.Errorf("Failed to send message to user %d: %v", userID, err)
	}
}

// sendQRCode sends a QR code image to a Telegram user
func (tb *TelegramBot) sendQRCode(userID int64, qrCode string) {
	// For now, send QR code as text. In production, you might want to generate an actual QR image
	message := fmt.Sprintf("📱 WhatsApp QR Code:\n\n```\n%s\n```\n\nScan this QR code with your WhatsApp mobile app:\n1. Open WhatsApp\n2. Go to Settings → Linked Devices\n3. Tap 'Link a Device'\n4. Scan this QR code", qrCode)
	tb.sendMessage(userID, message)
}

// GetBotToken gets the bot token from environment variable
func GetBotToken() string {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Errorf("TELEGRAM_BOT_TOKEN environment variable is required")
		os.Exit(1)
	}
	return token
}