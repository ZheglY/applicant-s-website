FROM python:3.12-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

RUN apt-get update \
    && apt-get install -y --no-install-recommends build-essential libpq-dev \
    && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копируем ВСЁ, а не только app
COPY . .

# Или если нужно именно так:
# COPY app ./app
# COPY alembic ./alembic
# Также нужны db? Но db уже внутри app/app, смотрите ниже

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]