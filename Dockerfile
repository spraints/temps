FROM golang:1.14.1

WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY cmd cmd
COPY pkg pkg
RUN go build -o /usr/bin/temps ./cmd/temps
COPY public /app/public
COPY templates /app/templates
ENV TEMPS_LISTEN_ADDR=:8080
ENV TEMPS_PUBLIC_PATH=/app/public
ENV TEMPS_TEMPLATES_PATH=/app/templates
ENTRYPOINT [ "/usr/bin/temps" ]
