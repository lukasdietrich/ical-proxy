package proxy

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	headerAge           = "Age"
	headerCacheControl  = "Cache-Control"
	headerContentLength = "Content-Length"
	headerContentType   = "Content-Type"
	headerDate          = "Date"
)

type cache struct {
	sync.Mutex

	expiration *cacheControl
	buffer     []byte
}

func (c *cache) respondFromCache(w http.ResponseWriter) error {
	if c.expiration == nil {
		sendStatus(w, http.StatusBadGateway)
		return nil
	}

	maxAge, age, _ := c.expiration.evaluate()

	h := w.Header()
	h.Add(headerAge, fmt.Sprintf("%d", age))
	h.Add(headerCacheControl, fmt.Sprintf("max-age=%d, public", maxAge))
	h.Add(headerContentLength, fmt.Sprintf("%d", len(c.buffer)))
	h.Add(headerContentType, "text/calendar")

	r := bytes.NewReader(c.buffer)
	_, err := r.WriteTo(w)
	return err

}

type cacheControl struct {
	date   time.Time
	maxAge int64
	age    int64
}

func parseCacheControl(h http.Header) (*cacheControl, error) {
	c := cacheControl{
		date:   time.Now(),
		maxAge: 3600,
		age:    0,
	}

	if err := c.parseDate(h); err != nil {
		return nil, err
	}

	if err := c.parseMaxAge(h); err != nil {
		return nil, err
	}

	if err := c.parseAge(h); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *cacheControl) evaluate() (maxAge, age int64, valid bool) {
	if c == nil {
		return 0, 0, false
	}

	delta := int64(time.Since(c.date) / time.Second)
	age = delta + c.age
	return c.maxAge, age, age < c.maxAge
}

func (c *cacheControl) parseDate(h http.Header) error {
	rawDate := h.Get(headerDate)
	if rawDate == "" {
		return nil
	}

	date, err := time.Parse(time.RFC1123, rawDate)
	if err == nil {
		c.date = date
	}

	return err
}

func (c *cacheControl) parseMaxAge(h http.Header) error {
	rawCacheControl := h.Get(headerCacheControl)
	for _, directive := range strings.Split(rawCacheControl, ",") {
		directive = strings.TrimSpace(directive)
		kv := strings.SplitN(directive, "=", 2)

		if kv[0] == "max-age" && len(kv) == 2 {
			maxAge, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return err
			}

			c.maxAge = maxAge
		}
	}

	return nil
}

func (c *cacheControl) parseAge(h http.Header) error {
	rawAge := h.Get(headerAge)
	if rawAge == "" {
		return nil
	}

	age, err := strconv.ParseInt(rawAge, 10, 64)
	if err == nil {
		c.age = age
	}

	return err
}
