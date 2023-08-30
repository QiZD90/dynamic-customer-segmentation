# Cache modules
FROM golang:alpine as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download


# Build and run the tests
FROM golang:alpine
COPY --from=modules /go/pkg /go/pkg
WORKDIR /test
COPY . .

CMD ["go", "test", "./integration-test"]