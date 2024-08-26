package main

import (
	"bufio"
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
	FinalOctet    string
	PacketsSent int
	PacketsRecv int
	AvgRtt      string
	Success     bool
	Addr string
}

func main() {
	// Create a buffered channel to collect the ping results from each goroutine.
	// The buffer size is set to the number of records to prevent blocking.
	results := make(chan PingResult, 254)
	// Create a WaitGroup to keep track of all the goroutines.
	// The WaitGroup ensures that the main function waits until all pings are done.
	var wg sync.WaitGroup

	for i := 2; i <= 255; i++ {
		wg.Add(1) // increment the waitgroup counter for each goroutine
		addr := fmt.Sprintf("%d", i)
		go func(addr string) { // create a goroutine for each record
			defer wg.Done() // decrement the waitgroup counter when the goroutine completes

			pinger, err := probing.NewPinger("129.97.154." + addr)
			if err != nil {
				log.Printf("Error creating pinger for address %s: %v", addr, err)
				return
			}
			pinger.SetPrivileged(true) // ICMP ping requests won't work without this
			pinger.Timeout = 10 * 1000 * 1000 * 1000 // unit nanoseconds
			pinger.Count = 3

			err = pinger.Run()
			if err != nil {
				log.Printf("Error running pinger for address %s: %v", addr, err)
				return
			}

			stats := pinger.Statistics()
			success := stats.PacketsSent == stats.PacketsRecv

			// Send the ping result to the results channel
			results <- PingResult{
				FinalOctet:    addr,
				PacketsSent: stats.PacketsSent,
				PacketsRecv: stats.PacketsRecv,
				AvgRtt:      stats.AvgRtt.String(),
				Success:     success,
				Addr: stats.Addr,
			}
		}(addr) // pass the building and IP address to the goroutine
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
		n_spaces := 5 - len(result.FinalOctet)
		spaces := strings.Repeat(" ", n_spaces)
		if result.Success {
			green.Printf("%s:%s %d packets transmitted,   %d packets received,   average time %v,   success: %v\n",
				result.Addr, spaces, result.PacketsSent, result.PacketsRecv, result.AvgRtt, result.Success)
		} else {
			red.Printf("%s:%s %d packets transmitted,   %d packets received,   average time %v,   success: %v\n",
				result.Addr, spaces, result.PacketsSent, result.PacketsRecv, result.AvgRtt, result.Success)
		}
	}
	fmt.Println("\nFinished pinging all local subnet IP addresses!")

	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}
