FROM golang:1.19-alpine AS go-build

RUN mkdir /src
WORKDIR /src

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine

RUN apk --no-cache --update add ca-certificates curl
WORKDIR /app
COPY static static
COPY --from=go-build /src/main /app/
EXPOSE 8080
ENTRYPOINT ./main