# website-checker

[![Go Report Card](https://goreportcard.com/badge/github.com/geoffreybauduin/website-checker)](https://goreportcard.com/report/github.com/geoffreybauduin/website-checker)

Performs multiple checks against your website, mostly:

- Goes through every `img`, `script`, `a`, `link[rel="stylesheet"]` tags, and stores the availability of the resource
- Stores a map of the dependencies between each resource
- Can perform validation of structured data using [Yandex Structured Data Validator](https://tech.yandex.com/validator/doc/dg/concepts/about-docpage/)

## Installation

```
go get -u github.com/geoffreybauduin/website-checker
go install github.com/geoffreybauduin/website-checker/cmd/website-checker
```

## Usage

```
usage: website-checker --urls=URLS [<flags>]

Checks 404 and other stuff by crawling your website

Flags:
  --help                         Show context-sensitive help (also try --help-long and --help-man).
  --workers=10                   Number of workers to perform the work
  --urls=URLS ...                URLs to check
  --ignore-urls=IGNORE-URLS ...  Ignore those URLs and do not attempt to fetch them. Expecting a regexp
  --no-external-inspection       Do not inspect external urls
  --check-structured-data=CHECK-STRUCTURED-DATA  
                                 Check structured data validity
  --yandex-api-key=YANDEX-API-KEY  
                                 Yandex API Key
```

### Explained examples

```
website-checker --urls http://localhost:1313 --no-external-inspection --ignore-urls "^https://docs\.google\.com" --workers=5 --check-structured-data=yandex --yandex-api-key=1234
```

- Will crawl the website located at http://localhost:1313 and all its dependencies.
- Will not fetch the dependencies from the pages that are outside of the host `localhost:1313`.
- Any url starting with `https://docs.google.com` will be ignored.
- Will perform 5 tasks in parallel
- Will fetch any `application/ld+json` script tags and perform validation of those against Yandex API, using the api key `1234`

## Contributing

See [CONTRIBUTING.md](https://github.com/geoffreybauduin/website-checker/blob/master/CONTRIBUTING.md)

## License

MIT License

Copyright (c) 2019 Geoffrey Bauduin

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
