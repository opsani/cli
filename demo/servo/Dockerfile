FROM golang:latest

# Install Python3 & Ansible
RUN set -ex \
    && apt-get update \
    && apt-get install -y --no-install-recommends python3-pip python3-setuptools python3-wheel \
    && apt-get install -y --no-install-recommends git

RUN go get github.com/tsenart/vegeta
RUN go install github.com/tsenart/vegeta
RUN mv /go/bin/vegeta /usr/local/bin/vegeta

LABEL \
  maintainer="Blake Watters <blake@opsani.com>" \
  org.opencontainers.image.title="servo-newrelic-big_bear-vegeta" \
  org.opencontainers.image.description="An Opsani Servo image with New Relic measure & BIG-BEAR adjust plugins + Vegeta load testing" \
  org.opencontainers.image.authors="Blake Watters <blake@opsani.com>" \
  org.opencontainers.image.url="https://github.com/opsani/servo-newrelic-big_bear-vegeta" \
  org.opencontainers.image.vendor="https://opsani.com" \
  org.opencontainers.image.licenses="MIT"

# Install Servo dependencies
RUN pip3 install wheel requests PyYAML python-dateutil

# Install the Servo plugins (Adjust: BIG-BEAR, Measure: Aggregation, New Relic, Vegeta)
ADD https://github.com/opsani/servo/raw/master/servo \
    https://github.com/opsani/servo/raw/master/adjust.py \
    https://github.com/opsani/servo-put/raw/master/adjust \
    https://github.com/opsani/servo/raw/master/measure.py \
    https://github.com/opsani/servo-magg/raw/master/measure \
    /servo/

# Install aggregated measure plugins
ADD https://github.com/opsani/servo/raw/master/measure.py /servo/measure.d/
ADD https://github.com/opsani/servo-prom/raw/master/measure /servo/measure.d/prom
ADD https://github.com/opsani/servo-vegeta/raw/master/measure /servo/measure.d/vegeta
RUN chmod +x /servo/measure.d/prom /servo/measure.d/vegeta

RUN chmod a+rwx /servo/adjust /servo/measure /servo/servo

ENV PYTHONUNBUFFERED=1

WORKDIR /servo
ENTRYPOINT [ "python3", "servo" ]
