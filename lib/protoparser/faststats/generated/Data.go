// Generated SBE (Simple Binary Encoding) message codec

package generated

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
)

type Data struct {
	Points []DataPoints
}
type DataPoints struct {
	TimeseriesId uint32
	TimeEpochNs  int64
	Value        float64
}

func (d *Data) Encode(_m *SbeGoMarshaller, _w io.Writer, doRangeCheck bool) error {
	if doRangeCheck {
		if err := d.RangeCheck(d.SbeSchemaVersion(), d.SbeSchemaVersion()); err != nil {
			return err
		}
	}
	var PointsNumInGroup uint16 = uint16(len(d.Points))
	if err := _m.WriteUint16(_w, PointsNumInGroup); err != nil {
		return err
	}
	var PointsBlockLength uint16 = 20
	if err := _m.WriteUint16(_w, PointsBlockLength); err != nil {
		return err
	}
	for _, prop := range d.Points {
		if err := prop.Encode(_m, _w); err != nil {
			return err
		}
	}
	return nil
}

func (d *Data) Decode(_m *SbeGoMarshaller, _r io.Reader, actingVersion uint16, blockLength uint16, doRangeCheck bool) error {
	if actingVersion > d.SbeSchemaVersion() && blockLength > d.SbeBlockLength() {
		io.CopyN(ioutil.Discard, _r, int64(blockLength-d.SbeBlockLength()))
	}

	if d.PointsInActingVersion(actingVersion) {
		var PointsNumInGroup uint16
		if err := _m.ReadUint16(_r, &PointsNumInGroup); err != nil {
			return err
		}
		var PointsBlockLength uint16
		if err := _m.ReadUint16(_r, &PointsBlockLength); err != nil {
			return err
		}
		if cap(d.Points) < int(PointsNumInGroup) {
			d.Points = make([]DataPoints, PointsNumInGroup)
		}
		d.Points = d.Points[:PointsNumInGroup]
		for i := range d.Points {
			if err := d.Points[i].Decode(_m, _r, actingVersion, uint(PointsBlockLength)); err != nil {
				return err
			}
		}
	}
	if doRangeCheck {
		if err := d.RangeCheck(actingVersion, d.SbeSchemaVersion()); err != nil {
			return err
		}
	}
	return nil
}

func (d *Data) RangeCheck(actingVersion uint16, schemaVersion uint16) error {
	for _, prop := range d.Points {
		if err := prop.RangeCheck(actingVersion, schemaVersion); err != nil {
			return err
		}
	}
	return nil
}

func DataInit(d *Data) {
	return
}

func (d *DataPoints) Encode(_m *SbeGoMarshaller, _w io.Writer) error {
	if err := _m.WriteUint32(_w, d.TimeseriesId); err != nil {
		return err
	}
	if err := _m.WriteInt64(_w, d.TimeEpochNs); err != nil {
		return err
	}
	if err := _m.WriteFloat64(_w, d.Value); err != nil {
		return err
	}
	return nil
}

func (d *DataPoints) Decode(_m *SbeGoMarshaller, _r io.Reader, actingVersion uint16, blockLength uint) error {
	if !d.TimeseriesIdInActingVersion(actingVersion) {
		d.TimeseriesId = d.TimeseriesIdNullValue()
	} else {
		if err := _m.ReadUint32(_r, &d.TimeseriesId); err != nil {
			return err
		}
	}
	if !d.TimeEpochNsInActingVersion(actingVersion) {
		d.TimeEpochNs = d.TimeEpochNsNullValue()
	} else {
		if err := _m.ReadInt64(_r, &d.TimeEpochNs); err != nil {
			return err
		}
	}
	if !d.ValueInActingVersion(actingVersion) {
		d.Value = d.ValueNullValue()
	} else {
		if err := _m.ReadFloat64(_r, &d.Value); err != nil {
			return err
		}
	}
	if actingVersion > d.SbeSchemaVersion() && blockLength > d.SbeBlockLength() {
		io.CopyN(ioutil.Discard, _r, int64(blockLength-d.SbeBlockLength()))
	}
	return nil
}

func (d *DataPoints) RangeCheck(actingVersion uint16, schemaVersion uint16) error {
	if d.TimeseriesIdInActingVersion(actingVersion) {
		if d.TimeseriesId < d.TimeseriesIdMinValue() || d.TimeseriesId > d.TimeseriesIdMaxValue() {
			return fmt.Errorf("Range check failed on d.TimeseriesId (%v < %v > %v)", d.TimeseriesIdMinValue(), d.TimeseriesId, d.TimeseriesIdMaxValue())
		}
	}
	if d.TimeEpochNsInActingVersion(actingVersion) {
		if d.TimeEpochNs < d.TimeEpochNsMinValue() || d.TimeEpochNs > d.TimeEpochNsMaxValue() {
			return fmt.Errorf("Range check failed on d.TimeEpochNs (%v < %v > %v)", d.TimeEpochNsMinValue(), d.TimeEpochNs, d.TimeEpochNsMaxValue())
		}
	}
	if d.ValueInActingVersion(actingVersion) {
		if d.Value < d.ValueMinValue() || d.Value > d.ValueMaxValue() {
			return fmt.Errorf("Range check failed on d.Value (%v < %v > %v)", d.ValueMinValue(), d.Value, d.ValueMaxValue())
		}
	}
	return nil
}

func DataPointsInit(d *DataPoints) {
	return
}

func (*Data) SbeBlockLength() (blockLength uint16) {
	return 0
}

func (*Data) SbeTemplateId() (templateId uint16) {
	return 1
}

func (*Data) SbeSchemaId() (schemaId uint16) {
	return 118
}

func (*Data) SbeSchemaVersion() (schemaVersion uint16) {
	return 0
}

func (*Data) SbeSemanticType() (semanticType []byte) {
	return []byte("")
}

func (*Data) SbeSemanticVersion() (semanticVersion string) {
	return ""
}

func (*DataPoints) TimeseriesIdId() uint16 {
	return 0
}

func (*DataPoints) TimeseriesIdSinceVersion() uint16 {
	return 0
}

func (d *DataPoints) TimeseriesIdInActingVersion(actingVersion uint16) bool {
	return actingVersion >= d.TimeseriesIdSinceVersion()
}

func (*DataPoints) TimeseriesIdDeprecated() uint16 {
	return 0
}

func (*DataPoints) TimeseriesIdMetaAttribute(meta int) string {
	switch meta {
	case 1:
		return ""
	case 2:
		return ""
	case 3:
		return ""
	case 4:
		return "required"
	}
	return ""
}

func (*DataPoints) TimeseriesIdMinValue() uint32 {
	return 0
}

func (*DataPoints) TimeseriesIdMaxValue() uint32 {
	return math.MaxUint32 - 1
}

func (*DataPoints) TimeseriesIdNullValue() uint32 {
	return math.MaxUint32
}

func (*DataPoints) TimeEpochNsId() uint16 {
	return 1
}

func (*DataPoints) TimeEpochNsSinceVersion() uint16 {
	return 0
}

func (d *DataPoints) TimeEpochNsInActingVersion(actingVersion uint16) bool {
	return actingVersion >= d.TimeEpochNsSinceVersion()
}

func (*DataPoints) TimeEpochNsDeprecated() uint16 {
	return 0
}

func (*DataPoints) TimeEpochNsMetaAttribute(meta int) string {
	switch meta {
	case 1:
		return ""
	case 2:
		return ""
	case 3:
		return ""
	case 4:
		return "required"
	}
	return ""
}

func (*DataPoints) TimeEpochNsMinValue() int64 {
	return math.MinInt64 + 1
}

func (*DataPoints) TimeEpochNsMaxValue() int64 {
	return math.MaxInt64
}

func (*DataPoints) TimeEpochNsNullValue() int64 {
	return math.MinInt64
}

func (*DataPoints) ValueId() uint16 {
	return 2
}

func (*DataPoints) ValueSinceVersion() uint16 {
	return 0
}

func (d *DataPoints) ValueInActingVersion(actingVersion uint16) bool {
	return actingVersion >= d.ValueSinceVersion()
}

func (*DataPoints) ValueDeprecated() uint16 {
	return 0
}

func (*DataPoints) ValueMetaAttribute(meta int) string {
	switch meta {
	case 1:
		return ""
	case 2:
		return ""
	case 3:
		return ""
	case 4:
		return "required"
	}
	return ""
}

func (*DataPoints) ValueMinValue() float64 {
	return -math.MaxFloat64
}

func (*DataPoints) ValueMaxValue() float64 {
	return math.MaxFloat64
}

func (*DataPoints) ValueNullValue() float64 {
	return math.NaN()
}

func (*Data) PointsId() uint16 {
	return 0
}

func (*Data) PointsSinceVersion() uint16 {
	return 0
}

func (d *Data) PointsInActingVersion(actingVersion uint16) bool {
	return actingVersion >= d.PointsSinceVersion()
}

func (*Data) PointsDeprecated() uint16 {
	return 0
}

func (*DataPoints) SbeBlockLength() (blockLength uint) {
	return 20
}

func (*DataPoints) SbeSchemaVersion() (schemaVersion uint16) {
	return 0
}
