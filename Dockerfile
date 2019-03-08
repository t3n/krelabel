FROM golang:1.12-alpine AS build

RUN apk add --update curl git \
    && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

ADD . /go/src/github.com/t3n/krelabel/

RUN cd /go/src/github.com/t3n/krelabel/ \
    && dep ensure \
    && CGO_ENABLED=0 go build -a -o build/krelabel main.go


FROM scratch
COPY --from=build /go/src/github.com/t3n/krelabel/build/krelabel /usr/local/bin/krelabel
CMD ["/usr/local/bin/krelabel"]
