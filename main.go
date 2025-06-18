package main

import (
	"context"
	"github.com/scrapeless-ai/sdk-go/scrapeless"
	"github.com/scrapeless-ai/sdk-go/scrapeless/actor"
	"github.com/scrapeless-ai/sdk-go/scrapeless/log"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/scraping"
)

func main() {
	ctx := context.Background()
	a := actor.New()
	var params map[string]interface{}
	inputErr := a.Input(params)
	if inputErr != nil {
		log.Errorf("parse input err: %v", inputErr)
	}
	client := scrapeless.New(scrapeless.WithScraping())
	scrape, err := client.Scraping.Scrape(ctx, scraping.ScrapingTaskRequest{
		Actor:        "google.search",
		Input:        params,
		ProxyCountry: "US",
	})
	if err != nil {
		log.Warnf("scraping google.search failed: %v", err)
	}
	dataset := client.Storage.Dataset
	id, _, err := dataset.CreateDataset(ctx, "google.search")
	if err != nil {
		log.Warnf("create dataset failed: %v", err)
	}
	items, err := dataset.AddItems(ctx, id, []map[string]any{
		{"title": "Top news headlines", "content": scrape},
	})
	if err != nil {
		log.Warnf("save dataset failed: %v", err)
	}
	if items {
		log.Infof("save dataset success")
	}
}
