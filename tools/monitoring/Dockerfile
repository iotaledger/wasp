FROM golang:1.16.7-buster as build

WORKDIR /wasp

COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download

COPY . ./

RUN go build -tags rocksdb

FROM gcr.io/distroless/cc

WORKDIR /

COPY --from=build /wasp/tools/monitoring/config.json /
COPY --from=build /wasp/wasp /

EXPOSE 7000/tcp
EXPOSE 9090/tcp
EXPOSE 5550/tcp
EXPOSE 2112/tcp

ENTRYPOINT ["/wasp"]

