package faststats

import (
	"bufio"
	"fmt"
	"io"

	"sync"

	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert/netstorage"
	"github.com/VictoriaMetrics/VictoriaMetrics/app/vminsert/relabel"

	"github.com/VictoriaMetrics/VictoriaMetrics/lib/auth"
	"github.com/VictoriaMetrics/VictoriaMetrics/lib/bytesutil"
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

func ParseStream(reader io.Reader, callback func(data generated.Data, metricInfo []MetricInfo, at *auth.Token) error) error {

	ctx := getStreamContext()
	ctx.ic.Reset()
	ctx.reader = bufio.NewReaderSize(reader, 64*1024)
	ctx.at = ctx.ic.GetLocalAuthToken(nil)
	ctx.at.Set(1, 0) // FIXME: implement an auth message
	ctx.callback = callback
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
	readCalls.Inc()

	// blocking IO is used throughout so framing may be discarded
	if _, ctx.err = ctx.reader.Discard(framingSize); ctx.err != nil {
		return false
	}

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

	if ctx.Error() != nil {
		readErrors.Inc()
		ctx.err = fmt.Errorf("cannot read Time Machine protocol data: %w", ctx.err)
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
}

type streamContext struct {
	reader     *bufio.Reader
	sbe        generated.SbeGoMarshaller
	header     generated.MessageHeader
	definition generated.TimeSeriesDefinition
	err        error
	at         *auth.Token
	callback   func(data generated.Data, metricInfo []MetricInfo, at *auth.Token) error

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
	ctx.at = nil
	ctx.appendOnlyMetricInfo = nil
	ctx.ic.Reset()
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

	metricInfo.MetricNameRaw = storage.MarshalMetricNameRaw(nil, ctx.at.AccountID, ctx.at.ProjectID, ic.Labels)
	// Assignment of a storage node during registration intentionally does not account for unavailable storage nodes because rerouting occurs later when a row is pushed inside the storage layer
	metricInfo.StorageNodeIdx = ic.GetStorageNodeIdxRaw(metricInfo.MetricNameRaw)
	return nil
}

var (
	readCalls  = metrics.NewCounter(`vm_protoparser_read_calls_total{type="FastStats"}`)
	readErrors = metrics.NewCounter(`vm_protoparser_read_errors_total{type="FastStats"}`)
	rowsRead   = metrics.NewCounter(`vm_protoparser_rows_read_total{type="FastStats"}`)
)

func getStreamContext() *streamContext {
	return streamContextPool.Get().(*streamContext)
}

func putStreamContext(ctx *streamContext) {
	ctx.reset()
	streamContextPool.Put(ctx)
}

var streamContextPool = sync.Pool{
	New: func() interface{} {
		ctx := streamContext{
			sbe: *generated.NewSbeGoMarshaller(),
		}
		ctx.reset()
		return &ctx
	},
}

type unmarshalWork struct {
	ctx *streamContext

	data               generated.Data
	readOnlyMetricInfo []MetricInfo
}

func (uw *unmarshalWork) reset() {
	uw.ctx = nil
	uw.data.Points = uw.data.Points[:0]
	uw.readOnlyMetricInfo = nil
}

func (uw *unmarshalWork) runCallback() {
	ctx := uw.ctx
	if err := ctx.callback(uw.data, uw.readOnlyMetricInfo, ctx.at); err != nil {
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
	return unmarshalWorkPool.Get().(*unmarshalWork)
}

func putUnmarshalWork(uw *unmarshalWork) {
	uw.reset()
	unmarshalWorkPool.Put(uw)
}

var unmarshalWorkPool = sync.Pool{
	New: func() interface{} { return new(unmarshalWork) },
}
