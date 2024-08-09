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

func CliModel() (outTable map[string][]string) {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	
	finalModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
	
	output := finalModel.(model).outputRow

	return output
}

const (
	txtCol  = lipgloss.Color("#e0b6d9")
	borderCol = lipgloss.Color("#767676")
	BgCol = lipgloss.Color("#242424")
	highlightCol = lipgloss.Color("#d60db5")
	notHighlightCol = lipgloss.Color("#c2c2c2")
	placeholderCol = lipgloss.Color("#767676")
	errorCol1 = lipgloss.Color("#bf3434")
	errorCol2 = lipgloss.Color("#ff0000")
	lieuHrCol = lipgloss.Color("#e0e34b")
)

var (
	bgColourStyle = lipgloss.NewStyle().Foreground(txtCol).Background(BgCol)
	inputStyle = lipgloss.NewStyle().
				Foreground(highlightCol).
				Background(BgCol)
)

type Styles struct {
	txt lipgloss.Style
	tableHeaderStyle lipgloss.Style
	tableContentStyle lipgloss.Style
	inputField lipgloss.Style
	backgroundColour lipgloss.Color
	borderColour lipgloss.Color
	choiceStyle lipgloss.Style
	highlightStyle lipgloss.Style
	errorStyle lipgloss.Style
	errorColour1 lipgloss.Color
	errorColour2 lipgloss.Color
	lieuHrColour lipgloss.Color
	lieuHrStyle lipgloss.Style
}

func DefaultStyles() *Styles {
	s := new(Styles)
	s.backgroundColour = lipgloss.Color(BgCol)
	s.borderColour = lipgloss.Color(borderCol)
	s.errorColour1 = lipgloss.Color(errorCol1)
	s.errorColour2 = lipgloss.Color(errorCol2)
	s.lieuHrColour = lipgloss.Color(lieuHrCol)
	s.choiceStyle = lipgloss.NewStyle().
					Background(s.backgroundColour).
					Foreground(notHighlightCol).
					MarginLeft(5).
					MarginBackground(s.backgroundColour)
	s.highlightStyle = s.choiceStyle.Foreground(highlightCol)
	s.txt = lipgloss.NewStyle().Foreground(txtCol).Background(BgCol)
	s.inputField = lipgloss.NewStyle().
				   BorderForeground(s.borderColour).
				   BorderBackground(s.backgroundColour).
				   BorderStyle(lipgloss.RoundedBorder()).
				   Padding(0)
	s.tableContentStyle = lipgloss.NewStyle().
	 		   	   Foreground(txtCol).
				   Background(BgCol).
				   BorderForeground(s.borderColour).
				   BorderBackground(s.backgroundColour).
				   BorderStyle(lipgloss.RoundedBorder()).
				   Padding(1)
	s.tableHeaderStyle = s.tableContentStyle.Bold(true)
	s.errorStyle = lipgloss.NewStyle().Foreground(s.errorColour1).Background(s.backgroundColour)
	s.lieuHrStyle = lipgloss.NewStyle().Foreground(s.lieuHrColour).Background(s.backgroundColour)
	return s
}


type model struct {
	width int
	height int
	mode string
	firstTimeSetup bool
	lrSkip int
	taskNum int
	prompt string
	errText string
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
	errTimer bool
	pad string

	iniPresent bool
	inidata *ini.File
	ref_date string
	act_hrs float64
	req_hrs float64

	styles *Styles

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

func initIni() *ini.File {
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
	_, err = sec.NewKey("act_hrs_since_ref", "0")
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

	return inidata
}

func updateIni(m *model) {
	hourSum := calcHours(*m)
	tdyWkDays, _ := CalcWkDays(m.ref_date)

	m.act_hrs += hourSum
	m.req_hrs = float64(7 * tdyWkDays)
	m.inidata.Section("lieu_hours").Key("act_hrs_since_ref").SetValue(strconv.FormatFloat(m.act_hrs, 'f', -1, 32))
	m.inidata.Section("lieu_hours").Key("req_hrs_since_ref").SetValue(strconv.FormatFloat(m.req_hrs, 'f', -1, 32))
	m.inidata.Section("lieu_hours").Key("ref_date").SetValue(m.ref_date)

	err := m.inidata.SaveTo("./data/options.ini")
	if err != nil {
		log.Fatal(err)
	}
}

func calcHours (m model) float64 {
	hourSum := 0.0
	for i := 0; i < len(m.outputRow["hours"]); i++ {
		f, _ := strconv.ParseFloat(m.outputRow["hours"][i], 32)
		hourSum += f
	}
	return hourSum
}

func applyBackground(m model, s string) string {
	return bgColourStyle.Width(m.width).Height(m.height).Render(s)
}

func initialModel() model {
	today := getDateDiff(0, 0, 0)

	ti := textinput.New()
	ti.PromptStyle = inputStyle
	ti.TextStyle = inputStyle
	ti.PlaceholderStyle = inputStyle.Foreground(placeholderCol)
	ti.Cursor.Style = inputStyle.Foreground(notHighlightCol)
	ti.CompletionStyle = lipgloss.NewStyle().Background(highlightCol)
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
		Foreground(txtCol).
		Background(BgCol).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderCol).
		BorderBackground(BgCol).
		BorderBottom(true).
		BorderLeft(false).
		BorderRight(true).
		Bold(true)
	s.Cell = s.Cell.
		Foreground(notHighlightCol).
		Background(BgCol).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderCol).
		BorderBackground(BgCol).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(true).
		Bold(false)
	t.SetStyles(s)

	iniPresent := true
	inidata, err := ini.Load("./data/options.ini")
	if err != nil {
		iniPresent = false
	}

	var refDate string
	var actHrs float64
	var reqHrs float64
	if iniPresent {
		refDate = inidata.Section("lieu_hours").Key("ref_date").String()
		actHrs = inidata.Section("lieu_hours").Key("act_hrs_since_ref").MustFloat64(0.0)
		reqHrs = inidata.Section("lieu_hours").Key("req_hrs_since_ref").MustFloat64(0.0)
	} else {
		refDate = today
		actHrs = 0.0
		reqHrs = 0.0
		inidata = initIni()
		}
		
	wkdaysSinceRef, _ := CalcWkDays(refDate)
	inidata.Section("lieu_hours").Key("req_hrs_since_ref").SetValue(strconv.FormatFloat(float64(wkdaysSinceRef * 7), 'f', -1, 32))

	var credsExist bool
	_, err = os.Stat("./data/pass.enc")
	if err == nil {
		credsExist = true
	} else {
		credsExist = false
	}
	var initMode string
	var prompt string
	pad := "   "
	if !credsExist {
		initMode = "firstTimeSetup"
		prompt = "\n" + pad + "Howdy partner! I see you haven't set up any login credentials yet.\n" +
						   pad + "This application automates the signing into the WR environment to update your timesheet.\n" +
						   pad + "Don't worry, your password will be encrypted and stored in the most secure location of all: your local computer.\n\n" +
						   pad + "What is your password for the WR Application?\n" + 
						   pad + "It is recommended that you create a one-off password for this app.\n" + 
						   pad + "To do so, sign in normally to the WR application, and navigate to System Maintenance > Security > Change Password\n\n"
		ti.Placeholder = "Enter password"
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
	} else {
		initMode = "dateSelect"
		prompt = "Which timesheet date would you like to update?\n\n"
	}


	return model{
		width: 0,
		height: 0,
		mode: initMode,
		firstTimeSetup: !credsExist,
		lrSkip: 3,
		taskNum: 1,
		prompt: prompt,
		errText: "\n",
		choices: strings.Split(today, "-"),
		WRchoices : WRnumsFiltered,
		workCatChoices: WRcatsFiltered,
		optionsChoices: []string{"Lieu Hours Setup", "Credentials Setup", "Back to Date Select"},
		highlighted: 2,
		dayDelta: 0,
		monthDelta: 0,
		yearDelta: 0,
		dayOfWeek: time.Now().Weekday().String(),
		selected: make(map[int]struct{}),
		textinput:  ti,
		table: t,
		errTimer: false,
		pad: pad,

		iniPresent: iniPresent,
		inidata: inidata,
		ref_date: refDate,
		act_hrs: actHrs,
		req_hrs: reqHrs,

		styles: DefaultStyles(),

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

	if m.mode == "firstTimeSetup" {
		return updateCredsSetup(m, msg)
	}

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

	if m.mode == "credsSetup" {
		return updateCredsSetup(m, msg)
	}

	if m.mode == "exitCLI" {
		return updateExitCLI(m, msg)
	}

	return m, nil
}

func (m model) View() string {
	var s string
	if m.mode == "dateSelect" {
	s = viewDateSelect(m)
	}

	if m.mode == "firstTimeSetup" {
		s = viewCredsSetup(m)
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

	if m.mode == "credsSetup" {
		s = viewCredsSetup(m)
	}

	if m.mode == "exitCLI" {
		s = viewExitCLI(m)
	}

	return s
}

func CalcWkDays(ref_date_str string) (int, error) {
	ref_date, err := time.Parse("2006-January-02", ref_date_str)
	if err != nil {
		return 0, err
	}
	today := time.Now()
	var businessDays int = 0
	for {
		if today.Before(ref_date) {
			return businessDays, nil
		}
		if (ref_date.Weekday() != time.Saturday && ref_date.Weekday() != time.Sunday) {
			businessDays++
	   }
	   ref_date = ref_date.AddDate(0, 0, 1)
	}
}

type errFlash struct {}

func resetColorAfterDuration(d time.Duration) tea.Cmd {
    return func() tea.Msg {
        time.Sleep(d)
        return errFlash{}
    }
}

func updateDateSelect(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	
	wkdaysSinceRef, _ := CalcWkDays(m.ref_date)
	m.inidata.Section("lieu_hours").Key("req_hrs_since_ref").SetValue(strconv.FormatFloat(float64(wkdaysSinceRef * 7), 'f', -1, 32))

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
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
			m.prompt = "\n" + m.pad + "What would you like to do?\n\n"
			m.textinput.Placeholder = "Enter current lieu hours"
		
		case "enter":
			m.outputRow["date"] = append(m.outputRow["date"], getDateDiff(m.yearDelta, m.monthDelta, m.dayDelta))
			m.mode = "WRselect"
			m.highlighted = 0
			m.prompt = "Please select a WR\n\n"
		}
	}

	return m, nil
	}

func updateWRSelect(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.outputRow = make(map[string][]string)
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
			m.prompt = "Please enter a description of your work"
			m.errText = "\n"
			m.textinput.Reset()
			m.textinput.Placeholder = "Enter description"
		
		case "esc":
			updateIni(&m)
			m.mode = "exitCLI"
		}
	}

	return m, textinput.Blink
}

func updateDescriptionInput(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.outputRow = make(map[string][]string)
			return m, tea.Quit
		case "enter":
			_, err := strconv.ParseFloat(m.textinput.Value(), 32)
			if err == nil {
				m.errText = "\nDescription cannot be a number!\n"
				m.styles.errorStyle = m.styles.errorStyle.Foreground(m.styles.errorColour2)
				return m, resetColorAfterDuration(500 * time.Millisecond)
				} else if m.textinput.Value() == "" {
					m.errText = "\nDescription cannot be blank!\n"
					m.styles.errorStyle = m.styles.errorStyle.Foreground(m.styles.errorColour2)
					return m, resetColorAfterDuration(500 * time.Millisecond)
			} else {
			m.outputRow["description"] = append(m.outputRow["description"], m.textinput.Value())
			m.mode = "workCatselect"
			m.prompt = "Please select a work category\n\n"
			m.errText = "\n"
			m.highlighted = 0
			}

		case "esc":
			updateIni(&m)
			m.mode = "exitCLI"
		}
	case errFlash:
		m.errTimer = false
		m.styles.errorStyle = m.styles.errorStyle.Foreground(m.styles.errorColour1)
	}

	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func updateWorkCatselect(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.outputRow = make(map[string][]string)
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
			m.prompt = "Please enter hours worked on this job"
			m.errText = "\n"
			m.textinput.Reset()
			m.textinput.Placeholder = "Enter hours worked"
		case "esc":
			updateIni(&m)
			m.mode = "exitCLI"
		}
	}

	return m, textinput.Blink
}

func updateHoursInput(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.outputRow = make(map[string][]string)
			return m, tea.Quit
		case "enter":
			f, err := strconv.ParseFloat(m.textinput.Value(), 32)
			if err != nil || f < 0.0 || f > 24.0 {
			// m.prompt = "Enter hours worked\nPlease enter a valid number!\n \n "
			m.errText = "\nPlease enter a valid number!\n"
			m.styles.errorStyle = m.styles.errorStyle.Foreground(m.styles.errorColour2)
			return m, resetColorAfterDuration(500 * time.Millisecond)
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
				m.prompt = "Please select a WR\n\n"
				m.highlighted = 0
				m.taskNum += 1
			}

		case "esc":
			updateIni(&m)
			m.mode = "exitCLI"
		}
	case errFlash:
		m.errTimer = false
		m.styles.errorStyle = m.styles.errorStyle.Foreground(m.styles.errorColour1)
	}


	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd

}

func updateOptions(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.outputRow = make(map[string][]string)
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
				m.prompt = "\n" + m.pad + "Before the beginning of today's work day, how many lieu hours did you have available?\n\n"
			} else if m.optionsChoices[m.highlighted] == "Credentials Setup" {
				m.mode = "credsSetup"
				m.prompt = "\n" + m.pad + "What is your password for the WR Application?\n" + 
						   m.pad + "It is recommended that you create a one-off password for this app.\n" + 
						   m.pad + "To do so, sign in normally to the WR application, and navigate to System Maintenance > Security > Change Password\n\n"
				m.textinput.Placeholder = "Enter password"
				m.textinput.EchoMode = textinput.EchoPassword
				m.textinput.EchoCharacter = '•'
			} else if m.optionsChoices[m.highlighted] == "Back to Date Select" {
				m.mode = "dateSelect"
				m.prompt = "Which timesheet date would you like to update?\n\n"
				m.textinput.Placeholder = "Enter current lieu hours"
			}
			m.highlighted = 2
		}
	}

	return m, textinput.Blink
}

func updateLieuHoursSetup(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.outputRow = make(map[string][]string)
			return m, tea.Quit
		case "enter":
			f, err := strconv.ParseFloat(m.textinput.Value(), 32)
			if err != nil {
			m.prompt = "Before the beginning of today's work day, how many lieu hours did you have available?\nPlease enter a valid number!\n\n"
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
				m.ref_date = date
				// act_hrs_since_ref
				_, err = sec.NewKey("act_hrs_since_ref", strconv.FormatFloat(f, 'f', -1, 32))
				if err != nil {
					log.Fatal(err)
				}
				m.act_hrs = f
				// req_hrs_since_ref
				_, err = sec.NewKey("req_hrs_since_ref", "0")
				if err != nil {
					log.Fatal(err)
				}
				m.req_hrs = 0

				err = inidata.SaveTo("./data/options.ini")
				if err != nil {
					log.Fatal(err)
				}

			m.mode = "dateSelect"
			m.textinput.Reset()
			if m.firstTimeSetup {
				m.prompt = "You're all set up, amigo! My last tidbit of wisdom for you is to change the default dimensions of this app window.\n" +
				m.pad + "Do this by right-clicking the top of the window, select \"Properties\", and edit the window width in the \"Layout\" tab. I like a width of 160, but you do you.\n" +
				m.pad + "Which timesheet date would you like to update?\n\n"
				m.firstTimeSetup = false
			} else {
			m.prompt = "Which timesheet date would you like to update?\n\n"
			}
			m.highlighted = 2
			}
		}
	}


	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd

}

func updateCredsSetup(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.outputRow = make(map[string][]string)
			return m, tea.Quit
		case "enter":
			if m.firstTimeSetup {
				m.mode = "lieuHoursSetup"
				m.prompt = "\n" + m.pad + "One last step! This application can track your lieu hours for you. Let's set it up.\n" +
							m.pad + "Before the beginning of today's work day, how many lieu hours did you have available?\n\n"
				m.textinput.Placeholder = "Enter current lieu hours"
			} else {
			m.mode = "dateSelect"
			m.prompt = "Which timesheet date would you like to update?\n\n"
			m.highlighted = 2
		}
		m.textinput.EchoMode = textinput.EchoNormal

			// encrypt password and write to file
			encryptedPass, err := encrypt([]byte(m.textinput.Value()))
			if err != nil {
				log.Fatal(err)
			}
			err = os.WriteFile("./data/pass.enc", encryptedPass, 0644)
			if err != nil {
				log.Fatal(err)
			}
			m.textinput.Reset()

		}
	}

	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func updateViewTable(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.outputRow = make(map[string][]string)
			return m, tea.Quit
		case "enter":
			m.outputRow["date"] = append(m.outputRow["date"], getDateDiff(m.yearDelta, m.monthDelta, m.dayDelta))
			m.mode = "WRselect"
			m.prompt = "Please select a WR\n\n"
			m.highlighted = 0
		}
	}
	return m, nil
}

func updateExitCLI(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.mode = "exitForGood"
			return m, tea.Quit
		}
	}
	return m, nil
}

func viewDateSelect(m model) string {
	var s string
	s += "\n" + m.pad + "CURRENT LIEU HOURS: " + m.styles.lieuHrStyle.Render(strconv.FormatFloat(float64(m.act_hrs - m.req_hrs), 'f', -1, 32))

	s += "\n\n" + m.pad + m.prompt
	for i, choice := range m.choices {
		if i == m.highlighted {
			s += m.styles.highlightStyle.Render(choice)
		} else {
			s += m.styles.choiceStyle.Render(choice)
		}
	}
	s += m.styles.choiceStyle.Render("   (" + m.dayOfWeek + ")\n")

	s += "\n\n" + m.pad + "ctl+c: quit | ←↑↓→: change date | enter: select date | o: options\n" 


	return applyBackground(m, s)
}

func viewWRSelect(m model) string {
	s := viewTable(m) + "\n\n"
	s += m.pad + "TASK " + strconv.Itoa(m.taskNum) + " - " + m.prompt
	for i, choice := range m.WRchoices {
		s += m.pad
		if m.highlighted == i {
			s += m.styles.highlightStyle.Render(choice)
		} else{
		s += m.styles.choiceStyle.Render(choice)
		}
		s += "\n"
	}

	s += "\n\n" + m.pad + "ctl+c: quit | ↑↓: navigate | enter: select option | esc: write to timesheet\n" 

	return applyBackground(m, s)
}

func viewDescriptionInput(m model) string {
	s := viewTable(m) + "\n\n"
	s += m.pad + "TASK " + strconv.Itoa(m.taskNum) + " - " + m.prompt + m.styles.errorStyle.Render(m.errText) + "\n" +
		  m.styles.inputField.Render(m.textinput.View()) + 
		  "\n\n" + m.pad + "q: quit | enter: submit | esc: write to timesheet\n"
	
	return applyBackground(m, s)
}

func viewWorkCatselect(m model) string {
	s := viewTable(m) + "\n\n"
	s += m.pad + "TASK " + strconv.Itoa(m.taskNum) + " - " + m.prompt
	for i, choice := range m.workCatChoices {
		s += m.pad
		if m.highlighted == i {
			s += m.styles.highlightStyle.Render(choice)
		} else{
		s += m.styles.choiceStyle.Render(choice)
		}
		s += "\n"
	}

	s += "\n\n" + m.pad + "ctl+c: quit | ↑↓: navigate | enter: select option | esc: write to timesheet\n" 

	return applyBackground(m, s)
}

func viewHoursInput(m model) string {
	var s string
	s += viewTable(m) + "\n\n"
	s += m.pad + "TASK " + strconv.Itoa(m.taskNum) + " - " + m.prompt + m.styles.errorStyle.Render(m.errText) + "\n" +
	m.styles.inputField.Render(m.textinput.View()) + 
	"\n\n" + m.pad + "ctl+c: quit | enter: submit | esc: write to timesheet\n"
	return applyBackground(m, s)
}

func viewTable(m model) string {
	var s string
	s += lipgloss.NewStyle().Padding(1, 2).Background(m.styles.backgroundColour).Render(m.table.View())

	hourSum := calcHours(m)

	s += fmt.Sprintf("\n   Total hours: %.2f\n", hourSum)
	return s
}

func viewOptions(m model) string {
	var s string
	s += m.prompt
	for i, choice := range m.optionsChoices {
		s += m.pad
		if m.highlighted == i {
			s += m.styles.highlightStyle.Render(choice)
		} else{
		s += m.styles.choiceStyle.Render(choice)
		}
		s += "\n"
	}

	s += "\n\n" + m.pad + "ctl+c: quit | ↑↓: navigate | enter: select option\n"

	return applyBackground(m, s)
}

func viewLieuHoursSetup(m model) string {
	var s string
	s += m.prompt + 
	m.styles.inputField.Render(m.textinput.View()) + 
	"\n\n" + m.pad + "ctl+c: quit | enter: submit\n"
	return applyBackground(m, s)
}

func viewCredsSetup(m model) string {
	var s string
	s += m.prompt + 
	m.styles.inputField.Render(m.textinput.View()) + 
	"\n\n" + m.pad + "ctl+c: quit | enter: submit\n"
	return applyBackground(m, s)
}

func viewExitCLI(m model) string {
	lieu := m.act_hrs - m.req_hrs
	
	var s string
	s += "\n" + "Your new balance of lieu hours is: "
	s += m.styles.lieuHrStyle.Render(strconv.FormatFloat(float64(lieu), 'f', -1, 32))
	s += lipgloss.NewStyle().Background(BgCol).Render("\n\nenter: continue\n")

	return applyBackground(m, s)
}