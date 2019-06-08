FROM golang:1.12 as builder

WORKDIR /go/src/github.com/trelore/geoapi
ADD . /go/src/github.com/trelore/geoapi

RUN CGO_ENABLED=0 go build -o app main.go

FROM gcr.io/distroless/base

WORKDIR /root/
COPY --from=builder /go/src/github.com/trelore/geoapi/app .

ENTRYPOINT [ "./app" ]