export const $ = (id) => document.getElementById(id);

const ESC = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' };

// Every piece of user-provided data inserted into an HTML template must go through esc().
export const esc = (value) => String(value ?? '').replace(/[&<>"']/g, (c) => ESC[c]);

const faFormat = new Intl.NumberFormat('fa-IR', { useGrouping: false });
export const faNum = (n) => faFormat.format(n);

// ---- Click / change delegation -------------------------------------------
// Views render elements with data-action="name" (clicks) or data-change="name"
// (form changes) and register handlers here. No inline onclick, so the page
// works under a strict Content-Security-Policy.

const clickHandlers = {};
const changeHandlers = {};

export function onAction(name, fn) {
  clickHandlers[name] = fn;
}

export function onChange(name, fn) {
  changeHandlers[name] = fn;
}

export function bindDelegation(root) {
  root.addEventListener('click', (e) => {
    const el = e.target.closest('[data-action]');
    if (el && root.contains(el) && clickHandlers[el.dataset.action]) {
      clickHandlers[el.dataset.action](el, e);
    }
  });
  root.addEventListener('change', (e) => {
    const el = e.target.closest('[data-change]');
    if (el && changeHandlers[el.dataset.change]) changeHandlers[el.dataset.change](el, e);
  });
}

// ---- Toasts ---------------------------------------------------------------

const toastRoot = document.createElement('div');
toastRoot.className = 'pointer-events-none fixed inset-x-0 bottom-4 z-50 flex flex-col items-center gap-2 px-4';
toastRoot.setAttribute('role', 'status');
toastRoot.setAttribute('aria-live', 'polite');
document.body.appendChild(toastRoot);

const toastColors = {
  info: 'bg-blue-600',
  success: 'bg-green-600',
  error: 'bg-red-600',
  warning: 'bg-amber-600',
};

function mountToast(toast) {
  toast.className += ' pointer-events-auto w-full max-w-sm rounded-lg px-5 py-3 text-white shadow-lg transition-all duration-300 translate-y-4 opacity-0';
  toastRoot.appendChild(toast);
  requestAnimationFrame(() => toast.classList.remove('translate-y-4', 'opacity-0'));
}

function dismiss(toast) {
  toast.classList.add('translate-y-4', 'opacity-0');
  setTimeout(() => toast.remove(), 300);
}

export function toast(message, type = 'info') {
  const el = document.createElement('div');
  el.className = toastColors[type] || toastColors.info;
  el.textContent = message;
  mountToast(el);
  setTimeout(() => dismiss(el), 3500);
}

export function confirmToast(message, confirmLabel, onConfirm) {
  const el = document.createElement('div');
  el.className = 'bg-gray-800 dark:bg-gray-700 flex flex-col gap-3';
  el.setAttribute('role', 'alertdialog');

  const text = document.createElement('p');
  text.textContent = message;

  const buttons = document.createElement('div');
  buttons.className = 'flex justify-end gap-2';
  const yes = document.createElement('button');
  yes.type = 'button';
  yes.className = 'rounded bg-red-600 px-3 py-1 text-sm font-bold hover:bg-red-700';
  yes.textContent = confirmLabel;
  const no = document.createElement('button');
  no.type = 'button';
  no.className = 'rounded bg-white/10 px-3 py-1 text-sm hover:bg-white/20';
  no.textContent = 'لغو';

  yes.addEventListener('click', () => {
    dismiss(el);
    onConfirm();
  });
  no.addEventListener('click', () => dismiss(el));

  buttons.append(no, yes);
  el.append(text, buttons);
  mountToast(el);
  no.focus();
}

// ---- Storage ----------------------------------------------------------------
// localStorage can throw (private mode, blocked storage); the app must keep working.

function store(key, value) {
  try {
    if (value) localStorage.setItem(key, value);
    else localStorage.removeItem(key);
  } catch {
    /* ignore */
  }
}

function read(key) {
  try {
    return localStorage.getItem(key) || '';
  } catch {
    return '';
  }
}

// Tokens are signed by the server and replace the password for later edits,
// so the password itself is never stored in the browser.
export const saveToken = (sessionId, name, token) => store(`bm:token:${sessionId}:${name}`, token);
export const loadToken = (sessionId, name) => (name ? read(`bm:token:${sessionId}:${name}`) : '');
export const saveName = (sessionId, name) => store(`bm:name:${sessionId}`, name);
export const loadName = (sessionId) => read(`bm:name:${sessionId}`);

// Older versions stored plaintext passwords under pwd_<session>_<name>.
export function purgeLegacyPasswords() {
  try {
    for (const key of Object.keys(localStorage)) {
      if (key.startsWith('pwd_')) localStorage.removeItem(key);
    }
  } catch {
    /* ignore */
  }
}
