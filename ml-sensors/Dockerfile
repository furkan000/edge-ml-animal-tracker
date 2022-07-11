FROM ubuntu:20.04

ENV TERM=xterm
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    libopencv-dev \
    python3-opencv \
    python3-pip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*


RUN pip3 install torch==1.12
RUN pip3 install opencv-python
RUN pip3 install torchvision==0.13



WORKDIR /app
ADD . /app


ENV PYTHONUNBUFFERED=1
ENV PYTHONPATH=/app