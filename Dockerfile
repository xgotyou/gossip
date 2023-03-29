FROM golang:latest as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

######## Start a new stage #######
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY miscellaneous/words /usr/share/dict/
EXPOSE 5051
CMD ["./main"]