package control

const ScoutFreeboard = "scout-freeboard.json"

type fbCell struct {
    Three int `json:"3"`
    Four  int `json:"4"`
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
    Units     string `json:"units"`
}

type fbWidget struct {
    Type     string           `json:"type"`
    Settings fbWidgetSettings `json:"settings"`
}

type fbPane struct {
    Width       int
    Row         fbCell
    Col         fbCell
    Col_width   int
    Widgets   []fbWidget
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
        Refresh: 2,
        Method: "GET"}
}

func NewDashboard() *fbModel {
    return &fbModel{
        Version: 1,
        Allow_edit: false,
        Plugins: nil,
        Panes: nil,
        Datasources: []fbDataSource{*newDataSource()},
        Columns: 3}
}
