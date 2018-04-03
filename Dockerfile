FROM golang:1.8 AS builder

WORKDIR /go/src/github.com/stevenkitter/api/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/stevenkitter/api/app .
COPY --from=builder /go/src/github.com/stevenkitter/api/index.tmpl .
CMD ["./app"]
