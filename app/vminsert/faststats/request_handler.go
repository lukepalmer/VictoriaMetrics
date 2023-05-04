package faststats

import (
	"io"

	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert/netstorage"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/auth"
	parser "github.com/VictoriaMetrics/VictoriaMetrics/lib/protoparser/faststats"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/protoparser/faststats/generated"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/tenantmetrics"
	"github.com/VictoriaMetrics/metrics"
)

var (
	rowsInserted       = metrics.NewCounter(`vm_rows_inserted_total{type="FastStats"}`)
	rowsTenantInserted = tenantmetrics.NewCounterMap(`vm_tenant_inserted_rows_total{type="influx"}`)
	rowsPerInsert      = metrics.NewHistogram(`vm_rows_per_insert{type="FastStats"}`)
)

func InsertHandler(at *auth.Token, r io.Reader) error {
	return parser.ParseStream(at, r, insertRows)
}

func insertRows(data generated.Data, metricInfos []parser.MetricInfo) error {
	ic := netstorage.GetInsertCtx()
	defer netstorage.PutInsertCtx(ic)
	perTenantRows := make(map[auth.Token]int)
	ic.Reset()
	var metricInfo *parser.MetricInfo = nil
	rowsTotal := len(data.Points)

	for _, point := range data.Points {
		metricInfo = &metricInfos[point.TimeseriesId]
		if err := ic.WriteDataPointExt(metricInfo.StorageNodeIdx, metricInfo.MetricNameRaw, parser.DivRoundClosest(point.TimeEpochNs, 1000000), point.Value); err != nil {
			return err
		}
		perTenantRows[*metricInfo.AtLocal]++
	}
	rowsInserted.Add(rowsTotal)

	rowsTenantInserted.MultiAdd(perTenantRows)
	rowsPerInsert.Update(float64(rowsTotal))
	return ic.FlushBufs()
}
