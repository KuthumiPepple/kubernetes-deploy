FROM golang:1.22.1
WORKDIR /home
COPY ./pkg /home
RUN go build -o library
CMD ["/home/library"]