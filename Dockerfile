FROM alpine:latest
LABEL Author="Chiranjit Datta 'myself.chiranjit@gmail.com'"
COPY go-blockchain /bin/
EXPOSE 8888
CMD ["/bin/go-blockchain"]