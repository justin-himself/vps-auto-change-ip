FROM python:latest as builder
COPY requirements.txt /requirements.txt
RUN \
  apt update &&\
  curl https://sh.rustup.rs -sSf > /tmp/rustc.sh &&\
  sh /tmp/rustc.sh -y &&\
  . $HOME/.cargo/env &&\
  mkdir /install &&\
  pip install --upgrade pip &&\
  pip install --prefix /install -r /requirements.txt

FROM python:slim
COPY --from=builder /install /usr/local
COPY app /app
VOLUME /config
WORKDIR /
CMD ["python3", "/app/main.py"]