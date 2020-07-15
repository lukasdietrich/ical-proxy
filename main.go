package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type calendarMap map[string]calendar

type calendar struct {
	URL          string
	RemoveFields []string
}

func readConfig(filename string) (calendarMap, error) {
	var calendars calendarMap
	_, err := toml.DecodeFile(filename, &calendars)
	return calendars, err
}

func main() {
	var (
		addr string
		conf string
	)

	flag.StringVar(&addr, "addr", ":8080", "Address to bind the http server to")
	flag.StringVar(&conf, "conf", "conf.toml", "Filename of the config file")
	flag.Parse()

	calendars, err := readConfig(conf)
	if err != nil {
		log.Fatalf("could not read config: %v", err)
	}

	log.Printf("server stopped: %v", listenAndServe(addr, calendars))
}

// listenAndServe starts a web server on addr handling incoming http requests.
func listenAndServe(addr string, calendars calendarMap) error {
	mux := chi.NewMux()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Get("/ical/{calendar}.ics", handleProxy(calendars))

	return http.ListenAndServe(addr, mux)
}

func handleProxy(calendars calendarMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "calendar")
		cal, ok := calendars[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, http.StatusText(http.StatusNotFound))
			return
		}

		if err := proxyCalendar(w, cal); err != nil {
			log.Printf("could not proxy data for %q: %v", name, err)
		}
	}
}

func proxyCalendar(w io.Writer, cal calendar) error {
	res, err := http.Get(cal.URL)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return filterCalendar(w, res.Body, cal)
}

func filterCalendar(w io.Writer, r io.Reader, cal calendar) error {
	scanner := bufio.NewScanner(r)
	blocked := false

filterLines:
	for scanner.Scan() {
		line := scanner.Text()

		if len(line) > 0 && line[0] == ' ' || line[0] == '\t' {
			if blocked {
				continue filterLines
			}
		} else if i := strings.IndexRune(line, ':'); i > 0 {
			key := line[:i]

			for _, removeField := range cal.RemoveFields {
				if strings.EqualFold(removeField, key) {
					blocked = true
					continue filterLines
				}
			}
		}

		blocked = false
		fmt.Fprintln(w, line)
	}

	return scanner.Err()
}
