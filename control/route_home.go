package control

import (
	"html/template"
	"net/http"
	"time"

	"github.com/shanebarnes/goto/logger"
	"github.com/shanebarnes/scout/mission"
)

type Link struct {
	Link string
	Text string
}

type HomePageData struct {
	Version string
	Links   []Link
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	defer logger.PrintlnTime(logger.Info, time.Time{}, "Prepared response to "+r.Method+" "+r.URL.Path)()
	tmpl := template.Must(template.ParseFiles("control/html_home.tmpl"))

	data := HomePageData{
		Version: mission.GetVersion(),
		Links: []Link{
			{Link: "/reports?pretty", Text: "reports"},
			{Link: "/statistics?pretty", Text: "statistics"},
			{Link: "/tasks?pretty", Text: "tasks"},
		},
	}

	tmpl.Execute(w, data)
}
