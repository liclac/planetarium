FROM golang:1.13-alpine

WORKDIR /go/src/github.com/liclac/planetarium
COPY . .

RUN apk add --no-cache git && go get -d -v ./... && apk del git
RUN go install -v ./...

VOLUME "/amber"
ENV PLT_ADDR=0.0.0.0:8000
EXPOSE 8000
CMD ["planetarium", "serve", "/amber"]
