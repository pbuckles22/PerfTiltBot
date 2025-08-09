# Recent Changes Test Checklist

## Overview
This checklist focuses on testing the recent improvements made to the PerfTiltBot queue system.

## Recent Changes to Test

### 1. Enhanced !remove Command (Multiple Arguments)

#### Basic Multiple Removal
- [ ] **Multiple Positions**: `!remove 4 5` - Remove users at positions 4 and 5
- [ ] **Multiple Usernames**: `!remove user1 user2` - Remove multiple users by name
- [ ] **Mixed Arguments**: `!remove 3 user1 7` - Mix positions and usernames
- [ ] **Single Argument**: `!remove 3` - Still works as before (backward compatibility)

#### Error Handling
- [ ] **Partial Success**: `!remove 1 nonexistentuser 3` - Some valid, some invalid
- [ ] **All Invalid**: `!remove 999 nonexistentuser` - All arguments invalid
- [ ] **Invalid Position**: `!remove 999` - Position out of range
- [ ] **Non-existent User**: `!remove nonexistentuser` - User not in queue

#### Response Format
- [ ] **Success Response**: Verify "Removed X user(s): user1 removed; user2 removed" format
- [ ] **Error Response**: Verify "Invalid position 999 (queue has 5); nonexistentuser not in queue" format
- [ ] **Mixed Response**: Verify partial success shows both successes and failures

### 2. Queue Display Defensive Check

#### Empty Username Detection
- [ ] **Normal Queue**: `!q` with normal usernames - Should display normally
- [ ] **Empty Username Bug**: If queue contains empty username, should show warning
- [ ] **Warning Format**: Verify "[Warning] Queue data error: empty username at position X" message
- [ ] **Position Accuracy**: Verify warning shows correct position number

### 3. Pause Queue Enforcement

#### Join While Paused
- [ ] **Pause Queue**: `!pausequeue` - Verify queue pauses
- [ ] **Regular User Join**: Regular user tries `!join` while paused - Should be blocked
- [ ] **Mod User Join**: Mod tries `!join` while paused - Should be allowed
- [ ] **VIP User Join**: VIP tries `!join` while paused - Should be blocked (not privileged)
- [ ] **Broadcaster Join**: Broadcaster tries `!join` while paused - Should be allowed

#### Error Messages
- [ ] **Regular User Error**: Verify "queue system is currently paused" message
- [ ] **Mod Success**: Verify mod can join and gets normal join message
- [ ] **Unpause Test**: `!unpausequeue` then regular user `!join` - Should work

### 4. Integration Tests

#### Command Interaction
- [ ] **Remove Then Display**: `!remove 1 2` then `!q` - Verify queue updates correctly
- [ ] **Pause Then Remove**: Pause queue, mod removes users, then display - Should work
- [ ] **Multiple Operations**: Complex sequence of joins, removes, pauses, displays

#### Edge Cases
- [ ] **Empty Queue Remove**: `!remove 1` on empty queue - Should show error
- [ ] **Single User Remove**: `!remove 1` with only one user - Should work
- [ ] **Remove All**: `!remove 1 2 3` with 3 users - Should clear queue
- [ ] **Remove More Than Available**: `!remove 1 2 3 4 5` with 3 users - Should remove 3, error on 4,5

## Test Scenarios

### Scenario 1: Multiple Removal Workflow
1. [ ] Add 5 users to queue: `!join user1`, `!join user2`, etc.
2. [ ] Display queue: `!q` - Verify all 5 users shown
3. [ ] Remove positions 2 and 4: `!remove 2 4`
4. [ ] Display queue: `!q` - Verify users 1, 3, 5 remain
5. [ ] Remove remaining users: `!remove 1 2 3`
6. [ ] Display queue: `!q` - Verify empty queue

### Scenario 2: Pause Enforcement Workflow
1. [ ] Add 3 users to queue
2. [ ] Pause queue: `!pausequeue`
3. [ ] Regular user tries to join - Should be blocked
4. [ ] Mod joins successfully - Should work
5. [ ] Display queue: `!q` - Verify 4 users (3 original + 1 mod)
6. [ ] Unpause queue: `!unpausequeue`
7. [ ] Regular user joins - Should work now
8. [ ] Display queue: `!q` - Verify 5 users

### Scenario 3: Error Handling Workflow
1. [ ] Add 3 users to queue
2. [ ] Try invalid removals: `!remove 999 nonexistentuser 2`
3. [ ] Verify response: "Invalid position 999 (queue has 3); nonexistentuser not in queue; user2 removed"
4. [ ] Display queue: `!q` - Verify only 2 users remain
5. [ ] Try removing all: `!remove 1 2 3 4`
6. [ ] Verify response: "Removed 2 user(s): user1 removed; user3 removed; Invalid position 3 (queue has 0); Invalid position 4 (queue has 0)"

## Success Criteria

### !remove Multiple Arguments
- [ ] Accepts multiple arguments (positions and/or usernames)
- [ ] Processes each argument independently
- [ ] Provides clear feedback for each operation
- [ ] Maintains backward compatibility with single argument
- [ ] Handles errors gracefully without stopping execution

### Queue Display
- [ ] Shows all users correctly in numbered format
- [ ] Detects and reports empty username errors
- [ ] Maintains correct position numbering
- [ ] Updates immediately after queue modifications

### Pause Enforcement
- [ ] Blocks regular users from joining when paused
- [ ] Allows mods and broadcasters to join when paused
- [ ] Provides clear error messages for blocked users
- [ ] Resumes normal operation after unpausing

## Bug Reporting Template

For any issues found:

```
**Issue**: [Brief description]
**Steps to Reproduce**:
1. [Step 1]
2. [Step 2]
3. [Step 3]

**Expected Behavior**: [What should happen]
**Actual Behavior**: [What actually happened]
**Environment**: [Bot version, channel, user privileges]
**Additional Notes**: [Any other relevant information]
```

## Quick Commands Reference

```bash
# Test multiple removal
!remove 1 2 3
!remove user1 user2
!remove 1 user2 3

# Test pause enforcement
!pausequeue
!join  # (as regular user - should fail)
!join  # (as mod - should work)
!unpausequeue

# Test queue display
!q
!queue

# Test error handling
!remove 999
!remove nonexistentuser
!remove 1 999 user1
``` 