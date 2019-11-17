FROM node:12-stretch AS frontend-builder
WORKDIR /var/app
COPY . .
RUN cd ./frontend && npm i && npm run build && rm -rf node_modules/

FROM golang:1.13-stretch AS backend-builder
WORKDIR /go/src/github.com/spherical/spherical
COPY --from=frontend-builder /var/app .
RUN go get .
RUN go build .

FROM alpine AS certs
RUN apk add --no-cache ca-certificates

FROM scratch AS final-build
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt 
WORKDIR /var/app
COPY --from=backend-builder /go/src/github.com/spherical/spherical .
CMD ./spherical
