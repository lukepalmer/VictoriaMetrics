// Generated SBE (Simple Binary Encoding) message codec

package generated

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
)

type Acknowledgement struct {
	SequenceNumber uint64
}

func (a *Acknowledgement) Encode(_m *SbeGoMarshaller, _w io.Writer, doRangeCheck bool) error {
	if doRangeCheck {
		if err := a.RangeCheck(a.SbeSchemaVersion(), a.SbeSchemaVersion()); err != nil {
			return err
		}
	}
	if err := _m.WriteUint64(_w, a.SequenceNumber); err != nil {
		return err
	}
	return nil
}

func (a *Acknowledgement) Decode(_m *SbeGoMarshaller, _r io.Reader, actingVersion uint16, blockLength uint16, doRangeCheck bool) error {
	if !a.SequenceNumberInActingVersion(actingVersion) {
		a.SequenceNumber = a.SequenceNumberNullValue()
	} else {
		if err := _m.ReadUint64(_r, &a.SequenceNumber); err != nil {
			return err
		}
	}
	if actingVersion > a.SbeSchemaVersion() && blockLength > a.SbeBlockLength() {
		io.CopyN(ioutil.Discard, _r, int64(blockLength-a.SbeBlockLength()))
	}
	if doRangeCheck {
		if err := a.RangeCheck(actingVersion, a.SbeSchemaVersion()); err != nil {
			return err
		}
	}
	return nil
}

func (a *Acknowledgement) RangeCheck(actingVersion uint16, schemaVersion uint16) error {
	if a.SequenceNumberInActingVersion(actingVersion) {
		if a.SequenceNumber < a.SequenceNumberMinValue() || a.SequenceNumber > a.SequenceNumberMaxValue() {
			return fmt.Errorf("Range check failed on a.SequenceNumber (%v < %v > %v)", a.SequenceNumberMinValue(), a.SequenceNumber, a.SequenceNumberMaxValue())
		}
	}
	return nil
}

func AcknowledgementInit(a *Acknowledgement) {
	return
}

func (*Acknowledgement) SbeBlockLength() (blockLength uint16) {
	return 8
}

func (*Acknowledgement) SbeTemplateId() (templateId uint16) {
	return 0
}

func (*Acknowledgement) SbeSchemaId() (schemaId uint16) {
	return 119
}

func (*Acknowledgement) SbeSchemaVersion() (schemaVersion uint16) {
	return 0
}

func (*Acknowledgement) SbeSemanticType() (semanticType []byte) {
	return []byte("")
}

func (*Acknowledgement) SbeSemanticVersion() (semanticVersion string) {
	return ""
}

func (*Acknowledgement) SequenceNumberId() uint16 {
	return 0
}

func (*Acknowledgement) SequenceNumberSinceVersion() uint16 {
	return 0
}

func (a *Acknowledgement) SequenceNumberInActingVersion(actingVersion uint16) bool {
	return actingVersion >= a.SequenceNumberSinceVersion()
}

func (*Acknowledgement) SequenceNumberDeprecated() uint16 {
	return 0
}

func (*Acknowledgement) SequenceNumberMetaAttribute(meta int) string {
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

func (*Acknowledgement) SequenceNumberMinValue() uint64 {
	return 0
}

func (*Acknowledgement) SequenceNumberMaxValue() uint64 {
	return math.MaxUint64 - 1
}

func (*Acknowledgement) SequenceNumberNullValue() uint64 {
	return math.MaxUint64
}
