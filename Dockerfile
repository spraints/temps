FROM golang:1.14.1 AS build

WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY cmd cmd
COPY pkg pkg
RUN go build -o /usr/bin/temps ./cmd/temps

FROM node:13.12.0 AS js
WORKDIR /src
COPY package.json .
COPY package-lock.json .
RUN npm install
COPY script/rollup script/rollup
COPY assets .
RUN script/rollup

FROM golang:1.14.1
COPY public /app/public
COPY templates /app/templates
COPY --from=build /usr/bin/temps /usr/bin/temps
COPY --from=js /src/public /app/public
ENV TEMPS_LISTEN_ADDR=:8080
ENV TEMPS_PUBLIC_PATH=/app/public
ENV TEMPS_TEMPLATES_PATH=/app/templates
ENTRYPOINT [ "/usr/bin/temps" ]
