FROM golang:1.16-alpine

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /app/esb-bridge

# populate the module cache
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build -o ./bin/esb-bridge-server ./cmd/server


# Exposes port 9815
EXPOSE 9815

# Run the binary program produced by `go install`
CMD ["./bin/esb-bridge-server", "-d", "/dev/ttyACM0", "-p", "9815"]