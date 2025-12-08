# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Changed
- Increased max participants per challenge from 10 to 50
- Increased task description character limit from 800 to 1200
- Centralized business logic limits (MaxParticipants, MaxTasksPerChallenge, MaxChallengesPerUser, MaxTaskDescriptionLength) in `internal/domain/limits.go` for easier configuration

## [0.2.0] - 2025-12-05

### Added
- **Templates System**: Super admins can now create reusable challenge templates
  - Create templates from existing challenges (copies name, description, settings, and tasks)
  - Edit templates (name, description, daily limit, sequential mode)
  - Add, edit, delete, and reorder template tasks
  - Randomize task order in templates
  - Users can create new challenges from templates
- **Super Admin system**
  - View all challenges in the system
  - Observer mode for any challenge
  - Modify settings for any challenge
  - Grant/revoke super admin privileges
- Skip button for join name input (allows joining without setting a custom name)

### Changed
- Renamed "Tasks: Sequential mode" to "Mode: Sequential" in admin panel for clarity
- Increased character limit for task descriptions from 400 to 800 characters

## [0.1.0] - 2025-12-01

### Added
- Challenge creation with customizable settings
- Task management (create, edit, delete, reorder)
- Daily task limits (1-50 tasks per day)
- Sequential mode (hide future tasks until previous ones are completed)
- Time zone synchronization for accurate daily limit resets
- Team progress visualization with leaderboard
- Deep link support for sharing challenges
- Notification system for task completions
- Admin controls for challenge owners
  - Rename challenges
  - Edit task details
  - Configure daily limits and sequential mode
