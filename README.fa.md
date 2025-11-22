# BiaMeet (بیا میت)

بیا میت یک ابزار ساده و کارآمد برای زمان‌بندی جلسات است.

## مرور کلی
- **بک‌اند**: Go + Fiber
- **پایگاه داده**: SQLite
- **فرانت‌اند**: Static HTML/JS + Tailwind CSS
- **منطقه زمانی**: تمام زمان‌ها به صورت UTC ذخیره می‌شوند و به صورت جلالی (هجری شمسی) نمایش داده می‌شوند.

## توسعه محلی

### پیش‌نیازها
- Go 1.21+
- Docker & Docker Compose
- Node.js (برای ابزارهای بیلد فرانت‌اند)

### راه‌اندازی
1. مخزن را کلون کنید.
2. به پوشه `backend` بروید و دستور `go mod download` را اجرا کنید.
3. به پوشه `frontend` بروید و اسکریپت‌های بیلد را اجرا کنید.

### اجرا با داکر
```bash
docker-compose up --build
```

## سیاست شاخه‌گذاری
- شاخه‌های ویژگی: `feat/<descriptor>-<ticket>`
- شاخه‌های اصلاح: `fix/<descriptor>-<ticket>`
- پیام‌های کامیت: Conventional Commits (مثلاً `feat(session): add create endpoint`)

## مایگریشن‌ها
مایگریشن‌های SQL در مسیر `/backend/db/migrations` قرار دارند.
