package tui

import (
	"context"

	"github.com/brady1408/dnd/internal/auth"
	"github.com/brady1408/dnd/internal/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"golang.org/x/crypto/ssh"
)

// Screen represents the current screen
type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenLogin
	ScreenRegister
	ScreenHome
	ScreenCreate
	ScreenSheet
	ScreenEdit
)

// App holds the application state
type App struct {
	// Dependencies
	queries     *db.Queries
	authService *auth.Service
	ctx         context.Context

	// Session state
	currentUser *db.User
	publicKey   ssh.PublicKey

	// Screen state
	screen       Screen
	screenModel  tea.Model
	width        int
	height       int
	err          error

	// Data
	characters []db.Character
	selectedCharacter *db.Character
}

// NewApp creates a new app instance
func NewApp(ctx context.Context, queries *db.Queries, publicKey ssh.PublicKey) *App {
	authService := auth.NewService(queries)

	app := &App{
		queries:     queries,
		authService: authService,
		ctx:         ctx,
		publicKey:   publicKey,
		screen:      ScreenWelcome,
		width:       80,
		height:      24,
	}

	// Try to auto-login with SSH key
	if publicKey != nil {
		user, err := authService.LoginWithPublicKey(ctx, publicKey)
		if err == nil {
			app.currentUser = user
			app.screen = ScreenHome
		}
	}

	return app
}

// Msg types for internal communication
type (
	// UserLoggedIn is sent when a user successfully logs in
	UserLoggedIn struct {
		User *db.User
	}

	// UserRegistered is sent when a user successfully registers
	UserRegistered struct {
		User *db.User
	}

	// CharactersLoaded is sent when characters are loaded
	CharactersLoaded struct {
		Characters []db.Character
	}

	// CharacterSelected is sent when a character is selected
	CharacterSelected struct {
		Character *db.Character
	}

	// CharacterCreated is sent when a character is created
	CharacterCreated struct {
		Character *db.Character
	}

	// CharacterUpdated is sent when a character is updated
	CharacterUpdated struct {
		Character *db.Character
	}

	// CharacterDeleted is sent when a character is deleted
	CharacterDeleted struct {
		ID uuid.UUID
	}

	// NavigateTo is sent to navigate to a screen
	NavigateTo struct {
		Screen Screen
	}

	// ErrorOccurred is sent when an error occurs
	ErrorOccurred struct {
		Err error
	}

	// WindowSizeMsg is sent when the window is resized
	WindowSizeMsg struct {
		Width  int
		Height int
	}
)

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	if a.currentUser != nil {
		return a.loadCharacters()
	}
	return nil
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if a.screen == ScreenWelcome || a.screen == ScreenHome {
				return a, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case UserLoggedIn:
		a.currentUser = msg.User
		a.screen = ScreenHome
		return a, a.loadCharacters()

	case UserRegistered:
		a.currentUser = msg.User
		a.screen = ScreenHome
		return a, a.loadCharacters()

	case CharactersLoaded:
		a.characters = msg.Characters

	case CharacterSelected:
		a.selectedCharacter = msg.Character
		a.screen = ScreenSheet

	case CharacterCreated:
		a.selectedCharacter = msg.Character
		a.screen = ScreenSheet
		return a, a.loadCharacters()

	case CharacterUpdated:
		a.selectedCharacter = msg.Character
		return a, a.loadCharacters()

	case CharacterDeleted:
		a.selectedCharacter = nil
		a.screen = ScreenHome
		return a, a.loadCharacters()

	case NavigateTo:
		a.screen = msg.Screen
		if msg.Screen == ScreenHome {
			return a, a.loadCharacters()
		}

	case ErrorOccurred:
		a.err = msg.Err
	}

	// Update the current screen model if we have one
	if a.screenModel != nil {
		var cmd tea.Cmd
		a.screenModel, cmd = a.screenModel.Update(msg)
		return a, cmd
	}

	return a, nil
}

// View implements tea.Model
func (a *App) View() string {
	// This is a placeholder - actual views are in screen files
	return "Loading..."
}

// loadCharacters loads characters for the current user
func (a *App) loadCharacters() tea.Cmd {
	return func() tea.Msg {
		if a.currentUser == nil {
			return CharactersLoaded{Characters: nil}
		}
		chars, err := a.queries.GetCharactersByUserID(a.ctx, a.currentUser.ID)
		if err != nil {
			return ErrorOccurred{Err: err}
		}
		return CharactersLoaded{Characters: chars}
	}
}

// Getters for use by screens

func (a *App) Queries() *db.Queries { return a.queries }
func (a *App) AuthService() *auth.Service { return a.authService }
func (a *App) Context() context.Context { return a.ctx }
func (a *App) CurrentUser() *db.User { return a.currentUser }
func (a *App) PublicKey() ssh.PublicKey { return a.publicKey }
func (a *App) Characters() []db.Character { return a.characters }
func (a *App) SelectedCharacter() *db.Character { return a.selectedCharacter }
func (a *App) Width() int { return a.width }
func (a *App) Height() int { return a.height }
