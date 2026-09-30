import { api } from './api.js';
import { $, esc, faNum, toast } from './util.js';

const KEY = 'bm:admin-token';

function savedToken() {
  try {
    return sessionStorage.getItem(KEY) || '';
  } catch {
    return '';
  }
}

function remember(token) {
  try {
    if (token) sessionStorage.setItem(KEY, token);
    else sessionStorage.removeItem(KEY);
  } catch {
    /* ignore */
  }
}

export function showAdmin(app) {
  const token = savedToken();
  if (token) load(app, token);
  else renderLogin(app);
}

function renderLogin(app) {
  app.innerHTML = `
    <form id="admin-login" class="card mx-auto max-w-sm p-6">
      <h1 class="mb-4 text-center text-xl font-bold text-gray-900 dark:text-white">داشبورد مدیریت</h1>
      <label for="admin-token" class="label">توکن مدیریت</label>
      <input id="admin-token" type="password" class="field mb-4" autocomplete="current-password" dir="ltr" required>
      <button type="submit" class="btn-primary w-full">ورود</button>
    </form>`;
  $('admin-login').addEventListener('submit', (e) => {
    e.preventDefault();
    load(app, $('admin-token').value.trim());
  });
  $('admin-token').focus();
}

async function load(app, token) {
  try {
    const stats = await api('GET', '/admin/stats', undefined, { Authorization: `Bearer ${token}` });
    remember(token);
    render(app, stats);
  } catch (err) {
    remember('');
    if (err.status === 404) {
      app.innerHTML = '<div class="card p-6 text-center text-gray-600 dark:text-gray-300">داشبورد مدیریت غیرفعال است. برای فعال‌سازی، متغیر ADMIN_TOKEN را روی سرور تنظیم کنید.</div>';
      return;
    }
    toast(err.message, 'error');
    renderLogin(app);
  }
}

function render(app, stats) {
  const tile = (value, label, color) => `
    <div class="rounded-xl border p-6 text-center ${color}">
      <div class="mb-2 text-4xl font-bold">${esc(faNum(value))}</div>
      <div class="font-medium text-gray-600 dark:text-gray-300">${label}</div>
    </div>`;

  app.innerHTML = `
    <div class="card p-6 sm:p-8">
      <h1 class="mb-8 border-b border-gray-200 pb-4 text-center text-2xl font-bold text-gray-900 dark:border-gray-700 dark:text-white">داشبورد مدیریت</h1>
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
        ${tile(stats.total_sessions, 'کل جلسات', 'border-blue-100 bg-blue-50 text-blue-700 dark:border-blue-900 dark:bg-blue-950/40 dark:text-blue-300')}
        ${tile(stats.total_timeslots, 'زمان‌های پیشنهادی', 'border-green-100 bg-green-50 text-green-700 dark:border-green-900 dark:bg-green-950/40 dark:text-green-300')}
        ${tile(stats.total_votes, 'آرای ثبت‌شده', 'border-purple-100 bg-purple-50 text-purple-700 dark:border-purple-900 dark:bg-purple-950/40 dark:text-purple-300')}
      </div>
      <div class="mt-8 text-center"><a href="/" class="btn-ghost">بازگشت به صفحه اصلی</a></div>
    </div>`;
}
