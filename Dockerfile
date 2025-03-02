# Multi-stage build for Go application

# Step 1: Build the Go application
FROM golang:latest AS builder

WORKDIR /chatbox

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build the Go binary with static linking (CGO_ENABLED=0)
RUN CGO_ENABLED=0 go build -o chatapp .

# Step 2: Set up the final image
FROM alpine:latest

# Install bash for debugging (optional)
RUN apk add --no-cache bash

WORKDIR /chatbox

# Copy the statically compiled binary from the builder
COPY --from=builder /chatbox/chatapp /chatbox/

# Copy client folders

COPY --from=builder /chatbox/templates /chatbox/templates
COPY --from=builder /chatbox/static /chatbox/static


RUN chmod +x /chatbox/chatapp

CMD ["/chatbox/chatapp"]
