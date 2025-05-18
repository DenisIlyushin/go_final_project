FROM golang:1.23.4-alpine3.20 AS builder

ARG TODO_PORT=7540
ARG TODO_DBFILE=scheduler.db
ARG TODO_PASSWORD=secret123
ENV TODO_PORT=${TODO_PORT}
ENV TODO_DBFILE=${TODO_DBFILE}
ENV TODO_PASSWORD=${TODO_PASSWORD}

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app


FROM alpine:3.20

ARG TODO_PORT=7540
ARG TODO_DBFILE=scheduler.db
ARG TODO_PASSWORD=secret123
ENV TODO_PORT=${TODO_PORT}
ENV TODO_DBFILE=${TODO_DBFILE}
ENV TODO_PASSWORD=${TODO_PASSWORD}

WORKDIR /app
RUN adduser -D -g '' appuser

COPY --from=builder /src/app .

RUN chmod +x /app/app && chown appuser:appuser /app -R
USER appuser

EXPOSE ${TODO_PORT}
CMD ["./app"]