# TODO List

## Queue Management
- [x] Add `!remove <user>` command to remove specific user from queue
- [x] Make username matching case-insensitive (e.g., @UserName should match "username" in queue)
- [x] Add `!pop` command to remove first position in queue
- [ ] Enhance `!pop` command to support removing multiple users (e.g., `!pop 2` removes first two users)
- [ ] Implement queue state persistence:
  - [ ] Add JSON serialization for queue state
  - [ ] Save state on queue changes (add/remove/move)
  - [ ] Implement periodic state saving
  - [ ] Add state loading on startup
  - [ ] Add state saving during shutdown
  - [ ] Add proper error handling and logging
- [ ] Add `!move <user> <position>` command to move user to specific position in queue
- [ ] Fix help command not showing queue commands when enabled

## UI/UX
- [ ] Improve message formatting for queue-related responses
- [ ] Add visual indicators for queue position changes
- [ ] Add confirmation messages for destructive actions

## System Features
- [ ] Add command cooldowns per user type
- [ ] Add configuration reload without restart
- [ ] Add command usage statistics
- [ ] Add automatic reconnection on disconnect
- [ ] Add error reporting to logging service
- [ ] Implement OAuth token refresh:
  - [ ] Add refresh token handling
  - [ ] Auto-refresh token before expiration
  - [ ] Handle refresh token errors gracefully
  - [ ] Add token refresh logging

## Permission System
- [ ] Add more granular permission levels beyond Mod/VIP
- [ ] Add custom command creation through chat
- [ ] Add command aliases configuration in bot.yaml

## Queue Analytics
- [ ] Add queue history tracking
- [ ] Add queue size limits
- [ ] Add timeout/expiry for queue entries
- [ ] Add queue usage statistics

## Documentation
- [ ] Add setup guide for new users
- [ ] Document all available commands
- [ ] Add configuration file documentation
- [ ] Add development guide for contributors 