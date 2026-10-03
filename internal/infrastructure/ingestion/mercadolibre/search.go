package mercadolibre

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/arangue/autocompare-ar/internal/domain"
)

// Search queries used cars in category MLA1744.
// skipped counts items MapItem rejected. An empty token is an error:
// search without a bearer returns 403.
func (c *Client) Search(ctx context.Context, query string, limit, offset int) ([]domain.Listing, int, error) {
	if strings.TrimSpace(c.token) == "" {
		return nil, 0, fmt.Errorf("mercadolibre: ML_ACCESS_TOKEN is empty")
	}
	if limit <= 0 || limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	u, err := url.Parse(strings.TrimRight(c.baseURL, "/") + "/sites/" + url.PathEscape(c.site) + "/search")
	if err != nil {
		return nil, 0, err
	}
	q := u.Query()
	q.Set("category", carsCategory)
	q.Set("q", query)
	q.Set("condition", "used")
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	c.mu.Lock()
	c.wait(time.Now())
	c.mu.Unlock()

	c.log.Info("mercadolibre search", "site", c.site, "category", carsCategory, "limit", limit, "offset", offset)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, 0, fmt.Errorf("mercadolibre search: status %d: %s", resp.StatusCode, truncate(body))
	}

	var payload struct {
		Results []rawItem `json:"results"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, 0, fmt.Errorf("mercadolibre search: %w", err)
	}

	listings := make([]domain.Listing, 0, len(payload.Results))
	skipped := 0
	for _, item := range payload.Results {
		listing, ok := MapItem(item)
		if !ok {
			skipped++
			continue
		}
		listings = append(listings, listing)
	}
	return listings, skipped, nil
}

func truncate(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 200 {
		return s[:200]
	}
	return s
}
