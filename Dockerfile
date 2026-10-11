FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/byecors .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=build /out/byecors /usr/local/bin/byecors

ENV PORT=3133
EXPOSE 3133

ENTRYPOINT ["byecors"]
