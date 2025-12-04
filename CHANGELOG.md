# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added
- **Templates System**: Super admins can now create reusable challenge templates
  - Create templates from existing challenges (copies name, description, settings, and tasks)
  - Edit templates (name, description, daily limit, sequential mode)
  - Add, edit, delete, and reorder template tasks
  - Randomize task order in templates
  - Users can create new challenges from templates
- Template service tests with 21 test cases

### Changed
- Renamed "Tasks: Sequential mode" to "Mode: Sequential" in admin panel for clarity
- Template Admin Panel now matches Challenge Admin Panel format

### Fixed
- Template task reordering now works correctly (uses batch updates to avoid unique constraint violations)
- Template task randomization now works correctly
- Skip and Cancel buttons now properly route back to template editing flow
- Template task images are now properly copied when creating template from challenge

## [1.0.0] - 2024

### Added
- Challenge creation with customizable settings
- Task management (create, edit, delete, reorder)
- Daily task limits (1-50 tasks per day)
- Sequential mode (hide future tasks until previous ones are completed)
- Time zone synchronization for accurate daily limit resets
- Team progress visualization with leaderboard
- Deep link support for sharing challenges
- Notification system for task completions
- Super Admin system
  - View all challenges in the system
  - Observer mode for any challenge
  - Modify settings for any challenge
  - Grant/revoke super admin privileges
- Admin controls for challenge owners
  - Rename challenges
  - Edit task details
  - Configure daily limits and sequential mode
