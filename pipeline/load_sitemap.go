package pipeline

import (
	"context"
	"encoding/xml"
	"io"
	"log"
	"sync"

	"github.com/maksymhryb/gocachewarmer/client"
	"github.com/maksymhryb/gocachewarmer/config"
)

type SitemapIndex struct {
	XMLName xml.Name     `xml:"sitemapindex"`
	Urls    []SitemapUrl `xml:"sitemap"`
}

type UrlSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Urls    []SitemapUrl `xml:"url"`
}

type SitemapUrl struct {
	Url     string `xml:"loc"`
	Lastmod string `xml:"lastmod,omitempty"`
}

func LoadSitemap(ctx context.Context, config *config.Config) <-chan string {
	outputCh := make(chan string)

	go func() {
		defer close(outputCh)
		defer log.Println("[Load Sitemap Stage] stopped")

		log.Println("[Load Sitemap Stage] started")

		if config.SitemapUrl == "" {
			panic("sitemap url is not specified")
		}
		cl := client.NewClient()
		var response, err = cl.Get(config.SitemapUrl)

		if err != nil {
			panic("Cannot load sitemap")
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			panic("Cannot load sitemap")
		}

		fromUrlSetToChannel := func(urlSet UrlSet) {
			for _, url := range urlSet.Urls {
				select {
				case outputCh <- url.Url:
				case <-ctx.Done():
					return
				}
			}
		}

		var sitemapIndex SitemapIndex
		var urlSet UrlSet
		err = xml.Unmarshal(body, &sitemapIndex)
		if err != nil {
			err = xml.Unmarshal(body, &urlSet)
			if err != nil {
				panic("Cannot load sitemap")
			}

			fromUrlSetToChannel(urlSet)
			return
		}

		startWorker := func(wg *sync.WaitGroup, url string) {
			defer wg.Done()
			response, err := cl.Get(url)
			if err != nil {
				panic("Cannot load sitemap")
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				panic("Cannot read sitemap")
			}
			var urlSet UrlSet
			err = xml.Unmarshal(body, &urlSet)
			if err != nil {
				panic("Cannot unmarshal sitemap")
			}
			fromUrlSetToChannel(urlSet)
		}

		var wg sync.WaitGroup
		for _, url := range sitemapIndex.Urls {
			wg.Add(1)
			go startWorker(&wg, url.Url)
		}
		wg.Wait()
	}()

	return outputCh
}
