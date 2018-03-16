FROM golang:1.8


COPY . /go/src/
WORKDIR /go/src/api
RUN go install -v 
CMD ["api"]