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

EXPOSE 7000
EXPOSE 9090
EXPOSE 5550
EXPOSE 2112
EXPOSE 4000

ENTRYPOINT ["/wasp"]

