-- Add size and gender columns to character_details
ALTER TABLE character_details ADD COLUMN size VARCHAR(20);
ALTER TABLE character_details ADD COLUMN gender VARCHAR(50);
