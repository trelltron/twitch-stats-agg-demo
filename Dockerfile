#build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .
RUN go get -d -v
RUN go build -o /go/bin/app -v

#test stage
RUN go test -v ./...

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=builder /go/bin/app /app
ENTRYPOINT /app
LABEL Name=twitchstatsaggdemo Version=0.0.1
EXPOSE 3000