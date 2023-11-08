FROM golang:alpine as builder

RUN mkdir -p /src
WORKDIR /src

RUN apk add --no-cache make

ADD . .

RUN cd v4/src && make build_reservation && mv bin/reservation /bin


FROM alpine

WORKDIR /bin/

COPY --from=builder /bin/reservation .

ENTRYPOINT ["/bin/reservation"]
