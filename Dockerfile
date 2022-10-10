FROM golang:alpine as build-env

RUN apk add git

COPY . /go/src/github.com/douban/qiniu-exporter
WORKDIR /go/src/github.com/douban/qiniu-exporter
# Build
ENV GOPATH=/go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -v -a -ldflags "-s -w" -o /go/bin/qiniu-exporter .

FROM library/alpine:3.15.0
RUN apk --no-cache add tzdata
COPY --from=build-env /go/bin/qiniu-exporter /usr/bin/qiniu-exporter
ENTRYPOINT ["qiniu-exporter"]
