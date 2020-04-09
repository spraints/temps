FROM golang:1.14.1

WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY cmd cmd
COPY pkg pkg
RUN go build -o /usr/bin/temps ./cmd/temps
ENV TEMPS_LISTEN_ADDR=:8080
ENTRYPOINT [ "/usr/bin/temps" ]
