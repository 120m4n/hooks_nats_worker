# syntax=docker/dockerfile:1
# Start from the latest golang base image
ARG GO_VERSION=1.21.6
FROM golang:${GO_VERSION} AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0  go build -o /bin/server .

# Start from scratch
FROM scratch AS final

# Copy the executable from the "build" stage.
COPY --from=build /bin/server /server

# What the container should run when it is started.
ENTRYPOINT [ "/server" ]