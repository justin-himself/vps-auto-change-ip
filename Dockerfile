FROM python:latest as builder
COPY requirements.txt /requirements.txt
RUN \
  mkdir /install &&\
  pip install --upgrade pip &&\
  pip install --prefix /install -r /requirements.txt

FROM python:slim
COPY --from=builder /install /usr/local
COPY app /app
VOLUME /config
WORKDIR /
CMD ["python3", "/app/main.py"]