package proxy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/lukasdietrich/ical-proxy/internal/rfc5545"
	"github.com/lukasdietrich/ical-proxy/internal/rules"
)

const (
	headerContentType  = "Content-Type"
	headerCacheControl = "Cache-Control"
)

type Origin struct {
	URL string `yaml:"url"`
}

type CalendarSpec struct {
	Origin Origin           `yaml:"origin"`
	Rules  []rules.RuleSpec `yaml:"rules"`
}

type Calendar struct {
	origin Origin
	rules  []rules.Rule
}

func NewCalendar(calendarSpec CalendarSpec) (*Calendar, error) {
	var ruleSlice []rules.Rule

	for _, ruleSpec := range calendarSpec.Rules {
		rule, err := rules.NewRule(ruleSpec)
		if err != nil {
			return nil, err
		}

		ruleSlice = append(ruleSlice, rule)
	}

	calendar := Calendar{
		origin: calendarSpec.Origin,
		rules:  ruleSlice,
	}

	return &calendar, nil
}

func (p *Calendar) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer handlePanic(w, r)

	if err := p.fetchAndProxy(w); err != nil {
		log.Printf("error while trying to proxy calendar %q: %v", r.RequestURI, err)
		sendStatus(w, http.StatusInternalServerError)
	}
}

func sendStatus(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	w.Write([]byte(http.StatusText(status)))
}

func handlePanic(w http.ResponseWriter, r *http.Request) {
	if p := recover(); p != nil {
		log.Printf("panic while trying to handle %q: %v", r.RequestURI, p)
		sendStatus(w, http.StatusInternalServerError)
	}
}

func (p *Calendar) fetchAndProxy(w http.ResponseWriter) error {
	res, err := http.Get(p.origin.URL)
	if err != nil {
		return fmt.Errorf("could not fetch origin %q: %w", p.origin.URL, err)
	}

	header := w.Header()
	header.Add(headerContentType, "text/calendar")
	header.Add(headerCacheControl, res.Header.Get(headerCacheControl))

	defer res.Body.Close()
	return p.applyRules(rfc5545.NewWriter(w), rfc5545.NewReader(res.Body))
}

func (p *Calendar) applyRules(w rfc5545.Writer, r rfc5545.Reader) error {
	for {
		line, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}

		if !p.matchAnyRule(line) {
			w.Write(line)
		}
	}

	return nil

}

func (p *Calendar) matchAnyRule(line *rfc5545.ContentLine) bool {
	for _, rule := range p.rules {
		if rule.Match(line) {
			return true
		}
	}

	return false
}
