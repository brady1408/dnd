package character

// Skills and their associated abilities
var Skills = map[string]string{
	"Acrobatics":      "dexterity",
	"Animal Handling": "wisdom",
	"Arcana":          "intelligence",
	"Athletics":       "strength",
	"Deception":       "charisma",
	"History":         "intelligence",
	"Insight":         "wisdom",
	"Intimidation":    "charisma",
	"Investigation":   "intelligence",
	"Medicine":        "wisdom",
	"Nature":          "intelligence",
	"Perception":      "wisdom",
	"Performance":     "charisma",
	"Persuasion":      "charisma",
	"Religion":        "intelligence",
	"Sleight of Hand": "dexterity",
	"Stealth":         "dexterity",
	"Survival":        "wisdom",
}

// SkillList is an ordered list of all skills
var SkillList = []string{
	"Acrobatics", "Animal Handling", "Arcana", "Athletics",
	"Deception", "History", "Insight", "Intimidation",
	"Investigation", "Medicine", "Nature", "Perception",
	"Performance", "Persuasion", "Religion", "Sleight of Hand",
	"Stealth", "Survival",
}

// Abilities is the list of ability names
var Abilities = []string{
	"Strength", "Dexterity", "Constitution",
	"Intelligence", "Wisdom", "Charisma",
}

// Classes available in 5e
var Classes = []string{
	"Barbarian", "Bard", "Cleric", "Druid", "Fighter",
	"Monk", "Paladin", "Ranger", "Rogue", "Sorcerer",
	"Warlock", "Wizard",
}

// Races available in 5e (PHB)
var Races = []string{
	"Dragonborn", "Dwarf", "Elf", "Gnome", "Half-Elf",
	"Half-Orc", "Halfling", "Human", "Tiefling",
}

// Backgrounds available in 5e (PHB)
var Backgrounds = []string{
	"Acolyte", "Charlatan", "Criminal", "Entertainer",
	"Folk Hero", "Guild Artisan", "Hermit", "Noble",
	"Outlander", "Sage", "Sailor", "Soldier", "Urchin",
}

// Alignments in 5e
var Alignments = []string{
	"Lawful Good", "Neutral Good", "Chaotic Good",
	"Lawful Neutral", "True Neutral", "Chaotic Neutral",
	"Lawful Evil", "Neutral Evil", "Chaotic Evil",
}

// ClassHitDice maps class to hit dice
var ClassHitDice = map[string]int{
	"Barbarian": 12,
	"Bard":      8,
	"Cleric":    8,
	"Druid":     8,
	"Fighter":   10,
	"Monk":      8,
	"Paladin":   10,
	"Ranger":    10,
	"Rogue":     8,
	"Sorcerer":  6,
	"Warlock":   8,
	"Wizard":    6,
}

// ClassSavingThrows maps class to proficient saving throws
var ClassSavingThrows = map[string][]string{
	"Barbarian": {"Strength", "Constitution"},
	"Bard":      {"Dexterity", "Charisma"},
	"Cleric":    {"Wisdom", "Charisma"},
	"Druid":     {"Intelligence", "Wisdom"},
	"Fighter":   {"Strength", "Constitution"},
	"Monk":      {"Strength", "Dexterity"},
	"Paladin":   {"Wisdom", "Charisma"},
	"Ranger":    {"Strength", "Dexterity"},
	"Rogue":     {"Dexterity", "Intelligence"},
	"Sorcerer":  {"Constitution", "Charisma"},
	"Warlock":   {"Wisdom", "Charisma"},
	"Wizard":    {"Intelligence", "Wisdom"},
}

// ClassSkillChoices maps class to available skill choices and how many to pick
type SkillChoice struct {
	Options []string
	Count   int
}

var ClassSkillChoices = map[string]SkillChoice{
	"Barbarian": {
		Options: []string{"Animal Handling", "Athletics", "Intimidation", "Nature", "Perception", "Survival"},
		Count:   2,
	},
	"Bard": {
		Options: SkillList, // Bards can choose any skill
		Count:   3,
	},
	"Cleric": {
		Options: []string{"History", "Insight", "Medicine", "Persuasion", "Religion"},
		Count:   2,
	},
	"Druid": {
		Options: []string{"Arcana", "Animal Handling", "Insight", "Medicine", "Nature", "Perception", "Religion", "Survival"},
		Count:   2,
	},
	"Fighter": {
		Options: []string{"Acrobatics", "Animal Handling", "Athletics", "History", "Insight", "Intimidation", "Perception", "Survival"},
		Count:   2,
	},
	"Monk": {
		Options: []string{"Acrobatics", "Athletics", "History", "Insight", "Religion", "Stealth"},
		Count:   2,
	},
	"Paladin": {
		Options: []string{"Athletics", "Insight", "Intimidation", "Medicine", "Persuasion", "Religion"},
		Count:   2,
	},
	"Ranger": {
		Options: []string{"Animal Handling", "Athletics", "Insight", "Investigation", "Nature", "Perception", "Stealth", "Survival"},
		Count:   3,
	},
	"Rogue": {
		Options: []string{"Acrobatics", "Athletics", "Deception", "Insight", "Intimidation", "Investigation", "Perception", "Performance", "Persuasion", "Sleight of Hand", "Stealth"},
		Count:   4,
	},
	"Sorcerer": {
		Options: []string{"Arcana", "Deception", "Insight", "Intimidation", "Persuasion", "Religion"},
		Count:   2,
	},
	"Warlock": {
		Options: []string{"Arcana", "Deception", "History", "Intimidation", "Investigation", "Nature", "Religion"},
		Count:   2,
	},
	"Wizard": {
		Options: []string{"Arcana", "History", "Insight", "Investigation", "Medicine", "Religion"},
		Count:   2,
	},
}

// RaceSpeed maps race to base walking speed
var RaceSpeed = map[string]int{
	"Dragonborn": 30,
	"Dwarf":      25,
	"Elf":        30,
	"Gnome":      25,
	"Half-Elf":   30,
	"Half-Orc":   30,
	"Halfling":   25,
	"Human":      30,
	"Tiefling":   30,
}

// Character represents a D&D 5e character
type Character struct {
	// Basic Info
	Name            string
	Class           string
	Level           int
	Race            string
	Background      string
	Alignment       string
	ExperiencePoints int

	// Ability Scores
	Strength     int
	Dexterity    int
	Constitution int
	Intelligence int
	Wisdom       int
	Charisma     int

	// Combat
	MaxHitPoints       int
	CurrentHitPoints   int
	TemporaryHitPoints int
	ArmorClass         int
	Speed              int

	// Proficiencies
	SavingThrowProficiencies []string
	SkillProficiencies       []string

	// Other
	Equipment      []string
	FeaturesTraits string
	Notes          string
}

// NewCharacter creates a new character with defaults
func NewCharacter() *Character {
	return &Character{
		Level:                    1,
		ExperiencePoints:         0,
		TemporaryHitPoints:       0,
		ArmorClass:               10,
		Speed:                    30,
		SavingThrowProficiencies: []string{},
		SkillProficiencies:       []string{},
		Equipment:                []string{},
	}
}

// SetClass sets the class and updates related attributes
func (c *Character) SetClass(class string) {
	c.Class = class
	if saves, ok := ClassSavingThrows[class]; ok {
		c.SavingThrowProficiencies = saves
	}
}

// SetRace sets the race and updates related attributes
func (c *Character) SetRace(race string) {
	c.Race = race
	if speed, ok := RaceSpeed[race]; ok {
		c.Speed = speed
	}
}

// CalculateMaxHP calculates max HP for level 1
func (c *Character) CalculateMaxHP() int {
	hitDie := ClassHitDice[c.Class]
	if hitDie == 0 {
		hitDie = 8 // default
	}
	return hitDie + AbilityModifier(c.Constitution)
}

// InitializeHP sets HP to max
func (c *Character) InitializeHP() {
	c.MaxHitPoints = c.CalculateMaxHP()
	c.CurrentHitPoints = c.MaxHitPoints
}
