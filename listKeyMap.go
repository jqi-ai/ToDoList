package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type listKeyMap struct {
	editItem     key.Binding
	saveArchives key.Binding
	toggleItem   key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		editItem: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "Edit the item"),
		),
		saveArchives: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "Save the archives"),
		),
		toggleItem: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Select item to archive"),
		),
	}
}
