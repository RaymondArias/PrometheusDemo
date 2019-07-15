FROM golang:latest
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN go build -o PrometheusDemo . 
CMD ["/app/PrometheusDemo"]