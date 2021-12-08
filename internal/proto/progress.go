package proto

import (
	"github.com/go-faster/errors"
)

type Progress struct {
	Rows      uint64
	Bytes     uint64
	TotalRows uint64

	WroteRows  uint64
	WroteBytes uint64
}

func (p *Progress) DecodeAware(r *Reader, version int) error {
	{
		v, err := r.UVarInt()
		if err != nil {
			return errors.Wrap(err, "rows")
		}
		p.Rows = v
	}
	{
		v, err := r.UVarInt()
		if err != nil {
			return errors.Wrap(err, "bytes")
		}
		p.Bytes = v
	}
	if FeatureClientWriteInfo.In(version) {
		v, err := r.UVarInt()
		if err != nil {
			return errors.Wrap(err, "total rows")
		}
		p.TotalRows = v
	}
	if FeatureClientWriteInfo.In(version) {
		{
			v, err := r.UVarInt()
			if err != nil {
				return errors.Wrap(err, "wrote rows")
			}
			p.WroteRows = v
		}
		{
			v, err := r.UVarInt()
			if err != nil {
				return errors.Wrap(err, "wrote bytes")
			}
			p.WroteBytes = v
		}
	}

	return nil
}
