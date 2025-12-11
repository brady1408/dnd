-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByPublicKey :one
SELECT * FROM users WHERE public_key = $1;

-- name: CreateUserWithPassword :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: CreateUserWithPublicKey :one
INSERT INTO users (public_key)
VALUES ($1)
RETURNING *;

-- name: CreateUserWithBoth :one
INSERT INTO users (email, password_hash, public_key)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateUserPublicKey :one
UPDATE users SET public_key = $2 WHERE id = $1 RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users SET password_hash = $2 WHERE id = $1 RETURNING *;

-- name: UpdateUserEmail :one
UPDATE users SET email = $2 WHERE id = $1 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- Character Queries

-- name: GetCharacterByID :one
SELECT * FROM characters WHERE id = $1;

-- name: GetCharactersByUserID :many
SELECT * FROM characters WHERE user_id = $1 ORDER BY updated_at DESC;

-- name: CreateCharacter :one
INSERT INTO characters (
    user_id, name, class, level, race, background, alignment, experience_points,
    strength, dexterity, constitution, intelligence, wisdom, charisma,
    max_hit_points, current_hit_points, temporary_hit_points,
    armor_class, speed,
    saving_throw_proficiencies, skill_proficiencies,
    equipment, features_traits, notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14,
    $15, $16, $17,
    $18, $19,
    $20, $21,
    $22, $23, $24
)
RETURNING *;

-- name: UpdateCharacterBasicInfo :one
UPDATE characters SET
    name = $2,
    class = $3,
    level = $4,
    race = $5,
    background = $6,
    alignment = $7,
    experience_points = $8
WHERE id = $1
RETURNING *;

-- name: UpdateCharacterAbilities :one
UPDATE characters SET
    strength = $2,
    dexterity = $3,
    constitution = $4,
    intelligence = $5,
    wisdom = $6,
    charisma = $7
WHERE id = $1
RETURNING *;

-- name: UpdateCharacterCombat :one
UPDATE characters SET
    max_hit_points = $2,
    current_hit_points = $3,
    temporary_hit_points = $4,
    armor_class = $5,
    speed = $6
WHERE id = $1
RETURNING *;

-- name: UpdateCharacterHitPoints :one
UPDATE characters SET
    current_hit_points = $2,
    temporary_hit_points = $3
WHERE id = $1
RETURNING *;

-- name: UpdateCharacterProficiencies :one
UPDATE characters SET
    saving_throw_proficiencies = $2,
    skill_proficiencies = $3
WHERE id = $1
RETURNING *;

-- name: UpdateCharacterEquipment :one
UPDATE characters SET equipment = $2 WHERE id = $1 RETURNING *;

-- name: UpdateCharacterNotes :one
UPDATE characters SET
    features_traits = $2,
    notes = $3
WHERE id = $1
RETURNING *;

-- name: UpdateCharacterAlignment :one
UPDATE characters SET alignment = $2
WHERE id = $1
RETURNING *;

-- name: DeleteCharacter :exec
DELETE FROM characters WHERE id = $1;

-- name: DeleteCharacterByUserID :exec
DELETE FROM characters WHERE id = $1 AND user_id = $2;

-- ============================================
-- Character Details Queries
-- ============================================

-- name: GetCharacterDetails :one
SELECT * FROM character_details WHERE character_id = $1;

-- name: CreateCharacterDetails :one
INSERT INTO character_details (character_id)
VALUES ($1)
RETURNING *;

-- name: UpdateCharacterDetails :one
UPDATE character_details SET
    age = $2,
    height = $3,
    weight = $4,
    eyes = $5,
    skin = $6,
    hair = $7,
    size = $8,
    gender = $9,
    faith_deity = $10,
    personality_traits = $11,
    ideals = $12,
    bonds = $13,
    flaws = $14,
    backstory = $15,
    allies_organizations = $16
WHERE character_id = $1
RETURNING *;

-- name: UpdateDeathSaves :one
UPDATE character_details SET
    death_save_successes = $2,
    death_save_failures = $3
WHERE character_id = $1
RETURNING *;

-- name: UpdateInspiration :one
UPDATE character_details SET inspiration = $2
WHERE character_id = $1
RETURNING *;

-- name: UpdateHitDiceUsed :one
UPDATE character_details SET hit_dice_used = $2
WHERE character_id = $1
RETURNING *;

-- ============================================
-- Character Attacks Queries
-- ============================================

-- name: GetCharacterAttacks :many
SELECT * FROM character_attacks WHERE character_id = $1 ORDER BY sort_order;

-- name: GetCharacterAttackByID :one
SELECT * FROM character_attacks WHERE id = $1;

-- name: CreateCharacterAttack :one
INSERT INTO character_attacks (
    character_id, sort_order, name, attack_bonus, damage, damage_type, range, properties, notes
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateCharacterAttack :one
UPDATE character_attacks SET
    name = $2,
    attack_bonus = $3,
    damage = $4,
    damage_type = $5,
    range = $6,
    properties = $7,
    notes = $8
WHERE id = $1
RETURNING *;

-- name: UpdateCharacterAttackSortOrder :exec
UPDATE character_attacks SET sort_order = $2 WHERE id = $1;

-- name: DeleteCharacterAttack :exec
DELETE FROM character_attacks WHERE id = $1;

-- name: DeleteAllCharacterAttacks :exec
DELETE FROM character_attacks WHERE character_id = $1;

-- ============================================
-- Character Actions Queries
-- ============================================

-- name: GetCharacterActions :many
SELECT * FROM character_actions WHERE character_id = $1 ORDER BY sort_order;

-- name: GetCharacterActionByID :one
SELECT * FROM character_actions WHERE id = $1;

-- name: CreateCharacterAction :one
INSERT INTO character_actions (
    character_id, sort_order, name, action_type, source, description, uses_per, uses_max, uses_current
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateCharacterAction :one
UPDATE character_actions SET
    name = $2,
    action_type = $3,
    source = $4,
    description = $5,
    uses_per = $6,
    uses_max = $7,
    uses_current = $8
WHERE id = $1
RETURNING *;

-- name: UpdateCharacterActionUses :one
UPDATE character_actions SET uses_current = $2
WHERE id = $1
RETURNING *;

-- name: DeleteCharacterAction :exec
DELETE FROM character_actions WHERE id = $1;

-- ============================================
-- Character Inventory Queries
-- ============================================

-- name: GetCharacterInventory :many
SELECT * FROM character_inventory WHERE character_id = $1 ORDER BY sort_order;

-- name: CreateCharacterInventoryItem :one
INSERT INTO character_inventory (
    character_id, sort_order, name, quantity, weight, location, notes, is_equipped
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateCharacterInventoryItem :one
UPDATE character_inventory SET
    name = $2,
    quantity = $3,
    weight = $4,
    location = $5,
    notes = $6,
    is_equipped = $7
WHERE id = $1
RETURNING *;

-- name: DeleteCharacterInventoryItem :exec
DELETE FROM character_inventory WHERE id = $1;

-- ============================================
-- Character Currency Queries
-- ============================================

-- name: GetCharacterCurrency :one
SELECT * FROM character_currency WHERE character_id = $1;

-- name: CreateCharacterCurrency :one
INSERT INTO character_currency (character_id)
VALUES ($1)
RETURNING *;

-- name: UpdateCharacterCurrency :one
UPDATE character_currency SET
    copper = $2,
    silver = $3,
    electrum = $4,
    gold = $5,
    platinum = $6
WHERE character_id = $1
RETURNING *;

-- ============================================
-- Character Magic Items Queries
-- ============================================

-- name: GetCharacterMagicItems :many
SELECT * FROM character_magic_items WHERE character_id = $1 ORDER BY sort_order;

-- name: CreateCharacterMagicItem :one
INSERT INTO character_magic_items (
    character_id, sort_order, name, rarity, attunement_required, is_attuned, weight, description
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateCharacterMagicItem :one
UPDATE character_magic_items SET
    name = $2,
    rarity = $3,
    attunement_required = $4,
    is_attuned = $5,
    weight = $6,
    description = $7
WHERE id = $1
RETURNING *;

-- name: ToggleMagicItemAttunement :one
UPDATE character_magic_items SET is_attuned = NOT is_attuned
WHERE id = $1
RETURNING *;

-- name: DeleteCharacterMagicItem :exec
DELETE FROM character_magic_items WHERE id = $1;

-- ============================================
-- Character Spellcasting Queries
-- ============================================

-- name: GetCharacterSpellcasting :one
SELECT * FROM character_spellcasting WHERE character_id = $1;

-- name: CreateCharacterSpellcasting :one
INSERT INTO character_spellcasting (character_id, spellcasting_class, spellcasting_ability)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateCharacterSpellcasting :one
UPDATE character_spellcasting SET
    spellcasting_class = $2,
    spellcasting_ability = $3,
    spell_save_dc = $4,
    spell_attack_bonus = $5
WHERE character_id = $1
RETURNING *;

-- name: UpdateSpellSlots :one
UPDATE character_spellcasting SET
    slots_1_max = $2, slots_1_used = $3,
    slots_2_max = $4, slots_2_used = $5,
    slots_3_max = $6, slots_3_used = $7,
    slots_4_max = $8, slots_4_used = $9,
    slots_5_max = $10, slots_5_used = $11,
    slots_6_max = $12, slots_6_used = $13,
    slots_7_max = $14, slots_7_used = $15,
    slots_8_max = $16, slots_8_used = $17,
    slots_9_max = $18, slots_9_used = $19
WHERE character_id = $1
RETURNING *;

-- ============================================
-- Character Spells Queries
-- ============================================

-- name: GetCharacterSpells :many
SELECT * FROM character_spells WHERE character_id = $1 ORDER BY level, name;

-- name: GetCharacterSpellsByLevel :many
SELECT * FROM character_spells WHERE character_id = $1 AND level = $2 ORDER BY name;

-- name: GetPreparedSpells :many
SELECT * FROM character_spells WHERE character_id = $1 AND is_prepared = true ORDER BY level, name;

-- name: CreateCharacterSpell :one
INSERT INTO character_spells (
    character_id, name, level, school, is_prepared, is_ritual, casting_time, range, components, duration, description, source
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: UpdateCharacterSpell :one
UPDATE character_spells SET
    name = $2,
    level = $3,
    school = $4,
    is_prepared = $5,
    is_ritual = $6,
    casting_time = $7,
    range = $8,
    components = $9,
    duration = $10,
    description = $11,
    source = $12
WHERE id = $1
RETURNING *;

-- name: ToggleSpellPrepared :one
UPDATE character_spells SET is_prepared = NOT is_prepared
WHERE id = $1
RETURNING *;

-- name: DeleteCharacterSpell :exec
DELETE FROM character_spells WHERE id = $1;

-- ============================================
-- Character Features Queries
-- ============================================

-- name: GetCharacterFeatures :many
SELECT * FROM character_features WHERE character_id = $1 ORDER BY sort_order;

-- name: GetCharacterFeaturesByType :many
SELECT * FROM character_features WHERE character_id = $1 AND source_type = $2 ORDER BY sort_order;

-- name: CreateCharacterFeature :one
INSERT INTO character_features (
    character_id, sort_order, name, source, source_type, description
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateCharacterFeature :one
UPDATE character_features SET
    name = $2,
    source = $3,
    source_type = $4,
    description = $5
WHERE id = $1
RETURNING *;

-- name: DeleteCharacterFeature :exec
DELETE FROM character_features WHERE id = $1;
