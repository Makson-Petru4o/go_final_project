FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY pkg ./pkg
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/go-todo .

FROM alpine:3
RUN addgroup -g 10001 todo && adduser -D -u 10001 -G todo todo && mkdir /data && chown todo:todo /data
WORKDIR /app
COPY --from=build /out/go-todo ./go-todo
COPY web ./web
ENV TODO_PORT=7540 TODO_DBFILE=/data/scheduler.db
EXPOSE 7540
USER todo
ENTRYPOINT ["/app/go-todo"]
