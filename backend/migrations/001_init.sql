-- Shalom Database Schema
-- PostgreSQL: Users, Meetings, Participants

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url TEXT DEFAULT '',
    status VARCHAR(20) DEFAULT 'offline',
    last_seen TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Meetings table
CREATE TABLE IF NOT EXISTS meetings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    host_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    meeting_code VARCHAR(16) UNIQUE NOT NULL,
    join_link VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'active', 'ended')),
    max_participants INT DEFAULT 500,
    scheduled_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Meeting participants
CREATE TABLE IF NOT EXISTS meeting_participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    meeting_id UUID NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'listener' CHECK (role IN ('host', 'cohost', 'speaker', 'listener', 'moderator', 'guest')),
    is_muted BOOLEAN DEFAULT true,
    is_video_on BOOLEAN DEFAULT false,
    is_screen_share BOOLEAN DEFAULT false,
    hand_raised BOOLEAN DEFAULT false,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    left_at TIMESTAMPTZ,
    UNIQUE(meeting_id, user_id)
);

-- Indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_meetings_host ON meetings(host_id);
CREATE INDEX idx_meetings_code ON meetings(meeting_code);
CREATE INDEX idx_meetings_status ON meetings(status);
CREATE INDEX idx_participants_meeting ON meeting_participants(meeting_id);
CREATE INDEX idx_participants_user ON meeting_participants(user_id);
CREATE INDEX idx_participants_active ON meeting_participants(meeting_id, left_at);
