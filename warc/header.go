package warc

import "net/textproto"

type Version string

const (
	Version1_0 Version = "WARC/1.0"
	Version1_1 Version = "WARC/1.1"
)

func (v Version) IsSupported() bool {
	return v == Version1_0 || v == Version1_1
}

type WARCType string

const (
	WARCTypeWARCInfo     WARCType = "warcinfo"
	WARCTypeResponse     WARCType = "response"
	WARCTypeResource     WARCType = "resource"
	WARCTypeRequest      WARCType = "request"
	WARCTypeMetadata     WARCType = "metadata"
	WARCTypeRevisit      WARCType = "revisit"
	WARCTypeConversion   WARCType = "conversion"
	WARCTypeContinuation WARCType = "continuation"
)

type Header struct {
	Version       Version
	WARCType      WARCType
	WARCRecordID  string
	WARCDate      string
	ContentLength int64
	Fields        textproto.MIMEHeader
}
