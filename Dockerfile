FROM golang:1.13.7-alpine as builder
WORKDIR /workspace
RUN apk --no-cache add git && \
    git clone https://github.com/negineri/backpacker.git && \
    go build

FROM alpine:latest
LABEL maintainer="harusoin@gmail.com"
WORKDIR /usr/local/backpacker
COPY --from=builder /workspace/backpacker .
CMD ["./backpacker"]