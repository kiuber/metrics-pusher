FROM golang:1.22.4-alpine AS builder
LABEL maintainer="kiuber <kiuber.zhang@gmail.com>"

RUN apk add --no-cache git

WORKDIR /tmp/go-app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 go test -v

RUN go build -o ./out/go-app .

FROM alpine:3.9
RUN apk add ca-certificates

COPY --from=builder /tmp/go-app/out/go-app /app/go-app

WORKDIR /app

CMD ["/app/go-app"]
