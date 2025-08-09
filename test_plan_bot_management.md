# PerfTiltBot Management Scripts Test Plan

## Overview
This test plan covers the PowerShell management scripts for the PerfTiltBot, including Docker image building, container management, and multi-channel bot deployment.

## Test Environment Setup
- **OS**: Windows 10/11 with PowerShell 7.0+
- **Docker**: Docker Desktop running
- **Bot**: PerfTiltBot source code in development directory
- **Channels**: Multiple test channels with configuration files

## Test Categories

### 1. PowerShell Script Management

#### 1.1 Basic Command Validation
- [ ] **No Arguments**: `.\run_bot.ps1` - Display usage instructions
- [ ] **Invalid Command**: `.\run_bot.ps1 invalid` - Display error and usage
- [ ] **Missing Channel**: `.\run_bot.ps1 start` - Error about missing channel
- [ ] **Missing Bot Name**: `.\run_bot.ps1 list-channels` - Error about missing bot name
- [ ] **Shortcut Syntax**: `.\run_bot.ps1 pbuckles` - Same as `.\run_bot.ps1 start pbuckles`

#### 1.2 Build Command
- [ ] **Build Image**: `.\run_bot.ps1 build` - Successfully build Docker image
- [ ] **Build Verification**: `docker images` shows `pbchatbot` image
- [ ] **Rebuild Image**: Build with existing image - Should rebuild without errors
- [ ] **Build Error Handling**: Test with Docker not running
- [ ] **Build with Dependencies**: Verify all Go dependencies are included

#### 1.3 Start Command
- [ ] **Start Single Bot**: `.\run_bot.ps1 start pbuckles` - Start bot for channel
- [ ] **Container Creation**: Verify container is created with correct name
- [ ] **Volume Mounting**: Verify secrets and config files are mounted correctly
- [ ] **Data Volume**: Verify channel-specific data volume is created
- [ ] **Environment Variables**: Verify CHANNEL_NAME and BOT_NAME are set
- [ ] **Existing Container**: Start when container already exists - Should stop and recreate
- [ ] **Missing Config**: Start with missing config file - Should show error
- [ ] **Missing Auth**: Start with missing bot auth file - Should show error

#### 1.4 Stop Commands
- [ ] **Stop Channel**: `.\run_bot.ps1 stop-channel pbuckles` - Stop specific channel
- [ ] **Stop All**: `.\run_bot.ps1 stop-all` - Stop all running bots
- [ ] **Stop Non-existent**: Stop channel that's not running - Should show message
- [ ] **Container Cleanup**: Verify containers are removed after stopping

#### 1.5 List Commands
- [ ] **List Running**: `.\run_bot.ps1 list` - Show all running bot instances
- [ ] **List Empty**: List when no bots are running - Should show "No running instances"
- [ ] **List Multiple**: List with multiple bots running - Should show all
- [ ] **List Channels by Bot**: `.\run_bot.ps1 list-channels BotWithTwoToes` - Show channels using specific bot

#### 1.6 Configuration Management
- [ ] **Update Bot Config**: `.\run_bot.ps1 update-bot BotWithTwoToes` - Edit bot configuration
- [ ] **Config Backup**: Verify backup is created before editing
- [ ] **Config Validation**: Test with invalid configuration - Should reject changes
- [ ] **Affected Channels**: Show which channels are affected by bot config changes

#### 1.7 Restart Commands
- [ ] **Restart All**: `.\run_bot.ps1 restart-all` - Stop and restart all bots
- [ ] **Restart Order**: Verify bots are restarted in sequence
- [ ] **Restart Timing**: Verify appropriate delays between restarts

### 2. Multi-Channel Bot Management

#### 2.1 Channel Configuration
- [ ] **Config File Structure**: Verify `*_config_secrets.yaml` format
- [ ] **Bot Reference**: Verify `bot_name` field references valid bot auth file
- [ ] **Channel Name**: Verify `channel` field matches Twitch channel name
- [ ] **Data Path**: Verify channel-specific data path configuration

#### 2.2 Bot Authentication
- [ ] **Auth File Structure**: Verify `*_auth_secrets.yaml` format
- [ ] **Required Fields**: Verify all required auth fields are present
- [ ] **Field Validation**: Test with missing required fields
- [ ] **Bot-Channel Mapping**: Verify bot auth files are correctly referenced

#### 2.3 Container Isolation
- [ ] **Separate Containers**: Verify each channel gets its own container
- [ ] **Container Names**: Verify naming convention `{BotName}-{ChannelName}`
- [ ] **Volume Isolation**: Verify each channel has separate data volume
- [ ] **Config Isolation**: Verify channel-specific config files are mounted

### 3. Docker Integration

#### 3.1 Image Management
- [ ] **Image Building**: Verify Docker image builds successfully
- [ ] **Image Size**: Monitor image size for optimization
- [ ] **Layer Caching**: Verify efficient use of Docker layer caching
- [ ] **Multi-stage Build**: Verify build process uses multi-stage approach

#### 3.2 Container Management
- [ ] **Container Startup**: Verify containers start successfully
- [ ] **Container Logs**: Verify logs are accessible via `docker logs`
- [ ] **Container Health**: Verify containers remain healthy during operation
- [ ] **Resource Usage**: Monitor CPU and memory usage

#### 3.3 Volume Management
- [ ] **Config Mounting**: Verify config files are mounted read-only
- [ ] **Data Persistence**: Verify data volumes persist across restarts
- [ ] **Volume Naming**: Verify volume naming convention
- [ ] **Volume Cleanup**: Verify volumes are cleaned up when containers are removed

### 4. Error Handling and Edge Cases

#### 4.1 File System Errors
- [ ] **Missing Config Files**: Test with non-existent config files
- [ ] **Corrupted Config Files**: Test with invalid YAML files
- [ ] **Permission Errors**: Test with insufficient file permissions
- [ ] **Disk Space**: Test with insufficient disk space

#### 4.2 Docker Errors
- [ ] **Docker Not Running**: Test when Docker Desktop is not running
- [ ] **Docker Permissions**: Test with insufficient Docker permissions
- [ ] **Network Issues**: Test with Docker network connectivity problems
- [ ] **Image Pull Errors**: Test when base images cannot be pulled

#### 4.3 Configuration Errors
- [ ] **Invalid YAML**: Test with malformed configuration files
- [ ] **Missing Fields**: Test with missing required configuration fields
- [ ] **Invalid Values**: Test with invalid configuration values
- [ ] **Channel Mismatch**: Test with mismatched channel names

### 5. Performance and Scalability

#### 5.1 Multiple Channels
- [ ] **Start Multiple Bots**: Test starting 5+ channels simultaneously
- [ ] **Resource Usage**: Monitor system resources with multiple bots
- [ ] **Startup Time**: Measure time to start multiple bots
- [ ] **Memory Usage**: Monitor memory usage per bot instance

#### 5.2 Large Scale Operations
- [ ] **Many Channels**: Test with 10+ channel configurations
- [ ] **Restart All**: Test restarting many bots simultaneously
- [ ] **List Performance**: Test listing many running bots
- [ ] **Stop All Performance**: Test stopping many bots simultaneously

### 6. Security and Permissions

#### 6.1 File Permissions
- [ ] **Config File Permissions**: Verify appropriate file permissions
- [ ] **Auth File Security**: Verify auth files are not world-readable
- [ ] **Volume Permissions**: Verify data volume permissions
- [ ] **Script Permissions**: Verify PowerShell script execution policy

#### 6.2 Container Security
- [ ] **Non-root User**: Verify containers run as non-root user
- [ ] **Read-only Mounts**: Verify config files are mounted read-only
- [ ] **Network Isolation**: Verify containers have appropriate network access
- [ ] **Resource Limits**: Verify containers have resource limits set

### 7. Integration Testing

#### 7.1 Full Lifecycle Test
1. [ ] Build Docker image
2. [ ] Start bot for channel A
3. [ ] Start bot for channel B
4. [ ] List running bots
5. [ ] Update bot configuration
6. [ ] Restart all bots
7. [ ] Stop channel A
8. [ ] List running bots
9. [ ] Stop all bots
10. [ ] Verify cleanup

#### 7.2 Multi-Channel Test
1. [ ] Create multiple channel configurations
2. [ ] Start bots for all channels
3. [ ] Verify all bots are running
4. [ ] Test channel isolation
5. [ ] Restart all bots
6. [ ] Verify all bots restart successfully
7. [ ] Stop all bots
8. [ ] Verify complete cleanup

### 8. Cross-Platform Compatibility

#### 8.1 PowerShell Compatibility
- [ ] **PowerShell 5.1**: Test with Windows PowerShell 5.1
- [ ] **PowerShell 7**: Test with PowerShell 7.0+
- [ ] **PowerShell Core**: Test with PowerShell Core on Linux/macOS
- [ ] **Execution Policy**: Test with different execution policies

#### 8.2 Docker Compatibility
- [ ] **Docker Desktop**: Test with Docker Desktop for Windows
- [ ] **Docker Engine**: Test with Docker Engine on Linux
- [ ] **Docker Compose**: Test with Docker Compose integration
- [ ] **Container Runtime**: Test with different container runtimes

## Test Execution

### Prerequisites
1. Docker Desktop is running
2. PowerShell 7.0+ is installed
3. Bot source code is available
4. Test channel configurations are created
5. Sufficient disk space for multiple containers

### Test Data Setup
```powershell
# Create test channel configurations
Copy-Item configs/channels/pbuckles_config_secrets.yaml configs/channels/test1_config_secrets.yaml
Copy-Item configs/channels/pbuckles_config_secrets.yaml configs/channels/test2_config_secrets.yaml

# Create test bot configurations
Copy-Item configs/bots/BotWithTwoToes_auth_secrets.yaml configs/bots/TestBot_auth_secrets.yaml
```

### Test Execution Order
1. Basic functionality tests
2. Error handling tests
3. Multi-channel tests
4. Performance tests
5. Security tests
6. Integration tests

## Success Criteria
- All management commands work as expected
- Docker containers start and stop correctly
- Multi-channel deployment works properly
- Error handling is appropriate and informative
- Performance remains acceptable with multiple channels
- Security requirements are met
- Cross-platform compatibility is maintained

## Bug Reporting
For any issues found during testing:
1. Document the exact PowerShell command used
2. Record the expected vs actual behavior
3. Include Docker logs and container information
4. Note the test environment and conditions
5. Include any error messages or stack traces

## Maintenance
- Update test plan as new features are added
- Review and update test cases based on bug reports
- Maintain test configurations and data
- Regular regression testing after updates
- Monitor Docker and PowerShell version compatibility 