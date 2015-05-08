FROM golang:1.4

RUN apt-get update \ 
  && apt-get install -y git mercurial libtool pkg-config build-essential autoconf automake uuid-dev \
  && ln -s /usr/bin/libtoolize /usr/bin/libtool

RUN mkdir /build && cd /build

# install libsodium
RUN git clone https://github.com/jedisct1/libsodium.git \
  && cd libsodium \
  && ./autogen.sh \
  && ./configure \
  && make && make check && make install \
  && cd /build

# install zeromq
RUN git clone https://github.com/zeromq/libzmq.git \
  && cd libzmq \
  && ./autogen.sh \
  && ./configure \
  && make \
  && make install \
  && ldconfig \
  && cd /build

# install czmq
RUN git clone https://github.com/zeromq/czmq.git \
  && cd czmq \
  && ./autogen.sh \
  && ./configure \
  && make && make check && make install \
  && ldconfig \
  && cd /build

RUN rm -rf /build

COPY . /go/src/github.com/digitalocean/logtalez
RUN cd /go/src/github.com/digitalocean/logtalez \
  && export GOPATH=/go \
  && cd cmd/logtalez \
  && go get \
  && go build -o /bin/logtalez \
  && rm -rf /go

VOLUME ["/etc/curve.d"]
ENTRYPOINT ["/bin/logtalez"]



