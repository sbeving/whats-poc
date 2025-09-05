#!/bin/bash

# Test script for WhatsApp Spoofing Bot

echo "🧪 Testing WhatsApp Spoofing Bot..."

# Test CLI mode build
echo "📋 Testing CLI mode..."
./whats-spoofing-bot --help > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ CLI mode help works"
else
    echo "❌ CLI mode help failed"
fi

# Test that the binary exists
if [ -f "./whats-spoofing-bot" ]; then
    echo "✅ Binary created successfully"
else
    echo "❌ Binary not found"
fi

# Test that Telegram flag is recognized
./whats-spoofing-bot --help 2>&1 | grep -q "telegram"
if [ $? -eq 0 ]; then
    echo "✅ Telegram flag is available"
else
    echo "❌ Telegram flag not found"
fi

echo ""
echo "🚀 To run in CLI mode:"
echo "   ./whats-spoofing-bot"
echo ""
echo "🤖 To run in Telegram bot mode:"
echo "   export TELEGRAM_BOT_TOKEN='your_bot_token'"
echo "   ./whats-spoofing-bot -telegram"
echo ""
echo "📝 For more information, see README.md"