FROM golang:1.24.1-alpine as build

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd ./
RUN go build -o app.exe .
RUN pwd
RUN tree

#

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/app.exe /app/app.exe

ENTRYPOINT [ "/app/app.exe" ]
