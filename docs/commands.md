# PerfTiltBot Commands Documentation

This document provides a comprehensive list of all available commands for the PerfTiltBot Twitch chat bot.

## Table of Contents
- [Command Overview](#command-overview)
- [Base Commands](#base-commands)
- [Queue Management Commands](#queue-management-commands)
- [Bot Control Commands](#bot-control-commands)
- [Authentication Commands](#authentication-commands)
- [Permissions](#permissions)
- [Cooldowns](#cooldowns)
- [Usage Examples](#usage-examples)

## Command Overview

All commands use the `!` prefix and are case-insensitive. The bot supports command aliases for convenience.

## Base Commands

These commands are available to all users and provide basic bot functionality.

### `!help`
**Description:** Shows the list of available commands  
**Usage:** `!help`  
**Permission:** Everyone  
**Cooldown:** None  
**Response:** Lists all available commands grouped by category

### `!ping`
**Description:** Check if the bot is alive  
**Usage:** `!ping`  
**Permission:** Everyone  
**Cooldown:** None  
**Response:** `Pong! 🏓`

### `!uptime`
**Aliases:** `!up`  
**Description:** Shows how long the bot has been running  
**Usage:** `!uptime`  
**Permission:** Everyone  
**Cooldown:** None  
**Response:** Displays bot uptime in hours, minutes, and seconds

## Queue Management Commands

### Basic Queue Commands

These commands are available to all users when the queue system is enabled.

#### `!join`
**Aliases:** `!j`  
**Description:** Join the queue  
**Usage:** 
- `!join` - Join the queue yourself
- `!join <username>` - Add a specific user (Moderators/VIPs only)
- `!join <user1> <user2> <user3>` - Add multiple users (Moderators/VIPs only)  
**Permission:** Everyone (self), Moderators/VIPs (others)  
**Cooldown:** 30s (Regular), 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms user has joined and shows their position

#### `!leave`
**Aliases:** `!l`  
**Description:** Leave the queue  
**Usage:** 
- `!leave` - Leave the queue yourself
- `!leave <username>` - Remove a specific user (Moderators/VIPs only)  
**Permission:** Everyone (self), Moderators/VIPs (others)  
**Cooldown:** 30s (Regular), 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms user has left the queue

#### `!queue`
**Aliases:** `!q`  
**Description:** Show the current queue  
**Usage:** `!queue`  
**Permission:** Everyone  
**Cooldown:** 30s (Regular), 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Lists all users in the queue with their positions

#### `!position`
**Aliases:** `!pos`  
**Description:** Show your position in the queue  
**Usage:** `!position`  
**Permission:** Everyone  
**Cooldown:** 30s (Regular), 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Shows your current position in the queue

### Queue Control Commands

These commands control the queue system state and are restricted to Moderators/VIPs.

#### `!startqueue`
**Aliases:** `!sq`  
**Description:** Start the queue system  
**Usage:** `!startqueue`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms queue system has been started

#### `!endqueue`
**Description:** End the queue system  
**Usage:** `!endqueue`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms queue system has been ended

#### `!enable`
**Aliases:** `!e`  
**Description:** Enable the queue system  
**Usage:** `!enable`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms queue system has been enabled

#### `!disable`
**Aliases:** `!d`  
**Description:** Disable the queue system  
**Usage:** `!disable`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms queue system has been disabled

#### `!pausequeue`
**Aliases:** `!pq`  
**Description:** Pause the queue system  
**Usage:** `!pausequeue`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms queue is paused and no new entries can be added

#### `!unpausequeue`
**Aliases:** `!uq`  
**Description:** Unpause the queue system  
**Usage:** `!unpausequeue`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms queue is now open again

### Queue Management Commands

These commands allow manipulation of users within the queue and are restricted to Moderators/VIPs.

#### `!pop`
**Aliases:** `!p`  
**Description:** Pop users from the queue  
**Usage:** 
- `!pop` - Pop 1 user (default)
- `!pop <number>` - Pop specified number of users  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Lists the users that were removed from the queue

#### `!move`
**Aliases:** `!m`, `!mv`  
**Description:** Move a user in the queue  
**Usage:** 
- `!move <username> <position>` - Move user to specific position
- `!move <position> <new_position>` - Move user at position to new position  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms user has been moved to the new position

#### `!remove`
**Aliases:** `!r`  
**Description:** Remove a user from the queue  
**Usage:** `!remove <username>`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms user has been removed from the queue

#### `!clearqueue`
**Aliases:** `!cq`  
**Description:** Clear all users from the queue  
**Usage:** `!clearqueue`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms queue has been cleared and shows number of users removed

#### `!clear`
**Aliases:** `!c`  
**Description:** Clear the queue  
**Usage:** `!clear`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms queue has been cleared

### Queue State Commands

These commands manage queue persistence and are restricted to Moderators/VIPs.

**Note:** The queue state is automatically saved after every modification (add, remove, move, pop, etc.), so manual saving is usually not necessary unless you want to create a backup.

#### `!savequeue`
**Aliases:** `!svq`  
**Description:** Manually save the queue state (creates a backup)  
**Usage:** `!savequeue`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms queue state has been saved

#### `!restorequeue`
**Aliases:** `!rq`  
**Description:** Load the queue state from the last auto-save or manual save  
**Usage:** `!restorequeue`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms queue state has been restored and shows number of users loaded

## Bot Control Commands

These commands control the bot itself and are restricted to Moderators/VIPs.

#### `!kill`
**Aliases:** `!k`  
**Description:** Shutdown the bot  
**Usage:** `!kill`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms bot shutdown has been initiated

#### `!restart`
**Aliases:** `!rs`  
**Description:** Restart the bot  
**Usage:** `!restart`  
**Permission:** Moderators/VIPs  
**Cooldown:** 15s (VIP), 5s (Mod), 0s (Broadcaster)  
**Response:** Confirms bot restart has been initiated

## Authentication Commands

These commands manage bot authentication and are restricted to the channel owner.

#### `!auth`
**Description:** Refreshes the bot's authentication token  
**Usage:** `!auth`  
**Permission:** Channel Owner only  
**Cooldown:** None  
**Response:** Confirms token has been refreshed successfully

## Permissions

The bot recognizes different user types with varying permission levels:

### User Types
- **Regular Users**: Basic queue access (join, leave, view)
- **VIPs**: Queue management, user manipulation, bot control
- **Moderators**: Same as VIPs
- **Broadcasters/Channel Owners**: All permissions plus authentication management

### Permission Matrix

| Command | Regular | VIP | Moderator | Broadcaster |
|---------|---------|-----|-----------|-------------|
| `!help`, `!ping`, `!uptime` | ✅ | ✅ | ✅ | ✅ |
| `!join` (self), `!leave` (self), `!queue`, `!position` | ✅ | ✅ | ✅ | ✅ |
| `!join` (others), `!leave` (others) | ❌ | ✅ | ✅ | ✅ |
| Queue control commands | ❌ | ✅ | ✅ | ✅ |
| Queue management commands | ❌ | ✅ | ✅ | ✅ |
| Bot control commands | ❌ | ✅ | ✅ | ✅ |
| `!auth` | ❌ | ❌ | ❌ | ✅ |

## Cooldowns

Commands have different cooldown periods based on user type to prevent spam:

| User Type | Cooldown Duration |
|-----------|-------------------|
| Regular Users | 30 seconds |
| VIPs | 15 seconds |
| Moderators | 5 seconds |
| Broadcasters | No cooldown |

**Note:** Cooldown messages are rate-limited to prevent spam. Users will only see a cooldown message once per cooldown period.

## Usage Examples

### For Regular Users

```bash
!help                    # See all available commands
!join                    # Join the queue
!queue                   # See who's in the queue
!position                # Check your position
!leave                   # Leave the queue
!ping                    # Test if bot is responsive
!uptime                  # See how long bot has been running
```

### For Moderators/VIPs

```bash
# Queue Management
!startqueue              # Start the queue system
!join username           # Add a specific user to queue
!move username 5         # Move user to position 5
!pop 3                   # Remove 3 users from front of queue
!remove username         # Remove a specific user
!clearqueue              # Clear all users from queue

# Queue Control
!pausequeue              # Pause queue (no new entries)
!unpausequeue            # Resume queue
!savequeue               # Save current queue state
!restorequeue            # Load saved queue state

# Bot Control
!kill                    # Shutdown the bot
!restart                 # Restart the bot
```

### For Channel Owner

```bash
!auth                    # Manually refresh authentication token
```

## Command Response Examples

### Successful Queue Join
```
@username has joined the queue at position 3! (Total in queue: 5)
```

### Queue Display
```
Current Queue (4): 1) UserA, 2) UserB, 3) UserC, 4) UserD
```

### Error Messages
```
Queue system is currently disabled.
@username is not in the queue!
This command can only be used by moderators.
@username, this command is on cooldown. Please wait 15s.
```

## Notes

- All commands are case-insensitive
- Commands use the `!` prefix
- The bot supports command aliases for convenience
- **Queue state is automatically saved after every modification** (add, remove, move, pop, etc.)
- Queue state can be manually saved with `!savequeue` and restored with `!restorequeue`
- Cooldowns are per-user and per-command
- The bot will automatically reconnect if the connection is lost
- Authentication tokens are automatically refreshed when needed
- **Timezone Configuration**: 
  - Debug logs are always in PST (America/Los_Angeles) for consistency
  - User-facing messages use the configured timezone (defaults to EST if not specified)
  - Configure timezone in channel config: `timezone: "America/New_York"` 