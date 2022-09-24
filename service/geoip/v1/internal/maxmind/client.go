package maxmind

import (
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Client represents an active maxmind object
type Client struct {
	http        *http.Client
	logrus      *logrus.Logger
	workDir     string
	downloadDir string
	licenseKey  string
	baseURL     string
}

// Config defines the config for maxmind
type Config struct {
	Logrus            *logrus.Logger
	DownloadDirectory string
	LicenseKey        string
	BaseURL           string
}

// New returns a maxmind client
func New(config Config) (*Client, error) {
	if config.LicenseKey == "" {
		return nil, errors.New("License key required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://download.maxmind.com"
	}
	_, err := url.ParseRequestURI(config.BaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid base URL")
	}

	return &Client{
		http:        http.DefaultClient,
		logrus:      config.Logrus,
		workDir:     os.TempDir(),
		downloadDir: config.DownloadDirectory,
		licenseKey:  config.LicenseKey,
		baseURL:     config.BaseURL,
	}, nil
}

// NewDownloader returns a new downloader instance
func (c *Client) NewDownloader(eid EditionID) (*Downloader, error) {
	return &Downloader{
		Client: c,
		eid:    eid,
		dlDir:  c.downloadDir,
	}, nil
}
