package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brady1408/dnd/internal/auth"
	"github.com/brady1408/dnd/internal/db"
	"github.com/brady1408/dnd/internal/tui/screens"
	"github.com/brady1408/dnd/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/jackc/pgx/v5/pgxpool"
	gossh "golang.org/x/crypto/ssh"
)

const (
	host = "0.0.0.0"
	port = "2222"
)

// Config holds application configuration
type Config struct {
	DatabaseURL string
	Host        string
	Port        string
}

func main() {
	// Load configuration
	cfg := Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://postgres:postgres@192.168.23.44:5434/dnd_character?sslmode=disable"),
		Host:        getEnv("HOST", host),
		Port:        getEnv("PORT", port),
	}

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	queries := db.New(pool)

	// Create SSH server
	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			// Accept all public keys - we do our own auth
			return true
		}),
		wish.WithPasswordAuth(func(ctx ssh.Context, password string) bool {
			// Accept all passwords - we do our own auth
			return true
		}),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler(queries)),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create SSH server: %v", err)
	}

	// Start server in goroutine
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Starting D&D Character Server on %s:%s", cfg.Host, cfg.Port)
	log.Printf("Connect with: ssh -p %s localhost", cfg.Port)

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatalf("SSH server error: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
}

func teaHandler(queries *db.Queries) bubbletea.Handler {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		pty, _, _ := s.Pty()

		// Get public key from session
		var publicKey gossh.PublicKey
		if s.PublicKey() != nil {
			publicKey = s.PublicKey()
		}

		m := NewMainModel(queries, publicKey, pty.Window.Width, pty.Window.Height)
		return m, []tea.ProgramOption{tea.WithAltScreen()}
	}
}

// MainModel is the root model for the application
type MainModel struct {
	queries   *db.Queries
	auth      *auth.Service
	ctx       context.Context
	publicKey gossh.PublicKey

	// Current screen
	screen    string
	user      *db.User
	chars     []db.Character
	selChar   *db.Character

	// Screen models
	welcome *screens.WelcomeScreen
	home    *screens.HomeScreen
	create  *screens.CreateScreen
	sheet   *screens.SheetScreen

	width  int
	height int
	err    error
}

func NewMainModel(queries *db.Queries, publicKey gossh.PublicKey, width, height int) *MainModel {
	ctx := context.Background()
	authService := auth.NewService(queries)

	m := &MainModel{
		queries:   queries,
		auth:      authService,
		ctx:       ctx,
		publicKey: publicKey,
		screen:    "welcome",
		width:     width,
		height:    height,
	}

	// Try auto-login with SSH key
	if publicKey != nil {
		user, err := authService.LoginWithPublicKey(ctx, publicKey)
		if err == nil {
			m.user = user
			m.screen = "home"
			m.home = screens.NewHomeScreen(ctx, queries, user)
		}
	}

	if m.screen == "welcome" {
		m.welcome = screens.NewWelcomeScreen(ctx, authService, publicKey)
	}

	return m
}

func (m *MainModel) Init() tea.Cmd {
	switch m.screen {
	case "welcome":
		return m.welcome.Init()
	case "home":
		return m.home.Init()
	case "create":
		return m.create.Init()
	case "sheet":
		return m.sheet.Init()
	}
	return nil
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	// Handle screen-specific messages
	case screens.UserLoggedInMsg:
		m.user = msg.User
		m.screen = "home"
		m.home = screens.NewHomeScreen(m.ctx, m.queries, m.user)
		return m, m.home.Init()

	case screens.CharactersLoadedMsg:
		m.chars = msg.Characters
		if m.home != nil {
			m.home.SetCharacters(msg.Characters)
		}

	case screens.NavigateToCreateMsg:
		m.screen = "create"
		m.create = screens.NewCreateScreen(m.ctx, m.queries, m.user.ID)
		return m, m.create.Init()

	case screens.CharacterSelectedMsg:
		m.selChar = &msg.Character
		m.screen = "sheet"
		m.sheet = screens.NewSheetScreen(m.ctx, m.queries, msg.Character)
		return m, m.sheet.Init()

	case screens.CharacterCreatedMsg:
		m.selChar = &msg.Character
		m.screen = "sheet"
		m.sheet = screens.NewSheetScreen(m.ctx, m.queries, msg.Character)
		return m, m.sheet.Init()

	case screens.CharacterUpdatedMsg:
		m.selChar = &msg.Character
		if m.sheet != nil {
			m.sheet.SetCharacter(msg.Character)
		}

	case screens.CharacterDeletedMsg:
		m.selChar = nil
		m.screen = "home"
		m.home = screens.NewHomeScreen(m.ctx, m.queries, m.user)
		return m, m.home.Init()

	case screens.NavigateBackMsg:
		switch m.screen {
		case "create", "sheet":
			m.screen = "home"
			m.home = screens.NewHomeScreen(m.ctx, m.queries, m.user)
			return m, m.home.Init()
		}

	case screens.LogoutMsg:
		m.user = nil
		m.screen = "welcome"
		m.welcome = screens.NewWelcomeScreen(m.ctx, m.auth, m.publicKey)
		return m, m.welcome.Init()
	}

	// Update current screen
	var cmd tea.Cmd
	switch m.screen {
	case "welcome":
		var newModel tea.Model
		newModel, cmd = m.welcome.Update(msg)
		m.welcome = newModel.(*screens.WelcomeScreen)
	case "home":
		var newModel tea.Model
		newModel, cmd = m.home.Update(msg)
		m.home = newModel.(*screens.HomeScreen)
	case "create":
		var newModel tea.Model
		newModel, cmd = m.create.Update(msg)
		m.create = newModel.(*screens.CreateScreen)
	case "sheet":
		var newModel tea.Model
		newModel, cmd = m.sheet.Update(msg)
		m.sheet = newModel.(*screens.SheetScreen)
	}

	return m, cmd
}

func (m *MainModel) View() string {
	var content string

	switch m.screen {
	case "welcome":
		content = m.welcome.View()
	case "home":
		content = m.home.View()
	case "create":
		content = m.create.View()
	case "sheet":
		content = m.sheet.View()
	default:
		content = "Loading..."
	}

	if m.err != nil {
		content += "\n" + styles.ErrorText.Render("Error: "+m.err.Error())
	}

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Ensure MainModel implements tea.Model
var _ tea.Model = (*MainModel)(nil)

// Ensure textinput is imported (used by screens)
var _ = textinput.Model{}
