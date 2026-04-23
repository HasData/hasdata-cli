package spec

import "encoding/json"

type CatalogEntry struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
	APIs  []API  `json:"apis"`
}

type API struct {
	ID    int    `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Price int    `json:"price"`
	Mode  string `json:"mode,omitempty"`
}

type Detail struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	Price       int        `json:"price"`
	Mode        string     `json:"mode,omitempty"`
	ModeMessage string     `json:"mode_message,omitempty"`
	Category    Category   `json:"category"`
	Schema      Schema     `json:"schema"`
}

type Category struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
	Order int    `json:"order"`
}

type Schema struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Method      string              `json:"method"`
	URL         string              `json:"url"`
	Cost        int                 `json:"cost"`
	Required    []string            `json:"required"`
	Properties  map[string]Property `json:"properties"`
}

type Property struct {
	Type                 string          `json:"type"`
	Title                string          `json:"title"`
	Description          string          `json:"description"`
	Default              json.RawMessage `json:"default,omitempty"`
	Enum                 []json.RawMessage `json:"enum,omitempty"`
	Items                *Items          `json:"items,omitempty"`
	AdditionalProperties json.RawMessage `json:"additionalProperties,omitempty"`
	UniqueItems          bool            `json:"uniqueItems,omitempty"`
}

type Items struct {
	Type string            `json:"type"`
	Enum []json.RawMessage `json:"enum,omitempty"`
}
