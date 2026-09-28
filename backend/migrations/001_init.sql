-- ============================================================
-- Shalom Database Schema
-- PostgreSQL 16+ with proper ENUMs, triggers, and indexes
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- ENUM TYPES
-- ============================================================

CREATE TYPE user_status AS ENUM ('online', 'offline', 'away', 'dnd');
CREATE TYPE meeting_status AS ENUM ('scheduled', 'active', 'ended', 'cancelled');
CREATE TYPE participant_role AS ENUM ('host', 'cohost', 'speaker', 'listener', 'moderator', 'guest');
CREATE TYPE room_type AS ENUM ('public', 'private', 'meeting', 'group', 'channel');
CREATE TYPE notification_type AS ENUM ('meeting_invite', 'meeting_start', 'message', 'mention', 'reaction', 'system', 'recording_ready');
CREATE TYPE recording_status AS ENUM ('recording', 'processing', 'ready', 'failed', 'deleted');
CREATE TYPE bandwidth_mode AS ENUM ('ultra_saving', 'economy', 'standard', 'hd');
CREATE TYPE file_type AS ENUM ('image', 'video', 'document', 'audio', 'other');

-- ============================================================
-- USERS
-- ============================================================

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           VARCHAR(255) UNIQUE NOT NULL,
    username        VARCHAR(50) UNIQUE NOT NULL,
    display_name    VARCHAR(100) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    avatar_url      TEXT DEFAULT '',
    bio             VARCHAR(500) DEFAULT '',
    phone           VARCHAR(20) DEFAULT '',
    status          user_status DEFAULT 'offline',
    last_seen       TIMESTAMPTZ DEFAULT NOW(),
    email_verified  BOOLEAN DEFAULT FALSE,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_last_seen ON users(last_seen);

-- ============================================================
-- USER SETTINGS (one-to-one)
-- ============================================================

CREATE TABLE user_settings (
    user_id             UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    notifications_enabled   BOOLEAN DEFAULT TRUE,
    sound_enabled           BOOLEAN DEFAULT TRUE,
    default_bandwidth       bandwidth_mode DEFAULT 'standard',
    language                VARCHAR(10) DEFAULT 'en',
    theme                   VARCHAR(20) DEFAULT 'dark',
    auto_reconnect          BOOLEAN DEFAULT TRUE,
    show_online_status      BOOLEAN DEFAULT TRUE,
    allow_direct_messages   BOOLEAN DEFAULT TRUE,
    camera_default_on       BOOLEAN DEFAULT FALSE,
    mic_default_muted       BOOLEAN DEFAULT TRUE,
    created_at              TIMESTAMPTZ DEFAULT NOW(),
    updated_at              TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- CONTACTS / FRIENDS
-- ============================================================

CREATE TABLE contacts (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contact_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    nickname    VARCHAR(100) DEFAULT '',
    is_blocked  BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, contact_id),
    CHECK (user_id != contact_id)
);

CREATE INDEX idx_contacts_user ON contacts(user_id);
CREATE INDEX idx_contacts_contact ON contacts(contact_id);

-- ============================================================
-- MEETINGS
-- ============================================================

CREATE TABLE meetings (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title             VARCHAR(255) NOT NULL,
    description       TEXT DEFAULT '',
    host_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    meeting_code      VARCHAR(16) UNIQUE NOT NULL,
    join_link         VARCHAR(500) NOT NULL,
    status            meeting_status DEFAULT 'scheduled',
    max_participants  INT DEFAULT 500,
    is_recording      BOOLEAN DEFAULT FALSE,
    is_waiting_room   BOOLEAN DEFAULT FALSE,
    is_breakout       BOOLEAN DEFAULT FALSE,
    parent_meeting_id UUID REFERENCES meetings(id) ON DELETE SET NULL,
    password          VARCHAR(255) DEFAULT '',
    scheduled_at      TIMESTAMPTZ,
    started_at        TIMESTAMPTZ,
    ended_at          TIMESTAMPTZ,
    duration_seconds  INT DEFAULT 0,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_meetings_host ON meetings(host_id);
CREATE INDEX idx_meetings_code ON meetings(meeting_code);
CREATE INDEX idx_meetings_status ON meetings(status);
CREATE INDEX idx_meetings_scheduled ON meetings(scheduled_at);
CREATE INDEX idx_meetings_parent ON meetings(parent_meeting_id);

-- ============================================================
-- MEETING PARTICIPANTS
-- ============================================================

CREATE TABLE meeting_participants (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    meeting_id      UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            participant_role DEFAULT 'listener',
    is_muted        BOOLEAN DEFAULT TRUE,
    is_video_on     BOOLEAN DEFAULT FALSE,
    is_screen_share BOOLEAN DEFAULT FALSE,
    hand_raised     BOOLEAN DEFAULT FALSE,
    is_banned       BOOLEAN DEFAULT FALSE,
    bandwidth_mode  bandwidth_mode DEFAULT 'standard',
    joined_at       TIMESTAMPTZ DEFAULT NOW(),
    left_at         TIMESTAMPTZ,
    duration_seconds INT DEFAULT 0,
    UNIQUE(meeting_id, user_id)
);

CREATE INDEX idx_participants_meeting ON meeting_participants(meeting_id);
CREATE INDEX idx_participants_user ON meeting_participants(user_id);
CREATE INDEX idx_participants_active ON meeting_participants(meeting_id, left_at);
CREATE INDEX idx_participants_role ON meeting_participants(role);

-- ============================================================
-- MEETING WAITING ROOM
-- ============================================================

CREATE TABLE waiting_room (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    meeting_id  UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    admitted    BOOLEAN DEFAULT FALSE,
    admitted_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    admitted_at TIMESTAMPTZ,
    UNIQUE(meeting_id, user_id)
);

CREATE INDEX idx_waiting_room_meeting ON waiting_room(meeting_id);

-- ============================================================
-- MEETING RECORDINGS
-- ============================================================

CREATE TABLE recordings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    meeting_id      UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    started_by      UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    status          recording_status DEFAULT 'recording',
    file_url        TEXT DEFAULT '',
    file_size       BIGINT DEFAULT 0,
    duration_seconds INT DEFAULT 0,
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    ended_at        TIMESTAMPTZ,
    processed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_recordings_meeting ON recordings(meeting_id);
CREATE INDEX idx_recordings_status ON recordings(status);

-- ============================================================
-- MEETING ATTENDANCE
-- ============================================================

CREATE TABLE attendance (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    meeting_id  UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at   TIMESTAMPTZ NOT NULL,
    left_at     TIMESTAMPTZ,
    duration    INT DEFAULT 0,
    UNIQUE(meeting_id, user_id, joined_at)
);

CREATE INDEX idx_attendance_meeting ON attendance(meeting_id);
CREATE INDEX idx_attendance_user ON attendance(user_id);

-- ============================================================
-- NOTIFICATIONS
-- ============================================================

CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        notification_type NOT NULL,
    title       VARCHAR(255) NOT NULL,
    body        TEXT DEFAULT '',
    data        JSONB DEFAULT '{}',
    is_read     BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_created ON notifications(created_at DESC);

-- ============================================================
-- FILE UPLOADS (metadata)
-- ============================================================

CREATE TABLE file_uploads (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_name   VARCHAR(255) NOT NULL,
    file_type   file_type DEFAULT 'other',
    file_size   BIGINT DEFAULT 0,
    mime_type   VARCHAR(100) DEFAULT '',
    storage_key VARCHAR(500) NOT NULL,
    url         TEXT NOT NULL,
    is_public   BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_files_uploader ON file_uploads(uploader_id);
CREATE INDEX idx_files_type ON file_uploads(file_type);

-- ============================================================
-- BROADCAST ANNOUNCEMENTS (Telegram-style channels)
-- ============================================================

CREATE TABLE channels (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            VARCHAR(255) NOT NULL,
    description     TEXT DEFAULT '',
    created_by      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_public       BOOLEAN DEFAULT TRUE,
    member_count    INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE channel_members (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id  UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        VARCHAR(20) DEFAULT 'member' CHECK (role IN ('admin', 'moderator', 'member')),
    joined_at   TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(channel_id, user_id)
);

CREATE INDEX idx_channel_members_channel ON channel_members(channel_id);
CREATE INDEX idx_channel_members_user ON channel_members(user_id);

-- ============================================================
-- BANDWIDTH LOG (for analytics)
-- ============================================================

CREATE TABLE bandwidth_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    meeting_id      UUID REFERENCES meetings(id) ON DELETE SET NULL,
    bandwidth_kbps  INT NOT NULL,
    mode            bandwidth_mode NOT NULL,
    network_quality VARCHAR(20) DEFAULT 'average',
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_bw_logs_user ON bandwidth_logs(user_id);
CREATE INDEX idx_bw_logs_meeting ON bandwidth_logs(meeting_id);
CREATE INDEX idx_bw_logs_created ON bandwidth_logs(created_at DESC);

-- ============================================================
-- AUTO-UPDATE TRIGGERS
-- ============================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_settings_updated_at BEFORE UPDATE ON user_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_meetings_updated_at BEFORE UPDATE ON meetings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_recordings_updated_at BEFORE UPDATE ON recordings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_channels_updated_at BEFORE UPDATE ON channels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
