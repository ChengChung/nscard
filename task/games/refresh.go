package games

import (
	"strconv"

	"github.com/chengchung/nscard/client/nintendo/games"
	"github.com/chengchung/nscard/db"
	"github.com/chengchung/nscard/task"
	"github.com/sirupsen/logrus"
)

func init() {
	task.RegisterTaskGenerator(&RefreshJPTitlesXMLToDBTaskGenerator{})
	task.RegisterTaskGenerator(&RefreshJPTitlesSearchToDBTaskGenerator{})
}

type RefreshJPTitlesXMLToDBTaskGenerator struct{}

func (g *RefreshJPTitlesXMLToDBTaskGenerator) New(cfg task.TaskConfig) task.TaskI {
	return &RefreshJPTitlesXMLToDBTask{
		name: cfg.TaskName,
	}
}

func (g *RefreshJPTitlesXMLToDBTaskGenerator) TaskType() string {
	return "RefreshJPTitlesXMLToDB"
}

type RefreshJPTitlesXMLToDBTask struct {
	name string
}

func (t *RefreshJPTitlesXMLToDBTask) TaskName() string {
	return t.name
}

func (t *RefreshJPTitlesXMLToDBTask) Run() error {
	logrus.Info("refreshing JP game XML data...")

	games, err := games.GetTitleListXML()
	if err != nil {
		return err
	}

	games_insert := make([]db.JPStoreGameTitle, 0, len(games))
	for _, game := range games {
		games_insert = append(games_insert, db.JPStoreGameTitle{
			InitialCode:      game.InitialCode,
			TitleName:        game.TitleName,
			MakerName:        game.MakerName,
			MakerKana:        game.MakerKana,
			Price:            game.Price,
			SalesDate:        game.SalesDate,
			SoftType:         game.SoftType,
			PlatformID:       game.PlatformID,
			DlIconFlg:        strconv.FormatInt(int64(game.DlIconFlg), 10),
			LinkURL:          game.LinkURL,
			ScreenshotImgFlg: strconv.FormatInt(int64(game.ScreenshotImgFlg), 10),
			ScreenshotImgURL: game.ScreenshotImgURL,
		})
	}

	err = db.BatchSaveJPStoreGameTitle(games_insert)
	if err != nil {
		return err
	}

	logrus.Infof("successfully refreshed JP game search data, total %d games", len(games_insert))

	return nil
}

type RefreshJPTitlesSearchToDBTaskGenerator struct{}

func (g *RefreshJPTitlesSearchToDBTaskGenerator) New(cfg task.TaskConfig) task.TaskI {
	return &RefreshJPTitlesSearchToDBTask{}
}

func (g *RefreshJPTitlesSearchToDBTaskGenerator) TaskType() string {
	return "RefreshJPTitlesSearchToDB"
}

type RefreshJPTitlesSearchToDBTask struct {
	name     string
	maxRetry int

	_fail_cnt int
}

func (t *RefreshJPTitlesSearchToDBTask) TaskName() string {
	return t.name
}

func (t *RefreshJPTitlesSearchToDBTask) Run() error {
	logrus.Info("refreshing JP game search data...")

	swtitles, err := games.GetTitleList(games.OptHardSwitch)
	if err != nil {
		return err
	}

	sw2titles, err := games.GetTitleList(games.OptHardSwitch2)
	if err != nil {
		return err
	}

	titles := make([]db.JPStoreGameTitleSearch, 0, len(swtitles)+len(sw2titles))
	fn := func(titles_search []games.Item) {
		for _, title := range titles_search {
			titles = append(titles, db.JPStoreGameTitleSearch{
				ID:     title.ID,
				Title:  title.Title,
				URL:    title.URL,
				TitleK: title.TitleK,
				NSUID:  title.NSUID,
				Hard:   string(title.Hard),
				IURL:   title.IURL,
				SIURL:  title.SIURL,
			})
		}
	}
	fn(swtitles)
	fn(sw2titles)

	err = db.BatchSaveJPStoreGameTitleSearch(titles)
	if err != nil {
		return err
	}

	logrus.Infof("successfully refreshed JP game XML data, total %d games", len(titles))

	return nil
}
