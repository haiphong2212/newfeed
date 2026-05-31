package platform

import "github.com/elastic/go-elasticsearch/v8"

func NewElasticsearch(url string) (*elasticsearch.Client, error) {
	return elasticsearch.NewClient(elasticsearch.Config{Addresses: []string{url}})
}
