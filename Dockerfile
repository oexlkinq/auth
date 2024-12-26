# build stage
FROM golang:1.23 AS build-stage

WORKDIR /app_src

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY *.go ./
RUN CGO_ENABLED=0 go build -v -o /app


# TODO: вернуть, когда появятся тесты
# Run the tests in the container
# FROM build-stage AS run-test-stage
# RUN go test -v


# build release stage
FROM scratch AS build-release-stage

WORKDIR /

COPY --from=build-stage /app /app
COPY .env /

ENV GIN_MODE=release
ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app"]
