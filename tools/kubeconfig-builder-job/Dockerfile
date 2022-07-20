FROM --platform=linux/amd64 golang:1.18.3-alpine3.16 AS builder

WORKDIR /go/src/app

COPY . .

RUN go get

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -installsuffix cgo -o kubeconfig-builder-job .

FROM --platform=linux/amd64 alpine:3.16

RUN apk --no-cache add ca-certificates

WORKDIR /usr/app/

COPY --from=builder /go/src/app .

ENTRYPOINT ["./kubeconfig-builder-job"]