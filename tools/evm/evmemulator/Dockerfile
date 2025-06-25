ARG GOLANG_IMAGE_TAG=1.24-bullseye

# Build stage
FROM golang:${GOLANG_IMAGE_TAG} AS build

# Get dependencies first. This improves caching behavior since they only need
# to be re-downloaded when go.mod changes.
COPY . /source
WORKDIR /source
RUN go mod download

WORKDIR /source/tools/evm/evmemulator
RUN go build -o /bin/evmemulator

EXPOSE 8545

ENTRYPOINT ["/bin/evmemulator"]
