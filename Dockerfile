FROM golang:1.24 AS gob

WORKDIR /build

COPY . /build/

RUN go mod download && CGO_ENABLED=0 go build -a -o restarter main.go

FROM alpine:3.21.3

COPY --from=gob /build/restarter /api
VOLUME [ "/config" ]

ENTRYPOINT [ "/restarter" ]
