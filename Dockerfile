
# builder image
FROM golang:1.19.2-alpine as builder
RUN mkdir /build
COPY . /build/
WORKDIR /build
RUN apk add --no-cache git
RUN go install -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -o booksbooksbooks .

FROM alpine:3.13.6
COPY --from=builder /build/ .
# COPY --from=builder /build/.env .
ENTRYPOINT [ "./booksbooksbooks" ]
# arguments that can be overridden
# CMD [ "3", "300" ]