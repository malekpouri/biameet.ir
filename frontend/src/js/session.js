import { api } from './api.js';
import { DAY_NAMES, WEEK_ORDER, dayName, faTime, formatDate, formatTime, readTime, timePicker } from './dates.js';
import { $, confirmToast, esc, faNum, loadName, loadToken, onAction, saveName, saveToken, toast } from './util.js';

let app;
let sessionId = '';
let session = null;
let selected = new Set();
let voterName = '';
let voterPassword = '';

// Selection styles are swapped in place on toggle instead of re-rendering the page.
const SLOT_ON = ['border-blue-500', 'bg-blue-50', 'dark:bg-blue-900/30'];
const SLOT_OFF = ['border-gray-200', 'dark:border-gray-700', 'hover:bg-gray-50', 'dark:hover:bg-gray-700/40'];
const CHECK_ON = ['border-blue-600', 'bg-blue-600', 'text-white'];
const CHECK_OFF = ['border-gray-300', 'text-transparent', 'dark:border-gray-600'];

export async function showSession(root, id) {
  app = root;
  sessionId = id;
  voterName = loadName(id);
  await refresh(true);
}

export function renderMessage(root, title, text) {
  root.innerHTML = `
    <div class="card p-8 text-center">
      <h1 class="mb-2 text-xl font-bold text-gray-900 dark:text-white">${esc(title)}</h1>
      <p class="mb-6 text-gray-600 dark:text-gray-400">${esc(text)}</p>
      <a href="/" class="btn-primary">ساخت جلسه جدید</a>
    </div>`;
}

async function refresh(initial = false) {
  try {
    session = await api('GET', `/sessions/${encodeURIComponent(sessionId)}`);
  } catch (err) {
    if (initial) renderMessage(app, 'جلسه در دسترس نیست', err.message);
    else toast(err.message, 'error');
    return;
  }

  const ids = new Set(session.timeslots.map((ts) => ts.id));
  for (const id of selected) if (!ids.has(id)) selected.delete(id);
  if (initial && voterName) preselect();
  render();
}

// Load the votes already cast under the typed name.
function preselect() {
  const mine = session.timeslots.filter((ts) => ts.votes.some((v) => v.voter_name === voterName));
  if (mine.length) selected = new Set(mine.map((ts) => ts.id));
}

function slotDate(ts) {
  return session.type === 'weekly' ? dayName(new Date(ts.start_utc)) : formatDate(ts.start_utc);
}

function slotLabel(ts) {
  return `${slotDate(ts)}، ${formatTime(ts.start_utc)} تا ${formatTime(ts.end_utc)}`;
}

function guide() {
  const { type, dynamic_config: cfg } = session;
  const passwordHint = 'برای اینکه بعداً بتوانید رای خود را ویرایش یا حذف کنید، یک رمز عبور انتخاب کنید.';
  if (type === 'dynamic') {
    return `
      <div class="note mb-5 border-blue-200 bg-blue-50 text-blue-900 dark:border-blue-900 dark:bg-blue-950/40 dark:text-blue-200">
        <p class="mb-1"><strong>تاریخ:</strong> ${formatDate(cfg.date_utc, true)} — <strong>بازه مجاز:</strong> ${faTime(cfg.min_time)} تا ${faTime(cfg.max_time)}</p>
        <p>زمان‌های موجود را انتخاب کنید یا زمان پیشنهادی خود را در این بازه اضافه کنید. ${passwordHint}</p>
      </div>`;
  }
  if (type === 'weekly') {
    const days = WEEK_ORDER.filter((d) => cfg.allowed_days.includes(d)).map((d) => DAY_NAMES[d]).join('، ');
    return `
      <div class="note mb-5 border-purple-200 bg-purple-50 text-purple-900 dark:border-purple-900 dark:bg-purple-950/40 dark:text-purple-200">
        <p class="mb-1"><strong>روزهای مجاز:</strong> ${days} — <strong>بازه مجاز:</strong> ${faTime(cfg.min_time)} تا ${faTime(cfg.max_time)}</p>
        <p>زمان‌های موجود را انتخاب کنید یا در یکی از این روزها زمان جدیدی پیشنهاد دهید. ${passwordHint}</p>
      </div>`;
  }
  return `
    <div class="note mb-5 border-gray-200 bg-gray-50 text-gray-700 dark:border-gray-700 dark:bg-gray-900/40 dark:text-gray-300">
      زمان‌های مناسب خود را انتخاب کنید و «ثبت رای» را بزنید. ${passwordHint}
    </div>`;
}

function slotItem(ts, maxVotes, open) {
  const count = ts.votes.length;
  const on = selected.has(ts.id);
  const best = count > 0 && count === maxVotes;
  const pct = maxVotes ? Math.round((count / maxVotes) * 100) : 0;
  // The server has the final say; this only hides the button where it would certainly fail.
  const deletable = open && session.type !== 'fixed' && ts.votes.every((v) => v.voter_name === ts.created_by);

  return `
    <li class="flex items-stretch gap-2">
      <button type="button" data-action="toggle" data-id="${esc(ts.id)}" aria-pressed="${on}" ${open ? '' : 'disabled'}
        class="block flex-1 rounded-lg border-2 p-3 text-right transition-colors disabled:cursor-default ${(on ? SLOT_ON : SLOT_OFF).join(' ')}">
        <span class="flex items-start justify-between gap-3">
          <span class="block">
            <span class="block text-sm text-gray-500 dark:text-gray-400">${slotDate(ts)}</span>
            <span class="block font-bold text-gray-900 dark:text-white">${formatTime(ts.start_utc)} تا ${formatTime(ts.end_utc)}</span>
            ${ts.created_by ? `<span class="mt-1 block text-xs text-gray-500 dark:text-gray-400">پیشنهاد: ${esc(ts.created_by)}</span>` : ''}
          </span>
          <span class="flex shrink-0 items-center gap-2">
            ${best ? '<span class="rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-800 dark:bg-green-900/60 dark:text-green-300">بیشترین رای</span>' : ''}
            <span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-700 dark:bg-gray-700 dark:text-gray-200">${faNum(count)} رای</span>
            <span class="check flex h-6 w-6 items-center justify-center rounded-full border-2 text-sm ${(on ? CHECK_ON : CHECK_OFF).join(' ')}" aria-hidden="true">✓</span>
          </span>
        </span>
        <span class="mt-3 block h-1.5 overflow-hidden rounded-full bg-gray-100 dark:bg-gray-700" aria-hidden="true">
          <span class="block h-full rounded-full ${best ? 'bg-green-500' : 'bg-blue-400'}" style="width:${pct}%"></span>
        </span>
        ${count ? `<span class="mt-2 block text-xs text-gray-600 dark:text-gray-400"><span class="font-medium">رای‌دهندگان:</span> ${ts.votes.map((v) => esc(v.voter_name)).join('، ')}</span>` : ''}
      </button>
      ${deletable ? `<button type="button" data-action="delete" data-id="${esc(ts.id)}" class="btn-ghost px-3 text-red-600 dark:text-red-400" aria-label="حذف زمان ${esc(slotLabel(ts))}"><span aria-hidden="true">🗑</span></button>` : ''}
    </li>`;
}

function proposeForm() {
  const { type, dynamic_config: cfg } = session;
  const [minH, minM] = cfg.min_time.split(':').map(Number);
  const startMin = Math.floor(minM / 15) * 15;
  const days = WEEK_ORDER.filter((d) => cfg.allowed_days && cfg.allowed_days.includes(d))
    .map((d) => `<option value="${d}">${DAY_NAMES[d]}</option>`).join('');

  return `
    <section class="mt-6 border-t border-gray-200 pt-6 dark:border-gray-700" aria-labelledby="propose-title">
      <h2 id="propose-title" class="mb-4 font-bold text-gray-900 dark:text-white">پیشنهاد زمان جدید</h2>
      ${type === 'weekly' ? `<label for="weekly-day" class="label">روز هفته</label><select id="weekly-day" class="field mb-4">${days}</select>` : ''}
      <div class="mb-4 flex items-center justify-center gap-4">
        ${timePicker('new_start', 'شروع', minH, startMin)}
        <span class="mt-5 text-gray-400" aria-hidden="true">←</span>
        ${timePicker('new_end', 'پایان', Math.min(minH + 1, 23), startMin)}
      </div>
      <button type="button" data-action="add-slot" class="btn-primary w-full">+ افزودن زمان</button>
      <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">زمانی که پیشنهاد می‌دهید خودکار با نام شما رای می‌گیرد.</p>
    </section>`;
}

function render() {
  const { title, creator_name, type, timeslots, expired, expires_at_utc } = session;
  const open = !expired;
  const maxVotes = timeslots.reduce((m, ts) => Math.max(m, ts.votes.length), 0);
  const best = maxVotes ? timeslots.filter((ts) => ts.votes.length === maxVotes) : [];
  const hasToken = !!loadToken(sessionId, voterName);

  app.innerHTML = `
    <article class="card p-5 sm:p-6">
      <div class="mb-1 flex items-start justify-between gap-3">
        <h1 class="min-w-0 break-words text-2xl font-bold text-gray-900 dark:text-white">${esc(title)}</h1>
        <button type="button" data-action="share" class="btn-ghost shrink-0 px-3 text-sm">
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M8.7 10.7l6.6-3.4M8.7 13.3l6.6 3.4M18 5a3 3 0 11-6 0 3 3 0 016 0zm0 14a3 3 0 11-6 0 3 3 0 016 0zM9 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
          اشتراک‌گذاری
        </button>
      </div>
      <p class="mb-5 text-sm text-gray-600 dark:text-gray-400">ایجاد شده توسط ${esc(creator_name)}</p>

      ${expired
        ? '<div class="note mb-5 border-red-200 bg-red-50 text-red-800 dark:border-red-900 dark:bg-red-950/40 dark:text-red-300">مهلت این جلسه به پایان رسیده است و امکان رای دادن یا افزودن زمان وجود ندارد.</div>'
        : guide()}
      ${expires_at_utc && !expired ? `<p class="-mt-3 mb-5 text-xs text-gray-500 dark:text-gray-400">این لینک تا ${formatDate(expires_at_utc)} معتبر است.</p>` : ''}

      ${open ? `
      <div class="mb-6 grid gap-4 sm:grid-cols-2">
        <div>
          <label for="voter-name" class="label">نام شما</label>
          <input id="voter-name" class="field" maxlength="64" autocomplete="nickname" placeholder="نام خود را وارد کنید" value="${esc(voterName)}">
        </div>
        <div>
          <label for="voter-password" class="label">رمز عبور <span class="font-normal text-gray-500 dark:text-gray-400">(اختیاری)</span></label>
          <input id="voter-password" type="password" class="field" maxlength="72" autocomplete="off"
            placeholder="${hasToken ? 'در این مرورگر ذخیره شده است' : 'برای ویرایش بعدی رای'}" value="${esc(voterPassword)}">
        </div>
      </div>` : ''}

      ${best.length ? `
      <div class="note mb-4 border-green-200 bg-green-50 text-green-900 dark:border-green-900 dark:bg-green-950/40 dark:text-green-200">
        <strong>بیشترین رای (${faNum(maxVotes)}):</strong> ${best.map(slotLabel).join(' / ')}
      </div>` : ''}

      <h2 class="mb-3 font-bold text-gray-900 dark:text-white">زمان‌ها <span class="text-sm font-normal text-gray-500 dark:text-gray-400">(${faNum(timeslots.length)})</span></h2>
      ${timeslots.length
        ? `<ul class="space-y-3">${timeslots.map((ts) => slotItem(ts, maxVotes, open)).join('')}</ul>`
        : '<p class="rounded-lg border-2 border-dashed border-gray-200 p-6 text-center text-sm text-gray-500 dark:border-gray-700 dark:text-gray-400">هنوز زمانی پیشنهاد نشده است.</p>'}

      ${open && type !== 'fixed' ? proposeForm() : ''}

      ${open ? '<button type="button" data-action="vote" class="btn-success mt-6 w-full py-3 text-lg font-bold">ثبت / ویرایش رای</button>' : ''}
    </article>`;

  const nameInput = $('voter-name');
  if (nameInput) {
    nameInput.addEventListener('input', () => {
      voterName = nameInput.value.trim();
      const mine = session.timeslots.filter((ts) => ts.votes.some((v) => v.voter_name === voterName));
      if (mine.length) {
        selected = new Set(mine.map((ts) => ts.id));
        app.querySelectorAll('[data-action="toggle"]').forEach((b) => paint(b, selected.has(b.dataset.id)));
      }
      $('voter-password').placeholder = loadToken(sessionId, voterName) ? 'در این مرورگر ذخیره شده است' : 'برای ویرایش بعدی رای';
    });
    $('voter-password').addEventListener('input', (e) => {
      voterPassword = e.target.value;
    });
  }
}

function paint(button, on) {
  button.setAttribute('aria-pressed', String(on));
  button.classList.remove(...(on ? SLOT_OFF : SLOT_ON));
  button.classList.add(...(on ? SLOT_ON : SLOT_OFF));
  const check = button.querySelector('.check');
  check.classList.remove(...(on ? CHECK_OFF : CHECK_ON));
  check.classList.add(...(on ? CHECK_ON : CHECK_OFF));
}

function requireName() {
  if (voterName) return true;
  toast('لطفاً نام خود را وارد کنید', 'warning');
  $('voter-name').focus();
  return false;
}

// Credentials for acting as `name`: the typed password, else the stored token.
function credentials(name) {
  return { password: voterPassword, token: loadToken(sessionId, name) };
}

function afterAuthSuccess(name, token) {
  if (token) {
    saveToken(sessionId, name, token);
    voterPassword = ''; // the token replaces it from now on
  }
  saveName(sessionId, name);
}

function showError(err, name) {
  toast(err.message, 'error');
  if (err.code === 'password_required' || err.code === 'invalid_password') {
    // A stored token that stopped working is useless; forget it.
    if (err.code === 'password_required' && name) saveToken(sessionId, name, '');
    const field = $('voter-password');
    if (field) field.focus();
  }
}

async function busy(button, fn) {
  if (button.disabled) return;
  button.disabled = true;
  try {
    await fn();
  } finally {
    if (button.isConnected) button.disabled = false;
  }
}

async function sendVote(button, votes) {
  const name = voterName;
  await busy(button, async () => {
    try {
      const res = await api('POST', `/sessions/${sessionId}/vote`, { voter_name: name, votes, ...credentials(name) });
      afterAuthSuccess(name, res.token);
      toast(votes.length ? 'رای شما ثبت شد' : 'رای‌های شما حذف شد', 'success');
      await refresh();
    } catch (err) {
      showError(err, name);
    }
  });
}

onAction('toggle', (el) => {
  const id = el.dataset.id;
  if (selected.has(id)) selected.delete(id);
  else selected.add(id);
  paint(el, selected.has(id));
});

onAction('vote', (el) => {
  if (!requireName()) return;
  const votes = [...selected].map((id) => ({ timeslot_id: id }));
  if (votes.length === 0) {
    confirmToast('هیچ زمانی انتخاب نشده است. همه رای‌های شما حذف شود؟', 'بله، حذف کن', () => sendVote(el, votes));
    return;
  }
  sendVote(el, votes);
});

onAction('add-slot', (el) => {
  if (!requireName()) return;
  const cfg = session.dynamic_config;
  const start = readTime('new_start');
  const end = readTime('new_end');

  // "HH:MM" strings compare correctly as text.
  if (start < cfg.min_time || end > cfg.max_time) {
    toast(`زمان انتخابی باید بین ${faTime(cfg.min_time)} و ${faTime(cfg.max_time)} باشد`, 'warning');
    return;
  }
  if (start >= end) {
    toast('زمان شروع باید قبل از زمان پایان باشد', 'warning');
    return;
  }

  let day;
  if (session.type === 'dynamic') {
    // date_utc is the chosen calendar day at 00:00 UTC; use that day in local time.
    const d = new Date(cfg.date_utc);
    day = new Date(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate());
  } else {
    // Weekly slots are anchored to the first matching weekday on or after creation.
    const target = Number($('weekly-day').value);
    day = new Date(session.created_at_utc);
    day.setHours(0, 0, 0, 0);
    day.setDate(day.getDate() + ((target - day.getDay() + 7) % 7));
  }
  const at = (hhmm) => {
    const [h, m] = hhmm.split(':').map(Number);
    return new Date(day.getFullYear(), day.getMonth(), day.getDate(), h, m).toISOString();
  };

  const name = voterName;
  busy(el, async () => {
    try {
      const res = await api('POST', `/sessions/${sessionId}/timeslots`, {
        start_utc: at(start),
        end_utc: at(end),
        created_by: name,
        ...credentials(name),
      });
      afterAuthSuccess(name, res.token);
      selected.add(res.id);
      toast('زمان جدید اضافه شد', 'success');
      await refresh();
    } catch (err) {
      showError(err, name);
    }
  });
});

onAction('delete', (el) => {
  const ts = session.timeslots.find((t) => t.id === el.dataset.id);
  if (!ts) return;
  confirmToast(`زمان «${slotLabel(ts)}» حذف شود؟`, 'بله، حذف کن', async () => {
    try {
      await api('DELETE', `/sessions/${sessionId}/timeslots/${encodeURIComponent(ts.id)}`, credentials(ts.created_by || ''));
      selected.delete(ts.id);
      toast('زمان حذف شد', 'success');
      await refresh();
    } catch (err) {
      showError(err, '');
    }
  });
});

onAction('share', async () => {
  const url = location.origin + location.pathname;
  if (navigator.share) {
    try {
      await navigator.share({ title: session.title, text: `دعوت به جلسه: ${session.title}`, url });
      return;
    } catch (err) {
      if (err && err.name === 'AbortError') return;
    }
  }
  try {
    await navigator.clipboard.writeText(url);
    toast('لینک جلسه کپی شد', 'success');
  } catch {
    toast(`لینک جلسه: ${url}`, 'info');
  }
});
