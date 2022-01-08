package rules

import (
	"regexp"

	"github.com/lukasdietrich/ical-proxy/internal/rfc5545"
)

type MatchSpec struct {
	Scope   string `yaml:"scope"`
	Pattern string `yaml:"pattern"`
}

type Matcher func(*rfc5545.ContentLine) bool

func NewMatcher(matchSpec MatchSpec) (Matcher, error) {
	scoper, err := NewScoper(matchSpec.Scope)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(matchSpec.Pattern)
	if err != nil {
		return nil, err
	}

	return newPatternMatcher(scoper, re), nil
}

func newPatternMatcher(scoper Scoper, re *regexp.Regexp) Matcher {
	return func(line *rfc5545.ContentLine) bool {
		for _, v := range scoper(line) {
			if !re.Match(*v) {
				return false
			}
		}

		return true
	}
}
