package faststats

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert/netstorage"
	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert/relabel"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/protoparser/common"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/protoparser/faststats/generated"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/storage"
)

type rowInfo struct {
	metricRow      storage.MetricRow
	storageNodeIdx int
}

func (r rowInfo) String() string {
	return fmt.Sprintf("%d / %s", r.storageNodeIdx, r.metricRow.String())
}

type rowsByTime []rowInfo

func (rows rowsByTime) Len() int { return len(rows) }
func (rows rowsByTime) Swap(i, j int) {
	rows[i], rows[j] = rows[j], rows[i]
}
func (rows rowsByTime) Less(i, j int) bool {
	return rows[i].metricRow.Timestamp < rows[j].metricRow.Timestamp
}

func TestParseStream(t *testing.T) {
	relabel.Init()
	netstorage.Init([]string{"host1", "host2", "host3"}, 0)
	common.StartUnmarshalWorkers()
	defer common.StopUnmarshalWorkers()

	file, err := os.Open("test_data.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	expected := []string{
		`2 / AccountID=0, ProjectID=0, one{French="un",Spanish="uno"} (Timestamp=1, Value=1.000000)`,
		`1 / AccountID=0, ProjectID=0, two{French="deux",Spanish="dos"} (Timestamp=2, Value=2.000000)`,
		`2 / AccountID=0, ProjectID=0, one{French="un",Spanish="uno"} (Timestamp=3, Value=1.100000)`,
		`2 / AccountID=0, ProjectID=0, one{French="un",Spanish="uno"} (Timestamp=4, Value=1.110000)`,
		`1 / AccountID=0, ProjectID=0, two{French="deux",Spanish="dos"} (Timestamp=5, Value=2.200000)`,
	}

	rows := rowsByTime{}
	callback := func(data generated.Data, metricInfos []MetricInfo) error {
		for _, point := range data.Points {
			row := rowInfo{
				metricRow:      storage.MetricRow{MetricNameRaw: metricInfos[point.TimeseriesId].MetricNameRaw, Timestamp: DivRoundClosest(point.TimeEpochNs, 1000000), Value: point.Value},
				storageNodeIdx: metricInfos[point.TimeseriesId].StorageNodeIdx}
			rows = append(rows, row)
		}
		return nil
	}

	err = ParseStream(nil, reader, callback)
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
