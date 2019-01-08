FROM golang AS build-env
RUN go get github.com/go-redis/redis github.com/lib/pq
WORKDIR /go/src/github.com/ezhdanovskiy/docker-go-multi
ADD . /go/src/github.com/ezhdanovskiy/docker-go-multi
RUN cd /go/src/github.com/ezhdanovskiy/docker-go-multi && CGO_ENABLED=0 go build -o goapp

FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=build-env /go/src/github.com/ezhdanovskiy/docker-go-multi/goapp /app

EXPOSE 8080
CMD ["./goapp"]