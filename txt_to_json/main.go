package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// error_check checks if the given error is not nil and panics if it is.
//
// Parameters:
// - e: the error to check.
func error_check(e error) {
	if e != nil {
		panic(e)
	}
}

// extractPrettyFormat extracts point information from a cleaned text. Stores point values
// in the given slices.
//
// Parameters:
// - cleanTxt: a slice of strings representing the cleaned text.
// - names: a pointer to a slice of strings to store the extracted names.
// - addresses: a pointer to a slice of strings to store the extracted addresses.
// - descriptions: a pointer to a slice of strings to store the extracted descriptions.
// - currentStates: a pointer to a slice of strings to store the extracted current states.
// - statuses: a pointer to a slice of strings to store the extracted statuses.
// - priorities: a pointer to a slice of strings to store the extracted priorities.
func extractPrettyFormat(cleanTxt []string, names, addresses, descriptions, currentStates, statuses, priorities *[]string) {
	// Extract point information
	re_virtual := regexp.MustCompile(`^\s+-Virtual-`) // ignore lines that match this
	re_ppcladdress := regexp.MustCompile(`PPCL Address`) // ignore lines that match this
	re_values := regexp.MustCompile(`(?m)^(.+?)   (.+?)   \((.+)\) (.+?)  +(\S+)  +(\S+)`) // extract these values!
	matchMap := make(map[string]struct{})
	for _, line := range cleanTxt {
		if _, match := matchMap[line]; match {
			continue
		}
		matchMap[line] = struct{}{}
		if re_virtual.MatchString(line) || re_ppcladdress.MatchString(line) {
			continue
		}
		allFields := re_values.FindStringSubmatch(line)
		if allFields != nil {
			*names = append(*names, strings.TrimSpace(allFields[1]))
			*addresses = append(*addresses, strings.TrimSpace(allFields[2]))
			*descriptions = append(*descriptions, strings.TrimSpace(allFields[3]))
			*currentStates = append(*currentStates, strings.TrimSpace(allFields[4]))
			*statuses = append(*statuses, strings.TrimSpace(allFields[5]))
			*priorities = append(*priorities, strings.TrimSpace(allFields[6]))
		}
	}
}

// main is the entry point of the program.
//
// It reads all the .txt files in the current directory and extracts the relevant information from them.
// It then prints the number of lines, names, addresses, descriptions, current states, statuses, and priorities.
// Finally, it creates a JSON file named "all_points.json" with the extracted information.
//
// No parameters.
// No return values.
func main() {
	fileNames, err := filepath.Glob("*.txt")
	error_check(err)

	// Read text from file into list format
	var txt []string
	for _, report := range fileNames {
		data, err := os.ReadFile(report)
		error_check(err)
		txt = append(txt, strings.Split(string(data), "\n")...)
	}

	// Get rid of useless lines in the text
	var cleanTxt []string
	re1 := regexp.MustCompile(`\(.+\)`)
	re2 := regexp.MustCompile(`\(\d+ Field Panels\)`)
	re3 := regexp.MustCompile(`\(.+Points\)`)

	for _, line := range txt {
		if match := re1.MatchString(line); match && !re2.MatchString(line) && !re3.MatchString(line) {
			cleanTxt = append(cleanTxt, line)
		}
	}

	var names, addresses, descriptions, currentStates, statuses, priorities []string
	extractPrettyFormat(cleanTxt, &names, &addresses, &descriptions, &currentStates, &statuses, &priorities)

	// Ensuring that all features got the same number of extractions
	fmt.Printf("%-15s %5d\n", "n lines:", len(cleanTxt))
	fmt.Println()
	fmt.Printf("%-15s %5d\n", "n names:", len(names))
	fmt.Printf("%-15s %5d\n", "n addresses:", len(addresses))
	fmt.Printf("%-15s %5d\n", "n descriptions:", len(descriptions))
	fmt.Printf("%-15s %5d\n", "n curr states:", len(currentStates))
	fmt.Printf("%-15s %5d\n", "n statuses:", len(statuses))
	fmt.Printf("%-15s %5d\n", "n priorities:", len(priorities))
	// ask for user input just so they can see the console output
	input := bufio.NewScanner(os.Stdin)
    input.Scan()

	// create map and convert to json file
	outDict := map[string][]string{
		"name":         names,
		"address":      addresses,
		"description":  descriptions,
		"current_state": currentStates,
		"status":       statuses,
		"priority":     priorities,
	}

	outFile, err := os.Create("all_points.json")
	error_check(err)
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.Encode(outDict)}