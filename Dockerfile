FROM golang:1.12-alpine
MAINTAINER luwei <luw2007@gmail.com>

ENV GOROOT /usr/local/go
ENV GOPATH /go
ENV PATH $PATH:$GOROOT/bin
ENV GO111MODULE on

RUN mkdir -p /rabbitid /rabbitid/etc /rabbitid/logs
WORKDIR /rabbitid

ADD . /go/src/github.com/luw2007/rabbitid

# 编译
RUN cd /go/src/github.com/luw2007/rabbitid && \
  go build -v -o /rabbitid/idHttp  cmd/idHttp/main.go && \
  go build -v -o /rabbitid/idRedis cmd/idRedis/main.go && \
# 配置
  cp etc/*.toml /rabbitid/etc

CMD ["/rabbitid/idRedis"]
