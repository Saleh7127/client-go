FROM golang:alpine AS builder
RUN mkdir /client-go
COPY . /client-go
WORKDIR /client-go
RUN go build .

FROM alpine
WORKDIR /client-go
COPY --from=builder /client-go/ /client-go/

# Expose port 8080 to the outside world
EXPOSE 7127

# Command to run the executable
ENTRYPOINT ["./client-go"]
CMD ["client"]