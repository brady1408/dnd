package screens

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/brady1408/dnd/internal/character"
	"github.com/brady1408/dnd/internal/db"
	"github.com/brady1408/dnd/internal/tui/components"
	"github.com/brady1408/dnd/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackc/pgx/v5/pgtype"
)

type SheetMode int

const (
	ModeView SheetMode = iota
	ModeEditHP
	ModeEditDamage
	ModeEditHeal
	ModeEditNotes
	ModeEditFeatures
	ModeEditBackground
	ModeAddAttack
	ModeAddAction
	ModeAddSpell
	ModeHelp
)

type SheetScreen struct {
	ctx     context.Context
	queries *db.Queries
	char    db.Character
	styles  *styles.Styles

	mode       SheetMode
	tab        int // 0=core, 1=combat, 2=spells, 3=inventory, 4=features, 5=background, 6=notes
	width      int
	height     int

	// Edit mode inputs
	hpInput         textinput.Model
	damageInput     textinput.Model
	healInput       textinput.Model
	notesInput      textarea.Model
	featuresInput   textarea.Model
	backgroundModal *components.ModalModel
	attackModal     *components.ModalModel
	actionModal     *components.ModalModel
	spellModal      *components.ModalModel
	editCursor      int

	// Table components
	skillsTable     *components.TableModel
	attacksTable    *components.TableModel
	actionsTable    *components.TableModel
	inventoryTable  *components.TableModel
	magicItemsTable *components.TableModel
	spellsTable     *components.TableModel
	featuresTable   *components.TableModel

	// Combat tab focus: 0=stats panel, 1=attacks, 2=actions
	combatFocus int

	// Inventory tab focus: 0=currency, 1=equipment, 2=magic items
	inventoryFocus int

	// Spells tab: which spell level is selected (0=cantrips, 1-9=spell levels)
	spellLevelFilter int

	// Features tab: filter by source type (empty=all, "class", "race", "background", "feat")
	featuresFilter string

	// Cached data from DB
	attacks      []db.CharacterAttack
	actions      []db.CharacterAction
	inventory    []db.CharacterInventory
	magicItems   []db.CharacterMagicItem
	currency     *db.CharacterCurrency
	spellcasting *db.CharacterSpellcasting
	spells       []db.CharacterSpell
	features     []db.CharacterFeature
	details      *db.CharacterDetail

	// Status message for user feedback
	statusMsg   string
	statusIsErr bool
}

type CharacterUpdatedMsg struct {
	Character db.Character
}

type StatusMsg struct {
	Message string
	IsError bool
}

type ClearStatusMsg struct{}

func NewSheetScreen(ctx context.Context, queries *db.Queries, char db.Character, s *styles.Styles) *SheetScreen {
	hpInput := textinput.New()
	hpInput.Placeholder = "HP"
	hpInput.Width = 10
	hpInput.CharLimit = 5

	damageInput := textinput.New()
	damageInput.Placeholder = "0"
	damageInput.Width = 6
	damageInput.CharLimit = 4

	healInput := textinput.New()
	healInput.Placeholder = "0"
	healInput.Width = 6
	healInput.CharLimit = 4

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

	// Create features table
	featuresTable := components.NewTable([]components.TableColumn{
		{Title: "Feature", Width: 25},
		{Title: "Source", Width: 15},
		{Title: "Type", Width: 12},
	}, s)
	featuresTable.SetVisibleRows(12)
	featuresTable.SetEmptyMessage("No features")

	sheet := &SheetScreen{
		ctx:             ctx,
		queries:         queries,
		char:            char,
		styles:          s,
		mode:            ModeView,
		hpInput:         hpInput,
		damageInput:     damageInput,
		healInput:       healInput,
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
		featuresTable:   featuresTable,
		combatFocus:     1, // Default to Attacks tab
		inventoryFocus:  1, // Default to Equipment tab
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
	sheet.refreshFeaturesTable()
	sheet.refreshDetails()

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

// refreshFeaturesTable loads features from DB and populates the table
func (s *SheetScreen) refreshFeaturesTable() {
	var features []db.CharacterFeature
	var err error

	if s.featuresFilter != "" {
		features, err = s.queries.GetCharacterFeaturesByType(s.ctx, db.GetCharacterFeaturesByTypeParams{
			CharacterID: s.char.ID,
			SourceType:  pgtype.Text{String: s.featuresFilter, Valid: true},
		})
	} else {
		features, err = s.queries.GetCharacterFeatures(s.ctx, s.char.ID)
	}

	if err != nil {
		s.features = nil
		s.featuresTable.SetRows(nil)
		return
	}

	s.features = features
	var rows []components.TableRow
	for _, feat := range features {
		source := ""
		if feat.Source.Valid {
			source = feat.Source.String
		}

		sourceType := ""
		if feat.SourceType.Valid {
			sourceType = feat.SourceType.String
		}

		rows = append(rows, components.TableRow{
			ID:    fmt.Sprintf("%x", feat.ID.Bytes),
			Cells: []string{feat.Name, source, sourceType},
			Data:  feat,
		})
	}

	s.featuresTable.SetRows(rows)
}

// refreshDetails loads character details from DB
func (s *SheetScreen) refreshDetails() {
	details, err := s.queries.GetCharacterDetails(s.ctx, s.char.ID)
	if err != nil {
		s.details = nil
		return
	}
	s.details = &details
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
	s.refreshFeaturesTable()
	s.refreshDetails()
}

func (s *SheetScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	case StatusMsg:
		s.statusMsg = msg.Message
		s.statusIsErr = msg.IsError
		return s, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return ClearStatusMsg{}
		})
	case ClearStatusMsg:
		s.statusMsg = ""
		s.statusIsErr = false
		return s, nil
	case components.ModalSaveMsg:
		if s.mode == ModeEditBackground {
			s.mode = ModeView
			return s, s.saveBackgroundDetails(msg.Values)
		}
		if s.mode == ModeAddAttack {
			s.mode = ModeView
			return s, s.saveAttack(msg.Values)
		}
		if s.mode == ModeAddAction {
			s.mode = ModeView
			return s, s.saveAction(msg.Values)
		}
		if s.mode == ModeAddSpell {
			s.mode = ModeView
			return s, s.saveSpell(msg.Values)
		}
	case components.ModalCancelMsg:
		if s.mode == ModeEditBackground {
			s.mode = ModeView
			s.backgroundModal = nil
		}
		if s.mode == ModeAddAttack {
			s.mode = ModeView
			s.attackModal = nil
		}
		if s.mode == ModeAddAction {
			s.mode = ModeView
			s.actionModal = nil
		}
		if s.mode == ModeAddSpell {
			s.mode = ModeView
			s.spellModal = nil
		}
		return s, nil
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
	case ModeEditDamage:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return s.updateEditDamage(keyMsg)
		}
	case ModeEditHeal:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return s.updateEditHeal(keyMsg)
		}
	case ModeEditNotes:
		return s.updateEditNotes(msg)
	case ModeEditFeatures:
		return s.updateEditFeatures(msg)
	case ModeEditBackground:
		if s.backgroundModal != nil {
			var cmd tea.Cmd
			s.backgroundModal, cmd = s.backgroundModal.Update(msg)
			return s, cmd
		}
	case ModeAddAttack:
		if s.attackModal != nil {
			var cmd tea.Cmd
			s.attackModal, cmd = s.attackModal.Update(msg)
			return s, cmd
		}
	case ModeAddAction:
		if s.actionModal != nil {
			var cmd tea.Cmd
			s.actionModal, cmd = s.actionModal.Update(msg)
			return s, cmd
		}
	case ModeAddSpell:
		if s.spellModal != nil {
			var cmd tea.Cmd
			s.spellModal, cmd = s.spellModal.Update(msg)
			return s, cmd
		}
	case ModeHelp:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			// Any key closes help
			switch keyMsg.String() {
			case "?", "esc", "q", "enter", " ":
				s.mode = ModeView
				return s, nil
			}
		}
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
			// Pass to the currently visible table
			if s.combatFocus == 2 {
				var cmd tea.Cmd
				s.actionsTable, cmd = s.actionsTable.Update(msg)
				return s, cmd
			} else {
				var cmd tea.Cmd
				s.attacksTable, cmd = s.attacksTable.Update(msg)
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
		case "a":
			// Add attack or action based on current sub-tab
			if s.combatFocus == 2 {
				s.openActionModal()
				s.mode = ModeAddAction
				return s, textinput.Blink
			} else {
				s.openAttackModal()
				s.mode = ModeAddAttack
				return s, textinput.Blink
			}
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
		case "a":
			s.openSpellModal()
			s.mode = ModeAddSpell
			return s, textinput.Blink
		case "d", "x":
			// Delete selected spell
			if row := s.spellsTable.GetSelectedRow(); row != nil {
				if spell, ok := row.Data.(db.CharacterSpell); ok {
					return s, s.deleteSpell(spell.ID)
				}
			}
		case "p", " ":
			// Toggle prepared status on selected spell
			if row := s.spellsTable.GetSelectedRow(); row != nil {
				if spell, ok := row.Data.(db.CharacterSpell); ok {
					return s, s.toggleSpellPrepared(spell.ID)
				}
			}
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

	// Handle Features tab navigation
	if s.tab == 4 {
		switch msg.String() {
		case "up", "down", "j", "k", "pgup", "pgdown", "home", "end", "g", "G":
			var cmd tea.Cmd
			s.featuresTable, cmd = s.featuresTable.Update(msg)
			return s, cmd
		case "1":
			if s.featuresFilter == "class" {
				s.featuresFilter = ""
			} else {
				s.featuresFilter = "class"
			}
			s.refreshFeaturesTable()
			return s, nil
		case "2":
			if s.featuresFilter == "race" {
				s.featuresFilter = ""
			} else {
				s.featuresFilter = "race"
			}
			s.refreshFeaturesTable()
			return s, nil
		case "3":
			if s.featuresFilter == "background" {
				s.featuresFilter = ""
			} else {
				s.featuresFilter = "background"
			}
			s.refreshFeaturesTable()
			return s, nil
		case "4":
			if s.featuresFilter == "feat" {
				s.featuresFilter = ""
			} else {
				s.featuresFilter = "feat"
			}
			s.refreshFeaturesTable()
			return s, nil
		}
	}

	switch msg.String() {
	case "tab", "right", "l":
		s.tab = (s.tab + 1) % 7
		s.updateTableFocus()
	case "shift+tab", "left", "h":
		s.tab = (s.tab + 6) % 7
		s.updateTableFocus()

	case "e":
		if s.tab == 1 { // Combat tab - edit HP
			s.mode = ModeEditHP
			s.hpInput.SetValue(fmt.Sprintf("%d", s.char.CurrentHitPoints))
			s.hpInput.Focus()
			return s, textinput.Blink
		} else if s.tab == 5 { // Background tab - edit background details
			s.openBackgroundModal()
			s.mode = ModeEditBackground
			return s, textinput.Blink
		} else if s.tab == 6 { // Notes tab - edit notes
			s.mode = ModeEditNotes
			s.notesInput.SetValue(s.char.Notes)
			s.notesInput.Focus()
			return s, textarea.Blink
		}

	case "-":
		if s.tab == 1 { // Combat tab - damage
			s.mode = ModeEditDamage
			s.damageInput.SetValue("")
			s.damageInput.Focus()
			return s, textinput.Blink
		}

	case "+", "=":
		if s.tab == 1 { // Combat tab - heal (= is unshifted +)
			s.mode = ModeEditHeal
			s.healInput.SetValue("")
			s.healInput.Focus()
			return s, textinput.Blink
		}

	case "f":
		if s.tab == 6 { // Notes tab - edit features & traits
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

	case "?":
		s.mode = ModeHelp
		return s, nil

	case "esc", "q":
		return s, func() tea.Msg { return NavigateBackMsg{} }
	}

	return s, nil
}

// updateTableFocus sets focus on the appropriate table based on current tab
func (s *SheetScreen) updateTableFocus() {
	s.skillsTable.SetFocused(s.tab == 0)
	// Combat tab: only the visible sub-tab table is focused
	s.attacksTable.SetFocused(s.tab == 1 && s.combatFocus != 2)
	s.actionsTable.SetFocused(s.tab == 1 && s.combatFocus == 2)
	s.spellsTable.SetFocused(s.tab == 2)
	// Inventory tab: only the visible sub-tab table is focused
	s.inventoryTable.SetFocused(s.tab == 3 && s.inventoryFocus != 2)
	s.magicItemsTable.SetFocused(s.tab == 3 && s.inventoryFocus == 2)
	s.featuresTable.SetFocused(s.tab == 4)
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

func (s *SheetScreen) updateEditDamage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		var damage int
		fmt.Sscanf(s.damageInput.Value(), "%d", &damage)
		if damage < 0 {
			damage = 0
		}

		newHP := int(s.char.CurrentHitPoints) - damage
		if newHP < 0 {
			newHP = 0
		}

		return s, s.updateHP(int32(newHP))

	case "esc":
		s.mode = ModeView
		return s, nil
	}

	var cmd tea.Cmd
	s.damageInput, cmd = s.damageInput.Update(msg)
	return s, cmd
}

func (s *SheetScreen) updateEditHeal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		var heal int
		fmt.Sscanf(s.healInput.Value(), "%d", &heal)
		if heal < 0 {
			heal = 0
		}

		newHP := int(s.char.CurrentHitPoints) + heal
		if newHP > int(s.char.MaxHitPoints) {
			newHP = int(s.char.MaxHitPoints)
		}

		return s, s.updateHP(int32(newHP))

	case "esc":
		s.mode = ModeView
		return s, nil
	}

	var cmd tea.Cmd
	s.healInput, cmd = s.healInput.Update(msg)
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
			return StatusMsg{Message: "Failed to update HP", IsError: true}
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
			return StatusMsg{Message: "Failed to save notes", IsError: true}
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
			return StatusMsg{Message: "Failed to save features", IsError: true}
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
	tabs := []string{"Core", "Combat", "Spells", "Inventory", "Features", "Background", "Notes"}
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
		b.WriteString(s.viewFeatures())
	case 5:
		b.WriteString(s.viewBackground())
	case 6:
		b.WriteString(s.viewNotes())
	}

	// Status message (if any)
	if s.statusMsg != "" {
		b.WriteString("\n\n")
		if s.statusIsErr {
			b.WriteString(s.styles.ErrorText.Render("✗ " + s.statusMsg))
		} else {
			b.WriteString(s.styles.SuccessText.Render("✓ " + s.statusMsg))
		}
	}

	// Help
	b.WriteString("\n\n")
	b.WriteString(s.styles.Help.Render(s.getHelp()))

	content := lipgloss.Place(s.width, s.height,
		lipgloss.Center, lipgloss.Center,
		b.String())

	// Overlay help if in help mode
	if s.mode == ModeHelp {
		return s.renderHelpOverlay(content)
	}

	// Overlay background modal if editing
	if s.mode == ModeEditBackground && s.backgroundModal != nil {
		return s.backgroundModal.ViewWithOverlay(content, s.width, s.height)
	}

	// Overlay attack modal if adding
	if s.mode == ModeAddAttack && s.attackModal != nil {
		return s.attackModal.ViewWithOverlay(content, s.width, s.height)
	}

	// Overlay action modal if adding
	if s.mode == ModeAddAction && s.actionModal != nil {
		return s.actionModal.ViewWithOverlay(content, s.width, s.height)
	}

	// Overlay spell modal if adding
	if s.mode == ModeAddSpell && s.spellModal != nil {
		return s.spellModal.ViewWithOverlay(content, s.width, s.height)
	}

	return content
}

func (s *SheetScreen) viewCore() string {
	// Ability scores
	abilities := []struct {
		name  string
		abbr  string
		score int32
	}{
		{"Strength", "STR", s.char.Strength},
		{"Dexterity", "DEX", s.char.Dexterity},
		{"Constitution", "CON", s.char.Constitution},
		{"Intelligence", "INT", s.char.Intelligence},
		{"Wisdom", "WIS", s.char.Wisdom},
		{"Charisma", "CHA", s.char.Charisma},
	}

	profBonus := character.ProficiencyBonus(int(s.char.Level))

	// Build left column: Ability Scores
	var leftCol strings.Builder
	leftCol.WriteString("Ability Scores\n\n")

	for _, a := range abilities {
		mod := character.AbilityModifier(int(a.score))
		modStr := character.FormatModifierInt(mod)
		leftCol.WriteString(fmt.Sprintf("  %-3s %2d  %3s\n", a.abbr, a.score, modStr))
	}

	// Build right column: Saving Throws
	var rightCol strings.Builder
	rightCol.WriteString("Saving Throws\n\n")

	for _, a := range abilities {
		proficient := false
		for _, p := range s.char.SavingThrowProficiencies {
			if strings.EqualFold(p, a.name) {
				proficient = true
				break
			}
		}

		mod := character.SavingThrow(int(a.score), int(s.char.Level), proficient)
		profMark := "○"
		if proficient {
			profMark = "●"
		}
		modStr := character.FormatModifierInt(mod)
		rightCol.WriteString(fmt.Sprintf("  %s %-3s %3s\n", profMark, a.abbr, modStr))
	}

	// Style columns with fixed width
	leftStyle := lipgloss.NewStyle().Width(18)
	rightStyle := lipgloss.NewStyle().Width(18)

	// Join columns horizontally
	topSection := lipgloss.JoinHorizontal(lipgloss.Top,
		leftStyle.Render(leftCol.String()),
		rightStyle.Render(rightCol.String()),
	)

	// Build the full view
	var b strings.Builder
	b.WriteString(topSection)
	b.WriteString("\n")
	b.WriteString("Proficiency Bonus: ")
	b.WriteString(s.styles.StatValue.Render(character.FormatModifierInt(profBonus)))
	b.WriteString("\n\n")

	// Skills section
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

	// Combat stats panel
	b.WriteString("Combat Stats\n\n")

	initiative := character.Initiative(int(s.char.Dexterity))
	hitDie := character.ClassHitDice[s.char.Class]

	// Compact single-line stats bar
	hpStr := fmt.Sprintf("%s/%s",
		hpStyle.Render(fmt.Sprintf("%d", s.char.CurrentHitPoints)),
		s.styles.HPMax.Render(fmt.Sprintf("%d", s.char.MaxHitPoints)))
	if s.char.TemporaryHitPoints > 0 {
		hpStr += fmt.Sprintf("+%d", s.char.TemporaryHitPoints)
	}

	switch s.mode {
	case ModeEditHP:
		b.WriteString(fmt.Sprintf("  HP: %s/%d  |  AC: %d  |  Init: %s  |  Speed: %d ft  |  HD: %dd%d\n",
			s.hpInput.View(),
			s.char.MaxHitPoints,
			s.char.ArmorClass,
			character.FormatModifierInt(initiative),
			s.char.Speed,
			s.char.Level,
			hitDie))
	case ModeEditDamage:
		b.WriteString(fmt.Sprintf("  HP: %s -%s  |  AC: %d  |  Init: %s  |  Speed: %d ft  |  HD: %dd%d\n",
			hpStr,
			s.damageInput.View(),
			s.char.ArmorClass,
			character.FormatModifierInt(initiative),
			s.char.Speed,
			s.char.Level,
			hitDie))
	case ModeEditHeal:
		b.WriteString(fmt.Sprintf("  HP: %s +%s  |  AC: %d  |  Init: %s  |  Speed: %d ft  |  HD: %dd%d\n",
			hpStr,
			s.healInput.View(),
			s.char.ArmorClass,
			character.FormatModifierInt(initiative),
			s.char.Speed,
			s.char.Level,
			hitDie))
	default:
		b.WriteString(fmt.Sprintf("  HP: %s  |  AC: %d  |  Init: %s  |  Speed: %d ft  |  HD: %dd%d\n",
			hpStr,
			s.char.ArmorClass,
			character.FormatModifierInt(initiative),
			s.char.Speed,
			s.char.Level,
			hitDie))
	}

	// Sub-tabs for Attacks/Actions
	b.WriteString("\n")
	if s.combatFocus != 2 {
		b.WriteString(s.styles.FocusedButton.Render("[1:Attacks]"))
	} else {
		b.WriteString(s.styles.Button.Render(" 1:Attacks "))
	}
	b.WriteString(" ")
	if s.combatFocus == 2 {
		b.WriteString(s.styles.FocusedButton.Render("[2:Actions]"))
	} else {
		b.WriteString(s.styles.Button.Render(" 2:Actions "))
	}
	b.WriteString("\n\n")

	// Show the selected table
	if s.combatFocus == 2 {
		b.WriteString(s.actionsTable.View())
	} else {
		b.WriteString(s.attacksTable.View())
	}

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

func (s *SheetScreen) viewFeatures() string {
	var b strings.Builder

	// Filter tabs
	filters := []struct {
		key   string
		label string
	}{
		{"", "All"},
		{"class", "Class"},
		{"race", "Race"},
		{"background", "Background"},
		{"feat", "Feats"},
	}

	for i, f := range filters {
		label := f.label
		if i > 0 {
			label = fmt.Sprintf("%d:%s", i, f.label)
		}
		if s.featuresFilter == f.key {
			b.WriteString(s.styles.FocusedButton.Render("[" + label + "]"))
		} else {
			b.WriteString(s.styles.Button.Render(" " + label + " "))
		}
	}
	b.WriteString("\n\n")

	// Features table
	b.WriteString(s.styles.Header.Render("Features & Traits"))
	b.WriteString("\n\n")
	b.WriteString(s.featuresTable.View())

	return lipgloss.NewStyle().
		Align(lipgloss.Left).
		Render(b.String())
}

func (s *SheetScreen) viewBackground() string {
	var b strings.Builder

	// Character identity
	b.WriteString(s.styles.Header.Render("Character Info"))
	b.WriteString("\n\n")

	labelWidth := 12
	b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Name:", s.char.Name))
	b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Race:", s.char.Race))
	b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Class:", s.char.Class))
	b.WriteString(fmt.Sprintf("  %*s %d\n", labelWidth, "Level:", s.char.Level))
	nextLevelXP := character.XPThresholds[int(s.char.Level)+1]
	if s.char.Level >= 20 {
		nextLevelXP = character.XPThresholds[20]
	}
	b.WriteString(fmt.Sprintf("  %*s %d / %d\n", labelWidth, "Experience:", s.char.ExperiencePoints, nextLevelXP))
	if s.char.Background.Valid {
		b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Background:", s.char.Background.String))
	}
	if s.char.Alignment.Valid {
		b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Alignment:", s.char.Alignment.String))
	}

	// Physical characteristics from details
	if s.details != nil {
		b.WriteString("\n")
		b.WriteString(s.styles.Header.Render("Physical Traits"))
		b.WriteString("\n\n")

		if s.details.Size.Valid {
			b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Size:", s.details.Size.String))
		}
		if s.details.Gender.Valid {
			b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Gender:", s.details.Gender.String))
		}
		if s.details.Age.Valid {
			b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Age:", s.details.Age.String))
		}
		if s.details.Height.Valid {
			b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Height:", s.details.Height.String))
		}
		if s.details.Weight.Valid {
			b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Weight:", s.details.Weight.String))
		}
		if s.details.Eyes.Valid {
			b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Eyes:", s.details.Eyes.String))
		}
		if s.details.Hair.Valid {
			b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Hair:", s.details.Hair.String))
		}
		if s.details.Skin.Valid {
			b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Skin:", s.details.Skin.String))
		}
		if s.details.FaithDeity.Valid {
			b.WriteString(fmt.Sprintf("  %*s %s\n", labelWidth, "Faith/Deity:", s.details.FaithDeity.String))
		}

		// Personality
		b.WriteString("\n")
		b.WriteString(s.styles.Header.Render("Personality"))
		b.WriteString("\n\n")

		if s.details.PersonalityTraits.Valid && s.details.PersonalityTraits.String != "" {
			b.WriteString(s.styles.Muted.Render("  Traits: "))
			b.WriteString(s.details.PersonalityTraits.String)
			b.WriteString("\n")
		}
		if s.details.Ideals.Valid && s.details.Ideals.String != "" {
			b.WriteString(s.styles.Muted.Render("  Ideals: "))
			b.WriteString(s.details.Ideals.String)
			b.WriteString("\n")
		}
		if s.details.Bonds.Valid && s.details.Bonds.String != "" {
			b.WriteString(s.styles.Muted.Render("  Bonds: "))
			b.WriteString(s.details.Bonds.String)
			b.WriteString("\n")
		}
		if s.details.Flaws.Valid && s.details.Flaws.String != "" {
			b.WriteString(s.styles.Muted.Render("  Flaws: "))
			b.WriteString(s.details.Flaws.String)
			b.WriteString("\n")
		}

		// Backstory
		if s.details.Backstory.Valid && s.details.Backstory.String != "" {
			b.WriteString("\n")
			b.WriteString(s.styles.Header.Render("Backstory"))
			b.WriteString("\n\n")
			b.WriteString("  ")
			b.WriteString(s.details.Backstory.String)
			b.WriteString("\n")
		}

		// Allies
		if s.details.AlliesOrganizations.Valid && s.details.AlliesOrganizations.String != "" {
			b.WriteString("\n")
			b.WriteString(s.styles.Header.Render("Allies & Organizations"))
			b.WriteString("\n\n")
			b.WriteString("  ")
			b.WriteString(s.details.AlliesOrganizations.String)
			b.WriteString("\n")
		}
	} else {
		b.WriteString("\n")
		b.WriteString(s.styles.Muted.Render("No detailed background information available"))
		b.WriteString("\n")
	}

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
	case ModeEditDamage:
		return "enter: apply damage • esc: cancel"
	case ModeEditHeal:
		return "enter: apply healing • esc: cancel"
	case ModeEditNotes, ModeEditFeatures:
		return "ctrl+s: save • esc: cancel"
	default:
		help := "tab/←→: switch tabs • q/esc: back"
		switch s.tab {
		case 0: // Core
			help += " • j/k: navigate skills"
		case 1: // Combat
			help += " • -: damage • +: heal • e: set HP • 1: attacks • 2: actions • a: add"
		case 2: // Spells
			help += " • 0-9: filter • a: add • d: delete • p: toggle prepared"
		case 3: // Inventory
			help += " • 1: equipment • 2: magic items • j/k: navigate"
		case 4: // Features
			help += " • 1-4: filter type • j/k: navigate"
		case 5: // Background
			help += " • e: edit details"
		case 6: // Notes
			help += " • e: edit notes • f: edit features"
		}
		return help + " • ?: help"
	}
}

// renderHelpOverlay renders the help overlay on top of the content
func (s *SheetScreen) renderHelpOverlay(background string) string {
	// Build the help content
	var b strings.Builder

	b.WriteString(s.styles.Title.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	// Global shortcuts
	b.WriteString(s.styles.Header.Render("Global"))
	b.WriteString("\n")
	b.WriteString("  tab / ← →     Switch tabs\n")
	b.WriteString("  q / esc       Back to character list\n")
	b.WriteString("  ?             Show this help\n")
	b.WriteString("\n")

	// Navigation
	b.WriteString(s.styles.Header.Render("Navigation"))
	b.WriteString("\n")
	b.WriteString("  j / ↓         Move down in list\n")
	b.WriteString("  k / ↑         Move up in list\n")
	b.WriteString("  g / Home      Go to first item\n")
	b.WriteString("  G / End       Go to last item\n")
	b.WriteString("  PgUp/PgDn     Page up/down\n")
	b.WriteString("\n")

	// Tab-specific
	b.WriteString(s.styles.Header.Render("Tab-Specific"))
	b.WriteString("\n")
	b.WriteString("  Combat:     -: damage, +: heal, e: set HP, 1/2: tabs, a: add\n")
	b.WriteString("  Spells:     0-9: filter, a: add, d: delete, p: prepared\n")
	b.WriteString("  Inventory:  1/2: switch tables\n")
	b.WriteString("  Features:   1-4: filter by type\n")
	b.WriteString("  Background: e: edit details\n")
	b.WriteString("  Notes:      e: edit notes, f: edit features\n")
	b.WriteString("\n")

	// Editing
	b.WriteString(s.styles.Header.Render("Editing"))
	b.WriteString("\n")
	b.WriteString("  Ctrl+S        Save changes\n")
	b.WriteString("  Esc           Cancel editing\n")
	b.WriteString("\n\n")

	b.WriteString(s.styles.Muted.Render("Press ? or Esc to close"))

	// Style the help box
	helpBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		Width(50).
		Render(b.String())

	// Center the help box
	centered := lipgloss.Place(s.width, s.height,
		lipgloss.Center, lipgloss.Center,
		helpBox)

	// Dim the background (simple approach - just return centered overlay)
	// For a true dimmed effect, we'd need to process each character
	return centered
}

// openBackgroundModal creates and shows the background edit modal
func (s *SheetScreen) openBackgroundModal() {
	// Get current values from details (or empty if nil)
	var size, gender, age, height, weight, eyes, hair, skin, faith string
	var traits, ideals, bonds, flaws, backstory, allies string

	if s.details != nil {
		if s.details.Size.Valid {
			size = s.details.Size.String
		}
		if s.details.Gender.Valid {
			gender = s.details.Gender.String
		}
		if s.details.Age.Valid {
			age = s.details.Age.String
		}
		if s.details.Height.Valid {
			height = s.details.Height.String
		}
		if s.details.Weight.Valid {
			weight = s.details.Weight.String
		}
		if s.details.Eyes.Valid {
			eyes = s.details.Eyes.String
		}
		if s.details.Hair.Valid {
			hair = s.details.Hair.String
		}
		if s.details.Skin.Valid {
			skin = s.details.Skin.String
		}
		if s.details.FaithDeity.Valid {
			faith = s.details.FaithDeity.String
		}
		if s.details.PersonalityTraits.Valid {
			traits = s.details.PersonalityTraits.String
		}
		if s.details.Ideals.Valid {
			ideals = s.details.Ideals.String
		}
		if s.details.Bonds.Valid {
			bonds = s.details.Bonds.String
		}
		if s.details.Flaws.Valid {
			flaws = s.details.Flaws.String
		}
		if s.details.Backstory.Valid {
			backstory = s.details.Backstory.String
		}
		if s.details.AlliesOrganizations.Valid {
			allies = s.details.AlliesOrganizations.String
		}
	}

	// Get alignment from character (stored in characters table, not details)
	alignment := ""
	if s.char.Alignment.Valid {
		alignment = s.char.Alignment.String
	}

	// Simple single-column layout
	fields := []components.FormField{
		{Key: "size", Label: "Size", Type: components.FieldSelect, Value: size, Options: []string{"Tiny", "Small", "Medium", "Large", "Huge", "Gargantuan"}},
		{Key: "alignment", Label: "Alignment", Type: components.FieldSelect, Value: alignment, Options: []string{"Lawful Good", "Neutral Good", "Chaotic Good", "Lawful Neutral", "Neutral", "Chaotic Neutral", "Lawful Evil", "Neutral Evil", "Chaotic Evil"}},
		{Key: "gender", Label: "Gender", Type: components.FieldText, Value: gender, Placeholder: "Gender"},
		{Key: "height", Label: "Height", Type: components.FieldText, Value: height, Placeholder: "5'10\""},
		{Key: "weight", Label: "Weight", Type: components.FieldText, Value: weight, Placeholder: "180 lbs"},
		{Key: "age", Label: "Age", Type: components.FieldText, Value: age, Placeholder: "25"},
		{Key: "faith", Label: "Faith/Deity", Type: components.FieldText, Value: faith, Placeholder: "Deity or faith"},
		{Key: "hair", Label: "Hair", Type: components.FieldText, Value: hair, Placeholder: "Hair color/style"},
		{Key: "eyes", Label: "Eyes", Type: components.FieldText, Value: eyes, Placeholder: "Eye color"},
		{Key: "skin", Label: "Skin", Type: components.FieldText, Value: skin, Placeholder: "Skin tone"},
		{Key: "traits", Label: "Personality", Type: components.FieldText, Value: traits, Placeholder: "Personality traits"},
		{Key: "ideals", Label: "Ideals", Type: components.FieldText, Value: ideals, Placeholder: "Ideals"},
		{Key: "bonds", Label: "Bonds", Type: components.FieldText, Value: bonds, Placeholder: "Bonds"},
		{Key: "flaws", Label: "Flaws", Type: components.FieldText, Value: flaws, Placeholder: "Flaws"},
		{Key: "backstory", Label: "Backstory", Type: components.FieldText, Value: backstory, Placeholder: "Character backstory"},
		{Key: "allies", Label: "Allies", Type: components.FieldText, Value: allies, Placeholder: "Allies & organizations"},
	}

	s.backgroundModal = components.NewModal("Edit Background", fields, s.styles)
	s.backgroundModal.SetSize(s.width, s.height)
	s.backgroundModal.Show()
}

// saveBackgroundDetails saves the background modal values to the database
func (s *SheetScreen) saveBackgroundDetails(values map[string]string) tea.Cmd {
	return func() tea.Msg {
		// Ensure character_details row exists
		if s.details == nil {
			details, err := s.queries.CreateCharacterDetails(s.ctx, s.char.ID)
			if err != nil {
				return StatusMsg{Message: fmt.Sprintf("Error creating details: %v", err), IsError: true}
			}
			s.details = &details
		}

		// Update the details
		params := db.UpdateCharacterDetailsParams{
			CharacterID:        s.char.ID,
			Age:                pgtype.Text{String: values["age"], Valid: values["age"] != ""},
			Height:             pgtype.Text{String: values["height"], Valid: values["height"] != ""},
			Weight:             pgtype.Text{String: values["weight"], Valid: values["weight"] != ""},
			Eyes:               pgtype.Text{String: values["eyes"], Valid: values["eyes"] != ""},
			Skin:               pgtype.Text{String: values["skin"], Valid: values["skin"] != ""},
			Hair:               pgtype.Text{String: values["hair"], Valid: values["hair"] != ""},
			Size:               pgtype.Text{String: values["size"], Valid: values["size"] != ""},
			Gender:             pgtype.Text{String: values["gender"], Valid: values["gender"] != ""},
			FaithDeity:         pgtype.Text{String: values["faith"], Valid: values["faith"] != ""},
			PersonalityTraits:  pgtype.Text{String: values["traits"], Valid: values["traits"] != ""},
			Ideals:             pgtype.Text{String: values["ideals"], Valid: values["ideals"] != ""},
			Bonds:              pgtype.Text{String: values["bonds"], Valid: values["bonds"] != ""},
			Flaws:              pgtype.Text{String: values["flaws"], Valid: values["flaws"] != ""},
			Backstory:          pgtype.Text{String: values["backstory"], Valid: values["backstory"] != ""},
			AlliesOrganizations: pgtype.Text{String: values["allies"], Valid: values["allies"] != ""},
		}

		updated, err := s.queries.UpdateCharacterDetails(s.ctx, params)
		if err != nil {
			return StatusMsg{Message: fmt.Sprintf("Error saving: %v", err), IsError: true}
		}
		s.details = &updated

		// Update alignment in characters table if changed
		if alignment := values["alignment"]; alignment != "" {
			charUpdated, err := s.queries.UpdateCharacterAlignment(s.ctx, db.UpdateCharacterAlignmentParams{
				ID:        s.char.ID,
				Alignment: pgtype.Text{String: alignment, Valid: true},
			})
			if err != nil {
				return StatusMsg{Message: fmt.Sprintf("Error saving alignment: %v", err), IsError: true}
			}
			s.char = charUpdated
		}

		return StatusMsg{Message: "Background saved", IsError: false}
	}
}

// openAttackModal creates and shows the attack add modal
func (s *SheetScreen) openAttackModal() {
	fields := []components.FormField{
		{Key: "name", Label: "Weapon Name", Type: components.FieldText, Required: true, Placeholder: "Longsword"},
		{Key: "attack_bonus", Label: "Attack Bonus", Type: components.FieldNumber, Placeholder: "+5"},
		{Key: "damage", Label: "Damage", Type: components.FieldText, Placeholder: "1d8+3"},
		{Key: "damage_type", Label: "Damage Type", Type: components.FieldSelect, Options: []string{"Slashing", "Piercing", "Bludgeoning", "Fire", "Cold", "Lightning", "Acid", "Poison", "Necrotic", "Radiant", "Force", "Psychic", "Thunder"}},
		{Key: "range", Label: "Range", Type: components.FieldText, Placeholder: "5 ft or 20/60 ft"},
		{Key: "properties", Label: "Properties", Type: components.FieldText, Placeholder: "Versatile, Finesse"},
		{Key: "notes", Label: "Notes", Type: components.FieldText, Placeholder: "Additional notes"},
	}

	s.attackModal = components.NewModal("Add Attack", fields, s.styles)
	s.attackModal.SetSize(s.width, s.height)
	s.attackModal.Show()
}

// saveAttack saves the attack modal values to the database
func (s *SheetScreen) saveAttack(values map[string]string) tea.Cmd {
	return func() tea.Msg {
		// Parse attack bonus
		var attackBonus pgtype.Int4
		if values["attack_bonus"] != "" {
			var bonus int
			// Handle both "+5" and "5" formats
			bonusStr := strings.TrimPrefix(values["attack_bonus"], "+")
			if _, err := fmt.Sscanf(bonusStr, "%d", &bonus); err == nil {
				attackBonus = pgtype.Int4{Int32: int32(bonus), Valid: true}
			}
		}

		// Calculate sort order (append to end)
		sortOrder := int32(len(s.attacks))

		params := db.CreateCharacterAttackParams{
			CharacterID: s.char.ID,
			SortOrder:   pgtype.Int4{Int32: sortOrder, Valid: true},
			Name:        values["name"],
			AttackBonus: attackBonus,
			Damage:      pgtype.Text{String: values["damage"], Valid: values["damage"] != ""},
			DamageType:  pgtype.Text{String: values["damage_type"], Valid: values["damage_type"] != ""},
			Range:       pgtype.Text{String: values["range"], Valid: values["range"] != ""},
			Properties:  pgtype.Text{String: values["properties"], Valid: values["properties"] != ""},
			Notes:       pgtype.Text{String: values["notes"], Valid: values["notes"] != ""},
		}

		_, err := s.queries.CreateCharacterAttack(s.ctx, params)
		if err != nil {
			return StatusMsg{Message: fmt.Sprintf("Error adding attack: %v", err), IsError: true}
		}

		// Refresh the attacks table
		s.refreshAttacksTable()
		s.attackModal = nil

		return StatusMsg{Message: "Attack added", IsError: false}
	}
}

// openActionModal creates and shows the action add modal
func (s *SheetScreen) openActionModal() {
	fields := []components.FormField{
		{Key: "name", Label: "Action Name", Type: components.FieldText, Required: true, Placeholder: "Second Wind"},
		{Key: "action_type", Label: "Type", Type: components.FieldSelect, Options: []string{"action", "bonus action", "reaction", "free", "movement", "other"}},
		{Key: "source", Label: "Source", Type: components.FieldText, Placeholder: "Fighter 1"},
		{Key: "uses_max", Label: "Max Uses", Type: components.FieldNumber, Placeholder: "1"},
		{Key: "uses_per", Label: "Recharge", Type: components.FieldSelect, Options: []string{"", "short rest", "long rest", "dawn", "dusk"}},
		{Key: "description", Label: "Description", Type: components.FieldText, Placeholder: "Regain 1d10 + level HP"},
	}

	s.actionModal = components.NewModal("Add Action", fields, s.styles)
	s.actionModal.SetSize(s.width, s.height)
	s.actionModal.Show()
}

// saveAction saves the action modal values to the database
func (s *SheetScreen) saveAction(values map[string]string) tea.Cmd {
	return func() tea.Msg {
		// Parse uses max
		var usesMax pgtype.Int4
		if values["uses_max"] != "" {
			var max int
			if _, err := fmt.Sscanf(values["uses_max"], "%d", &max); err == nil {
				usesMax = pgtype.Int4{Int32: int32(max), Valid: true}
			}
		}

		// Calculate sort order (append to end)
		sortOrder := int32(len(s.actions))

		params := db.CreateCharacterActionParams{
			CharacterID: s.char.ID,
			SortOrder:   pgtype.Int4{Int32: sortOrder, Valid: true},
			Name:        values["name"],
			ActionType:  pgtype.Text{String: values["action_type"], Valid: values["action_type"] != ""},
			Source:      pgtype.Text{String: values["source"], Valid: values["source"] != ""},
			Description: pgtype.Text{String: values["description"], Valid: values["description"] != ""},
			UsesPer:     pgtype.Text{String: values["uses_per"], Valid: values["uses_per"] != ""},
			UsesMax:     usesMax,
			UsesCurrent: usesMax, // Start with full uses
		}

		_, err := s.queries.CreateCharacterAction(s.ctx, params)
		if err != nil {
			return StatusMsg{Message: fmt.Sprintf("Error adding action: %v", err), IsError: true}
		}

		// Refresh the actions table
		s.refreshActionsTable()
		s.actionModal = nil

		return StatusMsg{Message: "Action added", IsError: false}
	}
}

// openSpellModal creates and shows the spell add modal
func (s *SheetScreen) openSpellModal() {
	// Default to current filter level, or 0 for cantrips if showing all
	defaultLevel := "0"
	if s.spellLevelFilter >= 0 {
		defaultLevel = fmt.Sprintf("%d", s.spellLevelFilter)
	}

	fields := []components.FormField{
		{Key: "name", Label: "Spell Name", Type: components.FieldText, Required: true, Placeholder: "Fireball"},
		{Key: "level", Label: "Level", Type: components.FieldSelect, Value: defaultLevel, Options: []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}},
		{Key: "school", Label: "School", Type: components.FieldSelect, Options: []string{"Abjuration", "Conjuration", "Divination", "Enchantment", "Evocation", "Illusion", "Necromancy", "Transmutation"}},
		{Key: "casting_time", Label: "Casting Time", Type: components.FieldText, Placeholder: "1 action"},
		{Key: "range", Label: "Range", Type: components.FieldText, Placeholder: "150 feet"},
		{Key: "components", Label: "Components", Type: components.FieldText, Placeholder: "V, S, M"},
		{Key: "duration", Label: "Duration", Type: components.FieldText, Placeholder: "Instantaneous"},
		{Key: "is_ritual", Label: "Ritual", Type: components.FieldCheckbox},
		{Key: "is_prepared", Label: "Prepared", Type: components.FieldCheckbox},
		{Key: "source", Label: "Source", Type: components.FieldText, Placeholder: "PHB"},
	}

	s.spellModal = components.NewModal("Add Spell", fields, s.styles)
	s.spellModal.SetSize(s.width, s.height)
	s.spellModal.Show()
}

// saveSpell saves the spell modal values to the database
func (s *SheetScreen) saveSpell(values map[string]string) tea.Cmd {
	return func() tea.Msg {
		// Parse level
		var level int32
		if values["level"] != "" {
			var lvl int
			if _, err := fmt.Sscanf(values["level"], "%d", &lvl); err == nil {
				level = int32(lvl)
			}
		}

		// Parse booleans
		isPrepared := values["is_prepared"] == "true"
		isRitual := values["is_ritual"] == "true"

		params := db.CreateCharacterSpellParams{
			CharacterID: s.char.ID,
			Name:        values["name"],
			Level:       level,
			School:      pgtype.Text{String: values["school"], Valid: values["school"] != ""},
			IsPrepared:  pgtype.Bool{Bool: isPrepared, Valid: true},
			IsRitual:    pgtype.Bool{Bool: isRitual, Valid: true},
			CastingTime: pgtype.Text{String: values["casting_time"], Valid: values["casting_time"] != ""},
			Range:       pgtype.Text{String: values["range"], Valid: values["range"] != ""},
			Components:  pgtype.Text{String: values["components"], Valid: values["components"] != ""},
			Duration:    pgtype.Text{String: values["duration"], Valid: values["duration"] != ""},
			Source:      pgtype.Text{String: values["source"], Valid: values["source"] != ""},
		}

		_, err := s.queries.CreateCharacterSpell(s.ctx, params)
		if err != nil {
			return StatusMsg{Message: fmt.Sprintf("Error adding spell: %v", err), IsError: true}
		}

		// Refresh the spells table
		s.refreshSpellsTable()
		s.spellModal = nil

		return StatusMsg{Message: "Spell added", IsError: false}
	}
}

// deleteSpell deletes a spell from the database
func (s *SheetScreen) deleteSpell(spellID pgtype.UUID) tea.Cmd {
	return func() tea.Msg {
		err := s.queries.DeleteCharacterSpell(s.ctx, spellID)
		if err != nil {
			return StatusMsg{Message: fmt.Sprintf("Error deleting spell: %v", err), IsError: true}
		}

		// Refresh the spells table
		s.refreshSpellsTable()

		return StatusMsg{Message: "Spell deleted", IsError: false}
	}
}

// toggleSpellPrepared toggles the prepared status of a spell
func (s *SheetScreen) toggleSpellPrepared(spellID pgtype.UUID) tea.Cmd {
	return func() tea.Msg {
		updated, err := s.queries.ToggleSpellPrepared(s.ctx, spellID)
		if err != nil {
			return StatusMsg{Message: fmt.Sprintf("Error toggling spell: %v", err), IsError: true}
		}

		// Refresh the spells table
		s.refreshSpellsTable()

		status := "unprepared"
		if updated.IsPrepared.Valid && updated.IsPrepared.Bool {
			status = "prepared"
		}
		return StatusMsg{Message: fmt.Sprintf("%s %s", updated.Name, status), IsError: false}
	}
}
