package card

import (
	"bytes"
	"text/template"
)

const (
	svg_MAIN = `
  <svg width="622" height="206" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg">
    <style type="text/css">
      {{.SVG_ANIMATION_STYLE}}
      {{.SVG_CARD_FRAME_CSS_STYLE}}
	  {{.SVG_DYNAMIC_BG_CSS_STYLE}}
    </style>
    <foreignObject id="frame1" width="622" height="206">
      <div class="card card-bg" xmlns="http://www.w3.org/1999/xhtml">
        <div class="user-header">
	      <div class="avatar" />
          <div class="user-info">
            <div class="normal-row">
              <div class="flag" />
              <div class="name">{{.NickName}}</div>
            </div>
            <div class="normal-row">
              <div class="name">{{.SwFriendCode}}</div>
              <div class="ns-logo" />
            </div>
          </div>
        </div>
        <div class="stats">
          <div class="stat">
            <div class="stat-value">{{.GamesTotal}}</div>
            <div class="stat-key"><div>Games</div><div>Owned</div></div>
          </div>
          <div class="stat">
            <div class="stat-value">{{.PlayTimeTotal}}</div>
            <div class="stat-key"><div>Total</div><div>PlayTime</div></div>
          </div>
          <div class="stat">
            <div class="stat-value">{{.GamesThisMonth}}</div>
            <div class="stat-key"><div>Titles</div><div>ThisMonth</div></div>
          </div>
          <div class="stat">
            <div class="stat-value">{{.GamesThisYear}}</div>
            <div class="stat-key"><div>Titles</div><div>ThisYear</div></div>
          </div>
        </div>
      </div>
    </foreignObject>
    <foreignObject id="frame2" width="622" height="206">
      <div class="card card-bg" xmlns="http://www.w3.org/1999/xhtml">
      <div class="user-header">
	      <div class="avatar" height="40" width="40" />
          <div class="user-info">
            <div class="normal-row">
              <div class="flag" />
              <div class="name">{{.NickName}}</div>
            </div>
            <div class="normal-row">
              <div class="name">🎮 {{.GamesTotal}}</div>
              <div class="name">⏰ {{.PlayTimeTotal}}</div>
              <div class="ns-logo" />
            </div>
          </div>
        </div>
        <div class="games">
        {{range .FirstNGames}}
          <div class="game-group">
            <img class="game" height="75" width="75" src="{{.ImageData}}" />
            <div class="playtime">{{.PlayTime}}</div>
          </div>
		{{end}}
        </div>
      </div>
    </foreignObject>
  </svg>
`
)

var t = template.New("root")

func init() {
	t = template.Must(t.New("card").Parse(svg_MAIN))
	t = template.Must(t.New("card_dyn_css").Parse(css_DYNAMIC_BG))
}

func RenderSVG(info CardInfo) (string, error) {
	type data struct {
		SVG_ANIMATION_STYLE      string
		SVG_CARD_FRAME_CSS_STYLE string
		SVG_DYNAMIC_BG_CSS_STYLE string
		NS_LOGO_BASE64           string
		CardInfo
	}
	d := data{
		SVG_ANIMATION_STYLE:      css_ANIMATION,
		SVG_CARD_FRAME_CSS_STYLE: css_CARD,
		NS_LOGO_BASE64:           NS_LOGO_BASE64,
		CardInfo:                 info,
	}

	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf, "card_dyn_css", d)
	if err != nil {
		return "", err
	}
	d.SVG_DYNAMIC_BG_CSS_STYLE = buf.String()

	buf.Reset()
	err = t.ExecuteTemplate(&buf, "card", d)

	return buf.String(), err
}
