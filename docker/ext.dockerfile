FROM python:3.12-slim
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 PIP_NO_CACHE_DIR=1
WORKDIR /app
# ARG ENTRY=main.py

COPY . .
RUN python -m pip install --upgrade pip && \
    if [ -f requirements.txt ]; then pip install -r requirements.txt; fi

CMD ["python","main.py"]
# EXPOSE 8000
# CMD ["sh","-c","python ${ENTRY}"]
