FROM golang:1.10.3 as builder
WORKDIR /go/src/github.com/skpr/prometheus-cloudwatch
COPY . /go/src/github.com/skpr/prometheus-cloudwatch
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin/prometheus-cloudwatch github.com/skpr/prometheus-cloudwatch

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/github.com/skpr/prometheus-cloudwatch/bin/prometheus-cloudwatch /usr/local/bin/prometheus-cloudwatch
CMD ["prometheus-cloudwatch"]
