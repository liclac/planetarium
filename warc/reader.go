package warc

import (
	"bufio"
	"io"
	"io/ioutil"
	"net/textproto"
	"strconv"

	"github.com/pkg/errors"
)

type Reader struct {
	R *textproto.Reader

	hdr   Header
	block io.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{R: textproto.NewReader(bufio.NewReader(r))}
}

func (r *Reader) Next() (Header, io.Reader, error) {
	if r.block != nil {
		if _, err := io.Copy(ioutil.Discard, r.block); err != nil {
			return Header{}, nil, err
		}
	}

	r.hdr = Header{}
	r.block = nil

	for r.hdr.Version == "" {
		versionLine, err := r.R.ReadLine()
		if err != nil {
			return Header{}, nil, err
		}
		r.hdr.Version = Version(versionLine)
		if r.hdr.Version != "" && !r.hdr.Version.IsSupported() {
			return Header{}, nil, errors.Errorf("unsupported version: '%s'", r.hdr.Version)
		}
	}

	fields, err := r.R.ReadMIMEHeader()
	if err != nil {
		return Header{}, nil, err
	}
	r.hdr.Fields = fields

	r.hdr.WARCType = WARCType(fields.Get("WARC-Type"))
	r.hdr.WARCRecordID = fields.Get("WARC-Record-ID")
	r.hdr.WARCDate = fields.Get("WARC-Date")

	contentLength, err := strconv.ParseInt(fields.Get("Content-Length"), 10, 64)
	if err != nil {
		return Header{}, nil, err
	}
	r.hdr.ContentLength = contentLength

	r.block = io.LimitReader(r.R.R, r.hdr.ContentLength)
	return r.hdr, r.block, nil
}
