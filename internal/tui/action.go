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
	Name     string
	Label    string
	Default  string
	Validate func(string) error
}
