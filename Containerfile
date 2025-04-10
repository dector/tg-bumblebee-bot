FROM golang:1.24.1-alpine as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./
RUN CGO_ENABLED=0 \
    go build -o app.exe .

#

FROM scratch

COPY --from=build /app/app.exe /app/app.exe

ENTRYPOINT [ "/app/app.exe" ]
