package faststats

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"sync"

	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert/netstorage"
	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert/relabel"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/auth"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/bytesutil"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/cgroup"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/protoparser/common"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/protoparser/faststats/generated"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/storage"
	"github.com/VictoriaMetrics/metrics"
)

// Integer division with rounding to the nearest integer (instead of truncating)
func DivRoundClosest(n int64, divisor int64) int64 {
	return ((n + (divisor / 2)) / divisor)
}

var (
	dataSbeTemplateID uint16 = (&generated.Data{}).SbeTemplateId() // this is really a constant
)

const (
	doRangeCheck  bool   = false
	actingVersion uint16 = 0
	framingSize          = 2
)

func ParseStream(at *auth.Token, reader io.Reader, callback func(data generated.Data, metricInfo []MetricInfo) error) error {

	ctx := getStreamContext()
	ctx.reader = bufio.NewReaderSize(reader, 64*1024)
	ctx.atLocal = ctx.ic.GetLocalAuthToken(at)
	ctx.callback = callback
	ctx.ic.Reset()
	defer putStreamContext(ctx)

	for ctx.Read() {
		// all work is in Read()
	}

	ctx.wg.Wait()
	if err := ctx.Error(); err != nil {
		return err
	}
	return ctx.callbackErr
}

func (ctx *streamContext) Read() bool {
	// temporary hack to work with both framed and non-framed protocols
	var bytes []byte
	if bytes, ctx.err = ctx.reader.Peek(framingSize); ctx.err != nil {
		return false
	}
	// this could either be message framing or it could be the SBE blocklength, which we know to be zero
	if binary.LittleEndian.Uint16(bytes) != 0 {
		print("framing", binary.LittleEndian.Uint16(bytes))
		// blocking IO is used throughout so framing may be discarded
		if _, ctx.err = ctx.reader.Discard(framingSize); ctx.err != nil {
			return false
		}
	}

	readCalls.Inc()
	if ctx.err = ctx.header.Decode(&ctx.sbe, ctx.reader, actingVersion); ctx.err != nil {
		return false
	}

	if ctx.header.TemplateId == ctx.definition.SbeTemplateId() {
		if ctx.err = ctx.definition.Decode(&ctx.sbe, ctx.reader, actingVersion, ctx.header.BlockLength, doRangeCheck); ctx.err != nil {
			return false
		}
		if ctx.err = ctx.registerTimeSeries(); ctx.err != nil {
			return false
		}
	} else if ctx.header.TemplateId == dataSbeTemplateID {
		uw := getUnmarshalWork()
		uw.ctx = ctx

		/* Safety rules for sharing metric names:
		- the stream may only write elements once
		- unmarshall work may only read elements that existed at the time this slice is assigned

		Since the array backing appendOnlyMetricInfo is only be appended to, the contents of the slice visible right now is always safe for unserialized concurrent reads.
		*/
		uw.readOnlyMetricInfo = ctx.appendOnlyMetricInfo
		// Decoding must happen within the stream context because we are doing (fast) direct buffer reads.
		if ctx.err = uw.data.Decode(&ctx.sbe, ctx.reader, actingVersion, ctx.header.BlockLength, doRangeCheck); ctx.err != nil {
			return false
		}
		ctx.wg.Add(1)
		common.ScheduleUnmarshalWork(uw)
	}

	if ctx.err != nil {
		if ctx.err != io.EOF {
			readErrors.Inc()
			ctx.err = fmt.Errorf("cannot read Time Machine protocol data: %w", ctx.err)
		}
		return false
	}
	if ctx.hasCallbackError() {
		return false
	}

	return true
}

type MetricInfo struct {
	MetricNameRaw  []byte
	StorageNodeIdx int
	AtLocal        *auth.Token
}

type streamContext struct {
	reader     *bufio.Reader
	sbe        generated.SbeGoMarshaller
	header     generated.MessageHeader
	definition generated.TimeSeriesDefinition
	err        error
	atLocal    *auth.Token
	callback   func(data generated.Data, metricInfo []MetricInfo) error

	/* The backing array may only be appended to.
	This allows concurrent reads from a slice pointed to populated elements without synchronization. */
	appendOnlyMetricInfo []MetricInfo

	/* We will not insert data using this context, but need access to the name and label handling within.
	This dependency could be cleaned up with refactoring. */
	ic netstorage.InsertCtx

	wg              sync.WaitGroup
	callbackErrLock sync.Mutex
	callbackErr     error
}

func (ctx *streamContext) Error() error {
	if ctx.err == io.EOF {
		return nil
	}
	return ctx.err
}

func (ctx *streamContext) hasCallbackError() bool {
	ctx.callbackErrLock.Lock()
	ok := ctx.callbackErr != nil
	ctx.callbackErrLock.Unlock()
	return ok
}

func (ctx *streamContext) reset() {
	ctx.reader = nil
	ctx.callback = nil
	ctx.err = nil
	ctx.callbackErr = nil
	ctx.atLocal = nil
	ctx.appendOnlyMetricInfo = nil
}

func (ctx *streamContext) registerTimeSeries() error {
	id := int(ctx.definition.Id)

	/* A slice of the array backing appendOnlyMetricNamesRaw will be shared for unsynchronized read-only access in another goroutine.
	This is helpful for performance and allowable so long as elements that will be read (visible in the read-only slice) are never written after they are shared.

	A straightforward way to accomplish this is to expect clients implementing this protocol to assign metrics to ids sequentially.
	New definitions may then be appended without synchronization and without altering the apparent contents of a previously shared read-only slice.
	*/
	metricInfoLen := len(ctx.appendOnlyMetricInfo)
	if id < metricInfoLen {
		return fmt.Errorf("client illegally registered a non-sequential id of %d; expected %d", id, metricInfoLen)
	}
	ctx.appendOnlyMetricInfo = bytesutil.ResizeWithCopyMayOverallocate(ctx.appendOnlyMetricInfo, id+1)
	metricInfo := &ctx.appendOnlyMetricInfo[id]

	ic := &ctx.ic
	ic.Labels = ic.Labels[:0]
	ic.AddLabelBytes(nil, ctx.definition.Name[:])
	for _, label := range ctx.definition.Labels {
		ic.AddLabelBytes(label.Key[:], label.Value[:])
	}
	if relabel.HasRelabeling() {
		ic.ApplyRelabeling()
	}
	ic.SortLabelsIfNeeded()

	metricInfo.MetricNameRaw = storage.MarshalMetricNameRaw(nil, ctx.atLocal.AccountID, ctx.atLocal.ProjectID, ic.Labels)
	// Assignment of a storage node during registration intentionally does not account for unavailable storage nodes because rerouting occurs later when a row is pushed inside the storage layer
	metricInfo.StorageNodeIdx = ic.GetStorageNodeIdxRaw(metricInfo.MetricNameRaw)
	metricInfo.AtLocal = ctx.atLocal
	return nil
}

var (
	readCalls  = metrics.NewCounter(`vm_protoparser_read_calls_total{type="FastStats"}`)
	readErrors = metrics.NewCounter(`vm_protoparser_read_errors_total{type="FastStats"}`)
	rowsRead   = metrics.NewCounter(`vm_protoparser_rows_read_total{type="FastStats"}`)
)

func getStreamContext() *streamContext {
	select {
	case ctx := <-streamContextPoolCh:
		return ctx
	default:
		if v := streamContextPool.Get(); v != nil {
			ctx := v.(*streamContext)

			return ctx
		}
		return &streamContext{
			sbe: *generated.NewSbeGoMarshaller(),
		}
	}
}

func putStreamContext(ctx *streamContext) {
	ctx.reset()
	select {
	case streamContextPoolCh <- ctx:
	default:
		streamContextPool.Put(ctx)
	}
}

var streamContextPool sync.Pool
var streamContextPoolCh = make(chan *streamContext, cgroup.AvailableCPUs())

type unmarshalWork struct {
	ctx *streamContext

	data               generated.Data
	readOnlyMetricInfo []MetricInfo
}

func (uw *unmarshalWork) runCallback() {
	ctx := uw.ctx
	if err := ctx.callback(uw.data, uw.readOnlyMetricInfo); err != nil {
		ctx.callbackErrLock.Lock()
		if ctx.callbackErr == nil {
			ctx.callbackErr = fmt.Errorf("error when processing imported data: %w", err)
		}
		ctx.callbackErrLock.Unlock()
	}
	ctx.wg.Done()
}

// Unmarshal implements common.UnmarshalWork
func (uw *unmarshalWork) Unmarshal() {

	nPoints := len(uw.data.Points)
	rowsRead.Add(nPoints)
	uw.runCallback()
	putUnmarshalWork(uw)
}

func getUnmarshalWork() *unmarshalWork {
	v := unmarshalWorkPool.Get()
	if v == nil {
		return &unmarshalWork{}
	}
	return v.(*unmarshalWork)
}

func putUnmarshalWork(uw *unmarshalWork) {
	unmarshalWorkPool.Put(uw)
}

var unmarshalWorkPool sync.Pool
