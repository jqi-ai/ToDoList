package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	errMsg error
)

type model struct {
	choices    []string        // features of the app
	cursor     int             // which feature cursor is pointing at
	list       list.Model      // list of existing toDo items
	listCursor int             // which toDo item is pointed in CheckMode
	mode       Mode            // default mode is homeMode
	textInput  textinput.Model // input box when creating new toDo item
	err        error
}

var MarkedStatus = "marked"

func initialModel() model {
	// init textInput
	ti := textinput.New()
	ti.CharLimit = 32
	ti.Focus()
	ti.Placeholder = "Xu Shu"
	ti.Width = 32

	// init list
	items := []list.Item{}
	l := list.New(items, ItemDelegate{}, DefaultWidth, ListHeight)
	l.Title = "Existing ToDo items:"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = TitleStyle
	l.Styles.PaginationStyle = PaginationStyle
	l.Styles.HelpStyle = HelpStyle
	// write file data to list
	file, err := os.Open("./data.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	index := 0
	for scanner.Scan() {
		l.InsertItem(index, Item{
			title:       scanner.Text(),
			description: "Placeholder",
		})
		index += 1
	}

	return model{
		choices:    []string{"Create new item", "Check old items"},
		cursor:     0,
		list:       l,
		listCursor: 0,
		mode:       HomeMode,
		textInput:  ti,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case HomeMode:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				// Save new list to file
				file, err := os.Create("data.txt")
				if err != nil {
					panic(err)
				}
				defer file.Close()

				for _, item := range m.list.Items() {
					listItem := item.(Item)
					_, err := fmt.Fprintln(file, listItem.Title())
					if err != nil {
						panic(err)
					}
				}
				return m, tea.Quit

			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}

			case "enter", " ":
				choice := m.choices[m.cursor]
				m.mode = ChoiceToModel[choice]
			}
		}
	case CheckMode:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.list.SetWidth(msg.Width)
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "ctrl+c":
				m.mode = HomeMode
			case "enter":
				currentItem := m.list.Items()[m.listCursor].(Item)
				currentStatus := currentItem.status
				if currentStatus == MarkedStatus {
					currentStatus = ""
				} else {
					currentStatus = MarkedStatus
				}

				m.list.SetItem(m.listCursor, Item{
					title:       currentItem.Title(),
					description: currentItem.Description(),
					status:      currentStatus,
				})
			case "up", "k":
				if m.listCursor > 0 {
					m.listCursor--
				}
			case "down", "j":
				if m.listCursor < len(m.list.Items())-1 {
					m.listCursor++
				}
			case "s":
				savedItems := []list.Item{}
				for _, listItem := range m.list.Items() {
					item := listItem.(Item)
					if item.status != MarkedStatus {
						savedItems = append(savedItems, item)
					}
				}
				m.list.SetItems(savedItems)
				m.mode = HomeMode
			case "e":
				m.mode = EditMode
				currentItem := m.list.Items()[m.listCursor].(Item)
				m.textInput.SetValue(currentItem.title)
			}
		}
		m.list, _ = m.list.Update(msg)
	case NewMode:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.mode = HomeMode
			case tea.KeyEnter:
				listLength := len(m.list.Items())
				m.list.InsertItem(listLength, Item{
					description: "Placeholder",
					status:      "",
					title:       m.textInput.Value(),
				})

				m.textInput.Reset()
				m.mode = HomeMode
			}
		case errMsg:
			m.err = msg
		}

		m.textInput, _ = m.textInput.Update(msg)
	case EditMode:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.mode = CheckMode
			case tea.KeyEnter:
				currentItem := m.list.Items()[m.listCursor].(Item)
				m.list.SetItem(m.listCursor, Item{
					title:       m.textInput.Value(),
					description: currentItem.description,
					status:      currentItem.status,
				})
				m.mode = CheckMode
			}
		case errMsg:
			m.err = msg
		}
		m.textInput, _ = m.textInput.Update(msg)
	}

	return m, nil
}

func (m model) View() string {
	switch m.mode {
	case HomeMode:
		s := "What to do next?\n\n"
		for i, choice := range m.choices {
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = ">" // cursor!
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}

		year, month, day := time.Now().Date()
		s += fmt.Sprintf("\nPress q to quit. Date: %s-%d-%d\n", month, day, year)
		return s
	case NewMode:
		return fmt.Sprintf("What's the plan for today?\n\n%s\n\n%s", m.textInput.View(), "(esc to quit)") + "\n"
	case CheckMode:
		return m.list.View() + "\nHelp: e to edit, s to remove marked items, q to quit"
	case EditMode:
		return fmt.Sprintf("Edit your plan:\n\n%s\n\n%s", m.textInput.View(), "(esc to quit)") + "\n"
	}
	return ""
}

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
