package screens

import (
	"context"
	"strings"

	"github.com/brady1408/dnd/internal/auth"
	"github.com/brady1408/dnd/internal/db"
	"github.com/brady1408/dnd/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/crypto/ssh"
)

type WelcomeMode int

const (
	ModeMenu WelcomeMode = iota
	ModeLogin
	ModeRegister
	ModeRegisterSSH
	ModeLoginSSH
)

type WelcomeScreen struct {
	ctx         context.Context
	authService *auth.Service
	publicKey   ssh.PublicKey
	styles      *styles.Styles

	mode        WelcomeMode
	menuIndex   int
	emailInput  textinput.Model
	passInput   textinput.Model
	focusIndex  int
	err         string
	width       int
	height      int
}

type UserLoggedInMsg struct {
	User *db.User
}

func NewWelcomeScreen(ctx context.Context, authService *auth.Service, publicKey ssh.PublicKey, s *styles.Styles) *WelcomeScreen {
	emailInput := textinput.New()
	emailInput.Placeholder = "Email"
	emailInput.CharLimit = 255
	emailInput.Width = 30

	passInput := textinput.New()
	passInput.Placeholder = "Password"
	passInput.EchoMode = textinput.EchoPassword
	passInput.EchoCharacter = '*'
	passInput.CharLimit = 100
	passInput.Width = 30

	return &WelcomeScreen{
		ctx:         ctx,
		authService: authService,
		publicKey:   publicKey,
		styles:      s,
		mode:        ModeMenu,
		emailInput:  emailInput,
		passInput:   passInput,
		width:       80,
		height:      24,
	}
}

func (w *WelcomeScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (w *WelcomeScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height

	case tea.KeyMsg:
		w.err = ""

		switch w.mode {
		case ModeMenu:
			return w.updateMenu(msg)
		case ModeLogin, ModeRegister:
			return w.updateForm(msg)
		case ModeRegisterSSH:
			return w.updateSSHRegister(msg)
		case ModeLoginSSH:
			return w.updateSSHLogin(msg)
		}
	}

	// Update inputs
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if w.mode == ModeLogin || w.mode == ModeRegister {
		if w.focusIndex == 0 {
			w.emailInput, cmd = w.emailInput.Update(msg)
			cmds = append(cmds, cmd)
		} else if w.focusIndex == 1 {
			w.passInput, cmd = w.passInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return w, tea.Batch(cmds...)
}

func (w *WelcomeScreen) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuItems := w.getMenuItems()

	switch msg.String() {
	case "up", "k":
		if w.menuIndex > 0 {
			w.menuIndex--
		}
	case "down", "j":
		if w.menuIndex < len(menuItems)-1 {
			w.menuIndex++
		}
	case "enter":
		switch menuItems[w.menuIndex] {
		case "Login with SSH Key":
			w.mode = ModeLoginSSH
		case "Login with Email":
			w.mode = ModeLogin
			w.focusIndex = 0
			w.emailInput.Focus()
			return w, textinput.Blink
		case "Register with Email":
			w.mode = ModeRegister
			w.focusIndex = 0
			w.emailInput.Focus()
			return w, textinput.Blink
		case "Register with SSH Key":
			w.mode = ModeRegisterSSH
		}
	case "q", "ctrl+c":
		return w, tea.Quit
	}

	return w, nil
}

func (w *WelcomeScreen) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "down":
		w.focusIndex++
		if w.focusIndex > 2 {
			w.focusIndex = 0
		}
		w.updateFocus()
		return w, nil

	case "shift+tab", "up":
		w.focusIndex--
		if w.focusIndex < 0 {
			w.focusIndex = 2
		}
		w.updateFocus()
		return w, nil

	case "enter":
		if w.focusIndex == 2 {
			// Submit button
			return w.submitForm()
		}
		// Move to next field
		w.focusIndex++
		if w.focusIndex > 2 {
			w.focusIndex = 0
		}
		w.updateFocus()
		return w, nil

	case "esc":
		w.mode = ModeMenu
		w.emailInput.SetValue("")
		w.passInput.SetValue("")
		return w, nil
	}

	// Pass key to focused input
	var cmd tea.Cmd
	if w.focusIndex == 0 {
		w.emailInput, cmd = w.emailInput.Update(msg)
	} else if w.focusIndex == 1 {
		w.passInput, cmd = w.passInput.Update(msg)
	}
	return w, cmd
}

func (w *WelcomeScreen) updateSSHRegister(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "y":
		if w.publicKey != nil {
			user, err := w.authService.RegisterWithPublicKey(w.ctx, w.publicKey)
			if err != nil {
				w.err = err.Error()
				return w, nil
			}
			return w, func() tea.Msg { return UserLoggedInMsg{User: user} }
		}
		w.err = "No SSH key detected"
		return w, nil

	case "esc", "n":
		w.mode = ModeMenu
		return w, nil
	}

	return w, nil
}

func (w *WelcomeScreen) updateSSHLogin(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "y":
		if w.publicKey != nil {
			user, err := w.authService.LoginWithPublicKey(w.ctx, w.publicKey)
			if err != nil {
				w.err = "SSH key not registered. Please register first."
				return w, nil
			}
			return w, func() tea.Msg { return UserLoggedInMsg{User: user} }
		}
		w.err = "No SSH key detected"
		return w, nil

	case "esc", "n":
		w.mode = ModeMenu
		return w, nil
	}

	return w, nil
}

func (w *WelcomeScreen) submitForm() (tea.Model, tea.Cmd) {
	email := strings.TrimSpace(w.emailInput.Value())
	pass := w.passInput.Value()

	if email == "" {
		w.err = "Email is required"
		return w, nil
	}
	if pass == "" {
		w.err = "Password is required"
		return w, nil
	}
	if len(pass) < 6 {
		w.err = "Password must be at least 6 characters"
		return w, nil
	}

	var user *db.User
	var err error

	if w.mode == ModeLogin {
		user, err = w.authService.LoginWithPassword(w.ctx, email, pass)
	} else {
		user, err = w.authService.RegisterWithPassword(w.ctx, email, pass)
	}

	if err != nil {
		w.err = err.Error()
		return w, nil
	}

	// Clear inputs
	w.emailInput.SetValue("")
	w.passInput.SetValue("")

	return w, func() tea.Msg { return UserLoggedInMsg{User: user} }
}

func (w *WelcomeScreen) updateFocus() {
	w.emailInput.Blur()
	w.passInput.Blur()

	switch w.focusIndex {
	case 0:
		w.emailInput.Focus()
	case 1:
		w.passInput.Focus()
	}
}

func (w *WelcomeScreen) getMenuItems() []string {
	items := []string{"Login with Email", "Register with Email"}
	if w.publicKey != nil {
		// Insert SSH login option at the beginning since it's the easiest
		items = []string{"Login with SSH Key", "Login with Email", "Register with Email", "Register with SSH Key"}
	}
	return items
}

func (w *WelcomeScreen) View() string {
	var b strings.Builder

	// Logo
	b.WriteString(w.styles.Logo.Render(styles.LogoText))
	b.WriteString("\n")

	switch w.mode {
	case ModeMenu:
		b.WriteString(w.renderMenu())
	case ModeLogin:
		b.WriteString(w.renderForm("Login"))
	case ModeRegister:
		b.WriteString(w.renderForm("Register"))
	case ModeRegisterSSH:
		b.WriteString(w.renderSSHRegister())
	case ModeLoginSSH:
		b.WriteString(w.renderSSHLogin())
	}

	// Error message
	if w.err != "" {
		b.WriteString("\n")
		b.WriteString(w.styles.ErrorText.Render("Error: " + w.err))
	}

	// Help
	b.WriteString("\n\n")
	switch w.mode {
	case ModeMenu:
		b.WriteString(w.styles.Help.Render("↑/↓: navigate • enter: select • q: quit"))
	default:
		b.WriteString(w.styles.Help.Render("tab: next field • enter: submit • esc: back"))
	}

	return lipgloss.Place(w.width, w.height,
		lipgloss.Center, lipgloss.Center,
		b.String())
}

func (w *WelcomeScreen) renderMenu() string {
	var b strings.Builder

	b.WriteString(w.styles.Title.Render("Welcome, Adventurer!"))
	b.WriteString("\n\n")

	menuItems := w.getMenuItems()
	for i, item := range menuItems {
		cursor := "  "
		style := w.styles.Unselected
		if i == w.menuIndex {
			cursor = "> "
			style = w.styles.Selected
		}
		b.WriteString(w.styles.Cursor.Render(cursor))
		b.WriteString(style.Render(item))
		b.WriteString("\n")
	}

	if w.publicKey != nil {
		b.WriteString("\n")
		b.WriteString(w.styles.SuccessText.Render("✓ SSH Key detected"))
	}

	return b.String()
}

func (w *WelcomeScreen) renderForm(title string) string {
	var b strings.Builder

	b.WriteString(w.styles.Title.Render(title))
	b.WriteString("\n\n")

	// Email field
	emailStyle := w.styles.InputField
	if w.focusIndex == 0 {
		emailStyle = w.styles.FocusedInput
	}
	b.WriteString("Email:\n")
	b.WriteString(emailStyle.Render(w.emailInput.View()))
	b.WriteString("\n\n")

	// Password field
	passStyle := w.styles.InputField
	if w.focusIndex == 1 {
		passStyle = w.styles.FocusedInput
	}
	b.WriteString("Password:\n")
	b.WriteString(passStyle.Render(w.passInput.View()))
	b.WriteString("\n\n")

	// Submit button
	btnStyle := w.styles.Button
	if w.focusIndex == 2 {
		btnStyle = w.styles.FocusedButton
	}
	b.WriteString(btnStyle.Render("[ " + title + " ]"))

	return b.String()
}

func (w *WelcomeScreen) renderSSHRegister() string {
	var b strings.Builder

	b.WriteString(w.styles.Title.Render("Register with SSH Key"))
	b.WriteString("\n\n")

	if w.publicKey != nil {
		keyStr := auth.NormalizePublicKey(w.publicKey)
		// Show truncated key
		if len(keyStr) > 50 {
			keyStr = keyStr[:50] + "..."
		}
		b.WriteString("Your SSH key:\n")
		b.WriteString(w.styles.Box.Render(keyStr))
		b.WriteString("\n\n")
		b.WriteString("Register with this key? (y/n)")
	} else {
		b.WriteString(w.styles.ErrorText.Render("No SSH key detected."))
		b.WriteString("\n")
		b.WriteString("Please connect with an SSH key or use email registration.")
	}

	return b.String()
}

func (w *WelcomeScreen) renderSSHLogin() string {
	var b strings.Builder

	b.WriteString(w.styles.Title.Render("Login with SSH Key"))
	b.WriteString("\n\n")

	if w.publicKey != nil {
		keyStr := auth.NormalizePublicKey(w.publicKey)
		// Show truncated key
		if len(keyStr) > 50 {
			keyStr = keyStr[:50] + "..."
		}
		b.WriteString("Your SSH key:\n")
		b.WriteString(w.styles.Box.Render(keyStr))
		b.WriteString("\n\n")
		b.WriteString("Login with this key? (y/n)")
	} else {
		b.WriteString(w.styles.ErrorText.Render("No SSH key detected."))
		b.WriteString("\n")
		b.WriteString("Please connect with an SSH key or use email login.")
	}

	return b.String()
}
