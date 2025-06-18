package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/chai2010/webp"
	"github.com/scrapeless-ai/sdk-go/scrapeless"
	"github.com/scrapeless-ai/sdk-go/scrapeless/actor"
	"github.com/scrapeless-ai/sdk-go/scrapeless/log"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/scraping"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/storage/dataset"
	"github.com/scrapeless-ai/sdk-go/scrapeless/services/storage/kv"
	"image/png"
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
	k := client.Storage.Kv
	kvId, _, cErr := k.CreateNamespace(ctx, "scraper.google.search")
	if cErr != nil {
		log.Warnf("kv--> create namespace failed: %v", cErr)
	}
	d := client.Storage.Dataset
	datasetId, _, err := d.CreateDataset(ctx, "scraper.google.search")
	if err != nil {
		log.Warnf("dataset--> create d failed: %v", err)
	}

	for i := 0; i < params.Limit; i++ {
		times := i + 1
		datasetSave(d, err, ctx, scrape, datasetId, times)
		kvSave(k, ctx, scrape, kvId, times)
		log.Infof("times--> %d", times)
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
	pngBytes, err := DownloadWebpAsPngBytes("https://banner2.cleanpng.com/20180408/vae/avgpocfjw.webp")
	if err != nil {
		log.Warnf("download image failed: %v", err)
	}
	value, sErr := object.Put(ctx, id, "demo.png", pngBytes)
	if sErr != nil {
		log.Warnf("object--> save object failed: %v", sErr)
	}
	log.Infof("object--> save object success, object: %v", value)
}

func kvSave(kvs kv.KV, ctx context.Context, scrape []byte, id string, times int) {
	str := string(scrape)
	var kks []kv.BulkItem
	for i := 0; i < 100; i++ {
		kks = append(kks, kv.BulkItem{
			Key:        fmt.Sprintf("%d-%d", times, i),
			Value:      str,
			Expiration: 3600,
		})
	}
	value, sErr := kvs.BulkSetValue(ctx, id, kks)
	if sErr != nil {
		log.Warnf("kv--> save kv failed: %v", sErr)
	}
	log.Infof("kv--> success count: %d", value)
}

func scrapingCrawl(client *scrapeless.Client, ctx context.Context, params map[string]interface{}) ([]byte, error) {
	scrape, err := client.Scraping.Scrape(ctx, scraping.ScrapingTaskRequest{
		Actor:        "scraper.google.search",
		Input:        params,
		ProxyCountry: "US",
	})
	if err != nil {
		log.Warnf("crawl--> scraping google.search failed: %v", err)
	}
	return scrape, err
}

func datasetSave(dataset dataset.Dataset, err error, ctx context.Context, scrape []byte, id string, times int) {
	var maps []map[string]any
	for i := 0; i < 100; i++ {
		maps = append(maps, map[string]any{"title": "scraper.google.search", "content": string(scrape), "times": times})
	}
	items, err := dataset.AddItems(ctx, id, maps)
	if err != nil {
		log.Warnf("dataset--> save dataset failed: %v", err)
	}
	if !items {
		log.Infof("dataset--> save dataset failed, isErr: %v", items)
	}
}

func DownloadWebpAsPngBytes(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// 下载 .webp 图片
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	// 解码 WebP -> image.Image
	img, err := webp.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("webp decode error: %w", err)
	}

	// 编码为 PNG 并写入 buffer
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("png encode error: %w", err)
	}

	return buf.Bytes(), nil
}
