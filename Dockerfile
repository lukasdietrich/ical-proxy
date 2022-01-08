FROM golang:alpine as build
	WORKDIR /github.com/lukasdietrich/ical-proxy
	COPY . .

	RUN go build ./cmd/ical-proxy

FROM alpine:latest
	WORKDIR /app
	COPY --from=build /github.com/lukasdietrich/ical-proxy/ical-proxy  ./

	RUN apk --no-cache add tzdata \
		&& adduser -D -H -u 1234 icalproxy

	EXPOSE 8080/tcp
	VOLUME [ "/data" ]
	USER icalproxy

	LABEL org.opencontainers.image.authors="Lukas Dietrich <lukas@lukasdietrich.com>"
	LABEL org.opencontainers.image.url="ghcr.io/lukasdietrich/ical-proxy"
	LABEL org.opencontainers.image.source="https://github.com/lukasdietrich/ical-proxy"

	CMD [ "/app/ical-proxy", "-config", "/data/config.yaml" ]
