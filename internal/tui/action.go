package tui

import "context"

type Action struct {
	Title  string
	Fields []Field
	Run    func(context.Context, map[string]string) (string, error)
}

type Field struct {
	Name    string
	Label   string
	Default string
}
