// Package s3 provides the TUI coordinator component for the S3 workflow.
// It delegates to sub-pages: menu (Upload/Browse selection), upload (placeholder),
// and browse (2-panel buckets + objects browser).
//
// Routes:
//   - menu   → upload (Enter on "Upload")
//   - menu   → browse (Enter on "Browse")
//   - upload → menu   (Esc → BackToS3MenuMsg)
//   - browse → menu   (Esc → BackToS3MenuMsg)
//
// The s3 coordinator owns the menu, upload, and browse pages as peer children.
// It handles navigation between them via custom messages. It is also
// responsible for loading the .env file (via godotenv) and creating the
// S3 client, which is passed down to child pages.
package s3

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/joho/godotenv"
	s3helper "github.com/IrwantoCia/utility/internal/helper/s3"
	"github.com/IrwantoCia/utility/internal/tui/common"
	"github.com/IrwantoCia/utility/internal/tui/s3/browse"
	"github.com/IrwantoCia/utility/internal/tui/s3/menu"
	"github.com/IrwantoCia/utility/internal/tui/s3/upload"
)

// S3 coordinates the S3 workflow, routing between the sub-menu,
// upload page, and browse page.
type S3 struct {
	lastWindow  tea.WindowSizeMsg
	menuModel   *menu.Menu
	uploadModel *upload.Upload
	browseModel *browse.Browse
	activePage  string

	client  *s3helper.S3 // cached S3 client
	envFile string       // env file used for last client init
}

var _ common.Component = (*S3)(nil)

// Close implements common.Component.
func (c *S3) Close() tea.Cmd { return nil }

// New creates a new S3 coordinator starting at the sub-menu.
func New() *S3 {
	return &S3{
		activePage: "",
		menuModel:  menu.New(),
	}
}

// Init initialises the active sub-page.
func (c *S3) Init() tea.Cmd {
	return c.menuModel.Init()
}

// Resize propagates window size changes to all initialised sub-models.
func (c *S3) Resize(ws tea.WindowSizeMsg) tea.Cmd {
	c.lastWindow = ws
	var cmds []tea.Cmd
	cmds = append(cmds, c.menuModel.Resize(ws))
	if c.uploadModel != nil {
		cmds = append(cmds, c.uploadModel.Resize(ws))
	}
	if c.browseModel != nil {
		cmds = append(cmds, c.browseModel.Resize(ws))
	}
	return tea.Batch(cmds...)
}

// View renders the active sub-page.
func (c *S3) View() string {
	switch c.activePage {
	case "upload":
		return c.uploadModel.View()
	case "browse":
		return c.browseModel.View()
	default:
		return c.menuModel.View()
	}
}

// Update dispatches messages to the active sub-page and intercepts
// custom messages to switch between pages.
func (c *S3) Update(msg tea.Msg) tea.Cmd {
	// Intercept page-switching messages regardless of current page.
	switch msg := msg.(type) {
	case menu.ShowUploadMsg:
		client, err := c.getClient(msg.EnvFile)
		c.activePage = "upload"
		if err != nil {
			c.uploadModel = upload.New(nil, err)
		} else {
			c.uploadModel = upload.New(client, nil)
		}
		c.uploadModel.Resize(c.lastWindow)
		return c.uploadModel.Init()
	case menu.ShowBrowseMsg:
		client, err := c.getClient(msg.EnvFile)
		c.activePage = "browse"
		if err != nil {
			c.browseModel = browse.New(nil, err)
		} else {
			c.browseModel = browse.New(client, nil)
		}
		c.browseModel.Resize(c.lastWindow)
		return c.browseModel.Init()
	case upload.BackToS3MenuMsg:
		c.activePage = ""
		return nil
	case browse.BackToS3MenuMsg:
		c.activePage = ""
		return nil
	}

	// Route remaining messages to the active sub-page.
	switch c.activePage {
	case "upload":
		return c.uploadModel.Update(msg)
	case "browse":
		return c.browseModel.Update(msg)
	default:
		return c.menuModel.Update(msg)
	}
}

// getClient returns a cached S3 client, creating one if the envFile has
// changed. When envFile is empty, returns nil with no error (caller can
// check for nil client).
func (c *S3) getClient(envFile string) (*s3helper.S3, error) {
	if envFile == "" {
		return nil, nil
	}
	if c.client != nil && c.envFile == envFile {
		return c.client, nil
	}

	// Load .env into process environment
	if err := godotenv.Load(envFile); err != nil {
		c.client = nil
		c.envFile = ""
		return nil, fmt.Errorf("loading %s: %w", envFile, err)
	}

	// Create client with env-var-based config
	client, err := s3helper.New(s3helper.Config{})
	if err != nil {
		c.client = nil
		c.envFile = ""
		return nil, fmt.Errorf("creating client: %w", err)
	}

	c.client = client
	c.envFile = envFile
	return c.client, nil
}
