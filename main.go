package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/scrapeless-ai/sdk-go/scrapeless"
	"github.com/scrapeless-ai/sdk-go/scrapeless/actor"
	"github.com/scrapeless-ai/sdk-go/scrapeless/log"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/scraping"
	"io"
	"net/http"
	"time"
)

type P struct {
	Limit        int    `json:"limit"`
	Q            string `json:"q"`
	Hl           string `json:"hl"`
	Gl           string `json:"gl"`
	GoogleDomain string `json:"google_domain"`
}

func main() {
	ctx := context.Background()
	a := actor.New()
	var params = &P{}
	inputErr := a.Input(params)
	if inputErr != nil {
		log.Errorf("parse input err: %v", inputErr)
	}
	result := toMapParams(params)

	client := scrapeless.New(scrapeless.WithScraping(), scrapeless.WithStorage(), scrapeless.WithUniversal())

	scrape, err := scrapingCrawl(client, ctx, result)

	for i := 0; i < params.Limit; i++ {
		datasetSave(client, err, ctx, scrape)
		kvSave(client, ctx, scrape)
	}
	objectSave(client, ctx)
}

//func scrapingUniversal(client *scrapeless.Client, ctx context.Context) {
//	maps := map[string]interface{}{
//		"url":       "https://www.scrapeless.com",
//		"method":    "GET",
//		"redirect":  true,
//		"js_render": true,
//		"js_instructions": [{"wait":100}],
//		"block":{"resources":["image", "font", "script"], "urls":["https://example.com"]}
//	}
//	scrape, err := client.Universal.Scrape(ctx, universal.UniversalTaskRequest{
//		Actor:        universal.ScraperUniversal,
//		Input:        maps,
//		ProxyCountry: "US",
//	})
//	if err != nil {
//		log.Warnf("scraping universal failed: %v", err)
//	}
//
//}

func toMapParams(params *P) map[string]interface{} {
	// 先编码为 JSON
	data, err := json.Marshal(*params)
	if err != nil {
		log.Errorf("json.Marshal err: %v", err)
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Errorf("json.Unmarshal err: %v", err)
	}
	return result
}

func objectSave(client *scrapeless.Client, ctx context.Context) {
	object := client.Storage.Object
	id, _, err := object.CreateBucket(ctx, "scraper.google.search", "scraper.google.search")
	if err != nil {
		log.Warnf("create bucket failed: %v", err)
	}
	bytes, err := DownloadImageAsBytes("https://banner2.cleanpng.com/20180408/vae/avgpocfjw.webp")
	if err != nil {
		log.Warnf("download image failed: %v", err)
	}
	value, sErr := object.Put(ctx, id, "demo.webp", bytes)
	if sErr != nil {
		log.Warnf("save value failed: %v", sErr)
	}
	log.Infof("save value success, object: %v", value)
}

func kvSave(client *scrapeless.Client, ctx context.Context, scrape []byte) {
	kv := client.Storage.Kv
	id, _, cErr := kv.CreateNamespace(ctx, "scraper.google.search")
	if cErr != nil {
		log.Warnf("create namespace failed: %v", cErr)
	}
	value, sErr := kv.SetValue(ctx, id, "scraper.google.search", string(scrape), 3600)
	if sErr != nil {
		log.Warnf("save value failed: %v", sErr)
	}
	if value {
		log.Infof("save value success")
	}
}

func scrapingCrawl(client *scrapeless.Client, ctx context.Context, params map[string]interface{}) ([]byte, error) {
	scrape, err := client.Scraping.Scrape(ctx, scraping.ScrapingTaskRequest{
		Actor:        "scraper.google.search",
		Input:        params,
		ProxyCountry: "US",
	})
	if err != nil {
		log.Warnf("scraping google.search failed: %v", err)
	}
	return scrape, err
}

func datasetSave(client *scrapeless.Client, err error, ctx context.Context, scrape []byte) {
	dataset := client.Storage.Dataset
	id, _, err := dataset.CreateDataset(ctx, "scraper.google.search")
	if err != nil {
		log.Warnf("create dataset failed: %v", err)
	}
	items, err := dataset.AddItems(ctx, id, []map[string]any{
		{"title": "scraper.google.search", "content": string(scrape)},
	})
	if err != nil {
		log.Warnf("save dataset failed: %v", err)
	}
	if items {
		log.Infof("save dataset success")
	}
}

func DownloadImageAsBytes(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %w", err)
	}

	return imgBytes, nil
}
