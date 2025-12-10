package character

import "strings"

// AbilityModifier calculates the modifier for an ability score
func AbilityModifier(score int) int {
	return (score - 10) / 2
}

// ProficiencyBonus returns the proficiency bonus for a given level
func ProficiencyBonus(level int) int {
	if level < 1 {
		return 2
	}
	return (level-1)/4 + 2
}

// SavingThrow calculates a saving throw bonus
func SavingThrow(abilityScore int, level int, proficient bool) int {
	bonus := AbilityModifier(abilityScore)
	if proficient {
		bonus += ProficiencyBonus(level)
	}
	return bonus
}

// SkillBonus calculates a skill bonus
func SkillBonus(abilityScore int, level int, proficient bool) int {
	bonus := AbilityModifier(abilityScore)
	if proficient {
		bonus += ProficiencyBonus(level)
	}
	return bonus
}

// Initiative calculates initiative bonus (just DEX modifier)
func Initiative(dexterity int) int {
	return AbilityModifier(dexterity)
}

// PassivePerception calculates passive perception
func PassivePerception(wisdom int, level int, proficient bool) int {
	return 10 + SkillBonus(wisdom, level, proficient)
}

// FormatModifier formats a modifier with +/- sign
func FormatModifier(mod int) string {
	if mod >= 0 {
		return "+" + strings.TrimPrefix(string(rune('0'+mod)), "0")
	}
	return string(rune('0'-mod))
}

// FormatModifierInt formats an int modifier with +/- sign
func FormatModifierInt(mod int) string {
	if mod >= 0 {
		return "+" + itoa(mod)
	}
	return itoa(mod)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	negative := i < 0
	if negative {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if negative {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// GetAbilityScore returns the ability score for a character by name
func (c *Character) GetAbilityScore(ability string) int {
	switch strings.ToLower(ability) {
	case "strength":
		return c.Strength
	case "dexterity":
		return c.Dexterity
	case "constitution":
		return c.Constitution
	case "intelligence":
		return c.Intelligence
	case "wisdom":
		return c.Wisdom
	case "charisma":
		return c.Charisma
	default:
		return 10
	}
}

// GetSkillBonus returns the skill bonus for a character
func (c *Character) GetSkillBonus(skill string) int {
	ability := Skills[skill]
	score := c.GetAbilityScore(ability)
	proficient := contains(c.SkillProficiencies, skill)
	return SkillBonus(score, c.Level, proficient)
}

// GetSavingThrow returns the saving throw bonus for a character
func (c *Character) GetSavingThrow(ability string) int {
	score := c.GetAbilityScore(ability)
	proficient := contains(c.SavingThrowProficiencies, ability)
	return SavingThrow(score, c.Level, proficient)
}

// GetInitiative returns the initiative bonus
func (c *Character) GetInitiative() int {
	return Initiative(c.Dexterity)
}

// GetProficiencyBonus returns the proficiency bonus
func (c *Character) GetProficiencyBonus() int {
	return ProficiencyBonus(c.Level)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// XPThresholds maps level to XP required
var XPThresholds = map[int]int{
	1:  0,
	2:  300,
	3:  900,
	4:  2700,
	5:  6500,
	6:  14000,
	7:  23000,
	8:  34000,
	9:  48000,
	10: 64000,
	11: 85000,
	12: 100000,
	13: 120000,
	14: 140000,
	15: 165000,
	16: 195000,
	17: 225000,
	18: 265000,
	19: 305000,
	20: 355000,
}

// LevelFromXP returns the level for a given XP amount
func LevelFromXP(xp int) int {
	level := 1
	for l := 20; l >= 1; l-- {
		if xp >= XPThresholds[l] {
			level = l
			break
		}
	}
	return level
}

// XPToNextLevel returns XP needed for next level
func XPToNextLevel(currentXP int) int {
	currentLevel := LevelFromXP(currentXP)
	if currentLevel >= 20 {
		return 0
	}
	return XPThresholds[currentLevel+1] - currentXP
}
