FROM golang:alpine as build
	WORKDIR /build
	COPY . .

	RUN go build

FROM alpine:latest
	WORKDIR /app
	COPY --from=build /build/ical-proxy  ./
	COPY LICENSE ./

	RUN apk --no-cache add ca-certificates

	VOLUME [ "/data" ]

	EXPOSE 8080

	CMD [ "/app/ical-proxy", "-addr", ":8080", "-conf", "/data/conf.toml" ]
