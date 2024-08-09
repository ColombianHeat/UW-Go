package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/go-vgo/robotgo"
	bolt "go.etcd.io/bbolt"
)

// TODO: Bold the text displaying user controls. eg: ctrl+c: quit
func main() {
	var outTable map[string][]string = CliModel()
	if len(outTable["hours"]) == 0 {
		fmt.Print("\nNo complete entries found. Exiting Program...")
		fmt.Scanln()
		os.Exit(0)
	}
	nExisting := updateDB(outTable)

	// read encrypted password from file
	encryptedPass, err := os.ReadFile("./data/pass.enc")
	if err != nil {
		log.Printf("\nFailed to read password file: %v", err)
		fmt.Scanln()
		os.Exit(1)
	}
	decryptedPass, err := decrypt(encryptedPass)
	if err != nil {
		log.Fatal(err)
	}
	

	openMSAccess()

	// Sign into MS Access, navigate to timesheet, and fill in timesheet
	robotgo.KeySleep = 100
	waitForProcess("Logon")
	robotgo.TypeStr(string(decryptedPass))
	robotgo.KeyTap("tab")
	robotgo.KeyTap("enter")
	waitForProcess("Work Requests")
	time.Sleep(3000 * time.Millisecond) //TODO: is there a way to not depend on a guess of absolute time?

	goToTimesheet()
	robotgo.ActiveName("Work Requests")
	robotgo.KeySleep = 25
	fillTimesheet(outTable, nExisting)
}

func openMSAccess() {
	acc := []string{"cmd.exe", `/C`, `C:\Program Files\PlantOps SQL Apps\WorkRequests.mdb`, `/wrkgrp`, `V:\SecurityDatabases\WorkRequests_SecurityDB.mdw`}
	cmd := exec.Command(acc[0], acc[1:]...)
	err := cmd.Start()
	
	if err != nil {
			log.Fatal(err)
		}
}

// NOTE: Not used. Might be useful for other projects, though
// Executes the timesheet.py Python script
// func runPython() {
// 	py := []string{`cmd.exe`, `/C`, `start`, `py`, `timesheet.py`}
// 	cmd := exec.Command(py[0], py[1:]...)
// 	cmd.Dir = `C:\Users\agilchri\OneDrive - University of Waterloo\Home\Programming\Python`
// 	err := cmd.Start()

// 	if err != nil {
// 		log.Fatalf("cmd.Run() failed with %s", err)
// 	}
// }

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

func fillTimesheet(cliData map[string][]string, nExisting int) {
	date := cliData["date"][0]
	robotgo.KeyTap("tab")
	robotgo.TypeStr(date)
	robotgo.KeyTap("tab")
	fmt.Println(nExisting)
	for i := nExisting; i > 0; i-- {
		for j := 0; j < 5; j++ {
			robotgo.KeyTap("tab")
		}
	}

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

func updateDB(outTable map[string][]string) int {
	// open db. If not exists, create
	db, err := bolt.Open("./data/dailyEntries.db", 0600, nil) // 0600 - read-write permission
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// create bucket, keys, values
	date := outTable["date"][0]
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, _ := tx.CreateBucketIfNotExists([]byte(date)) // create date bucket
		for i := range outTable["hours"] {
			desc := outTable["description"][i]
			hrs := outTable["hours"][i]
			if err := bucket.Put([]byte(desc), []byte(hrs)); err != nil {
				return fmt.Errorf("put key-value pair: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("update transaction failed: %v", err)
	}

	// read db, get number of existing entries for this date
	// will be used outside of func to press "tab" this many extra times when updating timesheet
	var totalEntries int
	var nExisting int
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(date))
		
		cursor := bucket.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			totalEntries ++
		}
		nExisting = totalEntries - len(outTable["hours"])
		return nil
	})
	
	return nExisting
}