package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg error
)

type model struct {
	choices   []string // Features of the app
	cursor    int      // Index the cursor pointed at
	err       error
	mode      Mode            // Default mode is homeMode
	textInput textinput.Model // Input while creating new toDo item
	toDoList  list.Model      // List of toDo items
}

var MarkedStatus = "marked"

func initialModel() model {

	// Init textInput
	textInput := textinput.New()
	textInput.CharLimit = 100
	textInput.Focus()
	textInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#33eeff"))

	// Init TODO list
	toDoList := list.New([]list.Item{}, ItemDelegate{}, DefaultWidth, ListHeight)
	toDoList.Title = "Existing ToDo items:"
	toDoList.SetShowStatusBar(true)
	toDoList.SetFilteringEnabled(true)
	toDoList.Styles.Title = TitleStyle
	toDoList.Styles.PaginationStyle = PaginationStyle
	toDoList.Styles.HelpStyle = HelpStyle
	// Write file to list
	file, err := os.Open("./data.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	index := 0
	for scanner.Scan() {
		toDoList.InsertItem(index, Item{
			title:       scanner.Text(),
			description: "",
		})
		index += 1
	}

	return model{
		choices:   []string{"Create new item", "Check old items"},
		cursor:    0,
		textInput: textInput,
		toDoList:  toDoList,
		mode:      HomeMode,
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
			case "q":
				// Save new list to file
				file, err := os.Create("data.txt")
				if err != nil {
					panic(err)
				}
				defer file.Close()

				for _, item := range m.toDoList.Items() {
					listItem := item.(Item)
					_, err := fmt.Fprintln(file, listItem.Title())
					if err != nil {
						panic(err)
					}
				}
				return m, tea.Quit

			case "k":
				if m.cursor > 0 {
					m.cursor--
				}

			case "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}

			case "enter":
				choice := m.choices[m.cursor]
				m.cursor = 0
				m.mode = ChoiceToModel[choice]
			}
		}
	case CheckMode:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q":
				m.cursor = 0
				m.mode = HomeMode
			case "enter":
				currentItem := m.toDoList.Items()[m.cursor].(Item)
				currentStatus := currentItem.status
				if currentStatus == MarkedStatus {
					currentStatus = ""
				} else {
					currentStatus = MarkedStatus
				}

				m.toDoList.SetItem(m.cursor, Item{
					title:       currentItem.Title(),
					description: currentItem.Description(),
					status:      currentStatus,
				})

			case "k":
				if m.cursor > 0 {
					m.cursor--
				}

			case "j":
				if m.cursor < len(m.toDoList.Items())-1 {
					m.cursor++
				}

			case "s":
				savedItems := []list.Item{}
				for _, listItem := range m.toDoList.Items() {
					item := listItem.(Item)
					if item.status != MarkedStatus {
						savedItems = append(savedItems, item)
					}
				}
				m.toDoList.SetItems(savedItems)
				m.mode = HomeMode
				m.cursor = 0

			case "e":
				m.mode = EditMode
				currentItem := m.toDoList.Items()[m.cursor].(Item)
				m.textInput.SetValue(currentItem.title)
			}
		}
		// TODO: What does it mean?
		m.toDoList, _ = m.toDoList.Update(msg)
	case NewMode:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.mode = HomeMode
			case tea.KeyEnter:
				listLength := len(m.toDoList.Items())
				m.toDoList.InsertItem(listLength, Item{
					description: "",
					status:      "",
					title:       m.textInput.Value(),
				})

				m.textInput.Reset()
				m.mode = HomeMode
			}
		case errMsg:
			m.err = msg
		}
		// TODO: ??
		m.textInput, _ = m.textInput.Update(msg)
	case EditMode:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.mode = CheckMode
			case tea.KeyEnter:
				currentItem := m.toDoList.Items()[m.cursor].(Item)
				m.toDoList.SetItem(m.cursor, Item{
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
		return s

	case NewMode:
		return fmt.Sprintf("What's the plan for today?\n\n%s\n\n%s", m.textInput.View(), "(esc to quit)") + "\n"

	case CheckMode:
		return m.toDoList.View()

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
