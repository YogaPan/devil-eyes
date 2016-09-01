FROM golang
MAINTAINER YogaPan <yogapan85321@gmail.com>

RUN go get -u github.com/kataras/iris/iris
RUN go get -u github.com/jinzhu/gorm
RUN go get -u github.com/iris-contrib/middleware/logger
RUN go get -u github.com/go-sql-driver/mysql