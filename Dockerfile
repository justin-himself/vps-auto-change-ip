FROM python:3.7 as builder
COPY requirements.txt /requirements.txt
RUN \
  mkdir /install &&\
  pip install --prefix /install -r /requirements.txt

FROM python:3.7-slim
COPY --from=builder /install /usr/local
COPY app /app
VOLUME /config
WORKDIR /
CMD ["python3", "/app/main.py"]