FROM golang:1.13-alpine as build

WORKDIR $GOPATH/ataxia-go/

RUN apk add git

# copy and download dependencies
COPY go.* ./
RUN go mod download

#compile app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/ataxia ./cmd/ataxia

#resulting app
FROM scratch as final
COPY --from=build go/ataxia-go/bin /ataxia/
COPY --from=build go/ataxia-go/data /ataxia/data
COPY --from=build go/ataxia-go/scripts /ataxia/scripts
WORKDIR /ataxia
EXPOSE 9000
ENTRYPOINT [ "./ataxia" ]
