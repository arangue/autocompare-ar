package mercadolibre

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Specs are doors and gearbox taken from a listing page.
type Specs struct {
	Doors        string
	Transmission string
	Body         string
}

// Tokens returns extra words for MatchTitle, for example "5P Manual Hatchback".
func (s Specs) Tokens() string {
	var parts []string
	if s.Doors != "" {
		parts = append(parts, s.Doors)
	}
	if s.Transmission != "" {
		parts = append(parts, s.Transmission)
	}
	if s.Doors == "" && s.Body != "" {
		parts = append(parts, s.Body)
	}
	return strings.Join(parts, " ")
}

const chromeUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"

var (
	specJSON  = regexp.MustCompile(`"id":"(Puertas|Transmisi(?:ón|\\u00f3n)|Tipo de carrocer(?:ía|\\u00eda))","text":"([^"]+)"`)
	htmlDoors = regexp.MustCompile(`(?is)Puertas:</span>.*?<span>(\d{1,2})</span>`)
	htmlTrans = regexp.MustCompile(`(?is)Transmisi[oó]n:</span>.*?<span>([^<]{3,20})</span>`)
	htmlBody  = regexp.MustCompile(`(?is)Tipo de carrocer[ií]a:</span>.*?<span>([^<]{3,20})</span>`)
)

// ParseSpecs reads Puertas, Transmisión, and Tipo de carrocería from listing HTML.
func ParseSpecs(html []byte) Specs {
	var specs Specs
	for _, m := range specJSON.FindAllSubmatch(html, -1) {
		id := decodeJSONString(string(m[1]))
		text := decodeJSONString(string(m[2]))
		switch {
		case id == "Puertas" && specs.Doors == "":
			specs.Doors = doorToken(text)
		case strings.HasPrefix(id, "Transmisi") && specs.Transmission == "":
			specs.Transmission = transmissionToken(text)
		case strings.HasPrefix(id, "Tipo de carrocer") && specs.Body == "":
			specs.Body = bodyToken(text)
		}
	}
	if specs.Doors == "" {
		if m := htmlDoors.FindSubmatch(html); m != nil {
			specs.Doors = doorToken(string(m[1]))
		}
	}
	if specs.Transmission == "" {
		if m := htmlTrans.FindSubmatch(html); m != nil {
			specs.Transmission = transmissionToken(string(m[1]))
		}
	}
	if specs.Body == "" {
		if m := htmlBody.FindSubmatch(html); m != nil {
			specs.Body = bodyToken(string(m[1]))
		}
	}
	return specs
}

// FetchSpecs loads one listing page and parses specs. An empty URL is a no-op.
func (c *Client) FetchSpecs(ctx context.Context, rawURL string) (Specs, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return Specs{}, nil
	}
	if i := strings.IndexByte(rawURL, '#'); i >= 0 {
		rawURL = rawURL[:i]
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return Specs{}, err
	}
	req.Header.Set("User-Agent", chromeUA)
	req.Header.Set("Accept-Language", "es-AR,es;q=0.9")
	req.Header.Set("Accept", "text/html")

	c.mu.Lock()
	c.wait(time.Now())
	c.mu.Unlock()

	resp, err := c.http.Do(req)
	if err != nil {
		return Specs{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return Specs{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Specs{}, fmt.Errorf("mercadolibre listing: status %d", resp.StatusCode)
	}
	return ParseSpecs(body), nil
}

func doorToken(text string) string {
	digits := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, text)
	if digits == "" {
		return ""
	}
	return digits + "P"
}

func transmissionToken(text string) string {
	t := strings.ToLower(text)
	switch {
	case strings.Contains(t, "manual"):
		return "Manual"
	case strings.Contains(t, "autom"):
		return "Automatica"
	default:
		return ""
	}
}

func bodyToken(text string) string {
	t := strings.ToLower(text)
	switch {
	case strings.Contains(t, "hatch"):
		return "Hatchback"
	case strings.Contains(t, "sed"):
		return "Sedan"
	default:
		return ""
	}
}

func decodeJSONString(s string) string {
	s = strings.ReplaceAll(s, `\u00f3`, "ó")
	s = strings.ReplaceAll(s, `\u00ed`, "í")
	s = strings.ReplaceAll(s, `\u00e1`, "á")
	s = strings.ReplaceAll(s, `\u00e9`, "é")
	s = strings.ReplaceAll(s, `\u00fa`, "ú")
	s = strings.ReplaceAll(s, `\u00f1`, "ñ")
	return s
}
