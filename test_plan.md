# PerfTiltBot Management Script Test Plan

## Prerequisites
- Docker Desktop running
- PowerShell 7.0 or higher
- At least two channel-specific secrets files (e.g., `pbuckles_secrets.yaml` and `test_channel_secrets.yaml`)

## Test Cases

### 1. Basic Command Validation
- [ ] Run script without arguments
  - Expected: Display usage instructions
- [ ] Run with invalid command
  - Expected: Display error message and usage instructions
- [ ] Run with missing channel name for start command
  - Expected: Display error message about missing channel name
- [ ] Run with missing channel name for stop-channel command
  - Expected: Display error message about missing channel name

### 2. Build Command
- [ ] Run `.\run_bot.ps1 build`
  - Expected: Successfully build Docker image
  - Verify: `docker images` shows perftiltbot image
- [ ] Run build with existing image
  - Expected: Rebuild image without errors

### 3. Start Command
- [ ] Start bot for channel with existing secrets file
  - Expected: Successfully start container
  - Verify: Container running with correct name
  - Verify: Correct secrets file mounted
  - Verify: Data volume created
- [ ] Start bot for channel without secrets file
  - Expected: Display error about missing secrets file
- [ ] Start bot for channel with existing running container
  - Expected: Stop and remove existing container
  - Expected: Start new container
- [ ] Start multiple bots for different channels
  - Expected: Each bot runs in separate container
  - Verify: Each has correct secrets file
  - Verify: Each has separate data volume

### 4. Stop Channel Command
- [ ] Stop running bot for specific channel
  - Expected: Container stopped and removed
  - Verify: Container no longer in `docker ps`
- [ ] Stop non-existent channel
  - Expected: Display message about no running instance
- [ ] Stop channel that was already stopped
  - Expected: Display message about no running instance

### 5. List Command
- [ ] List with no running bots
  - Expected: Display "No running bot instances found"
- [ ] List with one running bot
  - Expected: Display channel name and container name
- [ ] List with multiple running bots
  - Expected: Display all channels and containers
  - Verify: Correct channel names extracted from container names

### 6. Stop All Command
- [ ] Stop all with no running bots
  - Expected: Display "No running bot instances found"
- [ ] Stop all with one running bot
  - Expected: Stop and remove container
- [ ] Stop all with multiple running bots
  - Expected: Stop and remove all containers
  - Verify: No perftiltbot containers running

### 7. Error Handling
- [ ] Test with invalid Docker commands
  - Expected: Appropriate error messages
- [ ] Test with Docker not running
  - Expected: Clear error message about Docker not running
- [ ] Test with insufficient permissions
  - Expected: Appropriate permission error messages

### 8. Integration Tests
- [ ] Full lifecycle test
  1. Build image
  2. Start bot for channel A
  3. Start bot for channel B
  4. List running bots
  5. Stop channel A
  6. List running bots
  7. Stop all bots
  8. List running bots
  - Expected: All steps complete successfully
  - Verify: Correct state at each step

## Test Environment Setup
1. Create test secrets files:
   ```powershell
   # Create test channel secrets
   Copy-Item configs/pbuckles_secrets.yaml configs/test_channel_secrets.yaml
   ```

2. Clean up before testing:
   ```powershell
   # Stop and remove all existing containers
   .\run_bot.ps1 stop-all
   ```

## Test Execution
1. Run through each test case in sequence
2. Document any failures or unexpected behavior
3. For each failure:
   - Note the exact command used
   - Note the expected vs actual behavior
   - Note any error messages
   - Take screenshots if relevant

## Success Criteria
- All test cases pass
- No unexpected error messages
- Containers start and stop correctly
- Secrets files are mounted correctly
- Data volumes are created and managed correctly
- List command shows accurate information
- Stop commands work as expected

## Notes
- Keep Docker Desktop running during testing
- Ensure sufficient disk space for multiple containers
- Monitor system resources during multiple container tests
- Document any performance issues or resource constraints 