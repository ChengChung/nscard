package games

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

type SearchResult struct {
	Status int    `json:"status"`
	Query  Query  `json:"query"`
	Result Result `json:"result"`
}

type Query struct {
	Limit int `json:"limit"`
	Page  int `json:"page"`
}

type Result struct {
	Total int    `json:"total"`
	Items []Item `json:"items"`
}

type OptHard string

const (
	//	it may not be correct except for switch and switch2
	OptHardSwitch     OptHard = "1_HAC"
	OptHard3DS        OptHard = "2_CTR"
	OptHardOther      OptHard = "9_other"
	OptHardWiiU       OptHard = "4_WUP"
	OptHardAmiibo     OptHard = "9_amiibo"
	OptHardSwitch2    OptHard = "05_BEE"
	OptHardSmartphone OptHard = "3_smartphone"
)

type Item struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	URL    string  `json:"url"`
	TitleK string  `json:"titlek"`
	NSUID  string  `json:"nsuid"`
	Hard   OptHard `json:"hard"`
	IURL   string  `json:"iurl"`
	SIURL  string  `json:"siurl"`
}

const baseURL = "https://search.nintendo.jp/nintendo_soft/search.json"

var concurrencyCh = make(chan struct{}, 10)

func GetTitleList(hardware OptHard) ([]Item, error) {
	limit := 400
	page := 1
	total_page := 0

	finalResult := make([]Item, 0)
	init := true

	handleReqFn := func(ctx context.Context, page int) ([]Item, error) {
		concurrencyCh <- struct{}{}
		defer func() { <-concurrencyCh }()

		link, err := url.Parse(baseURL)
		if err != nil {
			panic(err)
		}

		query := link.Query()
		query.Set("opt_hard[]", string(hardware))
		query.Set("sort", "sodate asc,titlek asc,score")
		query.Set("fq", "!(sform_s:DLC) AND !(sform_s:hard)")
		query.Set("limit", strconv.FormatInt(int64(limit), 10))
		query.Set("page", strconv.FormatInt(int64(page), 10))
		link.RawQuery = query.Encode()

		results := make([]Item, 0)

		req, err := http.NewRequestWithContext(ctx, "GET", link.String(), nil)
		if err != nil {
			return results, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return results, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			bytes, err := io.ReadAll(resp.Body) // consume the body to avoid resource leak
			if err != nil {
				logrus.Errorf("Error reading response body: %v", err)
			} else {
				logrus.Errorf("Error response: %s", string(bytes))
			}

			return results, errors.New(resp.Status)
		}

		var searchResult SearchResult
		err = json.NewDecoder(resp.Body).Decode(&searchResult)
		if err != nil {
			return results, err
		}

		if init {
			init = false
			limit = searchResult.Query.Limit
			total_page = (searchResult.Result.Total + limit - 1) / limit
			logrus.Infof("Total items: %d, Total pages: %d, Items per page: %d\n", searchResult.Result.Total, total_page, limit)
		}

		results = append(results, searchResult.Result.Items...)
		if limit != searchResult.Query.Limit {
			err := errors.New("limit changed during pagination")
			return results, err
		}

		logrus.Infof("Total pages: %d, Current page: %d, Items on this page: %d", total_page, page, len(searchResult.Result.Items))

		return results, nil
	}

	httpctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if items, err := handleReqFn(httpctx, page); err != nil {
		logrus.Errorf("Error fetching first page: %v", err)
		return nil, err
	} else {
		finalResult = append(finalResult, items...)
	}

	resChannel := make(chan *struct {
		item []Item
		err  error
	}, total_page)

	wg := sync.WaitGroup{}
	// we allow one more to page to be queryed to ensure we get all items
	for i := 2; i <= total_page+1; i++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			items, err := handleReqFn(httpctx, page)
			if err != nil && !errors.Is(err, context.Canceled) {
				cancel()
			}
			resChannel <- &struct {
				item []Item
				err  error
			}{item: items, err: err}
		}(i)
	}

	go func() {
		wg.Wait()
		close(resChannel)
	}()

	var err error
	for res := range resChannel {
		logrus.Infof("Received results for page, items: %d, error: %v", len(res.item), res.err)
		if res.err == nil {
			finalResult = append(finalResult, res.item...)
		}
	}

	return finalResult, err
}
