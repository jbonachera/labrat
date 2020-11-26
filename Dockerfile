FROM node:alpine as web
WORKDIR /src
COPY ./www .
WORKDIR /src/
RUN yarn install
RUN yarn run build

FROM golang:alpine as builder
ENV CGO_ENABLED=0
RUN mkdir -p $GOPATH/src/github.com/jbonachera
WORKDIR $GOPATH/src/github.com/jbonachera/labrat
COPY go.* ./
RUN go mod download
COPY *.go ./
RUN go test ./...
ARG BUILT_VERSION="snapshot"
RUN go build -buildmode=exe -ldflags="-s -w -X main.BuiltVersion=${BUILT_VERSION}" \
       -a -o /bin/labrat .

FROM alpine as prod
ENTRYPOINT ["/usr/bin/labrat"]
RUN apk -U add ca-certificates && \
    rm -rf /var/cache/apk/*
COPY --from=web /src/build/ /usr/share/www/
COPY --from=builder /bin/labrat /usr/bin/labrat
