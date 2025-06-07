package games

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

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

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func GetTitleList(hardware OptHard) ([]Item, error) {
	link, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	limit := 400
	page := 1
	total_page := 0

	query := link.Query()
	query.Set("opt_hard[]", string(hardware))
	query.Set("sort", "sodate asc,titlek asc,score")
	query.Set("fq", "!(sform_s:DLC) AND !(sform_s:hard)")

	results := make([]Item, 0)
	init := true
	for {
		query.Set("limit", strconv.FormatInt(int64(limit), 10))
		query.Set("page", strconv.FormatInt(int64(page), 10))
		link.RawQuery = query.Encode()

		req, err := http.NewRequest("GET", link.String(), nil)
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
				fmt.Println("Error reading response body:", err)
			} else {
				fmt.Println("Error response:", string(bytes))
			}

			return results, err
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
			fmt.Printf("Total items: %d, Total pages: %d, Items per page: %d\n", searchResult.Result.Total, total_page, limit)
		}

		results = append(results, searchResult.Result.Items...)
		if limit != searchResult.Query.Limit {
			err := errors.New("limit changed during pagination")
			return results, err
		}

		//	we allow one more pages to be queryed
		if page > total_page ||
			len(searchResult.Result.Items) < limit {
			fmt.Printf("Total pages: %d, Current page: %d, Items on this page: %d\n", total_page, page, len(searchResult.Result.Items))
			break
		}

		fmt.Printf("Total pages: %d, Current page: %d, Items on this page: %d\n", total_page, page, len(searchResult.Result.Items))

		page++
	}

	return results, nil
}
