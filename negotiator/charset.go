// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package negotiator

import (
	"sort"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
)

type acceptCharset struct {
	charset string
	q       float64
	i       int
}

type acceptCharsets []acceptCharset

func (acs acceptCharsets) filter(f func(ac acceptCharset) bool) acceptCharsets {
	result := make(acceptCharsets, 0, len(acs))
	for _, ac := range acs {
		if f(ac) {
			result = append(result, ac)
		}
	}
	return result
}

func (acs acceptCharsets) toCharsets() []string {
	result := make([]string, len(acs), len(acs))
	for i, ac := range acs {
		result[i] = ac.charset
	}
	return result
}

type acceptCharsetBy func(ac1, ac2 *acceptCharset) bool

func (by acceptCharsetBy) sort(acs acceptCharsets) {
	as := &acceptCharsetSorter{acs, by}
	sort.Sort(as)
}

type acceptCharsetSorter struct {
	acs acceptCharsets
	by  func(ac1, ac2 *acceptCharset) bool
}

func (s *acceptCharsetSorter) Len() int {
	return len(s.acs)
}

func (s *acceptCharsetSorter) Swap(i, j int) {
	s.acs[i], s.acs[j] = s.acs[j], s.acs[i]
}

func (s *acceptCharsetSorter) Less(i, j int) bool {
	return s.by(&s.acs[i], &s.acs[j])
}

type specificity struct {
	i int
	o int
	q float64
	s int
}

type specificities []specificity

func (ss specificities) filter(f func(s specificity) bool) specificities {
	result := make(specificities, 0, len(ss))
	for _, s := range ss {
		if f(s) {
			result = append(result, s)
		}
	}
	return result
}

func (ss specificities) indexOf(s specificity) int {
	for i, v := range ss {
		if v == s {
			return i
		}
	}
	return -1
}

type specificityBy func(s1, s2 *specificity) bool

func (by specificityBy) sort(specs specificities) {
	ss := &specificitySorter{specs, by}
	sort.Sort(ss)
}

type specificitySorter struct {
	ss specificities
	by func(s1, s2 *specificity) bool
}

func (s *specificitySorter) Len() int {
	return len(s.ss)
}

func (s *specificitySorter) Swap(i, j int) {
	s.ss[i], s.ss[j] = s.ss[j], s.ss[i]
}

func (s *specificitySorter) Less(i, j int) bool {
	return s.by(&s.ss[i], &s.ss[j])
}

var simpleCharsetRegExp = regexp2.MustCompile("^\\s*([^\\s;]+)\\s*(?:;(.*))?$", regexp2.None)

// PreferredCharsets gets the preferred charsets from an Accept-Charset header.
// RFC 2616 sec 14.2: no header = *, so you should pass * if no Accept-Charset field in header.
func PreferredCharsets(accept string, provided ...string) []string {
	acs := parseAcceptCharset(accept)

	if len(provided) == 0 {
		// sorted list of all charsets
		filteredAcs := acs.filter(isAcceptCharsetQuality)
		acceptCharsetBy(func(ac1, ac2 *acceptCharset) bool {
			return ac1.q > ac2.q || ac1.i < ac2.i
		}).sort(filteredAcs)
		return filteredAcs.toCharsets()
	}

	// sorted list of accepted charsets
	priorities := getSpecificities(provided, acs)
	filteredPriorities := priorities.filter(isSpecificityQuality)
	specificityBy(func(s1, s2 *specificity) bool {
		return s1.q > s2.q || s1.s < s2.s || s1.o < s2.o || s1.i < s2.i
	}).sort(filteredPriorities)

	results := make([]string, 0, len(filteredPriorities))
	for _, v := range filteredPriorities {
		i := priorities.indexOf(v)
		if i >= 0 {
			results = append(results, provided[i])
		}
	}

	return results
}

// Parses the Accept-Charset header to slice with type acceptCharset.
func parseAcceptCharset(accept string) acceptCharsets {
	accepts := strings.Split(accept, ",")
	length := len(accepts)
	results := make(acceptCharsets, 0, length)

	for i := 0; i < length; i++ {
		charset := parseCharset(strings.Trim(accepts[i], " "), i)
		if charset != nil {
			results = append(results, *charset)
		}
	}

	return results
}

// Parse a charset from the Accept-Charset header.
func parseCharset(s string, i int) *acceptCharset {
	match, err := simpleCharsetRegExp.FindStringMatch(s)
	if match == nil || match.GroupCount() == 0 || err != nil {
		return nil
	}

	charset, q := match.Groups()[1].String(), 1.0
	if match.Groups()[2].String() != "" {
		params := strings.Split(match.Groups()[2].String(), ";")
		for j := 0; j < len(params); j++ {
			p := strings.Split(strings.Trim(params[j], " "), "=")
			if p[0] == "q" {
				q1, err := strconv.ParseFloat(p[1], 64)
				if err != nil {
					return nil
				}
				q = q1
				break
			}
		}
	}

	return &acceptCharset{charset, q, i}
}

// Get the priority of a charset.
func getCharsetPriority(charset string, acs acceptCharsets, index int) specificity {
	priority := specificity{o: -1, q: 0, s: 0}

	for i := 0; i < len(acs); i++ {
		spec := specify(charset, acs[i], index)
		if spec != nil {
			s, q, o := priority.s-spec.s, priority.q-spec.q, priority.o-spec.o
			if s < 0 || q < 0 || o < 0 {
				priority = *spec
			}
		}
	}

	return priority
}

// Get the specificity of the charset.
func specify(charset string, ac acceptCharset, index int) *specificity {
	s := 0
	if strings.ToLower(ac.charset) == strings.ToLower(charset) {
		s |= 1
	} else if ac.charset != "*" {
		return nil
	}
	return &specificity{index, ac.i, ac.q, s}
}

func isAcceptCharsetQuality(spec acceptCharset) bool {
	return spec.q > 0
}

func isSpecificityQuality(s specificity) bool {
	return s.q > 0
}

func getSpecificities(types []string, acs acceptCharsets) specificities {
	result := make(specificities, len(types), len(types))
	for i, v := range types {
		result[i] = getCharsetPriority(v, acs, i)
	}
	return result
}
