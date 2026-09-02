// Package tui contains Poli's interactive terminal request workbench.
package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	curlpkg "github.com/jojipanackal/poli/internal/curl"
	httppkg "github.com/jojipanackal/poli/internal/http"
	modelpkg "github.com/jojipanackal/poli/internal/model"
	"github.com/jojipanackal/poli/internal/store"
)

const (
	defaultImportedName = "Imported Request"
	defaultDraftName    = "New Request"
)

type screen uint8

const (
	screenBrowse screen = iota
	screenEdit
	screenImport
)

type pane uint8

const (
	paneList pane = iota
	paneEditor
	paneResponse
)

type editorField uint8

const (
	fieldName editorField = iota
	fieldMethod
	fieldURL
	fieldHeaders
	fieldBody
)

type responseMsg struct {
	response httppkg.Response
	err      error
	saveErr  error
}

type appModel struct {
	group        string
	requests     []modelpkg.Request
	selected     int
	current      modelpkg.Request
	originalName string
	isNew        bool

	screen screen
	pane   pane
	field  editorField

	nameInput    textinput.Model
	methodInput  textinput.Model
	urlInput     textinput.Model
	headersInput textarea.Model
	bodyInput    textarea.Model

	importNameInput textinput.Model
	importInput     textarea.Model

	responseViewport viewport.Model
	responseCode     int
	responseStatus   string
	responseHeaders  map[string]string
	responseBody     string
	responseDuration time.Duration
	showHeaders      bool

	running bool
	status  string
	width   int
	height  int
}

var (
	ink    = lipgloss.Color("#171717")
	paper  = lipgloss.Color("#F7F1E8")
	muted  = lipgloss.Color("#6B6258")
	lime   = lipgloss.Color("#C7FF3D")
	pink   = lipgloss.Color("#FF7297")
	purple = lipgloss.Color("#7652F4")
	yellow = lipgloss.Color("#FFD84D")
	blue   = lipgloss.Color("#8BD8FF")
)

var (
	brandStyle  = lipgloss.NewStyle().Bold(true).Foreground(ink).Background(lime).Padding(0, 1)
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(ink)
	mutedStyle  = lipgloss.NewStyle().Foreground(muted)
	labelStyle  = lipgloss.NewStyle().Bold(true).Foreground(ink)
	keyStyle    = lipgloss.NewStyle().Bold(true).Foreground(purple)
	activeStyle = lipgloss.NewStyle().Bold(true).Foreground(ink).Background(lime)
	statusStyle = lipgloss.NewStyle().Foreground(ink).Background(yellow).Padding(0, 1)
	errStyle    = lipgloss.NewStyle().Foreground(ink).Background(pink).Padding(0, 1)
	codeStyle   = lipgloss.NewStyle().Foreground(paper).Background(purple).Padding(0, 1)
)

// Run starts the request workbench for group. initialCurl can be used to open
// the editor with a cURL command already imported.
func Run(group, initialCurl string) error {
	m, err := newModel(group, initialCurl)
	if err != nil {
		return err
	}

	_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func newModel(group, initialCurl string) (*appModel, error) {
	m := &appModel{
		group:            group,
		selected:         -1,
		pane:             paneList,
		responseViewport: viewport.New(0, 0),
		responseHeaders:  make(map[string]string),
	}
	m.setupInputs()

	requests, err := store.ListRequests(group)
	if err != nil {
		return nil, err
	}
	m.requests = requests

	if strings.TrimSpace(initialCurl) != "" {
		req, err := curlpkg.Parse(defaultImportedName, initialCurl)
		if err != nil {
			return nil, fmt.Errorf("parse cURL: %w", err)
		}
		m.current = req
		m.originalName = ""
		m.isNew = true
		m.screen = screenEdit
		m.pane = paneEditor
		m.syncInputs()
		m.focusField(fieldName)
		m.status = "cURL imported — edit the request, then press ctrl+s"
		return m, nil
	}

	if len(requests) > 0 {
		m.selectRequest(0)
	} else {
		m.newDraft()
	}
	return m, nil
}

func (m *appModel) setupInputs() {
	m.nameInput = textinput.New()
	m.nameInput.Prompt = ""
	m.nameInput.CharLimit = 100

	m.methodInput = textinput.New()
	m.methodInput.Prompt = ""
	m.methodInput.CharLimit = 16

	m.urlInput = textinput.New()
	m.urlInput.Prompt = ""
	m.urlInput.CharLimit = 2000

	m.headersInput = textarea.New()
	m.headersInput.Prompt = ""
	m.headersInput.ShowLineNumbers = false
	m.headersInput.CharLimit = 10000
	m.headersInput.SetHeight(5)

	m.bodyInput = textarea.New()
	m.bodyInput.Prompt = ""
	m.bodyInput.ShowLineNumbers = false
	m.bodyInput.CharLimit = 50000
	m.bodyInput.SetHeight(8)

	m.importNameInput = textinput.New()
	m.importNameInput.Prompt = ""
	m.importNameInput.CharLimit = 100

	m.importInput = textarea.New()
	m.importInput.Prompt = ""
	m.importInput.ShowLineNumbers = false
	m.importInput.CharLimit = 50000
	m.importInput.SetHeight(12)
}

func (m *appModel) Init() tea.Cmd {
	return nil
}

func (m *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeInputs()
		m.resizeViewport()
		return m, nil
	case responseMsg:
		m.running = false
		if msg.err != nil {
			m.status = "Request failed: " + msg.err.Error()
			m.responseViewport.SetContent(errorContent(msg.err))
			return m, nil
		}
		m.setResponse(msg.response)
		if msg.saveErr != nil {
			m.status = "Finished, but could not save response: " + msg.saveErr.Error()
		} else {
			m.status = fmt.Sprintf("Finished %s in %s", msg.response.Status, formatDuration(msg.response.Duration))
		}
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.screen {
	case screenEdit:
		return m.updateEdit(msg)
	case screenImport:
		return m.updateImport(msg)
	default:
		return m.updateBrowse(msg)
	}
}

func (m *appModel) updateBrowse(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	key := keyMsg.String()

	if m.pane == paneResponse && isViewportKey(key) {
		var cmd tea.Cmd
		m.responseViewport, cmd = m.responseViewport.Update(msg)
		return m, cmd
	}

	switch key {
	case "q", "esc":
		return m, tea.Quit
	case "tab", "right", "l":
		m.pane = (m.pane + 1) % 3
	case "shift+tab", "left":
		m.pane = (m.pane + 2) % 3
	case "h":
		m.showHeaders = !m.showHeaders
		m.renderResponseContent()
	case "up", "k":
		m.moveSelection(-1)
	case "down", "j":
		m.moveSelection(1)
	case "enter", "e":
		if m.selected >= 0 {
			m.current = m.requests[m.selected]
			m.originalName = m.current.Name
			m.isNew = false
			m.syncInputs()
		}
		m.screen = screenEdit
		m.pane = paneEditor
		m.focusField(fieldName)
	case "n":
		m.newDraft()
	case "i":
		m.startImport()
	case "r", "ctrl+r":
		return m, m.runCurrent()
	case "s", "ctrl+s":
		if err := m.saveCurrent(); err != nil {
			m.status = "Save failed: " + err.Error()
		} else {
			m.status = "Saved " + m.current.Name
		}
	}
	return m, nil
}

func (m *appModel) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	key := keyMsg.String()

	switch key {
	case "esc":
		m.blurInputs()
		m.screen = screenBrowse
		m.pane = paneEditor
		return m, nil
	case "ctrl+s":
		if err := m.saveCurrent(); err != nil {
			m.status = "Save failed: " + err.Error()
			return m, nil
		}
		m.screen = screenBrowse
		m.pane = paneEditor
		m.status = "Saved " + m.current.Name
		return m, nil
	case "ctrl+r":
		return m, m.runCurrent()
	case "tab":
		m.focusField((m.field + 1) % 5)
		return m, nil
	case "shift+tab":
		m.focusField((m.field + 4) % 5)
		return m, nil
	case "enter":
		if m.field <= fieldURL {
			m.focusField((m.field + 1) % 5)
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.field {
	case fieldName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case fieldMethod:
		m.methodInput, cmd = m.methodInput.Update(msg)
	case fieldURL:
		m.urlInput, cmd = m.urlInput.Update(msg)
	case fieldHeaders:
		m.headersInput, cmd = m.headersInput.Update(msg)
	case fieldBody:
		m.bodyInput, cmd = m.bodyInput.Update(msg)
	}
	return m, cmd
}

func (m *appModel) updateImport(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	key := keyMsg.String()

	switch key {
	case "esc":
		m.screen = screenBrowse
		m.pane = paneList
		return m, nil
	case "ctrl+s":
		name := strings.TrimSpace(m.importNameInput.Value())
		if name == "" {
			name = defaultImportedName
		}
		req, err := curlpkg.Parse(name, m.importInput.Value())
		if err != nil {
			m.status = "Import failed: " + err.Error()
			return m, nil
		}
		m.current = req
		m.originalName = ""
		m.isNew = true
		m.screen = screenEdit
		m.pane = paneEditor
		m.syncInputs()
		m.focusField(fieldName)
		m.status = "cURL imported — edit the request, then press ctrl+s"
		return m, nil
	case "tab":
		m.importNameInput.Blur()
		m.importInput.Focus()
		return m, nil
	case "shift+tab":
		m.importInput.Blur()
		m.importNameInput.Focus()
		return m, nil
	}

	var cmd tea.Cmd
	if m.importNameInput.Focused() {
		m.importNameInput, cmd = m.importNameInput.Update(msg)
	} else {
		m.importInput, cmd = m.importInput.Update(msg)
	}
	return m, cmd
}

func (m *appModel) moveSelection(delta int) {
	if len(m.requests) == 0 {
		return
	}
	if m.selected < 0 {
		m.selected = 0
	}
	m.selected += delta
	if m.selected < 0 {
		m.selected = len(m.requests) - 1
	}
	if m.selected >= len(m.requests) {
		m.selected = 0
	}
	m.selectRequest(m.selected)
}

func (m *appModel) selectRequest(index int) {
	if index < 0 || index >= len(m.requests) {
		return
	}
	m.selected = index
	m.current = m.requests[index]
	m.originalName = m.current.Name
	m.isNew = false
	m.syncInputs()

	if saved, err := store.LoadResponse(m.group, m.current.Name); err == nil {
		m.setSavedResponse(saved)
	} else {
		m.clearResponse()
	}
}

func (m *appModel) newDraft() {
	now := time.Now()
	m.current = modelpkg.Request{
		Name:      defaultDraftName,
		Method:    "GET",
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.originalName = ""
	m.isNew = true
	m.screen = screenEdit
	m.pane = paneEditor
	m.syncInputs()
	m.clearResponse()
	m.focusField(fieldName)
	m.status = "New request — fill in the fields, then press ctrl+s"
}

func (m *appModel) startImport() {
	m.screen = screenImport
	m.pane = paneEditor
	m.importNameInput.SetValue(defaultImportedName)
	m.importInput.SetValue("")
	m.importNameInput.Focus()
	m.importInput.Blur()
	m.status = "Paste a cURL command; press ctrl+s to import"
}

func (m *appModel) syncInputs() {
	m.nameInput.SetValue(m.current.Name)
	m.methodInput.SetValue(strings.ToUpper(m.current.Method))
	m.urlInput.SetValue(m.current.URL)
	m.headersInput.SetValue(formatHeaders(m.current.Headers))
	m.bodyInput.SetValue(m.current.Body)
}

func (m *appModel) focusField(field editorField) {
	m.blurInputs()
	m.field = field
	switch field {
	case fieldName:
		m.nameInput.Focus()
	case fieldMethod:
		m.methodInput.Focus()
	case fieldURL:
		m.urlInput.Focus()
	case fieldHeaders:
		m.headersInput.Focus()
	case fieldBody:
		m.bodyInput.Focus()
	}
}

func (m *appModel) blurInputs() {
	m.nameInput.Blur()
	m.methodInput.Blur()
	m.urlInput.Blur()
	m.headersInput.Blur()
	m.bodyInput.Blur()
}

func (m *appModel) saveCurrent() error {
	name := strings.TrimSpace(m.nameInput.Value())
	if name == "" {
		return fmt.Errorf("request name cannot be empty")
	}
	url := strings.TrimSpace(m.urlInput.Value())
	if url == "" {
		return fmt.Errorf("request URL cannot be empty")
	}
	method := strings.ToUpper(strings.TrimSpace(m.methodInput.Value()))
	if method == "" {
		method = "GET"
	}

	headers, err := parseHeaders(m.headersInput.Value())
	if err != nil {
		return err
	}

	if m.current.CreatedAt.IsZero() {
		m.current.CreatedAt = time.Now()
	}
	m.current.Name = name
	m.current.Method = method
	m.current.URL = url
	m.current.Headers = headers
	m.current.Body = m.bodyInput.Value()
	m.current.UpdatedAt = time.Now()

	if m.originalName != "" && m.originalName != name {
		if _, err := store.LoadRequest(m.group, name); err == nil {
			return fmt.Errorf("a request named %q already exists", name)
		}
	}
	if err := store.SaveRequest(m.group, m.current); err != nil {
		return err
	}
	if m.originalName != "" && m.originalName != name {
		if err := store.DeleteRequest(m.group, m.originalName); err != nil {
			return fmt.Errorf("saved as %q, but could not remove old request %q: %w", name, m.originalName, err)
		}
	}

	m.originalName = name
	m.isNew = false
	return m.reloadRequests(name)
}

func (m *appModel) reloadRequests(preferred string) error {
	requests, err := store.ListRequests(m.group)
	if err != nil {
		return err
	}
	m.requests = requests
	m.selected = -1
	for i, req := range requests {
		if req.Name == preferred {
			m.selected = i
			break
		}
	}
	if m.selected == -1 && len(requests) > 0 {
		m.selected = 0
	}
	return nil
}

func (m *appModel) runCurrent() tea.Cmd {
	if err := m.saveCurrent(); err != nil {
		m.status = "Save failed: " + err.Error()
		return nil
	}

	req := m.current
	m.running = true
	m.screen = screenBrowse
	m.pane = paneResponse
	m.status = "Running " + req.Method + " " + req.URL
	return func() tea.Msg {
		resp, err := httppkg.Execute(req)
		if err != nil {
			return responseMsg{err: err}
		}

		flatHeaders := flattenHeaders(resp.Headers)
		saved := store.SavedResponse{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Headers:    flatHeaders,
			Body:       resp.Body,
			DurationMs: resp.Duration.Milliseconds(),
			Timestamp:  time.Now(),
		}
		return responseMsg{
			response: resp,
			saveErr:  store.SaveResponse(m.group, req.Name, saved),
		}
	}
}

func (m *appModel) setResponse(resp httppkg.Response) {
	m.responseCode = resp.StatusCode
	m.responseStatus = resp.Status
	m.responseHeaders = flattenHeaders(resp.Headers)
	m.responseBody = resp.Body
	m.responseDuration = resp.Duration
	m.renderResponseContent()
}

func (m *appModel) setSavedResponse(resp store.SavedResponse) {
	m.responseCode = resp.StatusCode
	m.responseStatus = resp.Status
	m.responseHeaders = resp.Headers
	m.responseBody = resp.Body
	m.responseDuration = time.Duration(resp.DurationMs) * time.Millisecond
	m.renderResponseContent()
}

func (m *appModel) clearResponse() {
	m.responseCode = 0
	m.responseStatus = ""
	m.responseHeaders = map[string]string{}
	m.responseBody = ""
	m.responseDuration = 0
	m.responseViewport.SetContent(mutedStyle.Render("No response yet.\n\nPress r to run the saved request."))
}

func (m *appModel) renderResponseContent() {
	if m.responseCode == 0 && m.responseStatus == "" {
		m.clearResponse()
		return
	}
	m.responseViewport.SetContent(formatResponse(
		m.responseCode,
		m.responseStatus,
		m.responseHeaders,
		m.responseBody,
		m.responseDuration,
		m.showHeaders,
	))
}

func (m *appModel) resizeInputs() {
	width := m.width - 10
	if width < 24 {
		width = 24
	}
	m.nameInput.Width = width
	m.methodInput.Width = minInt(width, 16)
	m.urlInput.Width = width
	m.headersInput.SetWidth(width)
	m.bodyInput.SetWidth(width)
	m.importNameInput.Width = width
	m.importInput.SetWidth(width)
}

func (m *appModel) resizeViewport() {
	width := m.width/3 - 6
	if m.width < 92 {
		width = m.width - 10
	}
	if width < 20 {
		width = 20
	}
	height := m.height - 9
	if height < 8 {
		height = 8
	}
	m.responseViewport.Width = width - 6
	m.responseViewport.Height = height
}

func (m *appModel) View() string {
	width := m.width
	if width <= 0 {
		width = 120
	}
	height := m.height
	if height <= 0 {
		height = 30
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		brandStyle.Render("POLI / "+m.group),
		"  ",
		mutedStyle.Render("interactive request workbench"),
	)

	var body string
	if m.screen == screenImport {
		body = m.renderImport(width-4, height-8)
	} else {
		body = m.renderWorkbench(width-4, height-8)
	}

	footer := m.renderFooter()
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m *appModel) renderWorkbench(width, height int) string {
	if width < 92 {
		panelWidth := maxInt(30, width)
		panelHeight := maxInt(8, height/3)
		return lipgloss.JoinVertical(
			lipgloss.Left,
			m.renderList(panelWidth, panelHeight),
			m.renderEditor(panelWidth, panelHeight+4),
			m.renderResponsePanel(panelWidth, panelHeight+4),
		)
	}

	listWidth := maxInt(24, width/5)
	responseWidth := maxInt(32, width/3)
	editorWidth := width - listWidth - responseWidth
	if editorWidth < 32 {
		editorWidth = 32
		listWidth = maxInt(20, width-editorWidth-responseWidth)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderList(listWidth, height),
		m.renderEditor(editorWidth, height),
		m.renderResponsePanel(responseWidth, height),
	)
}

func (m *appModel) renderList(width, height int) string {
	lines := []string{titleStyle.Render("REQUESTS"), mutedStyle.Render(m.group)}
	if len(m.requests) == 0 {
		lines = append(lines, "", mutedStyle.Render("No saved requests."), "", "Press n to create one.")
	} else {
		lines = append(lines, "")
		for i, req := range m.requests {
			line := fmt.Sprintf("r%-2d %-6s %s", i+1, strings.ToUpper(req.Method), req.Name)
			if i == m.selected {
				lines = append(lines, activeStyle.Render("> "+line))
			} else {
				lines = append(lines, "  "+line)
			}
		}
	}
	return m.panel("COLLECTION", strings.Join(lines, "\n"), width, height, m.pane == paneList)
}

func (m *appModel) renderEditor(width, height int) string {
	if m.screen == screenEdit {
		m.setEditorInputWidth(width)
		content := strings.Join([]string{
			titleStyle.Render("EDIT REQUEST"),
			fieldView("NAME", m.nameInput.View()),
			fieldView("METHOD", m.methodInput.View()),
			fieldView("URL", m.urlInput.View()),
			fieldView("HEADERS", m.headersInput.View()),
			fieldView("BODY", m.bodyInput.View()),
		}, "\n")
		return m.panel("REQUEST", content, width, height, true)
	}

	if m.current.Name == "" {
		return m.panel("REQUEST", mutedStyle.Render("Press n to create a request or i to import cURL."), width, height, m.pane == paneEditor)
	}

	headers := "none"
	if len(m.current.Headers) > 0 {
		headers = fmt.Sprintf("%d", len(m.current.Headers))
	}
	body := "empty"
	if strings.TrimSpace(m.current.Body) != "" {
		body = fmt.Sprintf("%d bytes", len(m.current.Body))
	}
	content := strings.Join([]string{
		titleStyle.Render("REQUEST"),
		"",
		labelStyle.Render(strings.ToUpper(m.current.Method)) + "  " + m.current.Name,
		mutedStyle.Render(m.current.URL),
		"",
		fieldView("HEADERS", headers),
		fieldView("BODY", body),
		"",
		mutedStyle.Render("Enter/e edit this request."),
	}, "\n")
	return m.panel("REQUEST", content, width, height, m.pane == paneEditor)
}

func (m *appModel) renderResponsePanel(width, height int) string {
	m.responseViewport.Width = maxInt(12, width-6)
	m.responseViewport.Height = maxInt(3, height-4)
	content := m.responseViewport.View()
	if m.running {
		content = codeStyle.Render("RUNNING...\n\n" + m.current.Method + " " + m.current.URL)
	}
	title := "RESPONSE"
	if m.showHeaders {
		title += " + HEADERS"
	}
	return m.panel(title, content, width, height, m.pane == paneResponse)
}

func (m *appModel) renderImport(width, height int) string {
	inputWidth := maxInt(16, width-6)
	m.importNameInput.Width = inputWidth
	m.importInput.SetWidth(inputWidth)
	content := strings.Join([]string{
		titleStyle.Render("IMPORT cURL"),
		mutedStyle.Render("Paste a full cURL command, including line continuations."),
		"",
		fieldView("REQUEST NAME", m.importNameInput.View()),
		fieldView("cURL", m.importInput.View()),
		"",
		mutedStyle.Render("tab switches fields · ctrl+s imports · esc cancels"),
	}, "\n")
	return m.panel("IMPORT", content, width, height, true)
}

func (m *appModel) setEditorInputWidth(width int) {
	inputWidth := maxInt(16, width-6)
	m.nameInput.Width = inputWidth
	m.methodInput.Width = minInt(inputWidth, 16)
	m.urlInput.Width = inputWidth
	m.headersInput.SetWidth(inputWidth)
	m.bodyInput.SetWidth(inputWidth)
}

func (m *appModel) renderFooter() string {
	keys := []string{
		keyStyle.Render("n") + " new",
		keyStyle.Render("i") + " import",
		keyStyle.Render("r") + " run",
		keyStyle.Render("e/enter") + " edit",
		keyStyle.Render("h") + " headers",
		keyStyle.Render("q") + " quit",
	}
	if m.screen != screenBrowse {
		keys = []string{
			keyStyle.Render("tab") + " next field",
			keyStyle.Render("ctrl+s") + " save",
			keyStyle.Render("ctrl+r") + " run",
			keyStyle.Render("esc") + " back",
		}
	}
	footer := strings.Join(keys, "  ")
	if m.status != "" {
		if strings.Contains(m.status, "failed") || strings.Contains(m.status, "could not") {
			footer += "\n" + errStyle.Render(m.status)
		} else {
			footer += "\n" + statusStyle.Render(m.status)
		}
	}
	return footer
}

func (m *appModel) panel(title, content string, width, height int, active bool) string {
	borderColor := muted
	if active {
		borderColor = purple
	}
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(maxInt(10, width-4)).
		Height(maxInt(5, height-2))
	return style.Render(titleStyle.Render(title) + "\n" + content)
}

func fieldView(label, value string) string {
	return labelStyle.Render(label) + "\n" + value
}

func formatHeaders(headers []modelpkg.Header) string {
	if len(headers) == 0 {
		return ""
	}
	lines := make([]string, 0, len(headers))
	for _, header := range headers {
		lines = append(lines, header.Key+": "+header.Value)
	}
	return strings.Join(lines, "\n")
}

func parseHeaders(raw string) ([]modelpkg.Header, error) {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	headers := make([]modelpkg.Header, 0, len(lines))
	for lineNumber, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			return nil, fmt.Errorf("invalid header on line %d; use Key: Value", lineNumber+1)
		}
		headers = append(headers, modelpkg.Header{
			Key:   strings.TrimSpace(parts[0]),
			Value: strings.TrimSpace(parts[1]),
		})
	}
	return headers, nil
}

func flattenHeaders(headers map[string][]string) map[string]string {
	flat := make(map[string]string, len(headers))
	for key, values := range headers {
		if len(values) > 0 {
			flat[key] = strings.Join(values, ", ")
		}
	}
	return flat
}

func formatResponse(statusCode int, status string, headers map[string]string, body string, duration time.Duration, showHeaders bool) string {
	var out strings.Builder
	out.WriteString(formatStatusLine(statusCode, status))
	if duration > 0 {
		out.WriteString("  " + formatDuration(duration))
	}
	out.WriteString("\n")

	if showHeaders && len(headers) > 0 {
		keys := make([]string, 0, len(headers))
		for key := range headers {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		out.WriteString("\nHEADERS\n")
		for _, key := range keys {
			out.WriteString(key + ": " + headers[key] + "\n")
		}
	}

	out.WriteString("\nBODY\n")
	if strings.TrimSpace(body) == "" {
		out.WriteString("<empty>")
	} else {
		out.WriteString(prettyBody(body))
	}
	return strings.TrimRight(out.String(), "\n")
}

func formatStatusLine(statusCode int, status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		if statusCode > 0 {
			return fmt.Sprintf("%d", statusCode)
		}
		return "NO RESPONSE"
	}
	if statusCode > 0 && strings.HasPrefix(status, fmt.Sprintf("%d ", statusCode)) {
		return status
	}
	if statusCode > 0 {
		return fmt.Sprintf("%d %s", statusCode, status)
	}
	return status
}

func prettyBody(body string) string {
	trimmed := strings.TrimSpace(body)
	var pretty bytes.Buffer
	if json.Indent(&pretty, []byte(trimmed), "", "  ") == nil {
		return pretty.String()
	}
	return body
}

func errorContent(err error) string {
	return errStyle.Render("REQUEST ERROR\n\n" + err.Error())
}

func formatDuration(duration time.Duration) string {
	if duration < time.Millisecond {
		return duration.String()
	}
	return fmt.Sprintf("%dms", duration.Milliseconds())
}

func isViewportKey(key string) bool {
	switch key {
	case "up", "down", "k", "j", "pgup", "pgdown", "ctrl+u", "ctrl+d", "home", "end":
		return true
	default:
		return false
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
