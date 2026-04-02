\c app_main

CREATE SCHEMA IF NOT EXISTS chat;
SET search_path TO chat;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================
-- Users table
-- =====================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    avatar_url TEXT,
    status VARCHAR(20) DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'away', 'dnd')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================
-- Rooms (channels/servers)
-- =====================
CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_private BOOLEAN DEFAULT FALSE,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================
-- Room members (many-to-many)
-- =====================
CREATE TABLE IF NOT EXISTS room_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'moderator', 'member')),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(room_id, user_id)
);

-- =====================
-- Messages
-- =====================
CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(26) PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,  -- removed NOT NULL
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'file', 'system')),
    is_edited BOOLEAN DEFAULT FALSE,
    parent_id VARCHAR(26) REFERENCES messages(id) ON DELETE SET NULL, -- for thread/reply support
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================
-- Direct Messages
-- =====================
CREATE TABLE IF NOT EXISTS direct_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'file')),
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================
-- Room invitations
-- =====================
CREATE TABLE IF NOT EXISTS room_invitations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invite_code VARCHAR(20) NOT NULL UNIQUE,
    max_uses INT DEFAULT NULL,       -- NULL = unlimited
    use_count INT DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- =====================
-- Indexes for performance
-- =====================
CREATE INDEX idx_messages_user_id ON messages(user_id);
CREATE INDEX idx_messages_room_id ON messages(room_id, id DESC);
CREATE INDEX idx_room_members_room_id ON room_members(room_id);
CREATE INDEX idx_room_members_user_id ON room_members(user_id);
CREATE INDEX idx_direct_messages_sender ON direct_messages(sender_id);
CREATE INDEX idx_direct_messages_receiver ON direct_messages(receiver_id);
CREATE INDEX idx_direct_messages_created_at ON direct_messages(sender_id, receiver_id, created_at DESC);