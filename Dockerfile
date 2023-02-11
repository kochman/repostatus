FROM alpine:20220316 as builder
RUN apk --update-cache --no-cache add go=1.20-r0 git
RUN mkdir /build
WORKDIR /build
COPY . .
RUN go build -o repostatus

FROM alpine:3.15.1
RUN apk --update-cache --no-cache add ca-certificates
RUN mkdir -p /app/static
WORKDIR /app
COPY --from=builder /build/repostatus /build/status.html /app/
COPY --from=builder /build/static/ /app/static/
EXPOSE 5000
CMD ["/app/repostatus"]
