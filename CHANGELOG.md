# Changelog

All notable changes to PerfTiltBot will be documented in this file.

## [Unreleased] - 2025-01-13

### Work in Progress
- **Multi-Channel Bot Enhancement**: Ongoing improvements to the multi-channel bot implementation
  - Added `internal/twitch/token_utils.go` for centralized token refresh logic
  - Enhanced token refresh safety checks to prevent ticker panic issues
  - Improved multi-channel bot connection handling and error recovery
  - Updated test harness for multi-channel testing scenarios
  - Enhanced command handlers for better channel isolation

### Current Status
- Feature branch `feature/multi-channel-bot` contains active development
- Core multi-channel functionality is implemented and working
- Token refresh improvements have been added and tested
- Integration testing and documentation updates are pending
- Ready for final testing and merge preparation

## [2.0.0] - 2025-01-10

### Added
- **Auto-Save Queue State**: Queue is automatically saved after every modification (add, remove, move, pop, etc.)
- **Timezone Configuration**: 
  - New `timezone` field in channel configuration
  - Debug logs always in PST (America/Los_Angeles) for consistency
  - User-facing messages use configurable timezone (defaults to EST)
- **New Timezone Utilities**: `internal/utils/timezone.go` for handling timezone conversions
- **Enhanced Documentation**: Updated commands documentation with timezone and auto-save information

### Changed
- **Improved Token Refresh**: 
  - Fixed panic issues with non-positive ticker intervals
  - Added safety checks for negative/zero intervals
  - Simplified and cleaned up debug logging
  - Added blank lines between debug sections for better readability
- **Queue Persistence**: Queue state is now automatically persisted after every operation
- **Logging Format**: 
  - Token logs now show: `[Token] Valid (expires in 2h15m35s, next check in 33m54s)`
  - Auth logs now show: `[Auth] Refreshing token...` and `[Auth] Token refreshed successfully`
  - Removed verbose startup schedule logging
- **Docker Configuration**: Updated timezone environment variable to PST

### Fixed
- **Token Refresh Panic**: Fixed "non-positive interval for Ticker.Reset" panic
- **Build Errors**: Removed unused variables causing compilation failures
- **Queue Data Loss**: Queue state is now automatically saved, preventing data loss on crashes

### Technical Improvements
- **Code Organization**: Separated timezone handling into utility package
- **Error Handling**: Better error handling for timezone loading
- **Performance**: Auto-save operations run in goroutines to avoid blocking
- **Configuration**: Added timezone field to config structure with proper defaults

## [1.0.0] - 2024-12-XX

### Initial Release
- Basic queue management functionality
- Command system with permissions
- Docker support
- Multi-channel support
- OAuth token management 