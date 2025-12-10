package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/brady1408/dnd/internal/db"
	"github.com/brady1408/dnd/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jackc/pgx/v5/pgtype"
)

type HomeScreen struct {
	ctx        context.Context
	queries    *db.Queries
	user       *db.User
	characters []db.Character

	selectedIndex int
	width         int
	height        int
	confirmDelete bool
}

type NavigateToCreateMsg struct{}
type CharacterSelectedMsg struct {
	Character db.Character
}
type CharacterDeletedMsg struct {
	ID pgtype.UUID
}
type LogoutMsg struct{}

func NewHomeScreen(ctx context.Context, queries *db.Queries, user *db.User) *HomeScreen {
	return &HomeScreen{
		ctx:     ctx,
		queries: queries,
		user:    user,
		width:   80,
		height:  24,
	}
}

func (h *HomeScreen) SetCharacters(chars []db.Character) {
	h.characters = chars
	if h.selectedIndex >= len(chars) && len(chars) > 0 {
		h.selectedIndex = len(chars) - 1
	}
}

func (h *HomeScreen) Init() tea.Cmd {
	return h.loadCharacters()
}

func (h *HomeScreen) loadCharacters() tea.Cmd {
	return func() tea.Msg {
		chars, err := h.queries.GetCharactersByUserID(h.ctx, h.user.ID)
		if err != nil {
			return nil
		}
		return CharactersLoadedMsg{Characters: chars}
	}
}

type CharactersLoadedMsg struct {
	Characters []db.Character
}

func (h *HomeScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height

	case CharactersLoadedMsg:
		h.characters = msg.Characters

	case tea.KeyMsg:
		if h.confirmDelete {
			return h.handleDeleteConfirm(msg)
		}
		return h.handleInput(msg)
	}

	return h, nil
}

func (h *HomeScreen) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if h.selectedIndex > 0 {
			h.selectedIndex--
		}

	case "down", "j":
		// +1 for "Create New Character" option
		maxIndex := len(h.characters)
		if h.selectedIndex < maxIndex {
			h.selectedIndex++
		}

	case "enter":
		if h.selectedIndex == len(h.characters) {
			// Create new character
			return h, func() tea.Msg { return NavigateToCreateMsg{} }
		}
		if h.selectedIndex < len(h.characters) {
			char := h.characters[h.selectedIndex]
			return h, func() tea.Msg { return CharacterSelectedMsg{Character: char} }
		}

	case "d", "delete":
		if h.selectedIndex < len(h.characters) {
			h.confirmDelete = true
		}

	case "l":
		return h, func() tea.Msg { return LogoutMsg{} }

	case "q", "ctrl+c":
		return h, tea.Quit
	}

	return h, nil
}

func (h *HomeScreen) handleDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if h.selectedIndex < len(h.characters) {
			charID := h.characters[h.selectedIndex].ID
			h.confirmDelete = false

			return h, func() tea.Msg {
				_ = h.queries.DeleteCharacter(h.ctx, charID)
				return CharacterDeletedMsg{ID: charID}
			}
		}

	case "n", "N", "esc":
		h.confirmDelete = false
	}

	return h, nil
}

func (h *HomeScreen) View() string {
	var b strings.Builder

	// Header
	b.WriteString(styles.Logo.Render(styles.LogoSmall))
	b.WriteString("\n")

	// User info
	userInfo := "Logged in"
	if h.user != nil && h.user.Email.Valid {
		userInfo = "Logged in as: " + h.user.Email.String
	}
	b.WriteString(styles.Subtitle.Render(userInfo))
	b.WriteString("\n\n")

	// Title
	b.WriteString(styles.Title.Render("Your Characters"))
	b.WriteString("\n\n")

	// Character list
	if len(h.characters) == 0 {
		b.WriteString(styles.Muted.Render("No characters yet. Create your first adventurer!"))
		b.WriteString("\n\n")
	} else {
		for i, char := range h.characters {
			cursor := "  "
			style := styles.Unselected
			if i == h.selectedIndex {
				cursor = "> "
				style = styles.Selected
			}

			line := fmt.Sprintf("%s%s - Level %d %s %s",
				cursor,
				char.Name,
				char.Level,
				char.Race,
				char.Class,
			)

			b.WriteString(style.Render(line))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Create new character option
	createCursor := "  "
	createStyle := styles.Unselected
	if h.selectedIndex == len(h.characters) {
		createCursor = "> "
		createStyle = styles.Selected
	}
	b.WriteString(styles.Cursor.Render(createCursor))
	b.WriteString(createStyle.Render("+ Create New Character"))
	b.WriteString("\n")

	// Delete confirmation
	if h.confirmDelete && h.selectedIndex < len(h.characters) {
		b.WriteString("\n")
		char := h.characters[h.selectedIndex]
		b.WriteString(styles.WarningText.Render(fmt.Sprintf(
			"Delete %s? This cannot be undone. (y/n)",
			char.Name,
		)))
	}

	// Help
	b.WriteString("\n\n")
	if h.confirmDelete {
		b.WriteString(styles.Help.Render("y: confirm delete • n: cancel"))
	} else {
		b.WriteString(styles.Help.Render("↑/↓: navigate • enter: select • d: delete • l: logout • q: quit"))
	}

	return lipgloss.Place(h.width, h.height,
		lipgloss.Center, lipgloss.Center,
		b.String())
}
