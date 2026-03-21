FROM python:3.11-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

RUN apt-get update \
    && apt-get install -y --no-install-recommends gcc libfreetype6-dev libpng-dev \
    && rm -rf /var/lib/apt/lists/*

COPY services/analytics-python/requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir -r /app/requirements.txt

COPY services/analytics-python /app

EXPOSE 50051

CMD ["python", "analytics_server.py"]
