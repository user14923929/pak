FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY . .
RUN go build -o pak .

# минимальный финальный образ — только бинарник
FROM alpine:3.19
COPY --from=builder /src/pak /usr/local/bin/pak
ENTRYPOINT ["pak"]
