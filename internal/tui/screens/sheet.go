package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/brady1408/dnd/internal/character"
	"github.com/brady1408/dnd/internal/db"
	"github.com/brady1408/dnd/internal/tui/components"
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
	styles  *styles.Styles

	mode       SheetMode
	tab        int // 0=core, 1=combat, 2=spells, 3=inventory, 4=notes
	width      int
	height     int

	// Edit mode inputs
	hpInput       textinput.Model
	notesInput    textarea.Model
	featuresInput textarea.Model
	editCursor    int

	// Table components
	skillsTable     *components.TableModel
	attacksTable    *components.TableModel
	actionsTable    *components.TableModel
	inventoryTable  *components.TableModel
	magicItemsTable *components.TableModel
	spellsTable     *components.TableModel

	// Combat tab focus: 0=stats panel, 1=attacks, 2=actions
	combatFocus int

	// Inventory tab focus: 0=currency, 1=equipment, 2=magic items
	inventoryFocus int

	// Spells tab: which spell level is selected (0=cantrips, 1-9=spell levels)
	spellLevelFilter int

	// Cached data from DB
	attacks      []db.CharacterAttack
	actions      []db.CharacterAction
	inventory    []db.CharacterInventory
	magicItems   []db.CharacterMagicItem
	currency     *db.CharacterCurrency
	spellcasting *db.CharacterSpellcasting
	spells       []db.CharacterSpell
}

type CharacterUpdatedMsg struct {
	Character db.Character
}

func NewSheetScreen(ctx context.Context, queries *db.Queries, char db.Character, s *styles.Styles) *SheetScreen {
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

	// Create skills table
	skillsTable := components.NewTable([]components.TableColumn{
		{Title: "Prof", Width: 4},
		{Title: "Skill", Width: 18},
		{Title: "Mod", Width: 5},
		{Title: "Ability", Width: 7},
	}, s)
	skillsTable.SetVisibleRows(12)
	skillsTable.SetEmptyMessage("No skills available")

	// Create attacks table
	attacksTable := components.NewTable([]components.TableColumn{
		{Title: "Weapon", Width: 15},
		{Title: "Atk", Width: 5},
		{Title: "Damage", Width: 12},
		{Title: "Type", Width: 10},
		{Title: "Range", Width: 8},
	}, s)
	attacksTable.SetVisibleRows(5)
	attacksTable.SetEmptyMessage("No attacks - press 'a' to add")

	// Create actions table
	actionsTable := components.NewTable([]components.TableColumn{
		{Title: "Action", Width: 18},
		{Title: "Type", Width: 8},
		{Title: "Uses", Width: 8},
		{Title: "Source", Width: 12},
	}, s)
	actionsTable.SetVisibleRows(5)
	actionsTable.SetEmptyMessage("No actions - press 'a' to add")

	// Create inventory table
	inventoryTable := components.NewTable([]components.TableColumn{
		{Title: "Item", Width: 20},
		{Title: "Qty", Width: 4},
		{Title: "Wt", Width: 6},
		{Title: "Location", Width: 12},
		{Title: "Eq", Width: 3},
	}, s)
	inventoryTable.SetVisibleRows(8)
	inventoryTable.SetEmptyMessage("No items - press 'a' to add")

	// Create magic items table
	magicItemsTable := components.NewTable([]components.TableColumn{
		{Title: "Item", Width: 22},
		{Title: "Rarity", Width: 10},
		{Title: "Att", Width: 4},
	}, s)
	magicItemsTable.SetVisibleRows(5)
	magicItemsTable.SetEmptyMessage("No magic items - press 'a' to add")

	// Create spells table
	spellsTable := components.NewTable([]components.TableColumn{
		{Title: "P", Width: 2},
		{Title: "Spell", Width: 20},
		{Title: "School", Width: 10},
		{Title: "Time", Width: 8},
		{Title: "Range", Width: 8},
	}, s)
	spellsTable.SetVisibleRows(10)
	spellsTable.SetEmptyMessage("No spells known")

	sheet := &SheetScreen{
		ctx:             ctx,
		queries:         queries,
		char:            char,
		styles:          s,
		mode:            ModeView,
		hpInput:         hpInput,
		notesInput:      notesInput,
		featuresInput:   featuresInput,
		width:           80,
		height:          24,
		skillsTable:     skillsTable,
		attacksTable:    attacksTable,
		actionsTable:    actionsTable,
		inventoryTable:  inventoryTable,
		magicItemsTable: magicItemsTable,
		spellsTable:     spellsTable,
	}

	// Populate tables
	sheet.refreshSkillsTable()
	sheet.refreshAttacksTable()
	sheet.refreshActionsTable()
	sheet.refreshInventoryTable()
	sheet.refreshMagicItemsTable()
	sheet.refreshCurrency()
	sheet.refreshSpellcasting()
	sheet.refreshSpellsTable()

	return sheet
}

// refreshSkillsTable populates the skills table with current character data
func (s *SheetScreen) refreshSkillsTable() {
	abilities := map[string]int32{
		"strength":     s.char.Strength,
		"dexterity":    s.char.Dexterity,
		"constitution": s.char.Constitution,
		"intelligence": s.char.Intelligence,
		"wisdom":       s.char.Wisdom,
		"charisma":     s.char.Charisma,
	}

	var rows []components.TableRow
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
		if proficient {
			profMark = "●"
		}

		abilityAbbr := strings.ToUpper(abilityName[:3])

		rows = append(rows, components.TableRow{
			ID:    skill,
			Cells: []string{profMark, skill, character.FormatModifierInt(mod), abilityAbbr},
			Data:  skill,
		})
	}

	s.skillsTable.SetRows(rows)
}

// refreshAttacksTable loads attacks from DB and populates the table
func (s *SheetScreen) refreshAttacksTable() {
	attacks, err := s.queries.GetCharacterAttacks(s.ctx, s.char.ID)
	if err != nil {
		s.attacks = nil
		s.attacksTable.SetRows(nil)
		return
	}

	s.attacks = attacks
	var rows []components.TableRow
	for _, atk := range attacks {
		atkBonus := ""
		if atk.AttackBonus.Valid {
			atkBonus = character.FormatModifierInt(int(atk.AttackBonus.Int32))
		}

		damage := ""
		if atk.Damage.Valid {
			damage = atk.Damage.String
		}

		damageType := ""
		if atk.DamageType.Valid {
			damageType = atk.DamageType.String
		}

		atkRange := ""
		if atk.Range.Valid {
			atkRange = atk.Range.String
		}

		rows = append(rows, components.TableRow{
			ID:    fmt.Sprintf("%x", atk.ID.Bytes),
			Cells: []string{atk.Name, atkBonus, damage, damageType, atkRange},
			Data:  atk,
		})
	}

	s.attacksTable.SetRows(rows)
}

// refreshActionsTable loads actions from DB and populates the table
func (s *SheetScreen) refreshActionsTable() {
	actions, err := s.queries.GetCharacterActions(s.ctx, s.char.ID)
	if err != nil {
		s.actions = nil
		s.actionsTable.SetRows(nil)
		return
	}

	s.actions = actions
	var rows []components.TableRow
	for _, act := range actions {
		actionType := "action"
		if act.ActionType.Valid {
			actionType = act.ActionType.String
		}

		uses := ""
		if act.UsesMax.Valid && act.UsesMax.Int32 > 0 {
			current := int32(0)
			if act.UsesCurrent.Valid {
				current = act.UsesCurrent.Int32
			}
			usesPer := ""
			if act.UsesPer.Valid {
				usesPer = act.UsesPer.String
			}
			uses = fmt.Sprintf("%d/%d %s", current, act.UsesMax.Int32, usesPer)
		}

		source := ""
		if act.Source.Valid {
			source = act.Source.String
		}

		rows = append(rows, components.TableRow{
			ID:    fmt.Sprintf("%x", act.ID.Bytes),
			Cells: []string{act.Name, actionType, uses, source},
			Data:  act,
		})
	}

	s.actionsTable.SetRows(rows)
}

// refreshInventoryTable loads inventory from DB and populates the table
func (s *SheetScreen) refreshInventoryTable() {
	inventory, err := s.queries.GetCharacterInventory(s.ctx, s.char.ID)
	if err != nil {
		s.inventory = nil
		s.inventoryTable.SetRows(nil)
		return
	}

	s.inventory = inventory
	var rows []components.TableRow
	for _, item := range inventory {
		qty := "1"
		if item.Quantity.Valid {
			qty = fmt.Sprintf("%d", item.Quantity.Int32)
		}

		weight := ""
		if item.Weight.Valid && item.Weight.Int != nil {
			weight = item.Weight.Int.String()
		}

		location := ""
		if item.Location.Valid {
			location = item.Location.String
		}

		equipped := " "
		if item.IsEquipped.Valid && item.IsEquipped.Bool {
			equipped = "●"
		}

		rows = append(rows, components.TableRow{
			ID:    fmt.Sprintf("%x", item.ID.Bytes),
			Cells: []string{item.Name, qty, weight, location, equipped},
			Data:  item,
		})
	}

	s.inventoryTable.SetRows(rows)
}

// refreshMagicItemsTable loads magic items from DB and populates the table
func (s *SheetScreen) refreshMagicItemsTable() {
	magicItems, err := s.queries.GetCharacterMagicItems(s.ctx, s.char.ID)
	if err != nil {
		s.magicItems = nil
		s.magicItemsTable.SetRows(nil)
		return
	}

	s.magicItems = magicItems
	var rows []components.TableRow
	for _, item := range magicItems {
		rarity := ""
		if item.Rarity.Valid {
			rarity = item.Rarity.String
		}

		attuned := " "
		if item.IsAttuned.Valid && item.IsAttuned.Bool {
			attuned = "●"
		} else if item.AttunementRequired.Valid && item.AttunementRequired.Bool {
			attuned = "○"
		}

		rows = append(rows, components.TableRow{
			ID:    fmt.Sprintf("%x", item.ID.Bytes),
			Cells: []string{item.Name, rarity, attuned},
			Data:  item,
		})
	}

	s.magicItemsTable.SetRows(rows)
}

// refreshCurrency loads currency from DB
func (s *SheetScreen) refreshCurrency() {
	currency, err := s.queries.GetCharacterCurrency(s.ctx, s.char.ID)
	if err != nil {
		s.currency = nil
		return
	}
	s.currency = &currency
}

// refreshSpellcasting loads spellcasting info from DB
func (s *SheetScreen) refreshSpellcasting() {
	spellcasting, err := s.queries.GetCharacterSpellcasting(s.ctx, s.char.ID)
	if err != nil {
		s.spellcasting = nil
		return
	}
	s.spellcasting = &spellcasting
}

// refreshSpellsTable loads spells from DB and populates the table
func (s *SheetScreen) refreshSpellsTable() {
	// Get spells filtered by level if filter is set
	var spells []db.CharacterSpell
	var err error

	if s.spellLevelFilter >= 0 {
		spells, err = s.queries.GetCharacterSpellsByLevel(s.ctx, db.GetCharacterSpellsByLevelParams{
			CharacterID: s.char.ID,
			Level:       int32(s.spellLevelFilter),
		})
	} else {
		spells, err = s.queries.GetCharacterSpells(s.ctx, s.char.ID)
	}

	if err != nil {
		s.spells = nil
		s.spellsTable.SetRows(nil)
		return
	}

	s.spells = spells
	var rows []components.TableRow
	for _, spell := range spells {
		prepared := " "
		if spell.IsPrepared.Valid && spell.IsPrepared.Bool {
			prepared = "●"
		}

		school := ""
		if spell.School.Valid {
			school = spell.School.String
		}

		castingTime := ""
		if spell.CastingTime.Valid {
			castingTime = spell.CastingTime.String
		}

		spellRange := ""
		if spell.Range.Valid {
			spellRange = spell.Range.String
		}

		rows = append(rows, components.TableRow{
			ID:    fmt.Sprintf("%x", spell.ID.Bytes),
			Cells: []string{prepared, spell.Name, school, castingTime, spellRange},
			Data:  spell,
		})
	}

	s.spellsTable.SetRows(rows)
}

func (s *SheetScreen) Init() tea.Cmd {
	return nil
}

// SetCharacter updates the character data without resetting the view state
func (s *SheetScreen) SetCharacter(char db.Character) {
	s.char = char
	s.refreshSkillsTable()
	s.refreshAttacksTable()
	s.refreshActionsTable()
	s.refreshInventoryTable()
	s.refreshMagicItemsTable()
	s.refreshCurrency()
	s.refreshSpellcasting()
	s.refreshSpellsTable()
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
	// Handle Core tab (skills table) navigation
	if s.tab == 0 {
		switch msg.String() {
		case "up", "down", "j", "k", "pgup", "pgdown", "home", "end", "g", "G":
			var cmd tea.Cmd
			s.skillsTable, cmd = s.skillsTable.Update(msg)
			return s, cmd
		}
	}

	// Handle Combat tab navigation
	if s.tab == 1 {
		switch msg.String() {
		case "up", "down", "j", "k", "pgup", "pgdown", "home", "end", "g", "G":
			// Pass to focused table
			if s.combatFocus == 1 {
				var cmd tea.Cmd
				s.attacksTable, cmd = s.attacksTable.Update(msg)
				return s, cmd
			} else if s.combatFocus == 2 {
				var cmd tea.Cmd
				s.actionsTable, cmd = s.actionsTable.Update(msg)
				return s, cmd
			}
		case "1":
			s.combatFocus = 1
			s.updateTableFocus()
			return s, nil
		case "2":
			s.combatFocus = 2
			s.updateTableFocus()
			return s, nil
		}
	}

	// Handle Spells tab navigation
	if s.tab == 2 {
		switch msg.String() {
		case "up", "down", "j", "k", "pgup", "pgdown", "home", "end", "g", "G":
			var cmd tea.Cmd
			s.spellsTable, cmd = s.spellsTable.Update(msg)
			return s, cmd
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Switch spell level filter (0=cantrips, 1-9=spell levels)
			level := int(msg.String()[0] - '0')
			if s.spellLevelFilter == level {
				s.spellLevelFilter = -1 // Toggle off to show all
			} else {
				s.spellLevelFilter = level
			}
			s.refreshSpellsTable()
			return s, nil
		}
	}

	// Handle Inventory tab navigation
	if s.tab == 3 {
		switch msg.String() {
		case "up", "down", "j", "k", "pgup", "pgdown", "home", "end", "g", "G":
			// Pass to focused table
			if s.inventoryFocus == 1 {
				var cmd tea.Cmd
				s.inventoryTable, cmd = s.inventoryTable.Update(msg)
				return s, cmd
			} else if s.inventoryFocus == 2 {
				var cmd tea.Cmd
				s.magicItemsTable, cmd = s.magicItemsTable.Update(msg)
				return s, cmd
			}
		case "1":
			s.inventoryFocus = 1
			s.updateTableFocus()
			return s, nil
		case "2":
			s.inventoryFocus = 2
			s.updateTableFocus()
			return s, nil
		}
	}

	switch msg.String() {
	case "tab", "right", "l":
		s.tab = (s.tab + 1) % 5
		s.updateTableFocus()
	case "shift+tab", "left", "h":
		s.tab = (s.tab + 4) % 5
		s.updateTableFocus()

	case "e":
		if s.tab == 1 { // Combat tab - edit HP
			s.mode = ModeEditHP
			s.hpInput.SetValue(fmt.Sprintf("%d", s.char.CurrentHitPoints))
			s.hpInput.Focus()
			return s, textinput.Blink
		} else if s.tab == 4 { // Notes tab - edit notes
			s.mode = ModeEditNotes
			s.notesInput.SetValue(s.char.Notes)
			s.notesInput.Focus()
			return s, textarea.Blink
		}

	case "f":
		if s.tab == 4 { // Notes tab - edit features & traits
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

// updateTableFocus sets focus on the appropriate table based on current tab
func (s *SheetScreen) updateTableFocus() {
	s.skillsTable.SetFocused(s.tab == 0)
	s.attacksTable.SetFocused(s.tab == 1 && s.combatFocus == 1)
	s.actionsTable.SetFocused(s.tab == 1 && s.combatFocus == 2)
	s.spellsTable.SetFocused(s.tab == 2)
	s.inventoryTable.SetFocused(s.tab == 3 && s.inventoryFocus == 1)
	s.magicItemsTable.SetFocused(s.tab == 3 && s.inventoryFocus == 2)
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
	b.WriteString(s.styles.Title.Render(header))
	b.WriteString("\n\n")

	// Tab bar
	tabs := []string{"Core", "Combat", "Spells", "Inventory", "Notes"}
	tabBar := ""
	for i, t := range tabs {
		if i == s.tab {
			tabBar += s.styles.FocusedButton.Render(" " + t + " ")
		} else {
			tabBar += s.styles.Button.Render(" " + t + " ")
		}
	}
	b.WriteString(tabBar)
	b.WriteString("\n\n")

	// Tab content
	switch s.tab {
	case 0:
		b.WriteString(s.viewCore())
	case 1:
		b.WriteString(s.viewCombat())
	case 2:
		b.WriteString(s.viewSpells())
	case 3:
		b.WriteString(s.viewInventory())
	case 4:
		b.WriteString(s.viewNotes())
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(s.styles.Help.Render(s.getHelp()))

	return lipgloss.Place(s.width, s.height,
		lipgloss.Center, lipgloss.Center,
		b.String())
}

func (s *SheetScreen) viewCore() string {
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

	b.WriteString(s.styles.Header.Render("Ability Scores"))
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

		b.WriteString(s.styles.Muted.Render(paddedName))
		b.WriteString("  ")
		b.WriteString(s.styles.StatValue.Render(paddedScore))
		b.WriteString("  ")
		b.WriteString(s.styles.StatMod.Render(paddedMod))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(s.styles.Header.Render("Saving Throws"))
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
		style := s.styles.NotProficient
		if proficient {
			profMark = "● "
			style = s.styles.Proficient
		}
		paddedName := fmt.Sprintf("%-*s", labelWidth, a.name)
		paddedMod := fmt.Sprintf("%*s", modWidth, character.FormatModifierInt(mod))
		b.WriteString(style.Render(profMark + paddedName + "  " + paddedMod))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("Proficiency Bonus: ")
	b.WriteString(s.styles.StatValue.Render(character.FormatModifierInt(profBonus)))
	b.WriteString("\n")

	// Skills section
	b.WriteString("\n")
	b.WriteString(s.styles.Header.Render("Skills"))
	b.WriteString("\n\n")
	b.WriteString(s.skillsTable.View())

	return b.String()
}

func (s *SheetScreen) viewCombat() string {
	var b strings.Builder

	// HP display
	hpPct := float64(s.char.CurrentHitPoints) / float64(s.char.MaxHitPoints)
	hpStyle := s.styles.HPCurrent
	if hpPct < 0.25 {
		hpStyle = s.styles.HPCritical
	} else if hpPct < 0.5 {
		hpStyle = s.styles.HPLow
	}

	// Right-align labels to align on the colon
	labelWidth := 14

	// Combat stats panel
	b.WriteString(s.styles.Header.Render("Combat Stats"))
	b.WriteString("\n\n")

	if s.mode == ModeEditHP {
		b.WriteString(fmt.Sprintf("%*s ", labelWidth, "Hit Points:"))
		b.WriteString(s.styles.FocusedInput.Render(s.hpInput.View()))
		b.WriteString(fmt.Sprintf(" / %d", s.char.MaxHitPoints))
	} else {
		b.WriteString(fmt.Sprintf("%*s ", labelWidth, "Hit Points:"))
		b.WriteString(hpStyle.Render(fmt.Sprintf("%d", s.char.CurrentHitPoints)))
		b.WriteString(" / ")
		b.WriteString(s.styles.HPMax.Render(fmt.Sprintf("%d", s.char.MaxHitPoints)))
	}

	if s.char.TemporaryHitPoints > 0 {
		b.WriteString(fmt.Sprintf(" (+%d temp)", s.char.TemporaryHitPoints))
	}
	b.WriteString("\n")

	// Other combat stats
	initiative := character.Initiative(int(s.char.Dexterity))

	b.WriteString(fmt.Sprintf("%*s ", labelWidth, "Armor Class:"))
	b.WriteString(s.styles.StatValue.Render(fmt.Sprintf("%d", s.char.ArmorClass)))
	b.WriteString("    ")
	b.WriteString(fmt.Sprintf("%*s ", 10, "Initiative:"))
	b.WriteString(s.styles.StatValue.Render(character.FormatModifierInt(initiative)))
	b.WriteString("    ")
	b.WriteString(fmt.Sprintf("%*s ", 6, "Speed:"))
	b.WriteString(s.styles.StatValue.Render(fmt.Sprintf("%d", s.char.Speed)))
	b.WriteString(" ft\n")

	// Hit dice
	hitDie := character.ClassHitDice[s.char.Class]
	b.WriteString(fmt.Sprintf("%*s %dd%d\n", labelWidth, "Hit Dice:", s.char.Level, hitDie))

	// Attacks section
	b.WriteString("\n")
	attacksHeader := "Attacks"
	if s.combatFocus == 1 {
		attacksHeader = "▶ Attacks"
	}
	b.WriteString(s.styles.Header.Render(attacksHeader))
	b.WriteString("\n\n")
	b.WriteString(s.attacksTable.View())

	// Actions section
	b.WriteString("\n")
	actionsHeader := "Actions"
	if s.combatFocus == 2 {
		actionsHeader = "▶ Actions"
	}
	b.WriteString(s.styles.Header.Render(actionsHeader))
	b.WriteString("\n\n")
	b.WriteString(s.actionsTable.View())

	// Wrap in a left-aligned box so the colon alignment works
	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		Render(b.String())
}

func (s *SheetScreen) viewSpells() string {
	var b strings.Builder

	// Spellcasting info header
	if s.spellcasting != nil {
		spellClass := "Unknown"
		if s.spellcasting.SpellcastingClass.Valid {
			spellClass = s.spellcasting.SpellcastingClass.String
		}
		ability := "—"
		if s.spellcasting.SpellcastingAbility.Valid {
			ability = strings.ToUpper(s.spellcasting.SpellcastingAbility.String[:3])
		}
		saveDC := "—"
		if s.spellcasting.SpellSaveDc.Valid {
			saveDC = fmt.Sprintf("%d", s.spellcasting.SpellSaveDc.Int32)
		}
		atkBonus := "—"
		if s.spellcasting.SpellAttackBonus.Valid {
			atkBonus = character.FormatModifierInt(int(s.spellcasting.SpellAttackBonus.Int32))
		}

		b.WriteString(fmt.Sprintf("%s | %s | Save DC: %s | Attack: %s\n\n",
			s.styles.StatValue.Render(spellClass),
			s.styles.Muted.Render(ability),
			s.styles.StatValue.Render(saveDC),
			s.styles.StatValue.Render(atkBonus),
		))

		// Spell slots display
		b.WriteString(s.styles.Header.Render("Spell Slots"))
		b.WriteString("\n\n")

		slotData := []struct {
			level int
			max   int32
			used  int32
		}{
			{1, s.spellcasting.Slots1Max.Int32, s.spellcasting.Slots1Used.Int32},
			{2, s.spellcasting.Slots2Max.Int32, s.spellcasting.Slots2Used.Int32},
			{3, s.spellcasting.Slots3Max.Int32, s.spellcasting.Slots3Used.Int32},
			{4, s.spellcasting.Slots4Max.Int32, s.spellcasting.Slots4Used.Int32},
			{5, s.spellcasting.Slots5Max.Int32, s.spellcasting.Slots5Used.Int32},
			{6, s.spellcasting.Slots6Max.Int32, s.spellcasting.Slots6Used.Int32},
			{7, s.spellcasting.Slots7Max.Int32, s.spellcasting.Slots7Used.Int32},
			{8, s.spellcasting.Slots8Max.Int32, s.spellcasting.Slots8Used.Int32},
			{9, s.spellcasting.Slots9Max.Int32, s.spellcasting.Slots9Used.Int32},
		}

		for _, slot := range slotData {
			if slot.max > 0 {
				remaining := slot.max - slot.used
				slots := ""
				for i := int32(0); i < slot.max; i++ {
					if i < remaining {
						slots += "[●]"
					} else {
						slots += "[ ]"
					}
				}
				b.WriteString(fmt.Sprintf("  %dst: %s\n", slot.level, slots))
			}
		}
	} else {
		b.WriteString(s.styles.Muted.Render("No spellcasting ability"))
		b.WriteString("\n")
	}

	// Spell level filter tabs
	b.WriteString("\n")
	levelNames := []string{"C", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	for i, name := range levelNames {
		if s.spellLevelFilter == i {
			b.WriteString(s.styles.FocusedButton.Render("[" + name + "]"))
		} else if s.spellLevelFilter == -1 {
			b.WriteString(s.styles.Button.Render(" " + name + " "))
		} else {
			b.WriteString(s.styles.Muted.Render(" " + name + " "))
		}
	}
	if s.spellLevelFilter == -1 {
		b.WriteString(" ")
		b.WriteString(s.styles.FocusedButton.Render("[All]"))
	}
	b.WriteString("\n\n")

	// Spells table
	levelLabel := "All Spells"
	if s.spellLevelFilter == 0 {
		levelLabel = "Cantrips"
	} else if s.spellLevelFilter > 0 {
		levelLabel = fmt.Sprintf("Level %d Spells", s.spellLevelFilter)
	}
	b.WriteString(s.styles.Header.Render(levelLabel))
	b.WriteString("\n\n")
	b.WriteString(s.spellsTable.View())

	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		Render(b.String())
}

func (s *SheetScreen) viewInventory() string {
	var b strings.Builder

	// Currency panel
	b.WriteString(s.styles.Header.Render("Currency"))
	b.WriteString("\n\n")

	if s.currency != nil {
		cp := int32(0)
		if s.currency.Copper.Valid {
			cp = s.currency.Copper.Int32
		}
		sp := int32(0)
		if s.currency.Silver.Valid {
			sp = s.currency.Silver.Int32
		}
		ep := int32(0)
		if s.currency.Electrum.Valid {
			ep = s.currency.Electrum.Int32
		}
		gp := int32(0)
		if s.currency.Gold.Valid {
			gp = s.currency.Gold.Int32
		}
		pp := int32(0)
		if s.currency.Platinum.Valid {
			pp = s.currency.Platinum.Int32
		}

		b.WriteString(fmt.Sprintf("  CP: %s  SP: %s  EP: %s  GP: %s  PP: %s\n",
			s.styles.StatValue.Render(fmt.Sprintf("%d", cp)),
			s.styles.StatValue.Render(fmt.Sprintf("%d", sp)),
			s.styles.StatValue.Render(fmt.Sprintf("%d", ep)),
			s.styles.StatValue.Render(fmt.Sprintf("%d", gp)),
			s.styles.StatValue.Render(fmt.Sprintf("%d", pp)),
		))
	} else {
		b.WriteString("  CP: 0  SP: 0  EP: 0  GP: 0  PP: 0\n")
	}

	// Calculate total weight
	totalWeight := 0.0
	for _, item := range s.inventory {
		if item.Weight.Valid && item.Weight.Int != nil {
			qty := int32(1)
			if item.Quantity.Valid {
				qty = item.Quantity.Int32
			}
			// Convert numeric weight to float using Float64Value
			f64Val, err := item.Weight.Float64Value()
			if err == nil && f64Val.Valid {
				totalWeight += f64Val.Float64 * float64(qty)
			}
		}
	}
	for _, item := range s.magicItems {
		if item.Weight.Valid && item.Weight.Int != nil {
			f64Val, err := item.Weight.Float64Value()
			if err == nil && f64Val.Valid {
				totalWeight += f64Val.Float64
			}
		}
	}

	b.WriteString(fmt.Sprintf("  Total Weight: %s lbs\n",
		s.styles.StatValue.Render(fmt.Sprintf("%.1f", totalWeight))))

	// Equipment table
	b.WriteString("\n")
	equipHeader := "Equipment"
	if s.inventoryFocus == 1 {
		equipHeader = "▶ Equipment"
	}
	b.WriteString(s.styles.Header.Render(equipHeader))
	b.WriteString("\n\n")
	b.WriteString(s.inventoryTable.View())

	// Magic Items table
	b.WriteString("\n")
	magicHeader := "Magic Items"
	if s.inventoryFocus == 2 {
		magicHeader = "▶ Magic Items"
	}
	b.WriteString(s.styles.Header.Render(magicHeader))
	b.WriteString("\n\n")
	b.WriteString(s.magicItemsTable.View())

	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		Render(b.String())
}

func (s *SheetScreen) viewNotes() string {
	var b strings.Builder

	b.WriteString(s.styles.Header.Render("Features & Traits"))
	b.WriteString("\n\n")

	if s.mode == ModeEditFeatures {
		b.WriteString(s.styles.FocusedInput.Render(s.featuresInput.View()))
	} else if s.char.FeaturesTraits != "" {
		b.WriteString(s.char.FeaturesTraits)
	} else {
		b.WriteString(s.styles.Muted.Render("No features or traits recorded."))
	}
	b.WriteString("\n\n")

	b.WriteString(s.styles.Header.Render("Notes"))
	b.WriteString("\n\n")

	if s.mode == ModeEditNotes {
		b.WriteString(s.styles.FocusedInput.Render(s.notesInput.View()))
	} else if s.char.Notes != "" {
		b.WriteString(s.char.Notes)
	} else {
		b.WriteString(s.styles.Muted.Render("No notes recorded."))
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
		switch s.tab {
		case 0: // Core
			help += " • j/k: navigate skills"
		case 1: // Combat
			help += " • e: edit HP • 1: attacks • 2: actions • j/k: navigate"
		case 2: // Spells
			help += " • 0-9: filter by level • j/k: navigate"
		case 3: // Inventory
			help += " • 1: equipment • 2: magic items • j/k: navigate"
		case 4: // Notes
			help += " • e: edit notes • f: edit features"
		}
		return help
	}
}
