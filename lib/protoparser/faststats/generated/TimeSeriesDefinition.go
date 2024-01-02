// Generated SBE (Simple Binary Encoding) message codec

package generated

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"unicode/utf8"
)

type TimeSeriesDefinition struct {
	Id     uint32
	Labels []TimeSeriesDefinitionLabels
	Name   []uint8
}
type TimeSeriesDefinitionLabels struct {
	Key   []uint8
	Value []uint8
}

func (t *TimeSeriesDefinition) Encode(_m *SbeGoMarshaller, _w io.Writer, doRangeCheck bool) error {
	if doRangeCheck {
		if err := t.RangeCheck(t.SbeSchemaVersion(), t.SbeSchemaVersion()); err != nil {
			return err
		}
	}
	if err := _m.WriteUint32(_w, t.Id); err != nil {
		return err
	}
	var LabelsNumInGroup uint16 = uint16(len(t.Labels))
	if err := _m.WriteUint16(_w, LabelsNumInGroup); err != nil {
		return err
	}
	var LabelsBlockLength uint16 = 0
	if err := _m.WriteUint16(_w, LabelsBlockLength); err != nil {
		return err
	}
	for _, prop := range t.Labels {
		if err := prop.Encode(_m, _w); err != nil {
			return err
		}
	}
	if err := _m.WriteUint16(_w, uint16(len(t.Name))); err != nil {
		return err
	}
	if err := _m.WriteBytes(_w, t.Name); err != nil {
		return err
	}
	return nil
}

func (t *TimeSeriesDefinition) Decode(_m *SbeGoMarshaller, _r io.Reader, actingVersion uint16, blockLength uint16, doRangeCheck bool) error {
	if !t.IdInActingVersion(actingVersion) {
		t.Id = t.IdNullValue()
	} else {
		if err := _m.ReadUint32(_r, &t.Id); err != nil {
			return err
		}
	}
	if actingVersion > t.SbeSchemaVersion() && blockLength > t.SbeBlockLength() {
		io.CopyN(ioutil.Discard, _r, int64(blockLength-t.SbeBlockLength()))
	}

	if t.LabelsInActingVersion(actingVersion) {
		var LabelsNumInGroup uint16
		if err := _m.ReadUint16(_r, &LabelsNumInGroup); err != nil {
			return err
		}
		var LabelsBlockLength uint16
		if err := _m.ReadUint16(_r, &LabelsBlockLength); err != nil {
			return err
		}
		if cap(t.Labels) < int(LabelsNumInGroup) {
			t.Labels = make([]TimeSeriesDefinitionLabels, LabelsNumInGroup)
		}
		t.Labels = t.Labels[:LabelsNumInGroup]
		for i := range t.Labels {
			if err := t.Labels[i].Decode(_m, _r, actingVersion, uint(LabelsBlockLength)); err != nil {
				return err
			}
		}
	}

	if t.NameInActingVersion(actingVersion) {
		var NameLength uint16
		if err := _m.ReadUint16(_r, &NameLength); err != nil {
			return err
		}
		if cap(t.Name) < int(NameLength) {
			t.Name = make([]uint8, NameLength)
		}
		t.Name = t.Name[:NameLength]
		if err := _m.ReadBytes(_r, t.Name); err != nil {
			return err
		}
	}
	if doRangeCheck {
		if err := t.RangeCheck(actingVersion, t.SbeSchemaVersion()); err != nil {
			return err
		}
	}
	return nil
}

func (t *TimeSeriesDefinition) RangeCheck(actingVersion uint16, schemaVersion uint16) error {
	if t.IdInActingVersion(actingVersion) {
		if t.Id < t.IdMinValue() || t.Id > t.IdMaxValue() {
			return fmt.Errorf("Range check failed on t.Id (%v < %v > %v)", t.IdMinValue(), t.Id, t.IdMaxValue())
		}
	}
	for _, prop := range t.Labels {
		if err := prop.RangeCheck(actingVersion, schemaVersion); err != nil {
			return err
		}
	}
	if !utf8.Valid(t.Name[:]) {
		return errors.New("t.Name failed UTF-8 validation")
	}
	return nil
}

func TimeSeriesDefinitionInit(t *TimeSeriesDefinition) {
	return
}

func (t *TimeSeriesDefinitionLabels) Encode(_m *SbeGoMarshaller, _w io.Writer) error {
	if err := _m.WriteUint16(_w, uint16(len(t.Key))); err != nil {
		return err
	}
	if err := _m.WriteBytes(_w, t.Key); err != nil {
		return err
	}
	if err := _m.WriteUint16(_w, uint16(len(t.Value))); err != nil {
		return err
	}
	if err := _m.WriteBytes(_w, t.Value); err != nil {
		return err
	}
	return nil
}

func (t *TimeSeriesDefinitionLabels) Decode(_m *SbeGoMarshaller, _r io.Reader, actingVersion uint16, blockLength uint) error {
	if actingVersion > t.SbeSchemaVersion() && blockLength > t.SbeBlockLength() {
		io.CopyN(ioutil.Discard, _r, int64(blockLength-t.SbeBlockLength()))
	}

	if t.KeyInActingVersion(actingVersion) {
		var KeyLength uint16
		if err := _m.ReadUint16(_r, &KeyLength); err != nil {
			return err
		}
		if cap(t.Key) < int(KeyLength) {
			t.Key = make([]uint8, KeyLength)
		}
		t.Key = t.Key[:KeyLength]
		if err := _m.ReadBytes(_r, t.Key); err != nil {
			return err
		}
	}

	if t.ValueInActingVersion(actingVersion) {
		var ValueLength uint16
		if err := _m.ReadUint16(_r, &ValueLength); err != nil {
			return err
		}
		if cap(t.Value) < int(ValueLength) {
			t.Value = make([]uint8, ValueLength)
		}
		t.Value = t.Value[:ValueLength]
		if err := _m.ReadBytes(_r, t.Value); err != nil {
			return err
		}
	}
	return nil
}

func (t *TimeSeriesDefinitionLabels) RangeCheck(actingVersion uint16, schemaVersion uint16) error {
	if !utf8.Valid(t.Key[:]) {
		return errors.New("t.Key failed UTF-8 validation")
	}
	if !utf8.Valid(t.Value[:]) {
		return errors.New("t.Value failed UTF-8 validation")
	}
	return nil
}

func TimeSeriesDefinitionLabelsInit(t *TimeSeriesDefinitionLabels) {
	return
}

func (*TimeSeriesDefinition) SbeBlockLength() (blockLength uint16) {
	return 4
}

func (*TimeSeriesDefinition) SbeTemplateId() (templateId uint16) {
	return 0
}

func (*TimeSeriesDefinition) SbeSchemaId() (schemaId uint16) {
	return 118
}

func (*TimeSeriesDefinition) SbeSchemaVersion() (schemaVersion uint16) {
	return 1
}

func (*TimeSeriesDefinition) SbeSemanticType() (semanticType []byte) {
	return []byte("")
}

func (*TimeSeriesDefinition) SbeSemanticVersion() (semanticVersion string) {
	return ""
}

func (*TimeSeriesDefinition) IdId() uint16 {
	return 0
}

func (*TimeSeriesDefinition) IdSinceVersion() uint16 {
	return 0
}

func (t *TimeSeriesDefinition) IdInActingVersion(actingVersion uint16) bool {
	return actingVersion >= t.IdSinceVersion()
}

func (*TimeSeriesDefinition) IdDeprecated() uint16 {
	return 0
}

func (*TimeSeriesDefinition) IdMetaAttribute(meta int) string {
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

func (*TimeSeriesDefinition) IdMinValue() uint32 {
	return 0
}

func (*TimeSeriesDefinition) IdMaxValue() uint32 {
	return math.MaxUint32 - 1
}

func (*TimeSeriesDefinition) IdNullValue() uint32 {
	return math.MaxUint32
}

func (*TimeSeriesDefinitionLabels) KeyMetaAttribute(meta int) string {
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

func (*TimeSeriesDefinitionLabels) KeySinceVersion() uint16 {
	return 0
}

func (t *TimeSeriesDefinitionLabels) KeyInActingVersion(actingVersion uint16) bool {
	return actingVersion >= t.KeySinceVersion()
}

func (*TimeSeriesDefinitionLabels) KeyDeprecated() uint16 {
	return 0
}

func (TimeSeriesDefinitionLabels) KeyCharacterEncoding() string {
	return "UTF-8"
}

func (TimeSeriesDefinitionLabels) KeyHeaderLength() uint64 {
	return 2
}

func (*TimeSeriesDefinitionLabels) ValueMetaAttribute(meta int) string {
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

func (*TimeSeriesDefinitionLabels) ValueSinceVersion() uint16 {
	return 0
}

func (t *TimeSeriesDefinitionLabels) ValueInActingVersion(actingVersion uint16) bool {
	return actingVersion >= t.ValueSinceVersion()
}

func (*TimeSeriesDefinitionLabels) ValueDeprecated() uint16 {
	return 0
}

func (TimeSeriesDefinitionLabels) ValueCharacterEncoding() string {
	return "UTF-8"
}

func (TimeSeriesDefinitionLabels) ValueHeaderLength() uint64 {
	return 2
}

func (*TimeSeriesDefinition) LabelsId() uint16 {
	return 1
}

func (*TimeSeriesDefinition) LabelsSinceVersion() uint16 {
	return 0
}

func (t *TimeSeriesDefinition) LabelsInActingVersion(actingVersion uint16) bool {
	return actingVersion >= t.LabelsSinceVersion()
}

func (*TimeSeriesDefinition) LabelsDeprecated() uint16 {
	return 0
}

func (*TimeSeriesDefinitionLabels) SbeBlockLength() (blockLength uint) {
	return 0
}

func (*TimeSeriesDefinitionLabels) SbeSchemaVersion() (schemaVersion uint16) {
	return 1
}

func (*TimeSeriesDefinition) NameMetaAttribute(meta int) string {
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

func (*TimeSeriesDefinition) NameSinceVersion() uint16 {
	return 0
}

func (t *TimeSeriesDefinition) NameInActingVersion(actingVersion uint16) bool {
	return actingVersion >= t.NameSinceVersion()
}

func (*TimeSeriesDefinition) NameDeprecated() uint16 {
	return 0
}

func (TimeSeriesDefinition) NameCharacterEncoding() string {
	return "UTF-8"
}

func (TimeSeriesDefinition) NameHeaderLength() uint64 {
	return 2
}
