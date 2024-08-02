package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/go-vgo/robotgo"
)

func main() {
	// TODO: Add support for a settings file
	// 		 Should have option for a first-time setup to track lieu hours
	// TODO: Add lipgloss to CLI
	// TODO: Possibly add support for custom themes??
	var outTable map[string][]string = cliModel()
	// outTable = map[string][]string{"WR": {"52963   Misc. Energy", "52491   ReCx: DMS, EC4, PHR"},
	// 			   "cat": {"Flex Time", "Trouble Shooting"},
	// 			   "date": {"2024-August-01"},
	// 			   "description": {"Did a thing!", "Did another thing"},
	// 			   "hours": {"2", "1.5"}}
	if len(outTable["hours"]) == 0 {
		fmt.Println("No complete entries found. Exiting Program...")
		os.Exit(0)
	}
	fmt.Println(outTable)

	openMSAccess()
	// runPython()

	// Sign into MS Access, navigate to timesheet, and fill in timesheet
	robotgo.KeySleep = 100
	waitForProcess("Logon")
	robotgo.TypeStr("AlAnIsBeSt")
	robotgo.KeyTap("tab")
	robotgo.KeyTap("enter")
	waitForProcess("Work Requests")
	time.Sleep(3000 * time.Millisecond) //TODO: is there a way to not depend on a guess of absolute time?
	goToTimesheet()
	// robotgo.ActiveName("main")
	robotgo.ActiveName("Work Requests")
	robotgo.KeySleep = 25
	fillTimesheet(outTable)
}

func openMSAccess() {
	acc := []string{"cmd.exe", `/C`, `C:\Program Files\PlantOps SQL Apps\WorkRequests.mdb`, `/wrkgrp`, `V:\SecurityDatabases\WorkRequests_SecurityDB.mdw`}
	cmd := exec.Command(acc[0], acc[1:]...)
	err := cmd.Start()
	
	if err != nil {
			log.Fatal(err)
		}
}

func runPython() {
	py := []string{`cmd.exe`, `/C`, `start`, `py`, `timesheet.py`}
	cmd := exec.Command(py[0], py[1:]...)
	cmd.Dir = `C:\Users\agilchri\OneDrive - University of Waterloo\Home\Programming\Python`
	err := cmd.Start()

	if err != nil {
		log.Fatalf("cmd.Run() failed with %s", err)
	}
}

func waitForProcess(processName string) {
	for {
		robotgo.ActiveName(processName)
		activeWindow := robotgo.GetTitle()
		fmt.Printf("Active window: %s\n", activeWindow)
		if activeWindow == processName {
			fmt.Printf("Process %s is now active.\n", processName)
			break
		}
		time.Sleep(150 * time.Millisecond)
	}
}

func goToTimesheet() {
	robotgo.KeyTap("tab")
	robotgo.KeyTap("tab")
	robotgo.KeyTap("tab")
	robotgo.KeyTap("enter")
	robotgo.KeyTap("tab")
	robotgo.KeyTap("tab")
	robotgo.KeyTap("enter")
}

func fillTimesheet(cliData map[string][]string) {
	date := cliData["date"][0]
	robotgo.KeyTap("tab")
	robotgo.TypeStr(date)
	robotgo.KeyTap("tab")

	for i := len(cliData["WR"]) - 1; i >= 0; i-- {
		WRnum := cliData["WR"][i][0:6]
		robotgo.TypeStr(WRnum)
		robotgo.KeyTap("tab")
		robotgo.TypeStr(cliData["hours"][i])
		robotgo.KeyTap("tab")
		robotgo.TypeStr(cliData["cat"][i])
		robotgo.KeyTap("tab")
		robotgo.TypeStr(cliData["description"][i])
		robotgo.KeyTap("tab")
		robotgo.KeyTap("tab")
	}
}