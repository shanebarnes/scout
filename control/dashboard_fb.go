package control

import (
    "strconv"

    "github.com/shanebarnes/goto/logger"
)

const ScoutFreeboard = "scout-freeboard.json"

type fbCell struct {
    Three int `json:"3"`
    //Four  int `json:"4"`
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

type fbWidgetTextSettings struct {
    Title     string `json:"title"`
    Size      string `json:"size"`
    Value     string `json:"value"`
    Sparkline bool   `json:"sparkline"`
    Animate   bool   `json:"animate"`
    Units     string `json:"units"`
}

type fbWidgetIndicatorSettings struct {
    Title   string `json:"title"`
    Value   string `json:"value"`
    OnText  string `json:"on_text"`
    OffText string `json:"off_text"`
}

type fbWidgetGaugeSettings struct {
    Title string `json:"title"`
    Value string `json:"value"`
    Units string `json:"units"`
    MinValue int `json:"min_value"`
    MaxValue int `json:"max_value"`
}

type fbWidget struct {
    Type     string      `json:"type"`
    Settings interface{} `json:"settings"`
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

func newWidgetIndicator(title, value string) *fbWidget {
    return &fbWidget{
        Type: "indicator",
        Settings: *newWidgetIndicatorSettings(title, value, "Running", "Stopped")}
}

func newWidgetIndicatorSettings(title, value, onText, offText string) *fbWidgetIndicatorSettings {
    return &fbWidgetIndicatorSettings{
        Title: title,
        Value: value,
        OnText: onText,
        OffText: offText}
}

func newWidgetGauge(title, value, units string) *fbWidget {
    return &fbWidget{
        Type: "gauge",
        Settings: *newWidgetGaugeSettings(title, value, units, 0, 100)}
}

func newWidgetGaugeSettings(title, value, units string, minVal, maxVal int) *fbWidgetGaugeSettings {
    return &fbWidgetGaugeSettings{
        Title: title,
        Value: value,
        Units: units,
        MinValue: minVal,
        MaxValue: maxVal}
}

func newWidgetSparkline(title, value, units string) *fbWidget {
    return &fbWidget{
        Type: "text_widget",
        Settings: *newWidgetTextSettings(title, value, units, true)}
}

func newWidgetText(title, value, units string) *fbWidget {
    return &fbWidget{
        Type: "text_widget",
        Settings: *newWidgetTextSettings(title, value, units, false)}
}

func newWidgetTextSettings(title, value, units string, sparkline bool) *fbWidgetTextSettings {
    return &fbWidgetTextSettings{
        Title: title,
        Size: "regular",
        Value: value,
        Sparkline: sparkline,
        Animate: true,
        Units: units}
}

func NewDashboard(reports *[][]Database) *fbModel {
    var panes []fbPane

    for i := range *reports {
        var target string = ""
        var widgets []fbWidget

        widgets = append(widgets, *newWidgetText("sample count", "datasources[\"reports\"][\"0\"][\"0\"][\"N\"]", ""))

        for j := range (*reports)[i] {
            db := (*reports)[i][j]
            target = db.Target

            for k := range db.Reports {
                value := ""
                xform := db.Reports[k].Xform

                switch db.Reports[k].Type {
                case "DIFF":
                    value = "datasources[\"reports\"][\"" + strconv.Itoa(i) + "\"][" + strconv.Itoa(j) + "][\"diff\"]"
                case "RATE":
                    value = "datasources[\"reports\"][\"" + strconv.Itoa(i) + "\"][" + strconv.Itoa(j) + "][\"rate\"]"
                case "RAW":
                    value = "datasources[\"reports\"][\"" + strconv.Itoa(i) + "\"][" + strconv.Itoa(j) + "][\"dpN\"][\"y\"]"
                default:
                }

                taskName := db.Task + " #" + db.Reports[k].Type
                switch db.Reports[k].Widget {
                case "gauge":
                    widgets = append(widgets, *newWidgetGauge(taskName, "(" + value + xform + ").toFixed(3)", db.Reports[k].Units))
                case "indicator":
                    widgets = append(widgets, *newWidgetIndicator(taskName, "(" + value + xform + ").toFixed(3)"))
                case "sparkline":
                    widgets = append(widgets, *newWidgetSparkline(taskName, "(" + value + xform + ").toFixed(3)", db.Reports[k].Units))
                case "text":
                    widgets = append(widgets, *newWidgetText(taskName, "(" + value + xform + ").toFixed(3)", db.Reports[k].Units))
                case "":
                    logger.PrintlnDebug("No dashboard widget specified for...")
                default:
                    logger.PrintlnError("Unknown dashboard widget: '" + db.Reports[k].Widget + "'")
                }
            }
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
