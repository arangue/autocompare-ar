package ingestion

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arangue/autocompare-ar/internal/domain"
)

// FileParser is the strategy interface for parsing vehicle listings and price tables from files.
type FileParser interface {
	CanParse(path string) bool
	Parse(ctx context.Context, path string) (listings []domain.Listing, skipped int, err error)
}

// ParserRegistry provides access to registered file parsers based on file type or extension.
type ParserRegistry struct {
	parsers []FileParser
}

// NewDefaultRegistry instantiates the default set of ingestion parsers.
func NewDefaultRegistry() *ParserRegistry {
	r := &ParserRegistry{}
	r.Register(&JSONParser{})
	r.Register(&CSVParser{})
	return r
}

func (r *ParserRegistry) Register(p FileParser) {
	r.parsers = append(r.parsers, p)
}

func (r *ParserRegistry) ParserFor(path string) (FileParser, error) {
	for _, p := range r.parsers {
		if p.CanParse(path) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("unsupported ingest file type %q", filepath.Ext(path))
}

// JSONParser parses JSON listing files.
type JSONParser struct{}

func (p *JSONParser) CanParse(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".json")
}

func (p *JSONParser) Parse(ctx context.Context, path string) ([]domain.Listing, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	return parseJSON(data)
}

// CSVParser parses CSV listing files.
type CSVParser struct{}

func (p *CSVParser) CanParse(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".csv")
}

func (p *CSVParser) Parse(ctx context.Context, path string) ([]domain.Listing, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	return parseCSV(data)
}
