package screens

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brady1408/dnd/internal/character"
	"github.com/brady1408/dnd/internal/db"
	"github.com/brady1408/dnd/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateStep int

const (
	StepBasicInfo CreateStep = iota
	StepRace
	StepClass
	StepAbilityMethod
	StepAbilityRoll
	StepAbilityArray
	StepAbilityPointBuy
	StepSkills
	StepReview
)

type CreateScreen struct {
	ctx     context.Context
	queries *db.Queries
	userID  pgtype.UUID

	step       CreateStep
	width      int
	height     int
	err        string

	// Basic info
	nameInput       textinput.Model
	backgroundInput textinput.Model
	alignmentIndex  int

	// Race & Class
	raceIndex  int
	classIndex int

	// Ability scores
	abilityMethodIndex int
	abilityRolls       character.AbilityRolls
	rolledScores       []int
	assignedScores     map[string]int
	assignIndex        int
	pointBuyState      *character.PointBuyState

	// Skills
	availableSkills   []string
	selectedSkills    []string
	skillsToSelect    int
	skillCursor       int
}

type CharacterCreatedMsg struct {
	Character db.Character
}

type NavigateBackMsg struct{}

func NewCreateScreen(ctx context.Context, queries *db.Queries, userID pgtype.UUID) *CreateScreen {
	nameInput := textinput.New()
	nameInput.Placeholder = "Character Name"
	nameInput.CharLimit = 100
	nameInput.Width = 30
	nameInput.Focus()

	bgInput := textinput.New()
	bgInput.Placeholder = "Background (optional)"
	bgInput.CharLimit = 50
	bgInput.Width = 30

	return &CreateScreen{
		ctx:            ctx,
		queries:        queries,
		userID:         userID,
		step:           StepBasicInfo,
		nameInput:      nameInput,
		backgroundInput: bgInput,
		assignedScores: make(map[string]int),
		width:          80,
		height:         24,
	}
}

func (c *CreateScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (c *CreateScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height

	case tea.KeyMsg:
		c.err = ""

		switch msg.String() {
		case "esc":
			if c.step == StepBasicInfo {
				return c, func() tea.Msg { return NavigateBackMsg{} }
			}
			c.previousStep()
			return c, nil

		case "ctrl+c":
			return c, tea.Quit
		}

		// Handle step-specific input
		switch c.step {
		case StepBasicInfo:
			return c.updateBasicInfo(msg)
		case StepRace:
			return c.updateRace(msg)
		case StepClass:
			return c.updateClass(msg)
		case StepAbilityMethod:
			return c.updateAbilityMethod(msg)
		case StepAbilityRoll:
			return c.updateAbilityRoll(msg)
		case StepAbilityArray:
			return c.updateAbilityArray(msg)
		case StepAbilityPointBuy:
			return c.updatePointBuy(msg)
		case StepSkills:
			return c.updateSkills(msg)
		case StepReview:
			return c.updateReview(msg)
		}
	}

	// Update text inputs
	var cmd tea.Cmd
	if c.step == StepBasicInfo {
		c.nameInput, cmd = c.nameInput.Update(msg)
	}

	return c, cmd
}

func (c *CreateScreen) previousStep() {
	switch c.step {
	case StepRace:
		c.step = StepBasicInfo
		c.nameInput.Focus()
	case StepClass:
		c.step = StepRace
	case StepAbilityMethod:
		c.step = StepClass
	case StepAbilityRoll, StepAbilityArray, StepAbilityPointBuy:
		c.step = StepAbilityMethod
	case StepSkills:
		// Go back to ability method selection
		c.step = StepAbilityMethod
	case StepReview:
		c.step = StepSkills
	}
}

func (c *CreateScreen) updateBasicInfo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "tab":
		if strings.TrimSpace(c.nameInput.Value()) == "" {
			c.err = "Name is required"
			return c, nil
		}
		c.step = StepRace
		c.nameInput.Blur()
		return c, nil

	case "up", "down":
		// Toggle between name and background inputs could be added here
	}

	var cmd tea.Cmd
	c.nameInput, cmd = c.nameInput.Update(msg)
	return c, cmd
}

func (c *CreateScreen) updateRace(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if c.raceIndex > 0 {
			c.raceIndex--
		}
	case "down", "j":
		if c.raceIndex < len(character.Races)-1 {
			c.raceIndex++
		}
	case "enter":
		c.step = StepClass
	}
	return c, nil
}

func (c *CreateScreen) updateClass(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if c.classIndex > 0 {
			c.classIndex--
		}
	case "down", "j":
		if c.classIndex < len(character.Classes)-1 {
			c.classIndex++
		}
	case "enter":
		c.step = StepAbilityMethod
	}
	return c, nil
}

func (c *CreateScreen) updateAbilityMethod(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	methods := []string{"Roll 4d6 (drop lowest)", "Standard Array", "Point Buy"}

	switch msg.String() {
	case "up", "k":
		if c.abilityMethodIndex > 0 {
			c.abilityMethodIndex--
		}
	case "down", "j":
		if c.abilityMethodIndex < len(methods)-1 {
			c.abilityMethodIndex++
		}
	case "enter":
		switch c.abilityMethodIndex {
		case 0:
			c.abilityRolls = character.RollAbilityScores()
			c.rolledScores = make([]int, len(c.abilityRolls.Totals))
			copy(c.rolledScores, c.abilityRolls.Totals)
			c.assignedScores = make(map[string]int)
			c.assignIndex = 0
			c.step = StepAbilityRoll
		case 1:
			c.rolledScores = character.GetStandardArray()
			c.assignedScores = make(map[string]int)
			c.assignIndex = 0
			c.step = StepAbilityArray
		case 2:
			c.pointBuyState = character.NewPointBuyState()
			c.assignIndex = 0
			c.step = StepAbilityPointBuy
		}
	}
	return c, nil
}

func (c *CreateScreen) updateAbilityRoll(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return c.updateAbilityAssignment(msg)
}

func (c *CreateScreen) updateAbilityArray(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	return c.updateAbilityAssignment(msg)
}

func (c *CreateScreen) updateAbilityAssignment(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if c.assignIndex > 0 {
			c.assignIndex--
		}
	case "down", "j":
		if c.assignIndex < len(character.Abilities)-1 {
			c.assignIndex++
		}
	case "1", "2", "3", "4", "5", "6":
		scoreIdx := int(msg.String()[0] - '1')
		if scoreIdx < len(c.rolledScores) {
			ability := character.Abilities[c.assignIndex]

			// Check if this score is already assigned
			for ab, idx := range c.assignedScores {
				if idx == scoreIdx {
					// Unassign from other ability
					delete(c.assignedScores, ab)
					break
				}
			}

			c.assignedScores[ability] = scoreIdx
		}
	case "enter":
		if len(c.assignedScores) == 6 {
			c.setupSkillSelection()
			c.step = StepSkills
		} else {
			c.err = "Please assign all 6 ability scores"
		}
	case "r":
		// Re-roll (only for roll method)
		if c.step == StepAbilityRoll {
			c.abilityRolls = character.RollAbilityScores()
			c.rolledScores = make([]int, len(c.abilityRolls.Totals))
			copy(c.rolledScores, c.abilityRolls.Totals)
			c.assignedScores = make(map[string]int)
		}
	}
	return c, nil
}

func (c *CreateScreen) updatePointBuy(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if c.assignIndex > 0 {
			c.assignIndex--
		}
	case "down", "j":
		if c.assignIndex < len(character.Abilities)-1 {
			c.assignIndex++
		}
	case "right", "l", "+", "=":
		ability := character.Abilities[c.assignIndex]
		c.pointBuyState.Increase(ability)
	case "left", "h", "-":
		ability := character.Abilities[c.assignIndex]
		c.pointBuyState.Decrease(ability)
	case "enter":
		c.setupSkillSelection()
		c.step = StepSkills
	}
	return c, nil
}

func (c *CreateScreen) setupSkillSelection() {
	className := character.Classes[c.classIndex]
	if choice, ok := character.ClassSkillChoices[className]; ok {
		c.availableSkills = choice.Options
		c.skillsToSelect = choice.Count
	} else {
		c.availableSkills = character.SkillList
		c.skillsToSelect = 2
	}
	c.selectedSkills = []string{}
	c.skillCursor = 0
}

func (c *CreateScreen) updateSkills(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if c.skillCursor > 0 {
			c.skillCursor--
		}
	case "down", "j":
		if c.skillCursor < len(c.availableSkills)-1 {
			c.skillCursor++
		}
	case " ", "x":
		skill := c.availableSkills[c.skillCursor]
		// Toggle selection
		found := false
		for i, s := range c.selectedSkills {
			if s == skill {
				c.selectedSkills = append(c.selectedSkills[:i], c.selectedSkills[i+1:]...)
				found = true
				break
			}
		}
		if !found && len(c.selectedSkills) < c.skillsToSelect {
			c.selectedSkills = append(c.selectedSkills, skill)
		}
	case "enter":
		if len(c.selectedSkills) == c.skillsToSelect {
			c.step = StepReview
		} else {
			c.err = fmt.Sprintf("Please select %d skills", c.skillsToSelect)
		}
	}
	return c, nil
}

func (c *CreateScreen) updateReview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "y":
		return c, c.createCharacter()
	case "n":
		c.step = StepBasicInfo
		c.nameInput.Focus()
	}
	return c, nil
}

func (c *CreateScreen) createCharacter() tea.Cmd {
	return func() tea.Msg {
		// Build character
		char := character.NewCharacter()
		char.Name = strings.TrimSpace(c.nameInput.Value())
		char.SetRace(character.Races[c.raceIndex])
		char.SetClass(character.Classes[c.classIndex])
		char.Background = character.Backgrounds[0] // Default
		if bg := strings.TrimSpace(c.backgroundInput.Value()); bg != "" {
			char.Background = bg
		}
		char.Alignment = character.Alignments[c.alignmentIndex]

		// Set ability scores
		if c.step == StepReview {
			if c.pointBuyState != nil {
				scores := c.pointBuyState.GetScores()
				char.Strength = scores[0]
				char.Dexterity = scores[1]
				char.Constitution = scores[2]
				char.Intelligence = scores[3]
				char.Wisdom = scores[4]
				char.Charisma = scores[5]
			} else {
				for i, ability := range character.Abilities {
					if scoreIdx, ok := c.assignedScores[ability]; ok {
						score := c.rolledScores[scoreIdx]
						switch i {
						case 0:
							char.Strength = score
						case 1:
							char.Dexterity = score
						case 2:
							char.Constitution = score
						case 3:
							char.Intelligence = score
						case 4:
							char.Wisdom = score
						case 5:
							char.Charisma = score
						}
					}
				}
			}
		}

		char.SkillProficiencies = c.selectedSkills
		char.InitializeHP()

		// Save to database
		equipmentJSON, _ := json.Marshal(char.Equipment)

		dbChar, err := c.queries.CreateCharacter(c.ctx, db.CreateCharacterParams{
			UserID:                   c.userID,
			Name:                     char.Name,
			Class:                    char.Class,
			Level:                    int32(char.Level),
			Race:                     char.Race,
			Background:               pgtype.Text{String: char.Background, Valid: char.Background != ""},
			Alignment:                pgtype.Text{String: char.Alignment, Valid: char.Alignment != ""},
			ExperiencePoints:         int32(char.ExperiencePoints),
			Strength:                 int32(char.Strength),
			Dexterity:                int32(char.Dexterity),
			Constitution:             int32(char.Constitution),
			Intelligence:             int32(char.Intelligence),
			Wisdom:                   int32(char.Wisdom),
			Charisma:                 int32(char.Charisma),
			MaxHitPoints:             int32(char.MaxHitPoints),
			CurrentHitPoints:         int32(char.CurrentHitPoints),
			TemporaryHitPoints:       int32(char.TemporaryHitPoints),
			ArmorClass:               int32(char.ArmorClass),
			Speed:                    int32(char.Speed),
			SavingThrowProficiencies: char.SavingThrowProficiencies,
			SkillProficiencies:       char.SkillProficiencies,
			Equipment:                equipmentJSON,
			FeaturesTraits:           char.FeaturesTraits,
			Notes:                    char.Notes,
		})

		if err != nil {
			return nil // Handle error
		}

		return CharacterCreatedMsg{Character: dbChar}
	}
}

func (c *CreateScreen) View() string {
	var b strings.Builder

	// Progress indicator
	steps := []string{"Info", "Race", "Class", "Abilities", "Skills", "Review"}
	stepIdx := c.currentStepIndex()
	progress := ""
	for i, s := range steps {
		if i == stepIdx {
			progress += styles.Selected.Render("[" + s + "]")
		} else if i < stepIdx {
			progress += styles.SuccessText.Render("✓" + s)
		} else {
			progress += styles.Muted.Render(" " + s + " ")
		}
		if i < len(steps)-1 {
			progress += " → "
		}
	}
	b.WriteString(progress)
	b.WriteString("\n\n")

	// Step content
	switch c.step {
	case StepBasicInfo:
		b.WriteString(c.viewBasicInfo())
	case StepRace:
		b.WriteString(c.viewRace())
	case StepClass:
		b.WriteString(c.viewClass())
	case StepAbilityMethod:
		b.WriteString(c.viewAbilityMethod())
	case StepAbilityRoll, StepAbilityArray:
		b.WriteString(c.viewAbilityAssignment())
	case StepAbilityPointBuy:
		b.WriteString(c.viewPointBuy())
	case StepSkills:
		b.WriteString(c.viewSkills())
	case StepReview:
		b.WriteString(c.viewReview())
	}

	// Error
	if c.err != "" {
		b.WriteString("\n")
		b.WriteString(styles.ErrorText.Render("Error: " + c.err))
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(styles.Help.Render(c.getHelp()))

	return lipgloss.Place(c.width, c.height,
		lipgloss.Center, lipgloss.Center,
		b.String())
}

func (c *CreateScreen) currentStepIndex() int {
	switch c.step {
	case StepBasicInfo:
		return 0
	case StepRace:
		return 1
	case StepClass:
		return 2
	case StepAbilityMethod, StepAbilityRoll, StepAbilityArray, StepAbilityPointBuy:
		return 3
	case StepSkills:
		return 4
	case StepReview:
		return 5
	}
	return 0
}

func (c *CreateScreen) viewBasicInfo() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Create Your Character"))
	b.WriteString("\n\n")

	b.WriteString("Name:\n")
	b.WriteString(styles.FocusedInput.Render(c.nameInput.View()))

	return b.String()
}

func (c *CreateScreen) viewRace() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Choose Your Race"))
	b.WriteString("\n\n")

	for i, race := range character.Races {
		cursor := "  "
		style := styles.Unselected
		if i == c.raceIndex {
			cursor = "> "
			style = styles.Selected
		}
		speed := character.RaceSpeed[race]
		b.WriteString(styles.Cursor.Render(cursor))
		b.WriteString(style.Render(fmt.Sprintf("%-12s (Speed: %d)", race, speed)))
		b.WriteString("\n")
	}

	return b.String()
}

func (c *CreateScreen) viewClass() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Choose Your Class"))
	b.WriteString("\n\n")

	for i, class := range character.Classes {
		cursor := "  "
		style := styles.Unselected
		if i == c.classIndex {
			cursor = "> "
			style = styles.Selected
		}
		hitDie := character.ClassHitDice[class]
		b.WriteString(styles.Cursor.Render(cursor))
		b.WriteString(style.Render(fmt.Sprintf("%-12s (Hit Die: d%d)", class, hitDie)))
		b.WriteString("\n")
	}

	return b.String()
}

func (c *CreateScreen) viewAbilityMethod() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Choose Ability Score Method"))
	b.WriteString("\n\n")

	methods := []struct {
		name string
		desc string
	}{
		{"Roll 4d6 (drop lowest)", "Roll 4d6, drop the lowest, 6 times"},
		{"Standard Array", "Use 15, 14, 13, 12, 10, 8"},
		{"Point Buy", "27 points to spend (scores 8-15)"},
	}

	for i, m := range methods {
		cursor := "  "
		style := styles.Unselected
		if i == c.abilityMethodIndex {
			cursor = "> "
			style = styles.Selected
		}
		b.WriteString(styles.Cursor.Render(cursor))
		b.WriteString(style.Render(m.name))
		b.WriteString("\n")
		b.WriteString(styles.Muted.Render("    " + m.desc))
		b.WriteString("\n")
	}

	return b.String()
}

func (c *CreateScreen) viewAbilityAssignment() string {
	var b strings.Builder

	title := "Assign Your Ability Scores"
	if c.step == StepAbilityRoll {
		title = "Assign Your Rolled Scores"
	}
	b.WriteString(styles.Title.Render(title))
	b.WriteString("\n\n")

	// Show available scores
	b.WriteString("Available scores: ")
	for i, score := range c.rolledScores {
		used := false
		for _, idx := range c.assignedScores {
			if idx == i {
				used = true
				break
			}
		}
		if used {
			b.WriteString(styles.Muted.Render(fmt.Sprintf("[%d]=%d ", i+1, score)))
		} else {
			b.WriteString(styles.SuccessText.Render(fmt.Sprintf("[%d]=%d ", i+1, score)))
		}
	}
	b.WriteString("\n\n")

	// Show abilities
	for i, ability := range character.Abilities {
		cursor := "  "
		style := styles.Unselected
		if i == c.assignIndex {
			cursor = "> "
			style = styles.Selected
		}

		scoreStr := "___"
		if scoreIdx, ok := c.assignedScores[ability]; ok {
			score := c.rolledScores[scoreIdx]
			mod := character.AbilityModifier(score)
			scoreStr = fmt.Sprintf("%2d (%s)", score, character.FormatModifierInt(mod))
		}

		b.WriteString(styles.Cursor.Render(cursor))
		b.WriteString(style.Render(fmt.Sprintf("%-14s: %s", ability, scoreStr)))
		b.WriteString("\n")
	}

	return b.String()
}

func (c *CreateScreen) viewPointBuy() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Point Buy"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Points remaining: %s\n\n",
		styles.StatValue.Render(fmt.Sprintf("%d", c.pointBuyState.PointsRemaining))))

	for i, ability := range character.Abilities {
		cursor := "  "
		style := styles.Unselected
		if i == c.assignIndex {
			cursor = "> "
			style = styles.Selected
		}

		score := c.pointBuyState.Scores[ability]
		mod := character.AbilityModifier(score)
		cost := character.PointBuyCosts[score]

		canInc := c.pointBuyState.CanIncrease(ability)
		canDec := c.pointBuyState.CanDecrease(ability)

		arrows := ""
		if canDec {
			arrows += "◀ "
		} else {
			arrows += "  "
		}
		if canInc {
			arrows += " ▶"
		}

		b.WriteString(styles.Cursor.Render(cursor))
		b.WriteString(style.Render(fmt.Sprintf("%-14s: %2d (%s) cost:%d %s",
			ability, score, character.FormatModifierInt(mod), cost, arrows)))
		b.WriteString("\n")
	}

	return b.String()
}

func (c *CreateScreen) viewSkills() string {
	var b strings.Builder

	className := character.Classes[c.classIndex]
	b.WriteString(styles.Title.Render(fmt.Sprintf("Choose %d Skills (%s)", c.skillsToSelect, className)))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Selected: %d/%d\n\n", len(c.selectedSkills), c.skillsToSelect))

	for i, skill := range c.availableSkills {
		cursor := "  "
		style := styles.Unselected
		if i == c.skillCursor {
			cursor = "> "
			style = styles.Selected
		}

		checkbox := "[ ]"
		for _, s := range c.selectedSkills {
			if s == skill {
				checkbox = "[x]"
				break
			}
		}

		b.WriteString(styles.Cursor.Render(cursor))
		b.WriteString(style.Render(fmt.Sprintf("%s %s", checkbox, skill)))
		b.WriteString("\n")
	}

	return b.String()
}

func (c *CreateScreen) viewReview() string {
	var b strings.Builder

	b.WriteString(styles.Title.Render("Review Your Character"))
	b.WriteString("\n\n")

	// Basic info
	b.WriteString(styles.Header.Render("Basic Info"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Name:       %s\n", c.nameInput.Value()))
	b.WriteString(fmt.Sprintf("Race:       %s\n", character.Races[c.raceIndex]))
	b.WriteString(fmt.Sprintf("Class:      %s\n", character.Classes[c.classIndex]))
	b.WriteString("\n")

	// Abilities
	b.WriteString(styles.Header.Render("Ability Scores"))
	b.WriteString("\n")

	for i, ability := range character.Abilities {
		var score int
		if c.pointBuyState != nil {
			score = c.pointBuyState.Scores[ability]
		} else if scoreIdx, ok := c.assignedScores[ability]; ok {
			score = c.rolledScores[scoreIdx]
		}
		mod := character.AbilityModifier(score)
		b.WriteString(fmt.Sprintf("%-14s: %2d (%s)\n", ability, score, character.FormatModifierInt(mod)))
		_ = i
	}
	b.WriteString("\n")

	// Skills
	b.WriteString(styles.Header.Render("Skill Proficiencies"))
	b.WriteString("\n")
	for _, skill := range c.selectedSkills {
		b.WriteString(fmt.Sprintf("  • %s\n", skill))
	}
	b.WriteString("\n")

	b.WriteString(styles.SuccessText.Render("Create this character? (y/n)"))

	return b.String()
}

func (c *CreateScreen) getHelp() string {
	switch c.step {
	case StepBasicInfo:
		return "enter: continue • esc: back"
	case StepRace, StepClass, StepAbilityMethod:
		return "↑/↓: select • enter: confirm • esc: back"
	case StepAbilityRoll:
		return "↑/↓: select ability • 1-6: assign score • r: re-roll • enter: confirm • esc: back"
	case StepAbilityArray:
		return "↑/↓: select ability • 1-6: assign score • enter: confirm • esc: back"
	case StepAbilityPointBuy:
		return "↑/↓: select • ←/→: adjust • enter: confirm • esc: back"
	case StepSkills:
		return "↑/↓: navigate • space: toggle • enter: confirm • esc: back"
	case StepReview:
		return "y: create • n: start over • esc: back"
	}
	return ""
}
