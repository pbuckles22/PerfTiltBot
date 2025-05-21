# Example Configurations

This directory contains example configuration files demonstrating how to set up bot and channel configurations.

## File Naming Convention

- Bot Authentication Files: `<bot_name>_auth_secrets.yaml`
  - Contains ONLY the bot's authentication credentials and API keys
  - Example: `mybot_auth_secrets.yaml`

- Channel Configuration Files: `<channel_name>_config_secrets.yaml`
  - Contains channel-specific settings and references which bot to use
  - Example: `mychannel_config_secrets.yaml`

## Single Bot, Multiple Channels

This setup shows how to use one bot for multiple channels:

1. **Bot Authentication** (`mybot_auth_secrets.yaml`):
   ```yaml
   bot_name: "mybot"
   oauth: "oauth:your_oauth_token"
   client_id: "your_client_id"
   client_secret: "your_client_secret"
   apis:
     twitch:
       client_id: "your_twitch_client_id"
       client_secret: "your_twitch_client_secret"
     openai:
       api_key: "your_openai_api_key"
     riot:
       api_key: "your_riot_api_key"
   ```

2. **Channel Configurations**:
   - `channel1_config_secrets.yaml`: Uses MyBot
     ```yaml
     bot_name: "mybot"  # References which bot's auth file to use
     channel: "channel1"  # The Twitch channel name
     data_path: "/app/data/channel1"  # Channel-specific data path
     commands:
       - name: "!command1"
         description: "Channel specific command"
       - name: "!command2"
         description: "Another channel command"
     ```
   - `channel2_config_secrets.yaml`: Also uses MyBot
     ```yaml
     bot_name: "mybot"
     channel: "channel2"
     data_path: "/app/data/channel2"
     commands:
       - name: "!custom1"
         description: "Different command for channel2"
     ```

To use this setup:
```bash
# Start bot for channel1
.\run_bot.ps1 start channel1

# Start bot for channel2
.\run_bot.ps1 start channel2

# List all channels using MyBot
.\run_bot.ps1 list-channels mybot
```

## Multiple Bots

This setup shows how to use different bots for different channels:

1. **Bot Authentication Files**:
   - `mybot_auth_secrets.yaml`: First bot's credentials
     ```yaml
     bot_name: "mybot"
     oauth: "oauth:your_oauth_token"
     client_id: "your_client_id"
     client_secret: "your_client_secret"
     apis:
       twitch:
         client_id: "your_twitch_client_id"
         client_secret: "your_twitch_client_secret"
       openai:
         api_key: "your_openai_api_key"
       riot:
         api_key: "your_riot_api_key"
     ```
   - `otherbot_auth_secrets.yaml`: Second bot's credentials
     ```yaml
     bot_name: "otherbot"
     oauth: "oauth:another_oauth_token"
     client_id: "another_client_id"
     client_secret: "another_client_secret"
     apis:
       twitch:
         client_id: "another_twitch_client_id"
         client_secret: "another_twitch_client_secret"
       openai:
         api_key: "another_openai_api_key"
     ```

2. **Channel Configurations**:
   - `channel1_config_secrets.yaml`: Uses MyBot
     ```yaml
     bot_name: "mybot"
     channel: "channel1"
     data_path: "/app/data/channel1"
     commands:
       - name: "!command1"
         description: "Channel specific command"
     ```
   - `channel3_config_secrets.yaml`: Uses OtherBot
     ```yaml
     bot_name: "otherbot"
     channel: "channel3"
     data_path: "/app/data/channel3"
     commands:
       - name: "!other1"
         description: "Command for channel3"
     ```

To use this setup:
```bash
# Start MyBot for channel1
.\run_bot.ps1 start channel1

# Start OtherBot for channel3
.\run_bot.ps1 start channel3

# List all channels using each bot
.\run_bot.ps1 list-channels mybot
.\run_bot.ps1 list-channels otherbot
```

## Configuration Structure

Each configuration file follows this pattern:

1. **Bot Authentication** (`<bot_name>_auth_secrets.yaml`):
   - Contains ONLY authentication credentials and API keys
   - No channel-specific settings
   ```yaml
   bot_name: "bot_name"
   oauth: "oauth:your_oauth_token"
   client_id: "your_client_id"
   client_secret: "your_client_secret"
   apis:
     twitch:
       client_id: "your_twitch_client_id"
       client_secret: "your_twitch_client_secret"
     openai:
       api_key: "your_openai_api_key"
     riot:
       api_key: "your_riot_api_key"
   ```

2. **Channel Configuration** (`<channel_name>_config_secrets.yaml`):
   - References which bot to use
   - Contains all channel-specific settings
   ```yaml
   bot_name: "bot_name"  # References which bot's auth file to use
   channel: "channel_name"  # The Twitch channel name
   data_path: "/app/data/channel_name"  # Channel-specific data path
   commands:
     - name: "!command"
       description: "Channel specific command"
   ```

## Security Notes

- These are example files. Replace the placeholder values with your actual credentials
- Never commit real credentials to version control
- Keep your bot's OAuth token and client credentials secure
- Use different bot configurations for different environments
- Store authentication files in a secure location
- Regularly rotate OAuth tokens and client credentials 