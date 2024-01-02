// Generated SBE (Simple Binary Encoding) message codec

package generated

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
)

type Authentication struct {
	AccountId uint32
	ProjectId uint32
}

func (a *Authentication) Encode(_m *SbeGoMarshaller, _w io.Writer, doRangeCheck bool) error {
	if doRangeCheck {
		if err := a.RangeCheck(a.SbeSchemaVersion(), a.SbeSchemaVersion()); err != nil {
			return err
		}
	}
	if err := _m.WriteUint32(_w, a.AccountId); err != nil {
		return err
	}
	if err := _m.WriteUint32(_w, a.ProjectId); err != nil {
		return err
	}
	return nil
}

func (a *Authentication) Decode(_m *SbeGoMarshaller, _r io.Reader, actingVersion uint16, blockLength uint16, doRangeCheck bool) error {
	if !a.AccountIdInActingVersion(actingVersion) {
		a.AccountId = a.AccountIdNullValue()
	} else {
		if err := _m.ReadUint32(_r, &a.AccountId); err != nil {
			return err
		}
	}
	if !a.ProjectIdInActingVersion(actingVersion) {
		a.ProjectId = a.ProjectIdNullValue()
	} else {
		if err := _m.ReadUint32(_r, &a.ProjectId); err != nil {
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

func (a *Authentication) RangeCheck(actingVersion uint16, schemaVersion uint16) error {
	if a.AccountIdInActingVersion(actingVersion) {
		if a.AccountId < a.AccountIdMinValue() || a.AccountId > a.AccountIdMaxValue() {
			return fmt.Errorf("Range check failed on a.AccountId (%v < %v > %v)", a.AccountIdMinValue(), a.AccountId, a.AccountIdMaxValue())
		}
	}
	if a.ProjectIdInActingVersion(actingVersion) {
		if a.ProjectId < a.ProjectIdMinValue() || a.ProjectId > a.ProjectIdMaxValue() {
			return fmt.Errorf("Range check failed on a.ProjectId (%v < %v > %v)", a.ProjectIdMinValue(), a.ProjectId, a.ProjectIdMaxValue())
		}
	}
	return nil
}

func AuthenticationInit(a *Authentication) {
	return
}

func (*Authentication) SbeBlockLength() (blockLength uint16) {
	return 8
}

func (*Authentication) SbeTemplateId() (templateId uint16) {
	return 2
}

func (*Authentication) SbeSchemaId() (schemaId uint16) {
	return 118
}

func (*Authentication) SbeSchemaVersion() (schemaVersion uint16) {
	return 1
}

func (*Authentication) SbeSemanticType() (semanticType []byte) {
	return []byte("")
}

func (*Authentication) SbeSemanticVersion() (semanticVersion string) {
	return ""
}

func (*Authentication) AccountIdId() uint16 {
	return 0
}

func (*Authentication) AccountIdSinceVersion() uint16 {
	return 0
}

func (a *Authentication) AccountIdInActingVersion(actingVersion uint16) bool {
	return actingVersion >= a.AccountIdSinceVersion()
}

func (*Authentication) AccountIdDeprecated() uint16 {
	return 0
}

func (*Authentication) AccountIdMetaAttribute(meta int) string {
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

func (*Authentication) AccountIdMinValue() uint32 {
	return 0
}

func (*Authentication) AccountIdMaxValue() uint32 {
	return math.MaxUint32 - 1
}

func (*Authentication) AccountIdNullValue() uint32 {
	return math.MaxUint32
}

func (*Authentication) ProjectIdId() uint16 {
	return 1
}

func (*Authentication) ProjectIdSinceVersion() uint16 {
	return 0
}

func (a *Authentication) ProjectIdInActingVersion(actingVersion uint16) bool {
	return actingVersion >= a.ProjectIdSinceVersion()
}

func (*Authentication) ProjectIdDeprecated() uint16 {
	return 0
}

func (*Authentication) ProjectIdMetaAttribute(meta int) string {
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

func (*Authentication) ProjectIdMinValue() uint32 {
	return 0
}

func (*Authentication) ProjectIdMaxValue() uint32 {
	return math.MaxUint32 - 1
}

func (*Authentication) ProjectIdNullValue() uint32 {
	return math.MaxUint32
}
