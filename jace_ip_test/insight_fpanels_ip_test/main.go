package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/gen2brain/beeep"
	probing "github.com/prometheus-community/pro-bing"
)

// NOTE: This script looks at the csv in the ./data directory and pings all Insight field panels contained.
// To update the list of panels, simply generate a Panel Configuration report which encompasses ALL field
// panels on campus, export to .txt, and drop the file in the ./data directory, and run the script.
// The panel list will be updated automatically. ONLY ONE TXT FILE SHOULD BE PRESENT AT ANY ONE TIME -- DELETE
// THE OLD ONE WHEN DROPPING IN A NEW ONE. Any panels which are taken offline (ie not simply a changed
// IP) will need to be deleted from the csv file manually.

type PingResult struct {
	Panel    string
	PacketsSent int
	PacketsRecv int
	AvgRtt      string
	Success     bool
	IP string
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

func main() {
	pingTimer := time.NewTicker(30 * time.Minute) // Run script and provide system notification every 30 minutes
	mainFnc()
	// for range pingTimer.C {
	// 	<- pingTimer.C
	// 	mainFnc()
	// }
	for range pingTimer.C { mainFnc() }
	}

	func arrUnion(a, b [][]string) [][]string {
		// Use a map to track which first elements have already been seen
		seen := make(map[string]bool)

		// Add all first elements from `a` to the map
		for _, row := range a {
			if len(row) > 0 {
				seen[row[0]] = true
			}
		}

		// Iterate over `b` and add rows whose first element is not yet in `seen`
		for _, row := range b {
			if len(row) > 0 && !seen[row[0]] {
				a = append(a, row)
				seen[row[0]] = true
			}
		}

		return a
}

func writeToCSV(filename string, data [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("could not write row: %w", err)
		}
	}
	return nil
}

	func mainFnc() {
		csvArr := updatePanelCSV() // defines list of panels and IPs from the local csv file
		matchArr := getIPs()

		// define the list of panels and IPs from Insight report
		var txtArr [][]string
		for _, matches := range matchArr {
			panel := strings.TrimSuffix(matches[1], "\r")
			ip := matches[2]
			txtArr = append(txtArr, []string{panel, ip})
		}

		// update the list of panels and IPs. If IP is different in the Insight report, update it for
		// eventual update in the local csv
		var updatedArr [][]string
		for _, txtPair := range txtArr {
			for _, csvPair := range csvArr {
				if txtPair[0] == csvPair[0] && txtPair[1] == csvPair[1] {
					updatedArr = append(updatedArr, csvPair)
				} else if txtPair[0] == csvPair[0] && txtPair[1] != csvPair[1] {
					updatedArr = append(updatedArr, txtPair)
				}
			}
		}

		// Update any missing panels in our working slice that are present in the Insight report or the 
		// csv file
		updatedArr = arrUnion(updatedArr, txtArr)
		updatedArr = arrUnion(updatedArr, csvArr)
		updatedArr = slices.Insert(updatedArr, 0, []string{"panel", "ip"}) // add column headers, just for writing to csv
		// write updatedArr to csv file
		err := writeToCSV("./data/fieldpanels.csv", updatedArr)
		if err != nil {
			log.Fatal(err)
		}
		updatedArr = updatedArr[1:] // remove column headers
		nSuccess, successArr := pingIPs(updatedArr) // mass ping all panels in the newly updated slice
		
		log.Printf("\nFinished pinging all Insight field panels. %d out of %d panels pinged successfully.\n\n", nSuccess, len(updatedArr))
		
		msg_str := ""
		for _, matches := range updatedArr {
			panel := strings.TrimSuffix(matches[0], "\r")
			ip := matches[1]
			if !contains(successArr, panel) {
				fmt.Printf("\nUnsuccessful ping at %s.      IP: %s", panel, ip)
				msg_str += (panel + " offline\n")
			}
		}

		if msg_str == "" {
			msg_str = "No issues!"
		}
		err = beeep.Notify("Insight Panels Ping", msg_str, "assets/information.png")
		if err != nil {
			log.Fatal(err)
		}
	}

	func updatePanelCSV() [][]string {
		csvFile, err := os.Open("./data/fieldpanels.csv")
		if err != nil {
			log.Fatal(err)
		}
		defer csvFile.Close()
		reader := csv.NewReader(csvFile)
		csvArr := [][]string{}
		for {
			records, err := reader.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
			log.Fatal(err)
			}
		// fmt.Println(records)
		csvArr = append(csvArr, records)
		}
	return csvArr[1:]
	}

	func parseFiles(dir string, file os.DirEntry) [][]string {
			f, err := os.ReadFile(dir + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		txt := string(f)

		// returns match only if IP address is listed within 25 lines of panel name
		re := regexp.MustCompile(`Panel System Name: *([\S].*)\n(?:.*\n){1,25}IP Address: *(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
		matchArr := re.FindAllStringSubmatch(txt, -1) // [panelName, upTo25Lines, ipAddress]
		cleanArr := [][]string{}

		for _, matches := range matchArr {
			if matches[2] != "0.0.0.0" {
				cleanArr = append(cleanArr, matches)
			}
		}
	return cleanArr
	}

	func getIPs() [][]string {
	dir := "./data/"
	files, _ := os.ReadDir(dir)
	var cleanArr [][]string
	for _, file := range files {
		if file.Name()[len(file.Name()) - 4:] == ".txt" {
			cleanArr = parseFiles(dir, file)
		}
	}
	return cleanArr
}

func pingIPs(matchArr [][]string) (int, []string) {
	resultsChan := make(chan PingResult, len(matchArr))
	nSuccess := 0
	successArr := []string{}
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	for _, matches := range matchArr {
		wg.Add(1)
		go func(panel, ip string){
			defer wg.Done()
			panel = strings.TrimSuffix(panel, "\r")
			pinger, err := probing.NewPinger(ip)
			if err != nil {
				log.Printf("Error creating pinger for %s: %v\n", panel, err)
				return
			}
			pinger.SetPrivileged(true) // ICMP ping requests won't work without this
			pinger.Timeout = 20 * time.Second // unit nanoseconds
			pinger.Interval = 6 * time.Second
			pinger.Count = 3
			
			err = pinger.Run()
			if err != nil {
				log.Printf("Error running pinger for %s: %v\n", panel, err)
				return
			}
			
			stats := pinger.Statistics()
			success := stats.PacketsRecv >= (stats.PacketsSent - 1) // condition to consider a panel as "online"
			
			resultsChan <- PingResult{
				Panel: panel,
				PacketsSent: stats.PacketsSent,
				PacketsRecv: stats.PacketsRecv,
				AvgRtt: stats.AvgRtt.String(),
				Success: success,
				IP: ip,
			}

			mu.Lock()
			if success {
				nSuccess++
				successArr = append(successArr, panel)
			}
			mu.Unlock()
			}(matches[0], matches[1])
		}
		
		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		// green := color.New(color.FgGreen)
		red := color.New(color.FgRed)
		for result := range resultsChan {
			n_spaces := 20 - len(result.Panel)
			spaces := strings.Repeat(" ", n_spaces)
			if result.Success {
				// green.Printf("%s:%s %d packets transmitted,   %d packets received,   average time %v,   success: %v\n",
				// result.Panel, spaces, result.PacketsSent, result.PacketsRecv, result.AvgRtt, result.Success)
				} else {
					red.Printf("%s:%s %d packets transmitted,   %d packets received,   average time %v,   success: %v\n",
					result.Panel, spaces, result.PacketsSent, result.PacketsRecv, result.AvgRtt, result.Success)
				}
			}
	return nSuccess, successArr
}