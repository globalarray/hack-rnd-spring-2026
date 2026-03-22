# Frontend

Новый клиент для ПрофДНК собран как отдельное приложение на `React + TypeScript + Vite`.

## Режимы работы

- `mock` — полный локальный demo-flow без backend, включен по умолчанию
- `bff` — обращение к `bff-go` по `VITE_API_BASE_URL`

## Переменные окружения

```bash
VITE_API_MODE=mock
VITE_API_BASE_URL=http://hack.benzo.cloud:8080
```

## Скрипты

```bash
npm install
npm run dev
```

## Демо-учетки в `mock`

- admin: `admin@profdnk.local` / `admin12345`
- psychologist: `psycho@profdnk.local` / `psych12345`
