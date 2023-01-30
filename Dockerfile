FROM golang:alpine as build
# RUN go install golang.org/x/tools/cmd/goimports@latest
# RUN go install honnef.co/go/tools/cmd/staticcheck@latest
# RUN go install github.com/gordonklaus/ineffassign@latest
#RUN go install github.com/gostaticanalysis/nilerr/cmd/nilerr@latest
COPY . /build
WORKDIR /build
# RUN test -z "$(gofmt -s -l . | grep -v '^vendor' | tee /dev/stderr)"
# RUN staticcheck ./...
#RUN test -z "$(nilerr ./... 2>&1 | tee /dev/stderr)"
# RUN ineffassign .
# RUN go install ./...
# RUN go test -race -v ./...
# RUN go vet ./...
# RUN test -z "$(go vet ./... | grep -v '^vendor' | tee /dev/stderr)"
RUN go build -o m5paper-dashboard-server
FROM alpine
RUN apk update && apk add chromium nss  freetype  harfbuzz  ca-certificates  ttf-freefont curl fontconfig \
		&& curl -O https://noto-website.storage.googleapis.com/pkgs/NotoSansCJKjp-hinted.zip \
		&& mkdir -p /usr/share/fonts/NotoSansCJKjp \
		&& unzip NotoSansCJKjp-hinted.zip -d /usr/share/fonts/NotoSansCJKjp/ \
		&& rm NotoSansCJKjp-hinted.zip \
		&& fc-cache -fv
COPY --from=build /build/m5paper-dashboard-server .
cmd ["/m5paper-dashboard-server"]
