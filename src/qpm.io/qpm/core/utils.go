// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package core

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	msg "qpm.io/common/messages"
	"strings"
	"text/template"
)

const (
	columnName        = "Name"
	columnDescription = "Description"
	columnAuthor      = "Author"
	columnVersion     = "Version"
	columnLicense     = "License"
)

func IntMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func IntMin(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// Pretty prints a table of package SearchResults.
func PrintSearchResults(results []*msg.SearchResult) {
	columnWidths := map[string]int{
		columnName:        len(columnName),
		columnDescription: len(columnDescription),
		columnAuthor:      len(columnAuthor),
		columnVersion:     len(columnVersion),
		columnLicense:     len(columnLicense),
	}

	if len(results) == 0 {
		fmt.Printf("No packages found.\n")
		return
	} else if len(results) < 1000 {
		// pre-process the list to get column widths
		for _, r := range results {
			columnWidths[columnName] = IntMin(IntMax(columnWidths[columnName], len(r.Name)), 40)
			columnWidths[columnDescription] = IntMin(IntMax(columnWidths[columnDescription], len(r.Description)), 80)
			columnWidths[columnAuthor] = IntMin(IntMax(columnWidths[columnAuthor], len(r.GetAuthor().Name)), 20)
			columnWidths[columnVersion] = IntMin(IntMax(columnWidths[columnVersion], len(r.Version)), 10)
			columnWidths[columnLicense] = IntMin(IntMax(columnWidths[columnLicense], len(msg.LicenseType_name[int32(r.License)])), 10)
		}
	} else {
		// Too many results to pre-process so use sensible defaults
		columnWidths[columnName] = 40
		columnWidths[columnDescription] = 60
		columnWidths[columnAuthor] = 20
		columnWidths[columnVersion] = 10
		columnWidths[columnLicense] = 10
	}

	width, _, err := terminal.GetSize(0)
	if err != nil {
		fmt.Printf("Couldn't get terminal width: %s\n", err.Error())
		// gracefully fallback to something sensible
		width = 110
	}

	const columnSpacing = 3

	fmt.Println("")
	columns := []string{columnName, columnDescription, columnAuthor, columnVersion, columnLicense}
	widths := make([]int, len(columns))
	for i, col := range columns {
		widths[i] = columnWidths[col]
	}

	// Print the headers
	printRow(width, columnSpacing, widths, columns)

	// Print a horizontal line
	fmt.Printf("%s\n", strings.Repeat("-", width))

	// Print the search results
	for _, r := range results {
		columns := []string{
			r.Name,
			r.Description,
			r.GetAuthor().Name,
			r.Version,
			msg.LicenseType_name[int32(r.License)],
		}
		printRow(width, columnSpacing, widths, columns)
	}
}

func printRow(screenWidth int, columnSpacing int, columnWidths []int, columns []string) {
	remaining := screenWidth
	for i, col := range columns {
		if remaining <= 0 {
			break
		}

		// convert to []rune since we want the char count not bytes
		runes := []rune(col)

		// truncate the string if we are out of space
		maxLength := IntMin(remaining, columnWidths[i])
		if len(runes) > maxLength {
			runes = runes[:maxLength]
		}

		fmt.Printf("%s", string(runes))
		w := columnWidths[i]
		remaining -= len(runes)
		toNextCol := IntMax(IntMin(w-len(runes)+columnSpacing, remaining), columnSpacing)
		fmt.Printf("%s", strings.Repeat(" ", toNextCol))
		remaining -= toNextCol
	}
	fmt.Printf("\n")
}

// Renders the given template using the fields contained in the data parameter and
// outputs the result to the given file. If the file already exists, it will be
// overwritten.
func WriteTemplate(filename string, tpl *template.Template, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tpl.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}
