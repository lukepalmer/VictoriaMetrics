package faststats

import (
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert/netstorage"
	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert/relabel"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/auth"
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

func run(t *testing.T, filename string, expectedAcks []int) {
	relabel.Init()
	netstorage.Init([]string{"host1", "host2", "host3"}, 0)
	common.StartUnmarshalWorkers()
	defer common.StopUnmarshalWorkers()

	expectedData := []string{
		`0 / AccountID=1, ProjectID=0, one{French="un",Spanish="uno"} (Timestamp=1, Value=1.000000)`,
		`1 / AccountID=1, ProjectID=0, two{French="deux",Spanish="dos"} (Timestamp=2, Value=2.000000)`,
		`0 / AccountID=1, ProjectID=0, one{French="un",Spanish="uno"} (Timestamp=3, Value=1.100000)`,
		`0 / AccountID=1, ProjectID=0, one{French="un",Spanish="uno"} (Timestamp=4, Value=1.110000)`,
		`1 / AccountID=1, ProjectID=0, two{French="deux",Spanish="dos"} (Timestamp=5, Value=2.200000)`,
	}

	rows := make(chan rowInfo, 10)
	callback := func(data generated.Data, metricInfos []MetricInfo, at *auth.Token) error {
		for _, point := range data.Points {
			row := rowInfo{
				metricRow:      storage.MetricRow{MetricNameRaw: metricInfos[point.TimeseriesId].MetricNameRaw, Timestamp: DivRoundClosest(point.TimeEpochNs, 1000000), Value: point.Value},
				storageNodeIdx: metricInfos[point.TimeseriesId].StorageNodeIdx}
			rows <- row
		}
		return nil
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	server, client := net.Pipe()

	go func() {
		client.Write(data)
	}()
	acks := make(chan int, 10)
	go func() {
		sbe := generated.NewSbeGoMarshaller()
		header := generated.SbeGoMessageHeader{}
		ack := generated.Acknowledgement{}
		for {
			io.CopyN(io.Discard, client, 2) // discard framing
			header.Decode(sbe, client)
			ack.Decode(sbe, client, ack.SbeSchemaVersion(), ack.SbeBlockLength(), true)
			acks <- int(ack.SequenceNumber)
		}
	}()
	var serverWait sync.WaitGroup
	serverWait.Add(1)
	go func() {
		streamErr := ParseStream(server, callback)
		if streamErr != nil {
			t.Error(streamErr)
		}
		serverWait.Done()
	}()

	// sort for deterministic test results because unmarshalling is concurrent
	sortedRows := rowsByTime{}
	for range expectedData {
		sortedRows = append(sortedRows, <-rows)
	}
	sort.Sort(sortedRows)
	sortedAcks := []int{}
	for range expectedAcks {
		sortedAcks = append(sortedAcks, <-acks)
	}
	sort.Ints(sortedAcks)
	if !reflect.DeepEqual(expectedAcks, sortedAcks) {
		t.Errorf("Unexpected acks: expected=%v, received=%v", expectedAcks, sortedAcks)
	}

	for i, row := range sortedRows {
		rowStr := row.String()
		if expectedData[i] != rowStr {
			t.Errorf("Unexpected in i=%v: %v", i, rowStr)
		}
	}

	client.Close()
	serverWait.Wait()
}

func TestParseStream(t *testing.T) {
	run(t, "test_data_v0.bin", []int{})
	run(t, "test_data_v1.bin", []int{0, 1})
}
