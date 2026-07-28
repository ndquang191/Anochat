CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT,
    name TEXT,
    avatar_url TEXT,
    is_active BOOLEAN DEFAULT FALSE,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    ban_count INTEGER NOT NULL DEFAULT 0,
    review_request_count INTEGER NOT NULL DEFAULT 0,
    review_requested BOOLEAN NOT NULL DEFAULT FALSE,
    banned_at TIMESTAMP,
    review_requested_at TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    nickname TEXT,
    nickname_updated_at TIMESTAMPTZ,
    is_male BOOLEAN,
    age INTEGER,
    is_hidden BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user1_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user2_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP,
    CONSTRAINT rooms_distinct_users CHECK (user1_id <> user2_id)
);

CREATE TABLE IF NOT EXISTS room_sessions (
    room_id UUID PRIMARY KEY,
    status VARCHAR(16) NOT NULL DEFAULT 'matched',
    matched_at TIMESTAMPTZ NOT NULL,
    connected_at TIMESTAMPTZ,
    disconnected_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS banned_words (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    word TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'General',
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uni_banned_words_word UNIQUE (word)
);

CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id UUID NOT NULL,
    reported_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS report_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    original_message_id UUID NOT NULL,
    sender_id UUID NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    CONSTRAINT idx_report_original_message UNIQUE (report_id, original_message_id)
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_is_deleted ON users(is_deleted);
CREATE INDEX IF NOT EXISTS idx_users_review_requested ON users(review_requested) WHERE is_active = FALSE;
CREATE INDEX IF NOT EXISTS idx_users_banned_admin_cursor ON users(review_requested DESC, review_requested_at DESC, banned_at DESC, created_at DESC, id DESC) WHERE is_active = FALSE AND is_deleted = FALSE;
CREATE INDEX IF NOT EXISTS idx_profiles_is_male ON profiles(is_male);
CREATE INDEX IF NOT EXISTS idx_profiles_age ON profiles(age);
CREATE INDEX IF NOT EXISTS idx_rooms_user1_id ON rooms(user1_id);
CREATE INDEX IF NOT EXISTS idx_rooms_user2_id ON rooms(user2_id);
CREATE INDEX IF NOT EXISTS idx_rooms_created_at ON rooms(created_at);
CREATE INDEX IF NOT EXISTS idx_room_sessions_matched_at ON room_sessions(matched_at);
CREATE INDEX IF NOT EXISTS idx_room_sessions_status ON room_sessions(status);
CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
CREATE INDEX IF NOT EXISTS idx_messages_room_created ON messages(room_id, created_at);
CREATE INDEX IF NOT EXISTS idx_messages_room_created_id ON messages(room_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_reports_reported_user_id ON reports(reported_user_id);
CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status);
CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at);
CREATE INDEX IF NOT EXISTS idx_reports_admin_grouping ON reports(status, reported_user_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_report_messages_report_id ON report_messages(report_id);
CREATE INDEX IF NOT EXISTS idx_report_messages_created_at ON report_messages(created_at);

UPDATE users
SET ban_count = 1,
    banned_at = COALESCE(banned_at, created_at)
WHERE is_active = FALSE
  AND is_deleted = FALSE
  AND ban_count = 0;

INSERT INTO room_sessions (room_id, status, matched_at)
SELECT id, 'matched', created_at
FROM rooms
ON CONFLICT (room_id) DO NOTHING;
