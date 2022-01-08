package rules

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/lukasdietrich/ical-proxy/internal/rfc5545"
)

type Scope = *[]byte
type ScopeSlice []Scope

type Scoper func(*rfc5545.ContentLine) ScopeSlice

func NewScoper(scope string) (Scoper, error) {
	const (
		name  = "name"
		value = "value"
		param = "param:"
	)

	if scope == name {
		return scopeName, nil
	}

	if scope == value {
		return scopeValue, nil
	}

	if strings.HasPrefix(param, scope) {
		return scopeParam([]byte(scope[len(param):])), nil
	}

	return nil, fmt.Errorf("unknown scope %q", scope)
}

func scopeName(line *rfc5545.ContentLine) ScopeSlice {
	return ScopeSlice{&line.Name}
}

func scopeValue(line *rfc5545.ContentLine) ScopeSlice {
	return ScopeSlice{&line.Value}
}

func scopeParam(name []byte) Scoper {
	return func(line *rfc5545.ContentLine) ScopeSlice {
		var scopeSlice ScopeSlice

		for _, param := range line.Params {
			if bytes.Equal(param.Name, name) {
				for i := range param.Values {
					scopeSlice = append(scopeSlice, &param.Values[i])
				}
			}
		}

		return scopeSlice
	}
}
