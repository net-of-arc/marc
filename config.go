//
// Copyright (C) 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/net-of-arc/marc
//

package marc

const defaultMaxSize = 4 << 20

// Config configures a Codec. The zero Config is valid: no schema and a 4 MiB
// input limit. New copies the Config, so later changes by the caller have no
// effect on the Codec.
type Config struct {
	// Schema answers how predicates render. nil means roles are inferred from
	// position on decode, the observed property or the fallback (meta for
	// literals, edge for references) is used on encode, and the default text
	// predicate is "text". The Codec keeps it by reference.
	Schema Schema

	// MaxSize is the largest input, in bytes, the decoder accepts. 0 means the
	// default of 4 MiB (4 194 304 bytes); a negative value is rejected by New
	// with ErrInvalidConfig.
	MaxSize int
}

func (cfg Config) maxSize() (int, error) {
	switch {
	case cfg.MaxSize < 0:
		return 0, ErrInvalidConfig.With(detail("MaxSize"))
	case cfg.MaxSize == 0:
		return defaultMaxSize, nil
	}
	return cfg.MaxSize, nil
}
