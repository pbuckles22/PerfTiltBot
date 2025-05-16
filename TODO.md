# TODO List

## Queue Management
- [ ] Add `!remove <user>` command to remove specific user from queue
- [ ] Add `!pop` command to remove first position in queue
- [ ] Implement queue state persistence (save/load queue state in case bot dies)
- [ ] Add `!move` command for repositioning users in queue
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