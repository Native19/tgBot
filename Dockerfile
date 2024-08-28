# syntax=docker/dockerfile:1

FROM golang:1.23

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy
COPY bot ./bot
COPY fileSaver ./fileSaver
COPY http ./http
COPY data ./data
COPY *.go ./
COPY .env ./

# Build
RUN go build -v -o /docker-tgBot

# Run
CMD ["/docker-tgBot"]