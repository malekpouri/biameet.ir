import { api } from './api.js';
import { DAY_NAMES, WEEK_ORDER, datePicker, isValidJalali, jalaliToISO, jalaliToUTCMidnight, readDate, readTime, timePicker } from './dates.js';
import { $, faNum, onAction, onChange, saveName, toast } from './util.js';

const TYPES = [
  { value: 'fixed', label: 'زمان‌های مشخص', help: 'چند زمان مشخص پیشنهاد می‌دهید و شرکت‌کنندگان به آن‌ها رای می‌دهند.' },
  { value: 'weekly', label: 'الگوی هفتگی', help: 'روزهای مجاز هفته و بازه ساعت را تعیین می‌کنید و شرکت‌کنندگان در این روزها زمان پیشنهاد می‌دهند.' },
  { value: 'dynamic', label: 'بازه زمانی', help: 'یک روز و بازه ساعت کلی تعیین می‌کنید و هر کس می‌تواند زمان دلخواهش را در این بازه پیشنهاد دهد.' },
];

let rowCounter = 0;

export function showCreate(app) {
  app.innerHTML = `
    <section class="card p-5 sm:p-6" aria-labelledby="create-title">
      <h2 id="create-title" class="mb-6 text-center text-2xl font-bold text-gray-900 dark:text-white">ایجاد جلسه جدید</h2>
      <div class="space-y-5">
        <div>
          <label for="session-title" class="label">عنوان جلسه</label>
          <input id="session-title" class="field" maxlength="200" placeholder="مثلاً: جلسه هفتگی تیم">
        </div>
        <div>
          <label for="creator-name" class="label">نام شما</label>
          <input id="creator-name" class="field" maxlength="64" autocomplete="name" placeholder="نام ایجاد کننده">
        </div>

        <fieldset>
          <legend class="label">نوع جلسه</legend>
          <div class="grid grid-cols-3 gap-2">
            ${TYPES.map((t, i) => `
              <label class="cursor-pointer">
                <input type="radio" name="session-type" value="${t.value}" data-change="session-type" class="peer sr-only"${i === 0 ? ' checked' : ''}>
                <span class="block h-full rounded-lg border-2 border-gray-200 p-2 text-center text-sm font-medium text-gray-700 transition-colors peer-checked:border-blue-500 peer-checked:bg-blue-50 peer-checked:text-blue-800 peer-focus-visible:ring-2 peer-focus-visible:ring-blue-500 dark:border-gray-600 dark:text-gray-200 dark:peer-checked:bg-blue-900/30 dark:peer-checked:text-blue-200">${t.label}</span>
              </label>`).join('')}
          </div>
          <p id="type-help" class="mt-2 text-sm text-gray-600 dark:text-gray-400">${TYPES[0].help}</p>
        </fieldset>

        <div id="section-fixed">
          <p class="label">زمان‌های پیشنهادی</p>
          <ul id="slot-rows" class="space-y-3"></ul>
          <button type="button" data-action="add-row" class="mt-3 w-full rounded-lg border-2 border-dashed border-blue-300 py-2 text-blue-600 transition-colors hover:bg-blue-50 dark:border-blue-800 dark:text-blue-400 dark:hover:bg-gray-700/50">+ افزودن زمان دیگر</button>
        </div>

        <div id="section-weekly" class="hidden space-y-4 rounded-lg bg-purple-50 p-4 dark:bg-purple-950/30">
          <fieldset>
            <legend class="label">روزهای هفته</legend>
            <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
              ${WEEK_ORDER.map((d) => `
                <label class="flex cursor-pointer items-center gap-2 rounded-lg border border-gray-200 bg-white p-2 dark:border-gray-600 dark:bg-gray-700">
                  <input type="checkbox" name="week-day" value="${d}" class="h-4 w-4">
                  <span class="text-sm">${DAY_NAMES[d]}</span>
                </label>`).join('')}
            </div>
          </fieldset>
          <div class="flex items-center justify-center gap-4">
            ${timePicker('weekly_start', 'از ساعت', 9)}
            <span class="mt-5 text-gray-400" aria-hidden="true">←</span>
            ${timePicker('weekly_end', 'تا ساعت', 17)}
          </div>
          <div>
            <label for="weeks" class="label">مدت اعتبار لینک (هفته)</label>
            <input id="weeks" type="number" class="field" value="4" min="1" max="52" inputmode="numeric">
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">پس از این مدت، امکان رای دادن و افزودن زمان بسته می‌شود.</p>
          </div>
        </div>

        <div id="section-dynamic" class="hidden space-y-4 rounded-lg bg-blue-50 p-4 dark:bg-blue-950/30">
          ${datePicker('dyn_date', 'تاریخ جلسه')}
          <div class="flex items-center justify-center gap-4">
            ${timePicker('dyn_min', 'از ساعت', 9)}
            <span class="mt-5 text-gray-400" aria-hidden="true">←</span>
            ${timePicker('dyn_max', 'تا ساعت', 17)}
          </div>
        </div>

        <button type="button" data-action="create" class="btn-success w-full py-3 text-lg font-bold">ایجاد جلسه</button>
      </div>
    </section>`;

  addRow();
  addRow();
}

function addRow() {
  const id = ++rowCounter;
  const li = document.createElement('li');
  li.className = 'rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-gray-700 dark:bg-gray-900/40';
  li.dataset.row = String(id);
  li.innerHTML = `
    <div class="mb-3 flex items-end gap-2">
      <div class="flex-1">${datePicker(`ts_${id}_date`, `تاریخ زمان ${faNum(id)}`)}</div>
      <button type="button" data-action="remove-row" data-row="${id}" class="btn-ghost px-3 text-red-600 dark:text-red-400" aria-label="حذف این زمان">✕</button>
    </div>
    <div class="flex items-center justify-center gap-4">
      ${timePicker(`ts_${id}_start`, 'شروع', 10)}
      <span class="mt-5 text-gray-400" aria-hidden="true">←</span>
      ${timePicker(`ts_${id}_end`, 'پایان', 11)}
    </div>`;
  $('slot-rows').appendChild(li);
}

onAction('add-row', addRow);

onAction('remove-row', (el) => {
  const row = el.closest('li');
  if (row) row.remove();
});

onChange('session-type', (el) => {
  const type = el.value;
  for (const t of TYPES) $(`section-${t.value}`).classList.toggle('hidden', t.value !== type);
  $('type-help').textContent = TYPES.find((t) => t.value === type).help;
});

function fail(message, focusId) {
  toast(message, 'warning');
  if (focusId) $(focusId).focus();
  return null;
}

function buildPayload() {
  const title = $('session-title').value.trim();
  const creator = $('creator-name').value.trim();
  if (!title) return fail('لطفاً عنوان جلسه را وارد کنید', 'session-title');
  if (!creator) return fail('لطفاً نام خود را وارد کنید', 'creator-name');

  const type = document.querySelector('input[name="session-type"]:checked').value;
  const payload = { title, creator_name: creator, type, timeslots: [] };

  if (type === 'fixed') {
    for (const row of $('slot-rows').children) {
      const id = row.dataset.row;
      const date = readDate(`ts_${id}_date`);
      const start = readTime(`ts_${id}_start`);
      const end = readTime(`ts_${id}_end`);
      if (!isValidJalali(date)) return fail('یکی از تاریخ‌ها معتبر نیست');
      if (start >= end) return fail('در هر زمان، ساعت شروع باید قبل از ساعت پایان باشد');
      payload.timeslots.push({ start_utc: jalaliToISO(date, start), end_utc: jalaliToISO(date, end) });
    }
    if (payload.timeslots.length < 2) return fail('لطفاً حداقل دو زمان مشخص کنید تا شرکت‌کنندگان حق انتخاب داشته باشند');
  } else if (type === 'weekly') {
    const days = [...document.querySelectorAll('input[name="week-day"]:checked')].map((cb) => Number(cb.value));
    if (!days.length) return fail('لطفاً حداقل یک روز هفته را انتخاب کنید');
    const min = readTime('weekly_start');
    const max = readTime('weekly_end');
    if (min >= max) return fail('ساعت شروع باید قبل از ساعت پایان باشد');
    const weeks = Math.min(52, Math.max(1, parseInt($('weeks').value, 10) || 4));
    payload.expires_at_utc = new Date(Date.now() + weeks * 7 * 24 * 3600 * 1000).toISOString();
    payload.dynamic_config = { min_time: min, max_time: max, allowed_days: days };
  } else {
    const date = readDate('dyn_date');
    const min = readTime('dyn_min');
    const max = readTime('dyn_max');
    if (!isValidJalali(date)) return fail('تاریخ جلسه معتبر نیست');
    if (min >= max) return fail('ساعت شروع باید قبل از ساعت پایان باشد');
    payload.dynamic_config = { date_utc: jalaliToUTCMidnight(date), min_time: min, max_time: max };
  }
  return payload;
}

onAction('create', async (el) => {
  const payload = buildPayload();
  if (!payload) return;
  el.disabled = true;
  try {
    const data = await api('POST', '/sessions', payload);
    // The creator lands on the session with their name already filled in.
    saveName(data.id, payload.creator_name);
    location.href = `/${encodeURIComponent(data.id)}`;
  } catch (err) {
    toast(err.message, 'error');
    el.disabled = false;
  }
});

