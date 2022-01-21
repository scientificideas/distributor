FROM golang:latest AS build

ARG VERSION
ARG COMMIT

ADD ./ /opt/gopath/src/distributor
WORKDIR /opt/gopath/src/distributor
RUN go build -v -o distributor -ldflags="-X 'main.Version=${VERSION}-${COMMIT}'" .

FROM debian:latest
RUN apt update && apt install ca-certificates -y
COPY --from=build /opt/gopath/src/distributor/distributor /opt/distributor/
WORKDIR /opt/distributor/
ENTRYPOINT ["./distributor", "-config-type=env"]