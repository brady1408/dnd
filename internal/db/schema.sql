-- D&D Character Tracker Schema

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

-- Character Details (extended info)
CREATE TABLE character_details (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- Physical characteristics
    age VARCHAR(20),
    height VARCHAR(20),
    weight VARCHAR(20),
    eyes VARCHAR(50),
    skin VARCHAR(50),
    hair VARCHAR(50),

    -- Personal details
    faith_deity VARCHAR(100),
    personality_traits TEXT,
    ideals TEXT,
    bonds TEXT,
    flaws TEXT,
    backstory TEXT,
    allies_organizations TEXT,

    -- Combat extras
    inspiration BOOLEAN DEFAULT FALSE,
    death_save_successes INTEGER DEFAULT 0 CHECK (death_save_successes >= 0 AND death_save_successes <= 3),
    death_save_failures INTEGER DEFAULT 0 CHECK (death_save_failures >= 0 AND death_save_failures <= 3),
    hit_dice_used INTEGER DEFAULT 0,

    UNIQUE(character_id)
);

CREATE INDEX idx_character_details_character_id ON character_details(character_id);

-- Character Attacks/Weapons
CREATE TABLE character_attacks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,

    name VARCHAR(100) NOT NULL,
    attack_bonus INTEGER DEFAULT 0,
    damage VARCHAR(50),
    damage_type VARCHAR(50),
    range VARCHAR(50),
    properties VARCHAR(200),
    notes TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_character_attacks_character_id ON character_attacks(character_id, sort_order);

-- Character Actions
CREATE TABLE character_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,

    name VARCHAR(100) NOT NULL,
    action_type VARCHAR(50) DEFAULT 'action',
    source VARCHAR(100),
    description TEXT,
    uses_per VARCHAR(50),
    uses_max INTEGER,
    uses_current INTEGER,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_character_actions_character_id ON character_actions(character_id, sort_order);

-- Character Inventory
CREATE TABLE character_inventory (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,

    name VARCHAR(200) NOT NULL,
    quantity INTEGER DEFAULT 1,
    weight DECIMAL(10,2),
    location VARCHAR(100),
    notes TEXT,
    is_equipped BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_character_inventory_character_id ON character_inventory(character_id, sort_order);

-- Character Currency
CREATE TABLE character_currency (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    copper INTEGER DEFAULT 0,
    silver INTEGER DEFAULT 0,
    electrum INTEGER DEFAULT 0,
    gold INTEGER DEFAULT 0,
    platinum INTEGER DEFAULT 0,

    UNIQUE(character_id)
);

CREATE INDEX idx_character_currency_character_id ON character_currency(character_id);

-- Character Magic Items
CREATE TABLE character_magic_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,

    name VARCHAR(200) NOT NULL,
    rarity VARCHAR(50),
    attunement_required BOOLEAN DEFAULT FALSE,
    is_attuned BOOLEAN DEFAULT FALSE,
    weight DECIMAL(10,2),
    description TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_character_magic_items_character_id ON character_magic_items(character_id, sort_order);

-- Character Spellcasting
CREATE TABLE character_spellcasting (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    spellcasting_class VARCHAR(50),
    spellcasting_ability VARCHAR(20),
    spell_save_dc INTEGER,
    spell_attack_bonus INTEGER,

    -- Spell slots (current/max for each level)
    slots_1_max INTEGER DEFAULT 0,
    slots_1_used INTEGER DEFAULT 0,
    slots_2_max INTEGER DEFAULT 0,
    slots_2_used INTEGER DEFAULT 0,
    slots_3_max INTEGER DEFAULT 0,
    slots_3_used INTEGER DEFAULT 0,
    slots_4_max INTEGER DEFAULT 0,
    slots_4_used INTEGER DEFAULT 0,
    slots_5_max INTEGER DEFAULT 0,
    slots_5_used INTEGER DEFAULT 0,
    slots_6_max INTEGER DEFAULT 0,
    slots_6_used INTEGER DEFAULT 0,
    slots_7_max INTEGER DEFAULT 0,
    slots_7_used INTEGER DEFAULT 0,
    slots_8_max INTEGER DEFAULT 0,
    slots_8_used INTEGER DEFAULT 0,
    slots_9_max INTEGER DEFAULT 0,
    slots_9_used INTEGER DEFAULT 0,

    UNIQUE(character_id)
);

CREATE INDEX idx_character_spellcasting_character_id ON character_spellcasting(character_id);

-- Character Spells
CREATE TABLE character_spells (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    name VARCHAR(200) NOT NULL,
    level INTEGER NOT NULL CHECK (level >= 0 AND level <= 9),
    school VARCHAR(50),
    is_prepared BOOLEAN DEFAULT FALSE,
    is_ritual BOOLEAN DEFAULT FALSE,
    casting_time VARCHAR(50),
    range VARCHAR(50),
    components VARCHAR(20),
    duration VARCHAR(100),
    description TEXT,
    source VARCHAR(100),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_character_spells_character_id ON character_spells(character_id, level);

-- Character Features
CREATE TABLE character_features (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,

    name VARCHAR(200) NOT NULL,
    source VARCHAR(100),
    source_type VARCHAR(50),
    description TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_character_features_character_id ON character_features(character_id, sort_order);
