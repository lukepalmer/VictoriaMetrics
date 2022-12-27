package faststats

import (
	"io"

	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert/common"
	parser "github.com/VictoriaMetrics/VictoriaMetrics/lib/protoparser/faststats"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/protoparser/faststats/generated"
	"github.com/VictoriaMetrics/metrics"
)

var (
	rowsInserted  = metrics.NewCounter(`vm_rows_inserted_total{type="FastStats"}`)
	rowsPerInsert = metrics.NewHistogram(`vm_rows_per_insert{type="FastStats"}`)
)

func InsertHandler(r io.Reader) error {
	return parser.ParseStream(r, insertRows)
}

func insertRows(data generated.Data, metricNamesRaw [][]byte) error {
	ic := common.GetInsertCtx()
	defer common.PutInsertCtx(ic)

	rowsTotal := len(data.Points)
	ic.Reset(rowsTotal)

	for _, point := range data.Points {
		if _, err := ic.WriteDataPointExt(metricNamesRaw[point.TimeseriesId], nil,
			parser.DivRoundClosest(point.TimeEpochNs, 1000000), point.Value); err != nil {
			return err
		}
	}
	rowsInserted.Add(rowsTotal)
	rowsPerInsert.Update(float64(rowsTotal))
	return ic.FlushBufs()
}
