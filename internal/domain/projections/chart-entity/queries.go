package chartview

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/chart"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
)

func QueryCharts(query projection.PaginationQuery, proj project.Project) []projection.PaginationResult[chart.Chart] {
	charts := utils.Values(proj.Charts)
	return projection.Paginate(charts, query)
}

func QueryChart(query projection.IDQuery, proj project.Project) *chart.Chart {
	return projection.GetFromMap(proj.Charts, query.ID)
}
