// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package util

import (
	"errors"
	"sort"
	"strconv"
	"strings"
)

type Range struct {
	Start int64
	End   int64
	index int
}

type Ranges struct {
	Type   string
	Ranges []*Range
}

type rangeSortBy func(r1, r2 *Range) bool

func (by rangeSortBy) sort(ranges []*Range) {
	rs := &rangeSorter{ranges, by}
	sort.Sort(rs)
}

type rangeSorter struct {
	ranges []*Range
	by     rangeSortBy
}

func (r *rangeSorter) Len() int {
	return len(r.ranges)
}

func (r *rangeSorter) Swap(i, j int) {
	r.ranges[i], r.ranges[j] = r.ranges[j], r.ranges[i]
}

func (r *rangeSorter) Less(i, j int) bool {
	return r.by(r.ranges[i], r.ranges[j])
}

// RangeParser parses "Range" header `str` relative to the given file `size`.
//
// The "combine" argument can be set to `true` and overlapping & adjacent ranges
// * will be combined into a single range.
func RangeParser(size int64, str string, combine bool) (Ranges, error) {
	str = strings.TrimSpace(str)
	index := strings.Index(str, "=")
	ranges := Ranges{}

	if index == -1 || index == len(str)-1 {
		return ranges, errors.New("malformed header string")
	}

	arr := strings.Split(str[index+1:], ",")
	ranges.Ranges = make([]*Range, 0, len(arr))

	// add ranges type
	ranges.Type = str[0:index]

	// parse all ranges
	for i, v := range arr {
		v = strings.TrimSpace(v)
		values := strings.Split(v, "-")
		start, err1 := strconv.ParseInt(values[0], 10, 64)
		end, err2 := strconv.ParseInt(values[1], 10, 64)

		if err1 != nil {
			start = size - end
			end = size - 1
		} else if err2 != nil {
			end = size - 1
		}

		// limit last-byte-pos to current length
		if end > size-1 {
			end = size - 1
		}

		// unsatisifiable
		if start > end || start < 0 {
			continue
		}

		ranges.Ranges = append(ranges.Ranges, &Range{start, end, i})
	}

	// unsatisifiable
	if len(ranges.Ranges) == 0 {
		return Ranges{}, errors.New("unsatisifiable range")
	}

	if combine {
		combineRanges(&ranges)
		return ranges, nil
	}

	return ranges, nil
}

// Combine overlapping & adjacent ranges.
func combineRanges(r *Ranges) {
	rangeSortBy(sortByRangeStart).sort(r.Ranges)
	length := len(r.Ranges)

	j := 0
	for i := 1; i < length; i++ {
		ra, current := r.Ranges[i], r.Ranges[j]

		if ra.Start > current.End+1 {
			// next range
			j++
			r.Ranges[j] = ra
		} else if ra.End > current.End {
			// extend range
			current.End = ra.End
			current.index = Min(current.index, ra.index)
		}
	}

	// trim ordered array
	r.Ranges = r.Ranges[0 : j+1]

	rangeSortBy(sortByRangeIndex).sort(r.Ranges)
}

func sortByRangeStart(r1, r2 *Range) bool {
	return r1.Start < r2.Start
}

func sortByRangeIndex(r1, r2 *Range) bool {
	return r1.index < r2.index
}
