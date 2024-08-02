package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/ini.v1"
)

func cliModel() (outTable map[string][]string) {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	
	finalModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
	
	output := finalModel.(model).outputRow

	return output
}

// const (
// 	hotPink  = lipgloss.Color("#FF06B7")
// 	darkGray = lipgloss.Color("#767676")
// )

// var (
// 	inputStyle    = lipgloss.NewStyle().Foreground(hotPink)
// 	continueStyle = lipgloss.NewStyle().Foreground(darkGray)
// )


type model struct {
	mode string
	lrSkip int
	taskNum int
	prompt string
	choices []string
	WRchoices []string
	workCatChoices []string
	optionsChoices []string
	highlighted int
	dayDelta int
	monthDelta int
	yearDelta int
	dayOfWeek string
	selected map[int]struct{}
	textinput  textinput.Model
	table table.Model

	outputRow map[string][]string
	err     error
}

func getDateDiff(deltaYears int, deltaMonths int, deltaDays int) string {
	now := time.Now()

	y, mnth, d := now.AddDate(deltaYears, deltaMonths, deltaDays).Date()
	return fmt.Sprintf("%d-%v-%02d", y, mnth, d)
}

func readCsv(relPath string) [][]string {
	csvFile, err := os.Open(relPath)
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	return records
}

func initialModel() model {
	today := getDateDiff(0, 0, 0)

	ti := textinput.New()
	ti.Placeholder = "Enter description"
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 100

	WRcats := readCsv("./data/wr_cats.csv")
	var WRcatsFiltered []string
	for _, WRcat := range WRcats {
		if WRcat[1] == "1" {
			WRcatsFiltered = append(WRcatsFiltered, WRcat[0])
		}
	}

	WRnums := readCsv("./data/wr_nums.csv")
	var WRnumsFiltered []string
	for _, WRnum := range WRnums {
		if WRnum[2] == "1" {
			WRnumsFiltered = append(WRnumsFiltered, WRnum[0] + "   " + WRnum[1])
		}
	}

	tableCols := []table.Column {
		{Title: "Date", Width: 20},
		{Title: "WR", Width: 30},
		{Title: "Description", Width: 50},
		{Title: "Work Category", Width: 30},
		{Title: "Hours", Width: 5},
	}
	tableRows := []table.Row{
		{"", "", "", "", ""},
	}
	t := table.New(
		table.WithColumns(tableCols),
		table.WithRows(tableRows),
		table.WithFocused(true),
		table.WithHeight(1),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Cell = s.Cell.
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(false)
	t.SetStyles(s)


	return model{
		mode: "dateSelect",
		lrSkip: 3,
		taskNum: 1,
		prompt: "Which timesheet date would you like to update?\n\n",
		choices: strings.Split(today, "-"),
		WRchoices : WRnumsFiltered,
		workCatChoices: WRcatsFiltered,
		optionsChoices: []string{"Lieu Hours Setup", "Back to Date Select"},
		highlighted: 2,
		dayDelta: 0,
		monthDelta: 0,
		yearDelta: 0,
		dayOfWeek: time.Now().Weekday().String(),
		selected: make(map[int]struct{}),
		textinput:  ti,
		table: t,

		outputRow: map[string][]string{
			"date": {},
			"WR":   {},
			"description": {},
			"cat":  {},
			"hours": {},
		},
		err:     nil,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.mode == "dateSelect" {
		return updateDateSelect(m, msg)
	}

	if m.mode == "WRselect" {
		return updateWRSelect(m, msg)
	}

	if m.mode == "descriptionInput" {
		return updateDescriptionInput(m, msg)
	}

	if m.mode == "hoursInput" {
		return updateHoursInput(m, msg)
	}

	if m.mode == "viewTable" {
		return updateViewTable(m, msg)
	}

	if m.mode == "workCatselect" {
		return updateWorkCatselect(m, msg)
	}

	if m.mode == "options" {
		return updateOptions(m, msg)
	}

	if m.mode == "lieuHoursSetup" {
		return updateLieuHoursSetup(m, msg)
	}

	if m.mode == "debugMode" {
		return updateDebugMode(m, msg)
	}

	return m, nil
}

func (m model) View() string {
	var s string
	if m.mode == "dateSelect" {
	s = viewDateSelect(m)
	}

	if m.mode == "WRselect" {
		s = viewWRSelect(m)
	}

	if m.mode == "descriptionInput" {
		s = viewDescriptionInput(m)
	}

	if m.mode == "hoursInput" {
		s = viewHoursInput(m)
	}

	if m.mode == "viewTable" {
		s = viewTable(m)
	}

	if m.mode == "workCatselect" {
		s = viewWorkCatselect(m)
	}

	if m.mode == "options" {
		s = viewOptions(m)
	}

	if m.mode == "lieuHoursSetup" {
		s = viewLieuHoursSetup(m)
	}

	if m.mode == "debugMode" {
		s = viewDebugMode(m)
	}

	return s
}

func calcWrkHrs(ref_date_str string) float32 {
	// FIXME: This is untested!!! 
	//        Returns number of business days right now, need to return hours (just x7?)
	inidata, err := ini.Load("./data/options.ini")
	if err != nil {
		log.Fatal(err)
	}
	ref_date_str = inidata.Section("lieu_hours").Key("ref_date").String()
	ref_date, err := time.Parse("2006-January-02", ref_date_str)
	if err != nil {
		log.Fatal(err)
	}
	today := time.Now()
	var businessDays float32 = 0
	for {
		if today.Equal(ref_date) {
			return businessDays
		}
		if (today.Weekday() != time.Saturday && today.Weekday() != time.Sunday) {
			businessDays++
	   }
	   today = today.Add(time.Hour*24)
	}
}

func updateDateSelect(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		
		case "left", "a":
			if m.highlighted > 0 {
				m.highlighted--
			}
		
		case "right", "d":
			if m.highlighted < len(m.choices) - 1 {
				m.highlighted++
			}

		case "up", "w":
			if m.highlighted == 2 {
				m.dayDelta ++
			} else if m.highlighted == 1 {
				m.monthDelta ++
			} else if m.highlighted == 0 {
				m.yearDelta ++
			}
			m.choices = strings.Split(getDateDiff(m.yearDelta, m.monthDelta, m.dayDelta), "-")
			date, _ := time.Parse("2006-January-02", getDateDiff(m.yearDelta, m.monthDelta, m.dayDelta))
			m.dayOfWeek = date.Weekday().String()

		case "down", "s":
			if m.highlighted == 2 {
				m.dayDelta --
			} else if m.highlighted == 1 {
				m.monthDelta --
			} else if m.highlighted == 0 {
				m.yearDelta --
			}
			m.choices = strings.Split(getDateDiff(m.yearDelta, m.monthDelta, m.dayDelta), "-")
			date, _ := time.Parse("2006-January-02", getDateDiff(m.yearDelta, m.monthDelta, m.dayDelta))
			m.dayOfWeek = date.Weekday().String()
		
		case "o":
			m.mode = "options"
			m.highlighted = 0
			m.prompt = "What would you like to do?\n \n"
		
		case "enter":
			m.outputRow["date"] = append(m.outputRow["date"], getDateDiff(m.yearDelta, m.monthDelta, m.dayDelta))
			m.mode = "WRselect"
			m.highlighted = 0
			m.prompt = "Please select a WR\n \n"
		}
	}

	return m, nil
	}

func updateWRSelect(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		
		case "up", "w":
			if m.highlighted > 0 {
				m.highlighted--
			}

		case "down", "s":
			if m.highlighted < len(m.WRchoices) - 1 {
				m.highlighted++
			}

		case "left", "a":
			if m.highlighted > 3 {
				m.highlighted -= m.lrSkip
			} else if m.highlighted <= 3 {
				m.highlighted = 0
			}
		
		case "right", "d":
			max := (len(m.WRchoices) - 1) - 3
			if m.highlighted < max {
				m.highlighted += m.lrSkip
			} else if m.highlighted >= max {
				m.highlighted = len(m.WRchoices) - 1
			}

		case "enter":
			m.outputRow["WR"] = append(m.outputRow["WR"], m.WRchoices[m.highlighted])
			m.mode = "descriptionInput"
			m.prompt = "Please enter a description of your work\n \n "
			m.textinput.Reset()
			m.textinput.Placeholder = "Enter description"
		
		case "esc":
			return m, tea.Quit
		}
	}

	return m, nil
}

func updateDescriptionInput(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			_, err := strconv.ParseFloat(m.textinput.Value(), 32)
			if err == nil {
				m.prompt = "Please enter a description of your work\nDescription cannot be a number!\n \n "
			} else if m.textinput.Value() == "" {
				m.prompt = "Please enter a description of your work\nDescription cannot be blank!\n \n "
			} else {
			m.outputRow["description"] = append(m.outputRow["description"], m.textinput.Value())
			m.mode = "workCatselect"
			m.prompt = "Please select a work category\n \n"
			m.highlighted = 0
			}

		case "esc":
			m.mode = "exitCLI"
		}
	}

	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func updateWorkCatselect(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		
		case "up", "w":
			if m.highlighted > 0 {
				m.highlighted--
			}

		case "down", "s":
			if m.highlighted < len(m.workCatChoices) - 1 {
				m.highlighted++
			}

		case "left", "a":
			if m.highlighted > 3 {
				m.highlighted -= m.lrSkip
			} else if m.highlighted <= 3 {
				m.highlighted = 0
			}
		
		case "right", "d":
			max := (len(m.workCatChoices) - 1) - 3
			if m.highlighted < max {
				m.highlighted += m.lrSkip
			} else if m.highlighted >= max {
				m.highlighted = len(m.workCatChoices) - 1
			}

		case "enter":
			m.outputRow["cat"] = append(m.outputRow["cat"], m.workCatChoices[m.highlighted])

			m.mode = "hoursInput"
			m.prompt = "Please enter hours worked on this job\n \n "
			m.textinput.Reset()
			m.textinput.Placeholder = "Enter hours worked"
		case "esc":
			m.mode = "exitCLI"
		}
	}

	return m, nil
}

func updateHoursInput(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			f, err := strconv.ParseFloat(m.textinput.Value(), 32)
			if err != nil || f < 0.0 || f > 24.0 {
			m.prompt = "Enter hours worked\nPlease enter a valid number!\n \n "
			} else {
				m.outputRow["hours"] = append(m.outputRow["hours"], m.textinput.Value())
				m.table.SetHeight(len(m.outputRow["WR"]))
				var rows []table.Row
				for i := 0; i < len(m.outputRow["WR"]); i++ {
					rows = append(rows, table.Row{m.outputRow["date"][0],
					m.outputRow["WR"][i], 
					m.outputRow["description"][i],
					m.outputRow["cat"][i],
					m.outputRow["hours"][i]})
				}
				m.table.SetRows(rows)
				
				m.mode = "WRselect"
				m.prompt = "Please select a WR\n \n"
				m.highlighted = 0
				m.taskNum += 1
			}

		case "esc":
			m.mode = "exitCLI"
		}
	}


	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd

}

func updateOptions(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		
		case "up", "w":
			if m.highlighted > 0 {
				m.highlighted--
			}

		case "down", "s":
			if m.highlighted < len(m.optionsChoices) - 1 {
				m.highlighted++
			}

		case "left", "a":
			if m.highlighted > 3 {
				m.highlighted -= m.lrSkip
			} else if m.highlighted <= 3 {
				m.highlighted = 0
			}
		
		case "right", "d":
			max := (len(m.optionsChoices) - 1) - 3
			if m.highlighted < max {
				m.highlighted += m.lrSkip
			} else if m.highlighted >= max {
				m.highlighted = len(m.optionsChoices) - 1
			}

		case "enter":
			if m.optionsChoices[m.highlighted] == "Lieu Hours Setup" {
				m.mode = "lieuHoursSetup"
				m.prompt = "Before the beginning of today's work day, how many lieu hours did you have available?\n \n "
			} else if m.optionsChoices[m.highlighted] == "Back to Date Select" {
				m.mode = "dateSelect"
				m.prompt = "Which timesheet date would you like to update?\n\n"
				m.textinput.Placeholder = "Enter current lieu hours"
			}
		}
	}

	return m, nil
}

func updateLieuHoursSetup(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			f, err := strconv.ParseFloat(m.textinput.Value(), 32)
			if err != nil {
			m.prompt = "Before the beginning of today's work day, how many lieu hours did you have available?\nPlease enter a valid number!\n \n "
			} else {
				// ini
				inidata := ini.Empty()
				sec, err := inidata.NewSection("lieu_hours")
				if err != nil {
					log.Fatal(err)
				}
				// ref_date
				y, mnth, d := time.Now().Date()
				date := fmt.Sprintf("%d-%v-%02d", y, mnth, d)
				_, err = sec.NewKey("ref_date", date)
				if err != nil {
					log.Fatal(err)
				}
				// act_hrs_since_ref
				_, err = sec.NewKey("act_hrs_since_ref", strconv.FormatFloat(f, 'f', -1, 32))
				if err != nil {
					log.Fatal(err)
				}
				// req_hrs_since_ref
				_, err = sec.NewKey("req_hrs_since_ref", "0")
				if err != nil {
					log.Fatal(err)
				}

				err = inidata.SaveTo("./data/options.ini")
				if err != nil {
					log.Fatal(err)
				}

			m.mode = "dateSelect"
			m.textinput.Reset()
			m.prompt = "Which timesheet date would you like to update?\n\n"
			m.highlighted = 0
			}
		}
	}


	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd

}

func updateViewTable(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.outputRow["date"] = append(m.outputRow["date"], getDateDiff(m.yearDelta, m.monthDelta, m.dayDelta))
			m.mode = "WRselect"
			m.prompt = "Please select a WR\n \n "
			m.highlighted = 0
		}
	}
	return m, nil
}


func updateDebugMode(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func viewDateSelect(m model) string {
	var s string
	s += m.prompt
	for _, choice := range m.choices {
		s += fmt.Sprintf("%s - ", choice)
	}
	s = s[:len(s) - 2] + "   (" + m.dayOfWeek + ")\n"
	switch m.highlighted {
	case 0:
		s += " ^"
	case 1:
		s += "         ^"
	case 2:
		s += "               ^"
	}
	s += "\n\nctl+c: quit | enter: select option | o: options\n" 

	return s
}

func viewWRSelect(m model) string {
	s := viewTable(m) + "\n\n"
	s += "TASK " + strconv.Itoa(m.taskNum) + " - " + m.prompt
	for i, choice := range m.WRchoices {
		cursor := " "
		if m.highlighted == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s[%d] %s\n", cursor, i + 1, choice)
	}

	s += "\n\nctl+c: quit | enter: select option | esc: write to timesheet\n" 

	return s
}

func viewDescriptionInput(m model) string {
	s := viewTable(m) + "\n\n"
	s += "TASK " + strconv.Itoa(m.taskNum) + " - " + m.prompt + m.textinput.View() + "\n\nq: quit | enter: submit | esc: write to timesheet\n"
	return s
}

func viewWorkCatselect(m model) string {
	s := viewTable(m) + "\n\n"
	s += "TASK " + strconv.Itoa(m.taskNum) + " - " + m.prompt
	for i, choice := range m.workCatChoices {
		cursor := " "
		if m.highlighted == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s[%d] %s\n", cursor, i + 1, choice)
	}

	s += "\n\nctl+c: quit | enter: select option | esc: write to timesheet\n" 

	return s
}

func viewHoursInput(m model) string {
	var s string
	s += viewTable(m) + "\n\n"
	s += "TASK " + strconv.Itoa(m.taskNum) + " - " + m.prompt + m.textinput.View() + "\n\nq: quit | enter: submit | esc: write to timesheet\n"
	return s
}

func viewTable(m model) string {
	var s string
	s += m.table.View()
	hourSum := 0.0

	for i := 0; i < len(m.outputRow["hours"]); i++ {
		f, _ := strconv.ParseFloat(m.outputRow["hours"][i], 32)
		hourSum += f
	}
	s += fmt.Sprintf("\nTotal hours: %.2f\n", hourSum)
	return s
}

func viewOptions(m model) string {
	var s string
	s += m.prompt
	for i, choice := range m.optionsChoices {
		cursor := " "
		if m.highlighted == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\n\nctl+c: quit | enter: select option\n"

	return s
}

func viewLieuHoursSetup(m model) string {
	var s string
	s += m.prompt + m.textinput.View() + "\n\nctl+c: quit | enter: submit\n"
	return s
}

func viewDebugMode(m model) string {
	var s string

	s += fmt.Sprintf("Selected Date: %s\n", m.outputRow["date"])
	s += fmt.Sprintf("Selected WR: %s\n", m.outputRow["WR"])
	s += fmt.Sprintf("Selected Description: %s\n", m.outputRow["description"])
	s += fmt.Sprintf("Selected hours worked: %s\n", m.outputRow["hours"])

	s += "\n\nPress ctl+c to quit\n"

	return s
}