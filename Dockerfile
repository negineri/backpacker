FROM golang:1.13.7-alpine as builder
WORKDIR /workspace
RUN apk --no-cache add git && \
    git clone https://github.com/negineri/backpacker.git . && \
    go build && \
    mkdir /cmds && \
    cp backpacker /cmds && \
    cp sync.sh /cmds

FROM alpine:latest
LABEL maintainer="harusoin@gmail.com"
WORKDIR /usr/local/backpacker
RUN apk --no-cache add rsync
COPY --from=builder /cmds/* .
CMD ["./backpacker"]