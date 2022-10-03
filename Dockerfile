FROM golang:1.19-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /sunbeam

FROM alpine AS runner

COPY --from=builder /sunbeam /sunbeam
RUN apk add --update nodejs npm
RUN npm install -g zx
COPY scripts /scripts
ENV SUNBEAM_COMMAND_DIR=/scripts
ENV TERM=xterm-256color
ENTRYPOINT ["/sunbeam"]
