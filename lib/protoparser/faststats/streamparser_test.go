package faststats

import (
	"bufio"
	"os"
	"sort"
	"testing"

	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert/relabel"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/protoparser/common"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/protoparser/faststats/generated"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/storage"
)

type rowsByTime []storage.MetricRow

func (rows rowsByTime) Len() int { return len(rows) }
func (rows rowsByTime) Swap(i, j int) {
	rows[i], rows[j] = rows[j], rows[i]
}
func (rows rowsByTime) Less(i, j int) bool {
	return rows[i].Timestamp < rows[j].Timestamp
}

func TestParseStream(t *testing.T) {
	relabel.Init()
	common.StartUnmarshalWorkers()
	defer common.StopUnmarshalWorkers()

	file, err := os.Open("test_data.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	expected := []string{
		`one{French="un",Spanish="uno"} (Timestamp=1, Value=1.000000)`,
		`two{French="deux",Spanish="dos"} (Timestamp=2, Value=2.000000)`,
		`one{French="un",Spanish="uno"} (Timestamp=3, Value=1.100000)`,
		`one{French="un",Spanish="uno"} (Timestamp=4, Value=1.110000)`,
		`two{French="deux",Spanish="dos"} (Timestamp=5, Value=2.200000)`,
	}

	rows := rowsByTime{}
	callback := func(data generated.Data, metricNamesRaw [][]byte) error {
		for _, point := range data.Points {
			row := storage.MetricRow{MetricNameRaw: metricNamesRaw[point.TimeseriesId], Timestamp: DivRoundClosest(point.TimeEpochNs, 1000000), Value: point.Value}
			rows = append(rows, row)
		}
		return nil
	}

	err = ParseStream(reader, callback)
	if err != nil {
		t.Error(err)
	}
	// for deterministic test results because unmarshalling is concurrent
	sort.Sort(rows)

	for i, row := range rows {
		rowStr := row.String()
		if expected[i] != rowStr {
			t.Errorf("Unexpected in i=%v: %v", i, rowStr)
		}
	}
}
