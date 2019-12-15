package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/geoffreybauduin/website-checker/checker"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	app := kingpin.New("website-checker", "Checks 404 and other stuff by crawling your website")
	workers := app.Flag("workers", "Number of workers to perform the work").Default("10").Int()
	urls := app.Flag("urls", "URLs to check").Required().Strings()
	ignore := app.Flag("ignore-urls", "Ignore those URLs and do not attempt to fetch them. Expecting a regexp").Strings()
	noExternalInspection := app.Flag("no-external-inspection", "Do not inspect external urls").Bool()
	checkStructuredData := app.Flag("check-structured-data", "Check structured data validity").Enum("yandex")
	yandexAPIKey := app.Flag("yandex-api-key", "Yandex API Key").String()

	if os.Getenv("DEBUG") == "1" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		exit(err)
		return
	}

	opts := &checker.Options{
		URLs:            make([]string, 0),
		IgnoreRegexps:   make([]*regexp.Regexp, 0),
		InspectExternal: true,
		Workers:         *workers,
	}
	if noExternalInspection != nil && *noExternalInspection {
		opts.InspectExternal = false
	}
	for _, url := range *urls {
		opts.URLs = append(opts.URLs, url)
	}
	if ignore != nil {
		for _, ig := range *ignore {
			rex, err := regexp.Compile(ig)
			if err != nil {
				exit(err)
				return
			}
			opts.IgnoreRegexps = append(opts.IgnoreRegexps, rex)
		}
	}
	if checkStructuredData != nil {
		switch *checkStructuredData {
		case "yandex":
			opts.StructuredDataCheckWithYandex = true
			if yandexAPIKey == nil {
				exit(fmt.Errorf("missing yandex-api-key parameter"))
			}
			opts.YandexAPIKey = *yandexAPIKey
		}
	}

	c := checker.New(opts)
	res, err := c.Run()
	if err != nil {
		exit(err)
		return
	}
	resp, errJSON := json.MarshalIndent(res, "", "    ")
	if errJSON != nil {
		exit(errJSON)
		return
	}
	fmt.Printf("%s\n", resp)
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	os.Exit(1)
}
