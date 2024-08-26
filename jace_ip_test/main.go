package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	probing "github.com/prometheus-community/pro-bing"
)

// PingResult holds results of a ping operation
type PingResult struct {
	Building    string
	PacketsSent int
	PacketsRecv int
	AvgRtt      string
	Success     bool
}

func main() {
	// read CSV file
	credsDir := `C:\Users\agilchri\OneDrive - University of Waterloo\Home\General Resources\BAS\Niagara\JACE_creds.csv`
	file, err := os.Open(credsDir)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	records = records[1:] // skip header

	// Create a buffered channel to collect the ping results from each goroutine.
	// The buffer size is set to the number of records to prevent blocking.
	results := make(chan PingResult, len(records))
	// Create a WaitGroup to keep track of all the goroutines.
	// The WaitGroup ensures that the main function waits until all pings are done.
	var wg sync.WaitGroup

	for _, record := range records {
		wg.Add(1) // increment the waitgroup counter for each goroutine
		go func(bldg, ip string) { // create a goroutine for each record
			defer wg.Done() // decrement the waitgroup counter when the goroutine completes

			pinger, err := probing.NewPinger(ip)
			if err != nil {
				log.Printf("Error creating pinger for %s: %v", bldg, err)
				return
			}
			pinger.SetPrivileged(true) // ICMP ping requests won't work without this
			pinger.Timeout = 10 * 1000 * 1000 * 1000 // unit nanoseconds
			pinger.Count = 3

			err = pinger.Run()
			if err != nil {
				log.Printf("Error running pinger for %s: %v", bldg, err)
				return
			}

			stats := pinger.Statistics()
			success := stats.PacketsSent == stats.PacketsRecv

			// Send the ping result to the results channel
			results <- PingResult{
				Building:    bldg,
				PacketsSent: stats.PacketsSent,
				PacketsRecv: stats.PacketsRecv,
				AvgRtt:      stats.AvgRtt.String(),
				Success:     success,
			}
		}(record[0], record[1]) // pass the building and IP address to the goroutine
	}

	// Close the channel once all goroutines are done
	go func() {
		wg.Wait() // wait for all goroutines to finish
		close(results) // close the channel
	}()

	// Collect and log the results
	// the range loop will continue until the channel is closed!
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	for result := range results {
		n_spaces := 5 - len(result.Building)
		spaces := strings.Repeat(" ", n_spaces)
		if result.Success {
			green.Printf("%s:%s %d packets transmitted,   %d packets received,   average time %v,   success: %v\n",
				result.Building, spaces, result.PacketsSent, result.PacketsRecv, result.AvgRtt, result.Success)
		} else {
			red.Printf("%s:%s %d packets transmitted,   %d packets received,   average time %v,   success: %v\n",
				result.Building, spaces, result.PacketsSent, result.PacketsRecv, result.AvgRtt, result.Success)
		}
	}
	fmt.Println("\nFinished pinging all JACE IP addresses!")

	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}
