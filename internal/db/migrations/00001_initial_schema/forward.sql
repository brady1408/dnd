-- Initial schema: users and characters tables

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table for authentication
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE,
    password_hash TEXT,
    public_key TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- At least one auth method must be configured
    CONSTRAINT auth_method_required CHECK (
        (email IS NOT NULL AND password_hash IS NOT NULL) OR
        public_key IS NOT NULL
    )
);

-- Index for faster lookups
CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_public_key ON users(public_key) WHERE public_key IS NOT NULL;

-- Characters table
CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Basic Info
    name VARCHAR(100) NOT NULL,
    class VARCHAR(50) NOT NULL,
    level INTEGER NOT NULL DEFAULT 1 CHECK (level >= 1 AND level <= 20),
    race VARCHAR(50) NOT NULL,
    background VARCHAR(50),
    alignment VARCHAR(50),
    experience_points INTEGER NOT NULL DEFAULT 0 CHECK (experience_points >= 0),

    -- Ability Scores (3-30 range for flexibility with magic items)
    strength INTEGER NOT NULL CHECK (strength >= 1 AND strength <= 30),
    dexterity INTEGER NOT NULL CHECK (dexterity >= 1 AND dexterity <= 30),
    constitution INTEGER NOT NULL CHECK (constitution >= 1 AND constitution <= 30),
    intelligence INTEGER NOT NULL CHECK (intelligence >= 1 AND intelligence <= 30),
    wisdom INTEGER NOT NULL CHECK (wisdom >= 1 AND wisdom <= 30),
    charisma INTEGER NOT NULL CHECK (charisma >= 1 AND charisma <= 30),

    -- Combat Stats
    max_hit_points INTEGER NOT NULL CHECK (max_hit_points >= 1),
    current_hit_points INTEGER NOT NULL,
    temporary_hit_points INTEGER NOT NULL DEFAULT 0 CHECK (temporary_hit_points >= 0),
    armor_class INTEGER NOT NULL DEFAULT 10,
    speed INTEGER NOT NULL DEFAULT 30,

    -- Proficiencies (stored as arrays)
    saving_throw_proficiencies TEXT[] NOT NULL DEFAULT '{}',
    skill_proficiencies TEXT[] NOT NULL DEFAULT '{}',

    -- Other
    equipment JSONB NOT NULL DEFAULT '[]',
    features_traits TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for user's characters
CREATE INDEX idx_characters_user_id ON characters(user_id);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_characters_updated_at
    BEFORE UPDATE ON characters
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
