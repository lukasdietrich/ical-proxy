package rules

import "github.com/lukasdietrich/ical-proxy/internal/rfc5545"

type RuleSpec struct {
	Match []MatchSpec `yaml:"match"`
}

type Rule []Matcher

func NewRule(ruleSpec RuleSpec) (Rule, error) {
	var rule Rule

	for _, matchSpec := range ruleSpec.Match {
		matcher, err := NewMatcher(matchSpec)
		if err != nil {
			return nil, err
		}

		rule = append(rule, matcher)
	}

	return rule, nil
}

func (r Rule) Match(line *rfc5545.ContentLine) bool {
	for _, matcher := range r {
		if !matcher(line) {
			return false
		}
	}

	return true
}
