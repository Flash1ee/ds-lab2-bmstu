FROM golang:alpine as builder

RUN mkdir -p /src
WORKDIR /src

RUN apk add --no-cache make

ADD . .

RUN cd v4/src && make build_library && mv bin/library /bin


FROM alpine

WORKDIR /bin/

COPY --from=builder /bin/library .

ENTRYPOINT ["/bin/library"]
