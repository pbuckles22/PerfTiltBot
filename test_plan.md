# PerfTiltBot Test Plans

## Bot Management Script Test Plan

### Prerequisites
- Docker Desktop running
- PowerShell 7.0 or higher
- At least two channel-specific secrets files (e.g., `pbuckles_secrets.yaml` and `test_channel_secrets.yaml`)

### Test Cases

#### 1. Basic Command Validation
- [ ] Run script without arguments
  - Expected: Display usage instructions
- [ ] Run with invalid command
  - Expected: Display error message and usage instructions
- [ ] Run with missing channel name for start command
  - Expected: Display error message about missing channel name
- [ ] Run with missing channel name for stop-channel command
  - Expected: Display error message about missing channel name

#### 2. Build Command
- [ ] Run `.\run_bot.ps1 build`
  - Expected: Successfully build Docker image
  - Verify: `docker images` shows perftiltbot image
- [ ] Run build with existing image
  - Expected: Rebuild image without errors

#### 3. Start Command
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

#### 4. Stop Channel Command
- [ ] Stop running bot for specific channel
  - Expected: Container stopped and removed
  - Verify: Container no longer in `docker ps`
- [ ] Stop non-existent channel
  - Expected: Display message about no running instance
- [ ] Stop channel that was already stopped
  - Expected: Display message about no running instance

#### 5. List Command
- [ ] List with no running bots
  - Expected: Display "No running bot instances found"
- [ ] List with one running bot
  - Expected: Display channel name and container name
- [ ] List with multiple running bots
  - Expected: Display all channels and containers
  - Verify: Correct channel names extracted from container names

#### 6. Stop All Command
- [ ] Stop all with no running bots
  - Expected: Display "No running bot instances found"
- [ ] Stop all with one running bot
  - Expected: Stop and remove container
- [ ] Stop all with multiple running bots
  - Expected: Stop and remove all containers
  - Verify: No perftiltbot containers running

#### 7. Error Handling
- [ ] Test with invalid Docker commands
  - Expected: Appropriate error messages
- [ ] Test with Docker not running
  - Expected: Clear error message about Docker not running
- [ ] Test with insufficient permissions
  - Expected: Appropriate permission error messages

#### 8. Integration Tests
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

## Docker Improvements Test Plan

### Prerequisites
- Docker Desktop running
- Go 1.21 or higher installed
- Git installed

### Test Cases

#### 1. Multi-stage Build
- [ ] Build image from scratch
  - Expected: Successfully build both stages
  - Verify: Final image size is smaller than single-stage build
  - Verify: No build tools in final image
- [ ] Build with cached layers
  - Expected: Faster build time
  - Verify: Correct use of layer caching

#### 2. Image Size and Layers
- [ ] Check final image size
  - Expected: Significantly smaller than previous version
  - Verify: No unnecessary files or tools included
- [ ] Inspect image layers
  - Expected: Minimal number of layers
  - Verify: Layers are properly ordered for caching

#### 3. Timezone Configuration
- [ ] Verify timezone setting
  - Expected: Container uses UTC timezone
  - Verify: Log timestamps are in UTC
- [ ] Test timezone persistence
  - Expected: Timezone setting survives container restart
  - Verify: No timezone-related errors in logs

#### 4. Directory Structure
- [ ] Verify directory creation
  - Expected: /app/configs and /app/data directories exist
  - Verify: Correct permissions on directories
- [ ] Test volume mounting
  - Expected: Secrets file mounts correctly
  - Expected: Data volume mounts correctly
  - Verify: No permission issues

#### 5. Application Behavior
- [ ] Test application startup
  - Expected: Application starts successfully
  - Verify: No missing dependency errors
- [ ] Test logging
  - Expected: Logs are properly formatted
  - Verify: Pacific Time timestamps in logs
- [ ] Test graceful shutdown
  - Expected: Application shuts down cleanly
  - Verify: No error messages during shutdown

#### 6. Security
- [ ] Check for sensitive files
  - Expected: No sensitive files in image
  - Verify: No credentials or secrets in image layers
- [ ] Verify file permissions
  - Expected: Appropriate permissions on all files
  - Verify: No world-writable files

#### 7. Performance
- [ ] Test container startup time
  - Expected: Fast startup time
  - Verify: No unnecessary delays
- [ ] Monitor resource usage
  - Expected: Low memory footprint
  - Expected: Minimal CPU usage when idle
  - Verify: No memory leaks

#### 8. Integration Tests
- [ ] Full deployment test
  1. Build image
  2. Run container with mounted volumes
  3. Verify application functionality
  4. Test logging and timezone
  5. Stop and restart container
  6. Verify data persistence
  - Expected: All steps complete successfully
  - Verify: Application works as expected

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
   
   # Remove existing images
   docker rmi perftiltbot
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
- Image size is optimized
- Timezone handling works correctly
- Security requirements are met

## Notes
- Keep Docker Desktop running during testing
- Ensure sufficient disk space for multiple containers
- Monitor system resources during multiple container tests
- Document any performance issues or resource constraints
- Test both PowerShell and shell scripts if applicable
- Verify cross-platform compatibility 