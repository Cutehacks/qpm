// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package core

import (
	"fmt"
	msg "qpm.io/common/messages"
	"strings"
	"net/http"
	"encoding/json"
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

type License struct {
	Key       string
	Name      string
	Permitted []string
	Forbidden []string
	Body      string
}

// Fetches a license info object from the Github license API.
func GetLicense(license string) (License, error) {
	var licenseInfo License
	url := "https://api.github.com/licenses/" + license
	client := &http.Client{}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return licenseInfo, err
	}
	request.Header.Set("Accept", "application/vnd.github.drax-preview+json")

	var response *http.Response
	response, err = client.Do(request)
	if err != nil {
		return licenseInfo, err
	}
	defer response.Body.Close()

	dec := json.NewDecoder(response.Body)
	err = dec.Decode(&licenseInfo);
	return licenseInfo, err
}
