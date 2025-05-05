from golang:alpine as build

	workdir /github.com/lukasdietrich/ical-proxy

	copy internal ./internal
	copy cmd ./cmd
	copy go* .

	run go build ./cmd/ical-proxy

FROM alpine:latest

	workdir /app

	copy --from=build /github.com/lukasdietrich/ical-proxy/ical-proxy  ./
	copy ./LICENSE ./

	run apk --no-cache add tzdata \
		&& adduser -D -H -u 1234 icalproxy

	expose 8080/tcp
	volume [ "/data" ]
	user icalproxy

	label org.opencontainers.image.authors="Lukas Dietrich <lukas@lukasdietrich.com>"
	label org.opencontainers.image.url="ghcr.io/lukasdietrich/ical-proxy"
	label org.opencontainers.image.source="https://github.com/lukasdietrich/ical-proxy"

	cmd [ "/app/ical-proxy", "-config", "/data/config.yml" ]
