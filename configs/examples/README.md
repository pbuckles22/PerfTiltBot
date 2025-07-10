# Example Configuration Files

This directory contains example configuration files for setting up PBChatBot.

## File Naming Conventions

1. **Bot Authentication Files**
   - Format: `<bot_name>_auth_secrets.yaml`
   - Example: `mybot_auth_secrets.yaml`

2. **Channel Configuration Files**
   - Format: `<channel_name>_config_secrets.yaml`
   - Example: `channel1_config_secrets.yaml`

## Setup Examples

### Single Bot, Multiple Channels

If you want to use one bot for multiple channels:

1. Create bot authentication file:
   ```yaml
   # configs/bots/mychatbot_auth_secrets.yaml
   bot_name: "mychatbot"
   oauth: "oauth:your_bot_oauth_token"
   client_id: "your_bot_client_id"
   client_secret: "your_bot_client_secret"
   refresh_token: "your_refresh_token"
   ```

2. Create channel configurations:
   ```yaml
   # configs/channels/channel1_config_secrets.yaml
   bot_name: "mychatbot"  # Must match the bot_name in auth file
   channel: "channel1"
   data_path: "/app/data"
   timezone: "America/New_York"  # Timezone for user-facing messages (optional, defaults to EST)
   
   commands:
     queue:
       max_size: 100
       default_position: 1
       default_pop_count: 1
     cooldowns:
       default: 5
       moderator: 2
       vip: 3
   ```

   ```yaml
   # configs/channels/channel2_config_secrets.yaml
   bot_name: "mychatbot"  # Same bot, different channel
   channel: "channel2"
   data_path: "/app/data"
   timezone: "America/Los_Angeles"  # Timezone for user-facing messages (optional, defaults to EST)
   
   commands:
     queue:
       max_size: 50
       default_position: 1
       default_pop_count: 1
     cooldowns:
       default: 10
       moderator: 3
       vip: 5
   ```

### Multiple Bots for Different Channels

If you want to use different bots for different channels:

1. Create bot authentication files:
   ```yaml
   # configs/bots/bot1_auth_secrets.yaml
   bot_name: "bot1"
   oauth: "oauth:bot1_oauth_token"
   client_id: "bot1_client_id"
   client_secret: "bot1_client_secret"
   refresh_token: "bot1_refresh_token"
   ```

   ```yaml
   # configs/bots/bot2_auth_secrets.yaml
   bot_name: "bot2"
   oauth: "oauth:bot2_oauth_token"
   client_id: "bot2_client_id"
   client_secret: "bot2_client_secret"
   refresh_token: "bot2_refresh_token"
   ```

2. Create channel configurations:
   ```yaml
   # configs/channels/channel1_config_secrets.yaml
   bot_name: "bot1"  # Uses bot1
   channel: "channel1"
   data_path: "/app/data"
   
   commands:
     queue:
       max_size: 100
       default_position: 1
       default_pop_count: 1
     cooldowns:
       default: 5
       moderator: 2
       vip: 3
   ```

   ```yaml
   # configs/channels/channel2_config_secrets.yaml
   bot_name: "bot2"  # Uses bot2
   channel: "channel2"
   data_path: "/app/data"
   timezone: "America/Los_Angeles"  # Timezone for user-facing messages (optional, defaults to EST)
   
   commands:
     queue:
       max_size: 50
       default_position: 1
       default_pop_count: 1
     cooldowns:
       default: 10
       moderator: 3
       vip: 5
   ```

## Important Notes

1. The command prefix is hardcoded to `!` in the bot
2. Each channel gets its own data volume for persistent storage
3. Cooldowns are channel-specific and can be customized per user type
4. Queue settings are channel-specific and can be customized
5. The bot name in the channel config must match the bot name in the auth file
6. **Timezone Configuration**: 
   - Debug logs are always in PST (America/Los_Angeles) for consistency
   - User-facing messages use the configured timezone (defaults to EST if not specified)
   - Common timezone options: `America/New_York` (EST/EDT), `America/Los_Angeles` (PST/PDT), `UTC`

## Security Notes

1. Never commit `*_auth_secrets.yaml` or `*_config_secrets.yaml` files to version control
2. Keep your bot's OAuth token and client credentials secure
3. Use different bot configurations for different environments (development, production)
4. Regularly rotate OAuth tokens and client credentials 