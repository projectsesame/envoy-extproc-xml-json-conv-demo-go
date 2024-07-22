FROM m.daocloud.io/docker.io/library/golang:1.21.6-bullseye

SHELL ["/bin/bash", "-c"]

RUN apt-get update && apt-get -y upgrade \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get -y clean

WORKDIR /build

COPY . .
RUN go mod tidy \
    && go mod download \
    && go build -o /extproc


FROM m.daocloud.io/docker.io/library/busybox:latest

COPY --from=0 /extproc /bin/extproc
RUN chmod +x /bin/extproc

ARG EXAMPLE=xml-json-conv

EXPOSE 50051

ENTRYPOINT [ "/bin/extproc" ]
CMD [ "conv", "--log-stream", "--log-phases" ]
