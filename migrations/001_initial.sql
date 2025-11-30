-- Initial schema for SquadChallengeBot

CREATE TABLE IF NOT EXISTS challenges (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    creator_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    challenge_id TEXT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    order_num INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    image_file_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(challenge_id, order_num)
);

CREATE TABLE IF NOT EXISTS participants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    challenge_id TEXT NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    telegram_id INTEGER NOT NULL,
    display_name TEXT NOT NULL,
    emoji TEXT NOT NULL,
    notify_enabled BOOLEAN DEFAULT 1,
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(challenge_id, telegram_id)
);

CREATE TABLE IF NOT EXISTS task_completions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    participant_id INTEGER NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    completed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(task_id, participant_id)
);

CREATE TABLE IF NOT EXISTS user_states (
    telegram_id INTEGER PRIMARY KEY,
    state TEXT NOT NULL DEFAULT 'idle',
    temp_data TEXT,
    current_challenge TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tasks_challenge ON tasks(challenge_id);
CREATE INDEX IF NOT EXISTS idx_participants_challenge ON participants(challenge_id);
CREATE INDEX IF NOT EXISTS idx_participants_telegram ON participants(telegram_id);
CREATE INDEX IF NOT EXISTS idx_completions_task ON task_completions(task_id);
CREATE INDEX IF NOT EXISTS idx_completions_participant ON task_completions(participant_id);
