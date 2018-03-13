FROM golang:1.8
WORKDIR /go/src/app
COPY . .
RUN go get github.com/gin-gonic/gin \
&& go install -v 
CMD ["app"]