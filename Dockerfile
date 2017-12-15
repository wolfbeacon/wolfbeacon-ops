FROM centos:7
RUN yum install epel-release -y && yum install -y wget && yum install -y git && yum -y install python-pip
RUN mkdir /root/.aws
COPY config/aws_credentials /root/.aws/credentials
RUN pip install awscli --upgrade --user
WORKDIR tmp
RUN wget https://redirector.gvt1.com/edgedl/go/go1.9.2.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.9.2.linux-amd64.tar.gz
RUN mkdir go
WORKDIR go
ENV GOPATH /tmp/go
RUN mkdir src
WORKDIR src
RUN mkdir github.com
WORKDIR github.com
RUN mkdir wolfbeacon
WORKDIR wolfbeacon
RUN mkdir wolfbeacon-ops
WORKDIR wolfbeacon-ops
COPY . .
RUN /usr/local/go/bin/go get ./...
RUN /usr/local/go/bin/go build



RUN pwd

CMD /tmp/go/src/github.com/wolfbeacon/wolfbeacon-ops/wolfbeacon-ops
