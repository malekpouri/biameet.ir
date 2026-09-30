import jalaali from 'jalaali-js';
import { faNum } from './util.js';

export const JALALI_MONTHS = ['فروردین', 'اردیبهشت', 'خرداد', 'تیر', 'مرداد', 'شهریور', 'مهر', 'آبان', 'آذر', 'دی', 'بهمن', 'اسفند'];

// Indexed by JS weekday (0 = Sunday).
export const DAY_NAMES = ['یکشنبه', 'دوشنبه', 'سه‌شنبه', 'چهارشنبه', 'پنج‌شنبه', 'جمعه', 'شنبه'];

// Weekday numbers in Iranian week order (Saturday first).
export const WEEK_ORDER = [6, 0, 1, 2, 3, 4, 5];

const timeFormat = new Intl.DateTimeFormat('fa-IR', { hour: '2-digit', minute: '2-digit', hour12: false });

export function dayName(date, useUTC = false) {
  return DAY_NAMES[useUTC ? date.getUTCDay() : date.getDay()];
}

function toJalali(date, useUTC) {
  return useUTC
    ? jalaali.toJalaali(date.getUTCFullYear(), date.getUTCMonth() + 1, date.getUTCDate())
    : jalaali.toJalaali(date);
}

// "شنبه ۱۲ مهر ۱۴۰۵". useUTC is for dates stored as UTC midnight (dynamic sessions).
export function formatDate(iso, useUTC = false) {
  if (!iso) return '';
  const date = new Date(iso);
  const j = toJalali(date, useUTC);
  return `${dayName(date, useUTC)} ${faNum(j.jd)} ${JALALI_MONTHS[j.jm - 1]} ${faNum(j.jy)}`;
}

export function formatTime(iso) {
  return iso ? timeFormat.format(new Date(iso)) : '';
}

export function faTime(hhmm) {
  return hhmm.replace(/\d/g, (d) => faNum(Number(d)));
}

export function jalaliToISO({ y, m, d }, hhmm) {
  const g = jalaali.toGregorian(y, m, d);
  const [h, min] = hhmm.split(':').map(Number);
  return new Date(g.gy, g.gm - 1, g.gd, h, min).toISOString();
}

export function jalaliToUTCMidnight({ y, m, d }) {
  const g = jalaali.toGregorian(y, m, d);
  return new Date(Date.UTC(g.gy, g.gm - 1, g.gd)).toISOString();
}

export function isValidJalali({ y, m, d }) {
  return jalaali.isValidJalaaliDate(y, m, d);
}

// ---- Pickers ----------------------------------------------------------------

const selectClass = 'rounded-lg border border-gray-300 bg-white p-2 text-center text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white';

export function timePicker(prefix, label, hour = 9, minute = 0) {
  const pad = (n) => String(n).padStart(2, '0');
  let hours = '';
  for (let i = 0; i < 24; i++) hours += `<option value="${pad(i)}"${i === hour ? ' selected' : ''}>${faNum(pad(i)).padStart(2, '۰')}</option>`;
  let minutes = '';
  for (let i = 0; i < 60; i += 15) minutes += `<option value="${pad(i)}"${i === minute ? ' selected' : ''}>${faNum(pad(i)).padStart(2, '۰')}</option>`;

  return `
    <fieldset class="flex flex-col items-center">
      <legend class="mb-1 text-xs text-gray-500 dark:text-gray-400">${label}</legend>
      <div class="flex items-center gap-1" dir="ltr">
        <select id="${prefix}_hour" class="${selectClass} w-16" aria-label="${label} — ساعت">${hours}</select>
        <span class="text-gray-500 dark:text-gray-400">:</span>
        <select id="${prefix}_minute" class="${selectClass} w-16" aria-label="${label} — دقیقه">${minutes}</select>
      </div>
    </fieldset>`;
}

export function readTime(prefix) {
  return `${document.getElementById(`${prefix}_hour`).value}:${document.getElementById(`${prefix}_minute`).value}`;
}

export function datePicker(prefix, label, date = new Date()) {
  const j = jalaali.toJalaali(date);
  let days = '';
  for (let i = 1; i <= 31; i++) days += `<option value="${i}"${i === j.jd ? ' selected' : ''}>${faNum(i)}</option>`;
  const months = JALALI_MONTHS.map((name, i) => `<option value="${i + 1}"${i + 1 === j.jm ? ' selected' : ''}>${name}</option>`).join('');
  let years = '';
  for (let y = j.jy; y <= j.jy + 1; y++) years += `<option value="${y}">${faNum(y)}</option>`;

  return `
    <fieldset>
      <legend class="label">${label}</legend>
      <div class="flex gap-2">
        <select id="${prefix}_day" class="${selectClass} flex-1" aria-label="روز">${days}</select>
        <select id="${prefix}_month" class="${selectClass} flex-[2]" aria-label="ماه">${months}</select>
        <select id="${prefix}_year" class="${selectClass} flex-1" aria-label="سال">${years}</select>
      </div>
    </fieldset>`;
}

export function readDate(prefix) {
  const v = (part) => parseInt(document.getElementById(`${prefix}_${part}`).value, 10);
  return { y: v('year'), m: v('month'), d: v('day') };
}
