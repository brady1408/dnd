-- Expanded character sheet tables

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
