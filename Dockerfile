FROM golang:1.10 as build
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
WORKDIR /go/src/github.com/thrasher-/gocryptotrader
COPY Gopkg.* ./
RUN dep ensure -vendor-only
COPY . .
RUN mv -vn config_example.json config.json \
 && GOARCH=386 GOOS=linux CGO_ENABLED=0 go build . \
 && mv gocryptotrader /go/bin/gocryptotrader

FROM alpine:latest
RUN apk update && apk add --no-cache ca-certificates
COPY --from=build /go/bin/gocryptotrader /app/
COPY --from=build /go/src/github.com/thrasher-/gocryptotrader/config.json /app/
EXPOSE 9050
EXPOSE 9051
CMD ["/app/gocryptotrader"]
