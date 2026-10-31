FROM node:22-bookworm-slim AS remotion-deps

WORKDIR /opt/openmontage/remotion-composer
COPY third_party/openmontage/remotion-composer/package.json third_party/openmontage/remotion-composer/package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci --omit=dev; else npm install --omit=dev; fi

FROM python:3.11-slim-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg nodejs \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY openmontage-runner/requirements.txt ./requirements.txt
COPY third_party/openmontage /app/third_party/openmontage
RUN pip install --no-cache-dir -r /app/third_party/openmontage/requirements.txt -r requirements.txt
COPY --from=remotion-deps /opt/openmontage/remotion-composer/node_modules /app/third_party/openmontage/remotion-composer/node_modules
COPY openmontage-runner/app /app/app
COPY openmontage-runner/entrypoint.sh /usr/local/bin/openmontage-runner
RUN chmod 0555 /usr/local/bin/openmontage-runner

ENV OPENMONTAGE_SOURCE_DIR=/app/third_party/openmontage
EXPOSE 8090
ENTRYPOINT ["openmontage-runner"]
