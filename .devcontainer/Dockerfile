FROM alpine:latest

RUN apk add --update

##
# git
##

RUN apk add git openssh-client nano less

##
# go
##

ARG GO_VERSION

RUN <<_EOF_ sh
arch=$(apk --print-arch)
goArch="amd64"
case \${arch} in
    aarch64) goArch="arm64" ;;
esac
wget -O /tmp/go.tar.gz https://dl.google.com/go/go${GO_VERSION}.linux-\${goArch}.tar.gz
tar -C /opt -xzf /tmp/go.tar.gz
_EOF_
ENV PATH="$PATH:/opt/go/bin"