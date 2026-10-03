package mercadolibre

import (
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	defaultSite     = "MLA"
	defaultBaseURL  = "https://api.mercadolibre.com"
	defaultInterval = time.Second
	carsCategory    = "MLA1744"
)

// Client calls the Mercado Libre search API. It does not touch the database
// or the worker.
type Client struct {
	http        *http.Client
	token       string
	site        string
	baseURL     string
	minInterval time.Duration
	log         *slog.Logger

	mu   sync.Mutex
	last time.Time
}

// New builds a client. An empty site becomes MLA.
func New(token, site string) *Client {
	if site == "" {
		site = defaultSite
	}
	return &Client{
		http:        &http.Client{Timeout: 15 * time.Second},
		token:       token,
		site:        site,
		baseURL:     defaultBaseURL,
		minInterval: defaultInterval,
		log:         slog.Default(),
	}
}

// NewFromEnv reads ML_ACCESS_TOKEN and ML_SITE_ID.
func NewFromEnv() *Client {
	return New(os.Getenv("ML_ACCESS_TOKEN"), os.Getenv("ML_SITE_ID"))
}

func (c *Client) wait(now time.Time) {
	if c.minInterval <= 0 {
		c.last = now
		return
	}
	if !c.last.IsZero() {
		if d := c.minInterval - now.Sub(c.last); d > 0 {
			time.Sleep(d)
			now = time.Now()
		}
	}
	c.last = now
}
