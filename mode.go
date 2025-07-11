package main

type Mode int

const (
	HomeMode Mode = iota
	NewMode
	CheckMode
	EditMode
)

var ChoiceToModel = map[string]Mode{
	"Create new item": NewMode,
	"Check old items": CheckMode,
}

