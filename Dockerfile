FROM docker.io/golang:1.19.4 AS builder
COPY . src
RUN cd src && make build

FROM docker.io/fedora:37
LABEL maintainer="Dalton Hubble <dghubble@gmail.com>"
LABEL org.opencontainers.image.title="scuttle",
LABEL org.opencontainers.image.source="https://github.com/poseidon/scuttle"
LABEL org.opencontainers.image.vendor="Poseidon Labs"
# AWS CLI v2 is an ugly pile of Python with dynamic linking to specific
# shared objects they zip up. And the zip doesn't even have a checksum
#
# Added for a customer using kubeconfig that exec's the aws cli
# https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
RUN dnf install -y curl zip && \
  curl -L https://awscli.amazonaws.com/awscli-exe-linux-x86_64-2.8.13.zip -o awscliv2.zip && \
  unzip awscliv2.zip && ./aws/install
COPY --from=builder /go/src/bin/scuttle /usr/local/bin
ENTRYPOINT ["/usr/local/bin/scuttle"]
