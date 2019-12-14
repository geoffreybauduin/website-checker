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

	if os.Getenv("DEBUG") == "1" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		exit(err)
		return
	}

	c := checker.New()
	for _, url := range *urls {
		c.AddURL(url)
	}
	if ignore != nil {
		for _, ig := range *ignore {
			rex, err := regexp.Compile(ig)
			if err != nil {
				exit(err)
				return
			}
			c.Ignore(rex)
		}
	}
	if noExternalInspection != nil && *noExternalInspection {
		c.NoExternalInspection()
	}
	c.Workers(*workers)

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
