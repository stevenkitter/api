FROM golang:1.8



WORKDIR /go/src/app
COPY . .
RUN go get -v github.com/gin-gonic/gin \
&& go get -v github.com/go-sql-driver/mysql \
&& go get -v github.com/gomodule/redigo/redis \
&& go install -v 
CMD ["app"]