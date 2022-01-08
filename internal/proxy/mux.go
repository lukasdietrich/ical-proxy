package proxy

import (
	"log"
	"net/http"
)

type CalendarMuxSpec map[string]CalendarSpec

func NewMux(muxSpec CalendarMuxSpec) (http.Handler, error) {
	mux := http.NewServeMux()

	for url, calendarSpec := range muxSpec {
		calendar, err := NewCalendar(calendarSpec)
		if err != nil {
			return nil, err
		}

		log.Printf("mount proxy on %q for %q", url, calendarSpec.Origin.URL)
		mux.Handle(url, calendar)
	}

	return mux, nil
}
