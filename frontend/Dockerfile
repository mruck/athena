FROM golang:1.12

RUN apt-get update && apt-get install -y \
    apt-transport-https \
    vim
RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -
RUN echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list
RUN apt-get update
RUN apt-get install -y kubectl

RUN mkdir /frontend
WORKDIR /frontend
ADD . .
RUN mkdir -p /root/.kube
CMD go run *.go
