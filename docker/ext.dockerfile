FROM python:3.12-slim AS base
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_NO_CACHE_DIR=1

# system deps only if needed (curl/tini/etc.)
RUN apt-get update && apt-get install -y --no-install-recommends tini && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# install deps first for Docker layer cache
COPY requirements.txt .
RUN pip install --upgrade pip && pip install -r requirements.txt

# now copy app code (so small changes don't bust dependency cache)
COPY . .

# run as non-root
RUN useradd -ms /bin/bash appuser
USER appuser

ENTRYPOINT ["/usr/bin/tini","--"]
CMD ["python","main.py"]
