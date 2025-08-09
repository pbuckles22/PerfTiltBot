# PerfTiltBot Queue System Test Plan

## Overview
This test plan covers the queue management system for the PerfTiltBot, including all commands, edge cases, and recent improvements.

## Current Status (2025-01-13)
**Multi-Channel Bot Development**: The bot has been enhanced with multi-channel support and improved token management. Recent changes include:
- Enhanced token refresh logic with safety checks
- Improved multi-channel connection handling
- Better error recovery and logging
- Channel-specific command and queue management
- Test harness updates for multi-channel scenarios

This test plan covers both single-channel and multi-channel operation scenarios.

## Test Environment Setup
- **Bot**: PerfTiltBot running on Twitch
- **Channel**: Test channel with moderator privileges
- **Test Users**: Multiple test accounts with different privilege levels (regular, VIP, mod, broadcaster)
- **Data Path**: `/app/data` (Docker container)

## Test Categories

### 1. Basic Queue Management

#### 1.1 Queue Lifecycle
- [ ] **Start Queue**: `!startqueue` - Verify queue system starts and is enabled
- [ ] **End Queue**: `!endqueue` - Verify queue system stops and clears all users
- [ ] **Enable/Disable**: `!enable` / `!disable` - Test manual enable/disable
- [ ] **Queue Status**: Verify queue state persists across bot restarts

#### 1.2 User Management
- [ ] **Join Queue**: `!join` - Regular user joins queue
- [ ] **Join with Username**: `!join username` - Add specific user to queue
- [ ] **Leave Queue**: `!leave` - User leaves queue
- [ ] **Leave with Username**: `!leave username` - Remove specific user (mod only)
- [ ] **Position Check**: `!position` - Show user's position in queue
- [ ] **Position by Username**: `!position username` - Show specific user's position
- [ ] **Position by Number**: `!position 3` - Show user at position 3

### 2. Queue Display and Information

#### 2.1 Queue Display (`!q` / `!queue`)
- [ ] **Empty Queue**: Verify "The queue is currently empty" message
- [ ] **Single User**: Verify "1) username" format
- [ ] **Multiple Users**: Verify "1) user1 2) user2 3) user3" format
- [ ] **Large Queue**: Test with 10+ users to verify display limits
- [ ] **Username Case**: Verify exact username case is preserved
- [ ] **Empty Username Bug**: Test defensive check for empty usernames
- [ ] **Position 1 Bug**: Verify first user always shows correctly

#### 2.2 Queue Statistics
- [ ] **Queue Size**: Verify accurate count in queue display
- [ ] **Position Accuracy**: Verify positions are 1-based and sequential
- [ ] **Real-time Updates**: Verify queue updates immediately after changes

### 3. Queue Control Commands

#### 3.1 Pause/Unpause System
- [ ] **Pause Queue**: `!pausequeue` - Verify queue pauses
- [ ] **Pause Message**: Verify "Queue is now paused" message
- [ ] **Join While Paused**: 
  - [ ] Regular user: Should be blocked with "queue system is currently paused"
  - [ ] Mod user: Should be allowed to join
  - [ ] VIP user: Should be blocked (not privileged for pause bypass)
- [ ] **Unpause Queue**: `!unpausequeue` - Verify queue resumes
- [ ] **Unpause Message**: Verify "Queue is now open again" message
- [ ] **Join After Unpause**: Verify regular users can join again

#### 3.2 Queue Manipulation
- [ ] **Pop Single**: `!pop` - Remove first user from queue
- [ ] **Pop Multiple**: `!pop 3` - Remove first 3 users
- [ ] **Pop More Than Available**: `!pop 10` with 5 users in queue
- [ ] **Clear Queue**: `!clear` - Remove all users
- [ ] **Clear Count**: Verify correct count of removed users

### 4. Enhanced Remove Command

#### 4.1 Single Removal
- [ ] **Remove by Username**: `!remove username` - Remove specific user
- [ ] **Remove by Position**: `!remove 3` - Remove user at position 3
- [ ] **Remove Invalid Position**: `!remove 999` - Error handling
- [ ] **Remove Non-existent User**: `!remove nonexistentuser` - Error handling

#### 4.2 Multiple Removal (NEW FEATURE)
- [ ] **Multiple Positions**: `!remove 4 5` - Remove users at positions 4 and 5
- [ ] **Multiple Usernames**: `!remove user1 user2` - Remove multiple users by name
- [ ] **Mixed Arguments**: `!remove 3 user1 7` - Mix positions and usernames
- [ ] **Partial Success**: `!remove 1 nonexistentuser 3` - Some valid, some invalid
- [ ] **All Invalid**: `!remove 999 nonexistentuser` - All arguments invalid
- [ ] **Response Format**: Verify "Removed X user(s): user1 removed; user2 removed" format

### 5. User Movement and Positioning

#### 5.1 Move Command
- [ ] **Move by Username**: `!move username 5` - Move user to position 5
- [ ] **Move by Position**: `!move 3 1` - Move user at position 3 to position 1
- [ ] **Move to Same Position**: `!move username 3` when user is at position 3
- [ ] **Move Invalid Position**: `!move username 999` - Error handling
- [ ] **Move Non-existent User**: `!move nonexistentuser 1` - Error handling

### 6. Privilege and Permission Testing

#### 6.1 User Privilege Levels
- [ ] **Regular User**: Test all commands with regular user
- [ ] **VIP User**: Test commands with VIP privileges
- [ ] **Moderator**: Test commands with moderator privileges
- [ ] **Broadcaster**: Test commands with broadcaster privileges

#### 6.2 Privilege-Specific Commands
- [ ] **Mod-Only Commands**: Verify only mods can use certain commands
- [ ] **Privileged Commands**: Verify VIPs and mods can use privileged commands
- [ ] **Pause Bypass**: Verify only mods can join when queue is paused

### 7. Cooldown System

#### 7.1 Cooldown Enforcement
- [ ] **Regular User Cooldown**: Test 30-second cooldown on commands
- [ ] **VIP User Cooldown**: Test 15-second cooldown on commands
- [ ] **Moderator Cooldown**: Test 5-second cooldown on commands
- [ ] **Broadcaster Cooldown**: Test no cooldown for broadcaster
- [ ] **Cooldown Messages**: Verify appropriate cooldown messages
- [ ] **Cooldown Persistence**: Verify cooldowns persist across commands

### 8. State Persistence and Recovery

#### 8.1 Save/Load System
- [ ] **Save Queue**: `!savequeue` - Save current queue state
- [ ] **Load Queue**: `!restorequeue` - Load saved queue state
- [ ] **Auto-Save**: Verify queue auto-saves after modifications
- [ ] **Crash Recovery**: `!restoreauto` - Test auto-recovery from crashes
- [ ] **State Persistence**: Verify queue state survives bot restarts

#### 8.2 Data Integrity
- [ ] **File Corruption**: Test behavior with corrupted state files
- [ ] **Missing Files**: Test behavior when state files are missing
- [ ] **Channel Mismatch**: Test behavior with wrong channel data

### 9. Error Handling and Edge Cases

#### 9.1 Invalid Inputs
- [ ] **Empty Commands**: `!join` with no arguments
- [ ] **Invalid Numbers**: `!pop abc` - Non-numeric arguments
- [ ] **Negative Numbers**: `!pop -1` - Negative values
- [ ] **Zero Values**: `!pop 0` - Zero values
- [ ] **Very Large Numbers**: `!pop 999999` - Extremely large values

#### 9.2 Race Conditions
- [ ] **Concurrent Joins**: Multiple users joining simultaneously
- [ ] **Concurrent Removes**: Multiple mods removing users simultaneously
- [ ] **Queue Modifications During Display**: Modify queue while displaying

#### 9.3 Edge Cases
- [ ] **Empty Queue Operations**: Try to pop/remove from empty queue
- [ ] **Single User Queue**: Operations with only one user in queue
- [ ] **Large Queue**: Operations with 100+ users in queue
- [ ] **Special Characters**: Usernames with special characters
- [ ] **Unicode Usernames**: Usernames with non-ASCII characters

### 10. Performance Testing

#### 10.1 Load Testing
- [ ] **Rapid Commands**: Send commands rapidly to test response time
- [ ] **Large Queue**: Test with maximum queue size (100 users)
- [ ] **Memory Usage**: Monitor memory usage with large queues
- [ ] **Response Time**: Measure response time for various commands

### 11. Integration Testing

#### 11.1 Bot Integration
- [ ] **Command Registration**: Verify all commands are properly registered
- [ ] **Help System**: `!help` - Verify all commands are listed
- [ ] **Ping Command**: `!ping` - Verify bot responsiveness
- [ ] **Uptime Command**: `!uptime` - Verify bot uptime tracking

#### 11.2 Multi-Channel Support
- [ ] **Channel Isolation**: Verify queues are separate per channel
- [ ] **Channel-Specific Data**: Verify data files are channel-specific
- [ ] **Cross-Channel Interference**: Ensure no cross-channel data leakage

## Test Execution

### Prerequisites
1. Bot is running and connected to Twitch
2. Test channel is configured
3. Multiple test accounts are available
4. Moderator privileges are granted to test accounts

### Test Execution Order
1. Basic functionality tests
2. Edge case testing
3. Performance testing
4. Integration testing
5. Regression testing

### Test Data
- **Test Users**: Create 10+ test accounts with various privilege levels
- **Test Scenarios**: Document specific test scenarios and expected outcomes
- **Test Results**: Record all test results, including failures and unexpected behavior

## Success Criteria
- All commands function as expected
- No data corruption or loss
- Proper error handling for all edge cases
- Performance remains acceptable under load
- Queue state persists correctly across restarts

## Bug Reporting
For any issues found during testing:
1. Document the exact steps to reproduce
2. Record the expected vs actual behavior
3. Include relevant log information
4. Note the test environment and conditions

## Maintenance
- Update test plan as new features are added
- Review and update test cases based on bug reports
- Maintain test data and scenarios
- Regular regression testing after updates 