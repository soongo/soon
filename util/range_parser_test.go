// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRangeParser(t *testing.T) {
	tests := []struct {
		desc           string
		size           int64
		str            string
		combine        bool
		expectedRanges Ranges
		err            error
	}{
		{
			desc: "should return error for invalid str",
			size: 200,
			str:  "malformed",
			err:  errors.New("malformed header string"),
		},
		{
			desc: "should return error if all specified ranges are invalid",
			size: 200,
			str:  "bytes=500-20",
			err:  errors.New("unsatisifiable range"),
		},
		{
			desc: "should return error if all specified ranges are invalid",
			size: 200,
			str:  "bytes=500-999",
			err:  errors.New("unsatisifiable range"),
		},
		{
			desc: "should return error if all specified ranges are invalid",
			size: 200,
			str:  "bytes=500-999,1000-1499",
			err:  errors.New("unsatisifiable range"),
		},
		{
			desc: "should parse str",
			size: 1000,
			str:  "bytes=0-499",
			expectedRanges: Ranges{
				Type:   "bytes",
				Ranges: []*Range{{Start: 0, End: 499}},
			},
		},
		{
			desc: "should parse str",
			size: 1000,
			str:  "bytes=40-80",
			expectedRanges: Ranges{
				Type:   "bytes",
				Ranges: []*Range{{Start: 40, End: 80}},
			},
		},
		{
			desc: "should cap end at size",
			size: 200,
			str:  "bytes=0-499",
			expectedRanges: Ranges{
				Type:   "bytes",
				Ranges: []*Range{{Start: 0, End: 199}},
			},
		},
		{
			desc: "should parse str asking for last n bytes",
			size: 1000,
			str:  "bytes=-400",
			expectedRanges: Ranges{
				Type:   "bytes",
				Ranges: []*Range{{Start: 600, End: 999}},
			},
		},
		{
			desc: "should parse str with only start",
			size: 1000,
			str:  "bytes=400-",
			expectedRanges: Ranges{
				Type:   "bytes",
				Ranges: []*Range{{Start: 400, End: 999}},
			},
		},
		{
			desc: `should parse "bytes=0-"`,
			size: 1000,
			str:  "bytes=0-",
			expectedRanges: Ranges{
				Type:   "bytes",
				Ranges: []*Range{{Start: 0, End: 999}},
			},
		},
		{
			desc: "should parse str with no bytes",
			size: 1000,
			str:  "bytes=0-0",
			expectedRanges: Ranges{
				Type:   "bytes",
				Ranges: []*Range{{Start: 0, End: 0}},
			},
		},
		{
			desc: "should parse str asking for last byte",
			size: 1000,
			str:  "bytes=-1",
			expectedRanges: Ranges{
				Type:   "bytes",
				Ranges: []*Range{{Start: 999, End: 999}},
			},
		},
		{
			desc: "should parse str with multiple ranges",
			size: 1000,
			str:  "bytes=40-80,81-90,-1",
			expectedRanges: Ranges{
				Type: "bytes",
				Ranges: []*Range{
					{Start: 40, End: 80, index: 0},
					{Start: 81, End: 90, index: 1},
					{Start: 999, End: 999, index: 2},
				},
			},
		},
		{
			desc: "should parse str with some invalid ranges",
			size: 200,
			str:  "bytes=0-499,1000-,500-999",
			expectedRanges: Ranges{
				Type:   "bytes",
				Ranges: []*Range{{Start: 0, End: 199}},
			},
		},
		{
			desc: "should parse non-byte range",
			size: 1000,
			str:  "items=0-5",
			expectedRanges: Ranges{
				Type:   "items",
				Ranges: []*Range{{Start: 0, End: 5}},
			},
		},
		{
			desc:    "should combine overlapping ranges when combine is true",
			size:    150,
			str:     "bytes=0-4,90-99,5-75,100-199,101-102",
			combine: true,
			expectedRanges: Ranges{
				Type: "bytes",
				Ranges: []*Range{
					{Start: 0, End: 75, index: 0},
					{Start: 90, End: 149, index: 1},
				},
			},
		},
		{
			desc:    "should retain original order when combine is true",
			size:    150,
			str:     "bytes=-1,20-100,0-1,101-120",
			combine: true,
			expectedRanges: Ranges{
				Type: "bytes",
				Ranges: []*Range{
					{Start: 149, End: 149, index: 0},
					{Start: 20, End: 120, index: 1},
					{Start: 0, End: 1, index: 2},
				},
			},
		},
		{
			desc:    "should ignore space",
			size:    150,
			str:     "  bytes= 0-4,  90-99, 5-75, 100-199,  101-102  ",
			combine: true,
			expectedRanges: Ranges{
				Type: "bytes",
				Ranges: []*Range{
					{Start: 0, End: 75, index: 0},
					{Start: 90, End: 149, index: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ranges, err := RangeParser(tt.size, tt.str, tt.combine)
			if tt.err != nil {
				require.Error(t, err)
				assert.Equal(t, tt.err, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedRanges.Type, ranges.Type)
			for i, r := range tt.expectedRanges.Ranges {
				assert.Equal(t, *r, *ranges.Ranges[i])
			}
		})
	}
}
