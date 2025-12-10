package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/brady1408/dnd/internal/character"
	"github.com/brady1408/dnd/internal/db"
	"github.com/brady1408/dnd/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SheetMode int

const (
	ModeView SheetMode = iota
	ModeEditHP
	ModeEditNotes
	ModeEditFeatures
)

type SheetScreen struct {
	ctx     context.Context
	queries *db.Queries
	char    db.Character

	mode       SheetMode
	tab        int // 0=stats, 1=skills, 2=combat, 3=notes
	width      int
	height     int

	// Edit mode inputs
	hpInput       textinput.Model
	notesInput    textarea.Model
	featuresInput textarea.Model
	editCursor    int
}

type CharacterUpdatedMsg struct {
	Character db.Character
}

func NewSheetScreen(ctx context.Context, queries *db.Queries, char db.Character) *SheetScreen {
	hpInput := textinput.New()
	hpInput.Placeholder = "HP"
	hpInput.Width = 10
	hpInput.CharLimit = 5

	notesInput := textarea.New()
	notesInput.Placeholder = "Enter notes here..."
	notesInput.SetWidth(50)
	notesInput.SetHeight(8)
	notesInput.CharLimit = 5000
	notesInput.ShowLineNumbers = false

	featuresInput := textarea.New()
	featuresInput.Placeholder = "Enter features & traits here..."
	featuresInput.SetWidth(50)
	featuresInput.SetHeight(8)
	featuresInput.CharLimit = 5000
	featuresInput.ShowLineNumbers = false

	return &SheetScreen{
		ctx:           ctx,
		queries:       queries,
		char:          char,
		mode:          ModeView,
		hpInput:       hpInput,
		notesInput:    notesInput,
		featuresInput: featuresInput,
		width:         80,
		height:        24,
	}
}

func (s *SheetScreen) Init() tea.Cmd {
	return nil
}

// SetCharacter updates the character data without resetting the view state
func (s *SheetScreen) SetCharacter(char db.Character) {
	s.char = char
}

func (s *SheetScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	}

	// Handle mode-specific updates
	switch s.mode {
	case ModeView:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return s.updateView(keyMsg)
		}
	case ModeEditHP:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return s.updateEditHP(keyMsg)
		}
	case ModeEditNotes:
		return s.updateEditNotes(msg)
	case ModeEditFeatures:
		return s.updateEditFeatures(msg)
	}

	return s, nil
}

func (s *SheetScreen) updateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "right", "l":
		s.tab = (s.tab + 1) % 4
	case "shift+tab", "left", "h":
		s.tab = (s.tab + 3) % 4

	case "e":
		if s.tab == 2 { // Combat tab - edit HP
			s.mode = ModeEditHP
			s.hpInput.SetValue(fmt.Sprintf("%d", s.char.CurrentHitPoints))
			s.hpInput.Focus()
			return s, textinput.Blink
		} else if s.tab == 3 { // Notes tab - edit notes
			s.mode = ModeEditNotes
			s.notesInput.SetValue(s.char.Notes)
			s.notesInput.Focus()
			return s, textarea.Blink
		}

	case "f":
		if s.tab == 3 { // Notes tab - edit features & traits
			s.mode = ModeEditFeatures
			s.featuresInput.SetValue(s.char.FeaturesTraits)
			s.featuresInput.Focus()
			return s, textarea.Blink
		}

	case "r":
		// Roll a d20
		roll := character.RollD20()
		// Display would need a message system
		_ = roll

	case "esc", "q":
		return s, func() tea.Msg { return NavigateBackMsg{} }
	}

	return s, nil
}

func (s *SheetScreen) updateEditHP(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		var hp int
		fmt.Sscanf(s.hpInput.Value(), "%d", &hp)
		if hp < 0 {
			hp = 0
		}
		if hp > int(s.char.MaxHitPoints) {
			hp = int(s.char.MaxHitPoints)
		}

		return s, s.updateHP(int32(hp))

	case "esc":
		s.mode = ModeView
		return s, nil
	}

	var cmd tea.Cmd
	s.hpInput, cmd = s.hpInput.Update(msg)
	return s, cmd
}

func (s *SheetScreen) updateEditNotes(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle special keys first
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+s":
			return s, s.updateNotes(s.notesInput.Value())
		case "esc":
			s.mode = ModeView
			return s, nil
		}
	}

	// Pass all other messages to textarea
	var cmd tea.Cmd
	s.notesInput, cmd = s.notesInput.Update(msg)
	return s, cmd
}

func (s *SheetScreen) updateHP(hp int32) tea.Cmd {
	return func() tea.Msg {
		updated, err := s.queries.UpdateCharacterHitPoints(s.ctx, db.UpdateCharacterHitPointsParams{
			ID:                 s.char.ID,
			CurrentHitPoints:   hp,
			TemporaryHitPoints: s.char.TemporaryHitPoints,
		})
		if err != nil {
			return nil
		}
		s.char = updated
		s.mode = ModeView
		return CharacterUpdatedMsg{Character: updated}
	}
}

func (s *SheetScreen) updateNotes(notes string) tea.Cmd {
	return func() tea.Msg {
		updated, err := s.queries.UpdateCharacterNotes(s.ctx, db.UpdateCharacterNotesParams{
			ID:             s.char.ID,
			FeaturesTraits: s.char.FeaturesTraits,
			Notes:          notes,
		})
		if err != nil {
			return nil
		}
		s.char = updated
		s.mode = ModeView
		return CharacterUpdatedMsg{Character: updated}
	}
}

func (s *SheetScreen) updateEditFeatures(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle special keys first
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+s":
			return s, s.updateFeatures(s.featuresInput.Value())
		case "esc":
			s.mode = ModeView
			return s, nil
		}
	}

	// Pass all other messages to textarea
	var cmd tea.Cmd
	s.featuresInput, cmd = s.featuresInput.Update(msg)
	return s, cmd
}

func (s *SheetScreen) updateFeatures(features string) tea.Cmd {
	return func() tea.Msg {
		updated, err := s.queries.UpdateCharacterNotes(s.ctx, db.UpdateCharacterNotesParams{
			ID:             s.char.ID,
			FeaturesTraits: features,
			Notes:          s.char.Notes,
		})
		if err != nil {
			return nil
		}
		s.char = updated
		s.mode = ModeView
		return CharacterUpdatedMsg{Character: updated}
	}
}

func (s *SheetScreen) View() string {
	var b strings.Builder

	// Header with character name
	header := fmt.Sprintf("%s - Level %d %s %s",
		s.char.Name, s.char.Level, s.char.Race, s.char.Class)
	b.WriteString(styles.Title.Render(header))
	b.WriteString("\n\n")

	// Tab bar
	tabs := []string{"Stats", "Skills", "Combat", "Notes"}
	tabBar := ""
	for i, t := range tabs {
		if i == s.tab {
			tabBar += styles.FocusedButton.Render(" " + t + " ")
		} else {
			tabBar += styles.Button.Render(" " + t + " ")
		}
	}
	b.WriteString(tabBar)
	b.WriteString("\n\n")

	// Tab content
	switch s.tab {
	case 0:
		b.WriteString(s.viewStats())
	case 1:
		b.WriteString(s.viewSkills())
	case 2:
		b.WriteString(s.viewCombat())
	case 3:
		b.WriteString(s.viewNotes())
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(styles.Help.Render(s.getHelp()))

	return lipgloss.Place(s.width, s.height,
		lipgloss.Center, lipgloss.Center,
		b.String())
}

func (s *SheetScreen) viewStats() string {
	var b strings.Builder

	// Ability scores
	abilities := []struct {
		name  string
		score int32
	}{
		{"Strength", s.char.Strength},
		{"Dexterity", s.char.Dexterity},
		{"Constitution", s.char.Constitution},
		{"Intelligence", s.char.Intelligence},
		{"Wisdom", s.char.Wisdom},
		{"Charisma", s.char.Charisma},
	}

	profBonus := character.ProficiencyBonus(int(s.char.Level))

	b.WriteString(styles.Header.Render("Ability Scores"))
	b.WriteString("\n\n")

	// Use fixed-width columns for alignment
	labelWidth := 14
	scoreWidth := 3
	modWidth := 4

	for _, a := range abilities {
		mod := character.AbilityModifier(int(a.score))
		// Pad the name manually before styling
		paddedName := fmt.Sprintf("%-*s", labelWidth, a.name)
		paddedScore := fmt.Sprintf("%*d", scoreWidth, a.score)
		paddedMod := fmt.Sprintf("%*s", modWidth, character.FormatModifierInt(mod))

		b.WriteString(styles.Muted.Render(paddedName))
		b.WriteString("  ")
		b.WriteString(styles.StatValue.Render(paddedScore))
		b.WriteString("  ")
		b.WriteString(styles.StatMod.Render(paddedMod))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.Header.Render("Saving Throws"))
	b.WriteString("\n\n")

	for _, a := range abilities {
		proficient := false
		for _, p := range s.char.SavingThrowProficiencies {
			if strings.EqualFold(p, a.name) {
				proficient = true
				break
			}
		}

		mod := character.SavingThrow(int(a.score), int(s.char.Level), proficient)
		profMark := "  "
		style := styles.NotProficient
		if proficient {
			profMark := "● "
			style = styles.Proficient
			paddedName := fmt.Sprintf("%-*s", labelWidth, a.name)
			paddedMod := fmt.Sprintf("%*s", modWidth, character.FormatModifierInt(mod))
			b.WriteString(style.Render(profMark + paddedName + "  " + paddedMod))
		} else {
			paddedName := fmt.Sprintf("%-*s", labelWidth, a.name)
			paddedMod := fmt.Sprintf("%*s", modWidth, character.FormatModifierInt(mod))
			b.WriteString(style.Render(profMark + paddedName + "  " + paddedMod))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("Proficiency Bonus: ")
	b.WriteString(styles.StatValue.Render(character.FormatModifierInt(profBonus)))
	b.WriteString("\n")

	return b.String()
}

func (s *SheetScreen) viewSkills() string {
	var b strings.Builder

	b.WriteString(styles.Header.Render("Skills"))
	b.WriteString("\n\n")

	abilities := map[string]int32{
		"strength":     s.char.Strength,
		"dexterity":    s.char.Dexterity,
		"constitution": s.char.Constitution,
		"intelligence": s.char.Intelligence,
		"wisdom":       s.char.Wisdom,
		"charisma":     s.char.Charisma,
	}

	skillWidth := 18
	modWidth := 4

	for _, skill := range character.SkillList {
		abilityName := character.Skills[skill]
		abilityScore := abilities[abilityName]

		proficient := false
		for _, p := range s.char.SkillProficiencies {
			if strings.EqualFold(p, skill) {
				proficient = true
				break
			}
		}

		mod := character.SkillBonus(int(abilityScore), int(s.char.Level), proficient)
		profMark := "  "
		style := styles.NotProficient
		if proficient {
			profMark = "● "
			style = styles.Proficient
		}

		// Abbreviate ability name
		abilityAbbr := strings.ToUpper(abilityName[:3])

		paddedSkill := fmt.Sprintf("%-*s", skillWidth, skill)
		paddedMod := fmt.Sprintf("%*s", modWidth, character.FormatModifierInt(mod))

		b.WriteString(style.Render(profMark + paddedSkill + "  " + paddedMod + "  (" + abilityAbbr + ")"))
		b.WriteString("\n")
	}

	return b.String()
}

func (s *SheetScreen) viewCombat() string {
	var b strings.Builder

	b.WriteString(styles.Header.Render("Combat"))
	b.WriteString("\n\n")

	// HP display
	hpPct := float64(s.char.CurrentHitPoints) / float64(s.char.MaxHitPoints)
	hpStyle := styles.HPCurrent
	if hpPct < 0.25 {
		hpStyle = styles.HPCritical
	} else if hpPct < 0.5 {
		hpStyle = styles.HPLow
	}

	// Right-align labels to align on the colon
	labelWidth := 14

	if s.mode == ModeEditHP {
		b.WriteString(fmt.Sprintf("%*s ", labelWidth, "Hit Points:"))
		b.WriteString(styles.FocusedInput.Render(s.hpInput.View()))
		b.WriteString(fmt.Sprintf(" / %d", s.char.MaxHitPoints))
	} else {
		b.WriteString(fmt.Sprintf("%*s ", labelWidth, "Hit Points:"))
		b.WriteString(hpStyle.Render(fmt.Sprintf("%d", s.char.CurrentHitPoints)))
		b.WriteString(" / ")
		b.WriteString(styles.HPMax.Render(fmt.Sprintf("%d", s.char.MaxHitPoints)))
	}

	if s.char.TemporaryHitPoints > 0 {
		b.WriteString(fmt.Sprintf(" (+%d temp)", s.char.TemporaryHitPoints))
	}
	b.WriteString("\n")

	// Other combat stats
	initiative := character.Initiative(int(s.char.Dexterity))

	b.WriteString(fmt.Sprintf("%*s ", labelWidth, "Armor Class:"))
	b.WriteString(styles.StatValue.Render(fmt.Sprintf("%d", s.char.ArmorClass)))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("%*s ", labelWidth, "Initiative:"))
	b.WriteString(styles.StatValue.Render(character.FormatModifierInt(initiative)))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("%*s ", labelWidth, "Speed:"))
	b.WriteString(styles.StatValue.Render(fmt.Sprintf("%d", s.char.Speed)))
	b.WriteString(" ft\n")

	// Hit dice
	hitDie := character.ClassHitDice[s.char.Class]
	b.WriteString(fmt.Sprintf("%*s %dd%d\n", labelWidth, "Hit Dice:", s.char.Level, hitDie))

	b.WriteString("\n")
	b.WriteString(styles.Header.Render("Quick Rolls"))
	b.WriteString("\n\n")

	// Attack bonus examples
	strMod := character.AbilityModifier(int(s.char.Strength))
	dexMod := character.AbilityModifier(int(s.char.Dexterity))
	profBonus := character.ProficiencyBonus(int(s.char.Level))

	b.WriteString(fmt.Sprintf("%*s ", labelWidth, "Melee Attack:"))
	b.WriteString(styles.StatValue.Render(character.FormatModifierInt(strMod + profBonus)))
	b.WriteString(" (STR + Prof)\n")

	b.WriteString(fmt.Sprintf("%*s ", labelWidth, "Ranged Attack:"))
	b.WriteString(styles.StatValue.Render(character.FormatModifierInt(dexMod + profBonus)))
	b.WriteString(" (DEX + Prof)\n")

	// Wrap in a left-aligned box so the colon alignment works
	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		Render(b.String())
}

func (s *SheetScreen) viewNotes() string {
	var b strings.Builder

	b.WriteString(styles.Header.Render("Features & Traits"))
	b.WriteString("\n\n")

	if s.mode == ModeEditFeatures {
		b.WriteString(styles.FocusedInput.Render(s.featuresInput.View()))
	} else if s.char.FeaturesTraits != "" {
		b.WriteString(s.char.FeaturesTraits)
	} else {
		b.WriteString(styles.Muted.Render("No features or traits recorded."))
	}
	b.WriteString("\n\n")

	b.WriteString(styles.Header.Render("Notes"))
	b.WriteString("\n\n")

	if s.mode == ModeEditNotes {
		b.WriteString(styles.FocusedInput.Render(s.notesInput.View()))
	} else if s.char.Notes != "" {
		b.WriteString(s.char.Notes)
	} else {
		b.WriteString(styles.Muted.Render("No notes recorded."))
	}

	return b.String()
}

func (s *SheetScreen) getHelp() string {
	switch s.mode {
	case ModeEditHP:
		return "enter: save • esc: cancel"
	case ModeEditNotes, ModeEditFeatures:
		return "ctrl+s: save • esc: cancel"
	default:
		help := "tab/←→: switch tabs • q/esc: back"
		if s.tab == 2 {
			help += " • e: edit HP"
		} else if s.tab == 3 {
			help += " • e: edit notes • f: edit features"
		}
		return help
	}
}
