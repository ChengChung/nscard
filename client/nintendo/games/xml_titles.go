package games

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	xml_url = "https://www.nintendo.co.jp/data/software/xml/switch.xml"
)

type JPGameTitleInfoList struct {
	XMLName   xml.Name    `xml:"TitleInfoList"`
	TitleInfo []TitleInfo `xml:"TitleInfo"`
}

type TitleInfo struct {
	InitialCode      string `xml:"InitialCode"`
	TitleName        string `xml:"TitleName"`
	MakerName        string `xml:"MakerName"`
	MakerKana        string `xml:"MakerKana"`
	Price            string `xml:"Price"`
	SalesDate        string `xml:"SalesDate"`
	SoftType         string `xml:"SoftType"`
	PlatformID       string `xml:"PlatformID"`
	DlIconFlg        int    `xml:"DlIconFlg"`
	LinkURL          string `xml:"LinkURL"`
	ScreenshotImgFlg int    `xml:"ScreenshotImgFlg"`
	ScreenshotImgURL string `xml:"ScreenshotImgURL"`
}

func GetTitleListXML() ([]TitleInfo, error) {
	resp, err := client.Get(xml_url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logrus.Errorf("Error reading response body: %v", err)
			return nil, err
		} else {
			logrus.Errorf("Error response: %s", string(bytes))
		}
		return nil, errors.New(resp.Status)
	}

	var list JPGameTitleInfoList
	if err = xml.NewDecoder(resp.Body).Decode(&list); err != nil {
		logrus.Errorf("Error decoding XML: %v", err)
		return nil, err
	}

	return list.TitleInfo, nil
}
