package proto

import "encoding/xml"

// JP Eshop software list https://www.nintendo.co.jp/data/software/xml/switch.xml
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

func ParseJPGameTitleInfoList(data []byte) (*JPGameTitleInfoList, error) {
	var list JPGameTitleInfoList
	if err := xml.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	return &list, nil
}
