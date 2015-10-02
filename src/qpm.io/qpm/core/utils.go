// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package core

import (
	"fmt"
	msg "qpm.io/common/messages"
	"strings"
	"text/template"
	"os"
)

// Pretty prints a table of package SearchResults.
func PrintSearchResults(results []*msg.SearchResult) {
	if len(results) == 0 {
		fmt.Printf("No packages found.\n")
	} else {
		fmt.Printf("\n%-40s %-20s\n", "Package", "Author")
		fmt.Printf("%s\n", strings.Repeat("-", 75))
	}

	for _, r := range results {
		fmt.Printf("%-40s %s\n",
			r.Name+"@"+r.Version,
			r.GetAuthor().Name+" <"+r.GetAuthor().Email+">",
		)
	}
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

