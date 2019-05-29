FROM golang:1.12.5-stretch
RUN go get -u github.com/erikfastermann/league-accounts
RUN go build -o /league-accounts github.com/erikfastermann/league-accounts
RUN cp -r $GOPATH/src/github.com/erikfastermann/league-accounts/template /template
CMD ["/league-accounts"]
