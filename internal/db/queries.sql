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

-- name: DeleteCharacter :exec
DELETE FROM characters WHERE id = $1;

-- name: DeleteCharacterByUserID :exec
DELETE FROM characters WHERE id = $1 AND user_id = $2;
