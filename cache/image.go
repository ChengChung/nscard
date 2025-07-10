package cache

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chengchung/nscard/db"
	"github.com/sirupsen/logrus"
)

type ImageRequest struct {
	TitleName string
	ImageUrl  string
}

type ImageRequestResponse struct {
	TitleName string
	ImageUrl  string
	Data      []byte
}

func GetOrCreateImageCache(requests []ImageRequest) ([]ImageRequestResponse, error) {
	urls := make([]string, len(requests))
	for i, req := range requests {
		urls[i] = req.ImageUrl
	}

	caches, err := db.GetGameTitleImgCache(urls)
	if err != nil {
		logrus.Errorf("failed to get image cache: %v", err)
	}

	url_img_map := make(map[string]string, len(caches))
	for _, cache := range caches {
		url_img_map[cache.ImgUrl] = cache.Data
	}

	results := make([]ImageRequestResponse, 0, len(requests))
	missing_requests := make([]ImageRequest, 0, len(requests)-len(caches))
	for _, req := range requests {
		if data, ok := url_img_map[req.ImageUrl]; ok {
			bytes, err := base64.StdEncoding.DecodeString(data)
			if err != nil {
				logrus.Errorf("failed to decode image data for %s: %v", req.ImageUrl, err)
				missing_requests = append(missing_requests, req)
				continue
			}
			results = append(results, ImageRequestResponse{
				TitleName: req.TitleName,
				ImageUrl:  req.ImageUrl,
				Data:      bytes,
			})
		} else {
			missing_requests = append(missing_requests, req)
		}
	}

	missing_urls := make([]string, len(missing_requests))
	for i, req := range missing_requests {
		missing_urls[i] = req.ImageUrl
	}
	images_fetched, err := batchGetImages(missing_urls)

	cache_to_save := make([]db.GameTitleImgCache, 0, len(images_fetched))
	for _, req := range missing_requests {
		if data, ok := images_fetched[req.ImageUrl]; ok {
			results = append(results, ImageRequestResponse{
				TitleName: req.TitleName,
				ImageUrl:  req.ImageUrl,
				Data:      data,
			})
			cache_to_save = append(cache_to_save, db.GameTitleImgCache{
				TitleName: req.TitleName,
				ImgUrl:    req.ImageUrl,
				Data:      base64.StdEncoding.EncodeToString(data),
			})
		}
	}

	if len(cache_to_save) > 0 {
		err := db.BatchSetGameTitleImageCache(cache_to_save)
		if err != nil {
			logrus.Errorf("failed to save image cache: %v", err)
		}
	}

	return results, err
}

var client = &http.Client{
	Timeout: 10 * time.Second,
}

var concurrencyCh = make(chan struct{}, 10)

func batchGetImages(urls []string) (map[string][]byte, error) {
	results := make(map[string][]byte, len(urls))

	type res struct {
		url  string
		data []byte
	}
	resCh := make(chan *res, len(urls))
	wg := &sync.WaitGroup{}

	var containsError atomic.Bool

	for _, url := range urls {
		concurrencyCh <- struct{}{}
		wg.Add(1)

		go func(imgUrl string) {
			defer func() { <-concurrencyCh }()
			defer wg.Done()

			resp, err := client.Get(imgUrl)
			if err != nil {
				logrus.Errorf("failed to fetch image from %s: %v", imgUrl, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				containsError.Store(true)
				logrus.Errorf("failed to fetch image from %s: status code %d", imgUrl, resp.StatusCode)
				return
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				containsError.Store(true)
				logrus.Errorf("failed to read image data from %s: %v", imgUrl, err)
				return
			}

			resCh <- &res{
				url:  imgUrl,
				data: data,
			}
		}(url)
	}

	go func() {
		wg.Wait()
		close(resCh)
	}()

	for res := range resCh {
		if res.data != nil {
			results[res.url] = res.data
		}
	}

	var err error
	if containsError.Load() {
		err = errors.New("some images failed to fetch")
	}

	return results, err
}
