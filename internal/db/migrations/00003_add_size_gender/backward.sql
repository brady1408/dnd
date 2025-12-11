-- Remove size and gender columns from character_details
ALTER TABLE character_details DROP COLUMN IF EXISTS gender;
ALTER TABLE character_details DROP COLUMN IF EXISTS size;
