package chart

type ChartType string

const (
	ChartTypeBar     ChartType = "bar"
	ChartTypeLine    ChartType = "line"
	ChartTypeArea    ChartType = "area"
	ChartTypePie     ChartType = "pie"
	ChartTypeHeatmap ChartType = "heatmap"
	ChartTypeTable   ChartType = "table"
)

type Orientation string

const (
	OrientationVertical   Orientation = "vertical"
	OrientationHorizontal Orientation = "horizontal"
)

type ChartSpec struct {
	Type        ChartType   `json:"type" validate:"required,oneof=bar line area pie heatmap table"`
	X           string      `json:"x" validate:"required"`
	Y           string      `json:"y" validate:"required"`
	Color       string      `json:"color,omitempty"`
	Series      string      `json:"series,omitempty"`
	Orientation Orientation `json:"orientation,omitempty" validate:"omitempty,oneof=vertical horizontal"`
	Stacked     bool        `json:"stacked,omitempty"`
	ShowLegend  bool        `json:"show_legend,omitempty"`
}

type Chart struct {
	ID      string `json:"id" validate:"required"`
	Healthy bool   `json:"healthy"`
	Version int    `json:"version"`
	ChartData
}

type ChartData struct {
	ProjectID   string    `json:"project_id" validate:"required"`
	Name        string    `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string    `json:"description" validate:"max=2000" normalize:"trim"`
	Query       string    `json:"query" validate:"required"`
	Spec        ChartSpec `json:"spec" validate:"required"`
	Pinned      bool      `json:"pinned"`
}

type UpdateChartData struct {
	Name        string     `json:"name,omitempty" validate:"omitempty,max=200" normalize:"trim"`
	Description string     `json:"description,omitempty" validate:"omitempty,max=2000" normalize:"trim"`
	Query       string     `json:"query,omitempty"`
	Spec        *ChartSpec `json:"spec,omitempty"`
}

func (c Chart) GetID() string {
	return c.ID
}

func (c Chart) GetProjectID() string {
	return c.ProjectID
}

func (c Chart) IsHealthy() bool {
	return c.Healthy
}

func (c Chart) WithUnhealthy() any {
	c.Healthy = false
	return &c
}

func (c Chart) GetVersion() int {
	return c.Version
}

func (c Chart) WithVersion(v int) any {
	c.Version = v
	return &c
}

func (c Chart) WithPinned(pinned bool) any {
	c.Pinned = pinned
	return &c
}
