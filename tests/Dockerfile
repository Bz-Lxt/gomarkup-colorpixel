FROM mcr.microsoft.com/playwright:v1.49.1-jammy
ENV TZ=Asia/Shanghai \
    PYTHONUNBUFFERED=1 \
    PYTHONIOENCODING=utf-8
WORKDIR /e2e
COPY package.json ./
RUN npm install --registry=https://registry.npmmirror.com
COPY playwright.config.ts e2e_flow.spec.ts api_smoke.py ./
CMD ["bash", "-lc", "python3 api_smoke.py && npx playwright test"]
