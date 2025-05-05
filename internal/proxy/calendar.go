package proxy

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/lukasdietrich/ical-proxy/internal/rfc5545"
	"github.com/lukasdietrich/ical-proxy/internal/rules"
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
	cache  cache
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

	if _, err := w.Write([]byte(http.StatusText(status))); err != nil {
		log.Printf("could not write status: %v", err)
	}
}

func handlePanic(w http.ResponseWriter, r *http.Request) {
	if p := recover(); p != nil {
		log.Printf("panic while trying to handle %q: %v", r.RequestURI, p)
		sendStatus(w, http.StatusInternalServerError)
	}
}

func (p *Calendar) fetchAndProxy(w http.ResponseWriter) error {
	p.cache.Lock()
	defer p.cache.Unlock()

	if err := p.updateCache(); err != nil {
		log.Printf("could not update cache: %v", err)
	}

	return p.cache.respondFromCache(w)
}

func (p *Calendar) updateCache() error {
	maxAge, age, valid := p.cache.expiration.evaluate()
	if valid && age < (maxAge/2) {
		return nil
	}

	res, err := http.Get(p.origin.URL)
	if err != nil {
		return fmt.Errorf("could not fetch origin %q: %w", p.origin.URL, err)
	}

	//nolint:errcheck
	defer res.Body.Close()
	var buffer bytes.Buffer

	if err := p.applyRules(rfc5545.NewWriter(&buffer), rfc5545.NewReader(res.Body)); err != nil {
		return fmt.Errorf("could not rewrite calendar: %w", err)
	}

	expiration, err := parseCacheControl(res.Header)
	if err != nil {
		return err
	}

	p.cache.expiration = expiration
	p.cache.buffer = buffer.Bytes()
	return nil
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

		if p.keepContentLine(line) {
			if err := w.Write(line); err != nil {
				return err
			}
		}
	}
}

func (p *Calendar) keepContentLine(line *rfc5545.ContentLine) bool {
	for _, rule := range p.rules {
		if rule.Match(line) {
			return false
		}
	}

	return true
}
