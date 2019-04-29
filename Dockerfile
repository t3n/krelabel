FROM golang:1.12-alpine AS build

RUN apk add --update --no-cache ca-certificates git

COPY . /go/krelabel/
WORKDIR /go/krelabel

RUN CGO_ENABLED=0 go build -a -o build/krelabel main.go


FROM scratch
COPY --from=build /go/krelabel/build/krelabel /usr/local/bin/krelabel
CMD ["/usr/local/bin/krelabel"]
