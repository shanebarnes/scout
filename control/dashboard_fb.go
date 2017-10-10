package control

import (
    "strconv"
)

const ScoutFreeboard = "scout-freeboard.json"

type fbCell struct {
    Three int `json:"3"`
//    Four  int `json:"4"`
}

type fbDataSourceSettings struct {
    Url            string `json:"url"`
    Use_thingproxy bool   `json:"use_thingproxy"`
    Refresh        int    `json:"refresh"`
    Method         string `json:"method"`
}

type fbDataSource struct {
    Name string                   `json:"name"`
    Type string                   `json:"type"`
    Settings fbDataSourceSettings `json:"settings"`
}

type fbWidgetSettings struct {
    Title     string `json:"title"`
    Size      string `json:"size"`
    Value     string `json:"value"`
    Sparkline bool   `json:"sparkline"`
    Animate   bool   `json:"animate"`
//    Units     string `json:"units"`
}

type fbWidget struct {
    Type     string           `json:"type"`
    Settings fbWidgetSettings `json:"settings"`
}

type fbPane struct {
    Title       string   `json:"title"`
    Width       int      `json:"width"`
    Row         fbCell   `json:"row"`
    Col         fbCell   `json:"col"`
    Col_width   int      `json:"col_width"`
    Widgets   []fbWidget `json:"widgets"`
}

type fbModel struct {
    Version       int          `json:"version"`
    Allow_edit    bool         `json:"allow_edit"`
    Plugins     []string       `json:"plugins"`
    Panes       []fbPane       `json:"panes"`
    Datasources []fbDataSource `json:"datasources"`
    Columns       int          `json:"columns"`
}

func newDataSource() *fbDataSource {
    return &fbDataSource{
        Name: "reports",
        Type: "JSON",
        Settings: *newDataSourceSettings()}
}

func newDataSourceSettings() *fbDataSourceSettings {
    return &fbDataSourceSettings{
        Url: "http://localhost:8080/reports",
        Use_thingproxy: true,
        Refresh: 5,
        Method: "GET"}
}

func newPane(title string, width, col, row int, widgets *[]fbWidget) *fbPane {
    return &fbPane{
        Title: title,
        Width: width,
        Row: fbCell{Three: row},
        Col: fbCell{Three: col},
        Col_width: width,
        Widgets: *widgets}
}

func newWidget(title, value string) *fbWidget {
    return &fbWidget{
        Type: "text_widget",
        Settings: *newWidgetSettings(title, value)}
}

func newWidgetSettings(title, value string) *fbWidgetSettings {
    return &fbWidgetSettings{
        Title: title,
        Size: "regular",
        Value: value,
        Sparkline: true,
        Animate: true}
}

func NewDashboard(reports *[][]Database) *fbModel {
    var panes []fbPane

    for i := range *reports {
        var widgets []fbWidget
        var target string = ""

        for j := range (*reports)[i] {
            target = (*reports)[i][j].Target
            value := "datasources[\"reports\"][\"" + strconv.Itoa(i) + "\"][" + strconv.Itoa(j) + "][\"dpN\"][\"y\"]"
            widgets = append(widgets, *newWidget((*reports)[i][j].Task, value))
        }

        panes = append(panes, *newPane(target, 1, i + 1, 1, &widgets))
    }

    return &fbModel{
        Version: 1,
        Allow_edit: false,
        Plugins: nil,
        Panes: panes,
        Datasources: []fbDataSource{*newDataSource()},
        Columns: len(*reports)}
}
