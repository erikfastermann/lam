FROM golang:1.12.5-stretch
RUN go get -u github.com/erikfastermann/lam
RUN go build -o /lam github.com/erikfastermann/lam
RUN cp -r $GOPATH/src/github.com/erikfastermann/lam/template /template
ENV LAM_TEMPLATE_GLOB=/template/*
RUN mkdir -p /var/lam/keypairs
CMD ["/lam"]
