package tui

import "context"

type Action struct {
	Title       string
	Description string
	Fields      []Field
	Run         func(context.Context, map[string]string) (Result, error)
	RunProgress func(context.Context, map[string]string, func(ProgressEvent)) (Result, error)
}

type Field struct {
	Name        string
	Label       string
	Default     string
	Placeholder string
	Toggle      bool
	Visible     func(map[string]string) bool
	Validate    func(string) error
}
