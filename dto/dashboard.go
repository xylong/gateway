package dto

type PanelOutput struct {
	ServiceNum      int64 `json:"service_num"`
	AppNum          int64 `json:"app_num"`
	CurrentQPS      int64 `json:"current_qps"`
	TodayRequestNum int64 `json:"today_request_num"`
}

type DashServiceStatItemOutput struct {
	Name     string `json:"name"`
	Value    int64  `json:"value"`
	LoadType int    `json:"load_type"`
}

type DashServiceStatOutput struct {
	Legend []string                    `json:"legend"`
	Data   []DashServiceStatItemOutput `json:"data"`
}
