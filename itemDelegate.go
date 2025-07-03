package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const ListHeight = 14
const DefaultWidth = 28

var (
	HelpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	PaginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	TitleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	markedItemStyle   = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#00ff33"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#dd00ff"))
)

type Item struct {
	description string
	status      string
	title       string
}

func (i Item) Description() string { return i.description }
func (i Item) FilterValue() string { return i.title }
func (i Item) Title() string       { return i.title }

type ItemDelegate struct{}

func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title())

	fn := itemStyle.Render
	if i.status == MarkedStatus {
		fn = func(s ...string) string {
			return markedItemStyle.Render(strings.Join(s, " "))
		}
	}

	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
