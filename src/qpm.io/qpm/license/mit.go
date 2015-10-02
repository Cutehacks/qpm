// Copyright 2015 Cutehacks AS. All rights reserved.
// License can be found in the LICENSE file.

package license

import (
	"strings"
	"strconv"
	"time"
	"net/http"
	"encoding/json"
	"qpm.io/common"
)

type License struct {
	Key       string
	Name      string
	Permitted []string
	Forbidden []string
	Body      string
}

// Fetches a license info object from the Github license API.
func GetLicense(identifier string, pkg *common.PackageWrapper) (License, error) {

	var info License
	url := "https://api.github.com/licenses/" + identifier
	client := &http.Client{}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return info, err
	}
	request.Header.Set("Accept", "application/vnd.github.drax-preview+json")

	var response *http.Response
	response, err = client.Do(request)
	if err != nil {
		return info, err
	}
	defer response.Body.Close()

	dec := json.NewDecoder(response.Body)
	err = dec.Decode(&info);
	if err != nil {
		return info, err
	}

	// FIXME: this is probably tied to the MIT license layout
	info.Body = strings.Replace(info.Body, "[year]", strconv.Itoa(time.Now().Year()), -1)
	info.Body = strings.Replace(info.Body, "[fullname]", pkg.Author.Name, -1)

	return info, err
}