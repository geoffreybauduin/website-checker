package checker

import (
	"context"
	"net/url"

	validator "github.com/geoffreybauduin/yandex-structured-data-validator"
)

type structureddatatask struct {
	txt  string
	from *url.URL
}

func (task *structureddatatask) Execute(c *checker) (*result, error) {
	c.result.Lock()
	res := c.result.Checked[task.from.String()]
	c.result.Unlock()
	resp := ResultStructuredData{
		Content: task.txt,
	}
	if c.options.StructuredDataCheckWithYandex {
		yandex := validator.New(c.options.YandexAPIKey)
		check, err := yandex.CheckDocument(context.TODO(), `<script type="application/ld+json">`+task.txt+`</script>`)
		if err != nil {
			return nil, err
		}
		resp.Yandex = &check
	}

	res.Lock()
	if res.StructuredData == nil {
		res.StructuredData = make([]ResultStructuredData, 0)
	}
	res.StructuredData = append(res.StructuredData, resp)
	res.Unlock()
	return res, nil
}
