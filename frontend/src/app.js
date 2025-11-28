const API_BASE = '/api/v1';

// State
let sessionData = null;
let selectedTimeslots = new Set();
let voterName = '';

// DOM Elements
const app = document.getElementById('app');

// Toast Container
const toastContainer = document.createElement('div');
toastContainer.className = 'fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 z-50 flex flex-col gap-2 w-full max-w-sm px-4';
document.body.appendChild(toastContainer);

// Theme Management


function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    const colors = {
        info: 'bg-blue-600',
        success: 'bg-green-600',
        error: 'bg-red-600',
        warning: 'bg-yellow-600'
    };
    toast.className = `${colors[type]} text-white px-6 py-3 rounded shadow-lg transform transition-all duration-300 translate-y-10 opacity-0`;
    toast.innerText = message;

    toastContainer.appendChild(toast);

    // Animate in
    requestAnimationFrame(() => {
        toast.classList.remove('translate-y-10', 'opacity-0');
    });

    // Remove after 3s
    setTimeout(() => {
        toast.classList.add('translate-y-10', 'opacity-0');
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// Utils
function getSessionIDFromURL() {
    const path = window.location.pathname;
    if (path === '/admin') return 'admin';
    const id = path.substring(1);
    if (id && /^[a-zA-Z0-9]{5}$/.test(id)) {
        return id;
    }
    return null;
}

function formatJalali(utcDateString) {
    if (!utcDateString) return '';
    const date = new Date(utcDateString);
    const j = jalaali.toJalaali(date);
    const time = date.toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit', hour12: false });
    return `${j.jy}/${j.jm}/${j.jd} ${time}`;
}

function formatTime(utcDateString) {
    if (!utcDateString) return '';
    const date = new Date(utcDateString);
    return date.toLocaleTimeString('fa-IR', { hour: '2-digit', minute: '2-digit', hour12: false });
}

// UI Helpers
function getJalaliMonths() {
    return ['فروردین', 'اردیبهشت', 'خرداد', 'تیر', 'مرداد', 'شهریور', 'مهر', 'آبان', 'آذر', 'دی', 'بهمن', 'اسفند'];
}

function getDayName(date) {
    const days = ['یکشنبه', 'دوشنبه', 'سه‌شنبه', 'چهارشنبه', 'پنج‌شنبه', 'جمعه', 'شنبه'];
    return days[date.getDay()];
}

function formatJalaliDate(utcDateString) {
    if (!utcDateString) return '';
    const date = new Date(utcDateString);
    const j = jalaali.toJalaali(date);
    const dayName = getDayName(date);
    return `${dayName} ${j.jy}/${j.jm}/${j.jd}`;
}

function renderTimePicker(prefix, defaultHour = 9, defaultMinute = 0, minHour = 0, maxHour = 23) {
    let hourOpts = '';
    for (let i = minHour; i <= maxHour; i++) {
        const val = String(i).padStart(2, '0');
        hourOpts += `<option value="${val}" ${i === defaultHour ? 'selected' : ''}>${val}</option>`;
    }

    let minOpts = '';
    for (let i = 0; i < 60; i += 15) {
        const val = String(i).padStart(2, '0');
        minOpts += `<option value="${val}" ${i === defaultMinute ? 'selected' : ''}>${val}</option>`;
    }

    return `
        <div class="flex gap-1 items-center justify-center bg-white dark:bg-gray-700 border dark:border-gray-600 rounded p-1" dir="ltr">
            <select id="${prefix}_hour" class="p-1 bg-transparent text-center font-mono w-12 outline-none dark:text-white">${hourOpts}</select>
            <span class="text-gray-500 dark:text-gray-400">:</span>
            <select id="${prefix}_minute" class="p-1 bg-transparent text-center font-mono w-12 outline-none dark:text-white">${minOpts}</select>
        </div>
    `;
}

function renderJalaliDatePicker(prefix, defaultDate = new Date()) {
    const j = jalaali.toJalaali(defaultDate);
    const months = getJalaliMonths();

    let daysOpts = '';
    for (let i = 1; i <= 31; i++) daysOpts += `<option value="${i}" ${i === j.jd ? 'selected' : ''}>${i}</option>`;

    let monthsOpts = '';
    months.forEach((m, i) => {
        monthsOpts += `<option value="${i + 1}" ${i + 1 === j.jm ? 'selected' : ''}>${m}</option>`;
    });

    let yearsOpts = '';
    const currentYear = j.jy;
    for (let i = currentYear; i <= currentYear + 1; i++) yearsOpts += `<option value="${i}" ${i === j.jy ? 'selected' : ''}>${i}</option>`;

    return `
        <div class="flex gap-1 items-center" dir="rtl">
            <select id="${prefix}_day" class="p-2 border rounded bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-white text-center text-sm flex-1">${daysOpts}</select>
            <select id="${prefix}_month" class="p-2 border rounded bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-white text-center text-sm flex-1">${monthsOpts}</select>
            <select id="${prefix}_year" class="p-2 border rounded bg-white dark:bg-gray-700 dark:border-gray-600 dark:text-white text-center text-sm flex-1">${yearsOpts}</select>
        </div>
    `;
}

function getJalaliDateFromPicker(prefix) {
    const y = parseInt(document.getElementById(`${prefix}_year`).value);
    const m = parseInt(document.getElementById(`${prefix}_month`).value);
    const d = parseInt(document.getElementById(`${prefix}_day`).value);
    return { y, m, d };
}

function getTimeFromPicker(prefix) {
    const h = document.getElementById(`${prefix}_hour`).value;
    const m = document.getElementById(`${prefix}_minute`).value;
    return `${h}:${m}`;
}

function jalaliToISO(jDate, timeStr) {
    const g = jalaali.toGregorian(jDate.y, jDate.m, jDate.d);
    // Time is HH:MM
    const [h, m] = timeStr.split(':').map(Number);
    // Create Date object (local time)
    const date = new Date(g.gy, g.gm - 1, g.gd, h, m);
    return date.toISOString();
}

// API Calls
async function fetchSession(id) {
    try {
        const res = await fetch(`${API_BASE}/sessions/${id}`);
        if (!res.ok) throw new Error('Session not found');
        sessionData = await res.json();
        renderSession();
    } catch (err) {
        app.innerHTML = `<div class="text-red-500 text-center mt-10">${err.message}</div>`;
    }
}

async function fetchAdminStats() {
    try {
        const res = await fetch(`${API_BASE}/admin/stats`);
        if (!res.ok) throw new Error('Failed to fetch stats');
        const stats = await res.json();
        renderAdminDashboard(stats);
    } catch (err) {
        showToast(err.message, 'error');
    }
}

async function submitVote() {
    if (!voterName) {
        showToast('لطفاً نام خود را وارد کنید', 'warning');
        return;
    }
    if (selectedTimeslots.size === 0) {
        showToast('لطفاً حداقل یک زمان را انتخاب کنید', 'warning');
        return;
    }

    const password = document.getElementById('voterPasswordInput').value;

    const votes = Array.from(selectedTimeslots).map(id => ({
        timeslot_id: id,
        note: 'Yes'
    }));

    try {
        const res = await fetch(`${API_BASE}/sessions/${sessionData.id}/vote`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                voter_name: voterName,
                password: password,
                votes: votes
            })
        });

        if (!res.ok) {
            const err = await res.json();
            if (err.error === 'password_required') {
                showToast('برای ویرایش رای، لطفاً رمز عبور خود را وارد کنید', 'warning');
                document.getElementById('voterPasswordInput').focus();
                return;
            }
            if (err.error === 'invalid_password') {
                showToast('رمز عبور اشتباه است', 'error');
                return;
            }
            throw new Error(err.error || 'خطا در ثبت رای');
        }

        if (password) {
            localStorage.setItem(`pwd_${sessionData.id}_${voterName}`, password);
        }

        showToast('رای شما با موفقیت ثبت شد', 'success');
        // Refresh data without reload
        fetchSession(sessionData.id);
        selectedTimeslots.clear();
    } catch (err) {
        showToast(err.message, 'error');
    }
}

// Renderers
function renderSession() {
    if (!sessionData) return;

    const { title, creator_name, timeslots = [], type, dynamic_config } = sessionData;

    const voteCounts = {};
    timeslots.forEach(ts => {
        voteCounts[ts.id] = (ts.votes || []).length;
    });

    let dynamicHeader = '';
    let dynamicInput = '';

    if (type === 'dynamic' && dynamic_config) {
        const date = new Date(dynamic_config.date_utc);
        const j = jalaali.toJalaali(date);
        dynamicHeader = `
            <div class="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg mb-6 border border-blue-100 dark:border-blue-800">
                <div class="text-center">
                    <span class="block text-gray-500 dark:text-gray-400 text-sm mb-1">تاریخ انتخابی</span>
                    <span class="text-xl font-bold text-blue-800 dark:text-blue-300">${j.jy}/${j.jm}/${j.jd}</span>
                </div>
                <div class="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
                    بازه مجاز: ${dynamic_config.min_time} تا ${dynamic_config.max_time}
                </div>
            </div>
        `;

        const minH = parseInt(dynamic_config.min_time.split(':')[0]);
        const maxTimeParts = dynamic_config.max_time.split(':');
        let maxH = parseInt(maxTimeParts[0]);
        // If max time is exactly on the hour (e.g. 17:00), the start time cannot be 17:xx.
        // So we reduce the max hour for the start picker by 1.
        // If max time is 17:30, start time can be 17:00 or 17:15, so we keep 17.
        const maxStartH = parseInt(maxTimeParts[1]) === 0 ? maxH - 1 : maxH;

        dynamicInput = `
            <div class="mt-8 border-t pt-6">
                <h3 class="font-semibold text-gray-700 dark:text-gray-300 mb-4">پیشنهاد زمان جدید</h3>
                <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded border dark:border-gray-600">
                    <div class="flex gap-4 items-center justify-center mb-4">
                        <div class="flex flex-col items-center">
                            <span class="text-xs text-gray-500 dark:text-gray-400 mb-1">از ساعت</span>
                            ${renderTimePicker('dynamic_start', 10, 0, minH, maxStartH)}
                        </div>
                        <div class="text-gray-400 mt-4">←</div>
                        <div class="flex flex-col items-center">
                            <span class="text-xs text-gray-500 dark:text-gray-400 mb-1">تا ساعت</span>
                            ${renderTimePicker('dynamic_end', 11, 0, minH, maxH)}
                        </div>
                    </div>
                    <button onclick="submitDynamicTimeslot()" class="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 text-sm font-bold">
                        افزودن زمان
                    </button>
                </div>
            </div>
        `;
    } else if (type === 'weekly' && dynamic_config) {
        const daysMap = ['یکشنبه', 'دوشنبه', 'سه‌شنبه', 'چهارشنبه', 'پنج‌شنبه', 'جمعه', 'شنبه'];
        const allowedDays = (dynamic_config.allowed_days || []).map(d => daysMap[d]).join('، ');

        const dateOptions = (dynamic_config.allowed_days || []).map(dayIdx => {
            return `<option value="${dayIdx}">${daysMap[dayIdx]}</option>`;
        }).join('');

        dynamicHeader = `
            <div class="bg-purple-50 dark:bg-purple-900/20 p-4 rounded-lg mb-6 border border-purple-100 dark:border-purple-800">
                <div class="text-center">
                    <span class="block text-gray-500 dark:text-gray-400 text-sm mb-1">الگوی هفتگی</span>
                    <span class="text-lg font-bold text-purple-800 dark:text-purple-300">${allowedDays}</span>
                </div>
                <div class="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
                    بازه مجاز: ${dynamic_config.min_time} تا ${dynamic_config.max_time}
                </div>
                <div class="mt-2 text-xs text-gray-500 dark:text-gray-400 text-center">
                    لطفاً یکی از روزهای مجاز را انتخاب کنید و ساعت پیشنهادی خود را وارد نمایید.
                </div>
            </div>
        `;

        const minH = parseInt(dynamic_config.min_time.split(':')[0]);
        const maxTimeParts = dynamic_config.max_time.split(':');
        let maxH = parseInt(maxTimeParts[0]);
        const maxStartH = parseInt(maxTimeParts[1]) === 0 ? maxH - 1 : maxH;

        dynamicInput = `
            <div class="mt-8 border-t pt-6">
                <h3 class="font-semibold text-gray-700 dark:text-gray-300 mb-4">پیشنهاد زمان جدید</h3>
                <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded border dark:border-gray-600">
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">انتخاب روز</label>
                        <select id="weekly_date_select" class="w-full p-2 border rounded dark:bg-gray-600 dark:border-gray-500 dark:text-white">
                            ${dateOptions}
                        </select>
                    </div>
                    <div class="flex gap-4 items-center justify-center mb-4">
                        <div class="flex flex-col items-center">
                            <span class="text-xs text-gray-500 dark:text-gray-400 mb-1">از ساعت</span>
                            ${renderTimePicker('dynamic_start', 10, 0, minH, maxStartH)}
                        </div>
                        <div class="text-gray-400 mt-4">←</div>
                        <div class="flex flex-col items-center">
                            <span class="text-xs text-gray-500 dark:text-gray-400 mb-1">تا ساعت</span>
                            ${renderTimePicker('dynamic_end', 11, 0, minH, maxH)}
                        </div>
                    </div>
                    <button onclick="submitDynamicTimeslot()" class="w-full bg-purple-600 text-white py-2 rounded hover:bg-purple-700 text-sm font-bold">
                        افزودن زمان
                    </button>
                </div>
            </div>
        `;
    }

    app.innerHTML = `
        <div class="max-w-2xl mx-auto bg-white dark:bg-gray-800 p-6 rounded-lg shadow">
            <div class="flex justify-between items-start mb-4">
                <h2 class="text-2xl font-bold text-gray-800 dark:text-white flex-1 text-center">${title}</h2>
                <div class="flex gap-2">
                    <button onclick="copyLink()" class="text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 p-2 rounded-full transition-colors" title="کپی لینک">
                        📋
                    </button>                    
                </div>
            </div>
            <p class="text-gray-600 dark:text-gray-400 mb-6 text-center">ایجاد شده توسط: ${creator_name}</p>

            ${type === 'fixed' ? `
                <div class="bg-gray-50 dark:bg-gray-700 p-3 rounded mb-6 text-sm text-gray-600 dark:text-gray-300 border dark:border-gray-600">
                    <span class="font-bold">راهنما:</span> لطفاً زمان‌های مناسب خود را از لیست زیر انتخاب کنید و دکمه ثبت رای را بزنید.
                </div>
            ` : ''}
            ${type === 'dynamic' ? `
                <div class="bg-blue-50 dark:bg-blue-900/20 p-3 rounded mb-6 text-sm text-blue-800 dark:text-blue-300 border border-blue-100 dark:border-blue-800">
                    <span class="font-bold">راهنما:</span> شما می‌توانید علاوه بر انتخاب زمان‌های موجود، زمان پیشنهادی خود را در بازه مجاز اضافه کنید.
                </div>
            ` : ''}
            ${type === 'weekly' ? `
                <div class="bg-purple-50 dark:bg-purple-900/20 p-3 rounded mb-6 text-sm text-purple-800 dark:text-purple-300 border border-purple-100 dark:border-purple-800">
                    <span class="font-bold">راهنما:</span> روزهای مجاز در بالا نمایش داده شده‌اند. می‌توانید از لیست پایین، یک روز خاص را انتخاب کرده و ساعت پیشنهادی خود را اضافه کنید.
                </div>
            ` : ''}

            ${dynamicHeader}

            <div class="mb-6">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">نام شما:</label>
                <input type="text" id="voterNameInput" class="w-full p-2 border rounded text-right dark:bg-gray-700 dark:border-gray-600 dark:text-white" placeholder="نام خود را وارد کنید...">
            </div>
            <div class="mb-6">
                 <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    رمز عبور (اختیاری)
                    <span class="text-xs text-gray-500 font-normal">- برای ویرایش بعدی الزامی است</span>
                 </label>
                 <input type="password" id="voterPasswordInput" class="w-full p-2 border rounded text-right dark:bg-gray-700 dark:border-gray-600 dark:text-white" placeholder="رمز عبور...">
            </div>

            <div class="space-y-3">
                <h3 class="font-semibold text-gray-700">زمان‌های موجود:</h3>
                ${timeslots.length === 0 ? '<p class="text-gray-400 text-sm text-center italic">هنوز زمانی ثبت نشده است</p>' : ''}
                ${timeslots.map(ts => {
        const isSelected = selectedTimeslots.has(ts.id);
        const count = voteCounts[ts.id] || 0;
        const voters = (ts.votes || []).map(v => v.voter_name);

        return `
                    <div class="timeslot-card border rounded p-3 cursor-pointer transition-colors ${isSelected ? 'bg-blue-50 border-blue-500' : 'hover:bg-gray-50'}"
                         onclick="toggleTimeslot('${ts.id}')">
                        <div class="flex flex-col sm:flex-row justify-between items-center gap-2">
                            <div>
                                <div class="text-sm text-gray-500 dark:text-gray-400 mb-1 text-right">
                                    ${type === 'weekly' ? getDayName(new Date(ts.start_utc)) : formatJalaliDate(ts.start_utc)}
                                </div>
                                <div class="font-bold text-gray-800 dark:text-white">${formatTime(ts.start_utc)} - ${formatTime(ts.end_utc)}</div>
                            </div>
                            <div class="flex items-center space-x-2 space-x-reverse">
                                <span class="bg-gray-200 text-gray-700 px-2 py-1 rounded text-xs">
                                    ${count} رای
                                </span>
                                ${isSelected ? '<span class="text-blue-600">✓</span>' : ''}
                            </div>
                        </div>
                        ${voters.length > 0 ? `
                        <div class="mt-2 text-xs text-gray-500 border-t pt-2">
                            <span class="font-semibold">رای‌دهندگان:</span> ${voters.join('، ')}
                        </div>
                        ` : ''}
                    </div>
                    `;
    }).join('')}
            </div>

            ${dynamicInput}

            <button onclick="submitVote()" class="mt-8 w-full bg-green-600 text-white py-3 rounded-lg hover:bg-green-700 font-bold shadow-lg transition-transform transform hover:scale-105">
                ثبت رای
            </button>
        </div>
    `;

    document.getElementById('voterNameInput').value = voterName;
    document.getElementById('voterNameInput').addEventListener('input', (e) => {
        voterName = e.target.value;
        const savedPwd = localStorage.getItem(`pwd_${sessionData.id}_${voterName}`);
        if (savedPwd) {
            document.getElementById('voterPasswordInput').value = savedPwd;
        }
    });
}

// Global handlers
window.toggleTimeslot = function (id) {
    if (selectedTimeslots.has(id)) {
        selectedTimeslots.delete(id);
    } else {
        selectedTimeslots.add(id);
    }
    renderSession();
};

window.copyLink = function () {
    const url = window.location.href;
    navigator.clipboard.writeText(url).then(() => {
        showToast('لینک کپی شد', 'success');
    }).catch(() => {
        showToast('خطا در کپی لینک', 'error');
    });
};

window.shareSession = function () {
    const url = window.location.href;
    if (navigator.share) {
        navigator.share({
            title: sessionData.title,
            text: `دعوت به جلسه: ${sessionData.title}`,
            url: url
        }).catch(console.error);
    } else {
        copyLink();
    }
};

window.submitDynamicTimeslot = async function () {
    const start = getTimeFromPicker('dynamic_start');
    const end = getTimeFromPicker('dynamic_end');

    // Construct dates
    let dateStr;
    if (sessionData.type === 'dynamic') {
        const sessionDate = new Date(sessionData.dynamic_config.date_utc);
        const yyyy = sessionDate.getFullYear();
        const mm = String(sessionDate.getMonth() + 1).padStart(2, '0');
        const dd = String(sessionDate.getDate()).padStart(2, '0');
        dateStr = `${yyyy}-${mm}-${dd}`;
    } else if (sessionData.type === 'weekly') {
        const dayIdx = parseInt(document.getElementById('weekly_date_select').value);
        if (isNaN(dayIdx)) {
            showToast('لطفاً یک روز را انتخاب کنید', 'warning');
            return;
        }

        // Find the first occurrence of this day on or after creation date to normalize
        const baseDate = new Date(sessionData.created_at_utc);
        baseDate.setHours(0, 0, 0, 0);

        // Calculate days to add
        // current day: baseDate.getDay()
        // target day: dayIdx
        let diff = dayIdx - baseDate.getDay();
        if (diff < 0) diff += 7;

        const targetDate = new Date(baseDate);
        targetDate.setDate(baseDate.getDate() + diff);

        const yyyy = targetDate.getFullYear();
        const mm = String(targetDate.getMonth() + 1).padStart(2, '0');
        const dd = String(targetDate.getDate()).padStart(2, '0');
        dateStr = `${yyyy}-${mm}-${dd}`;
    }

    const startDateTimeStr = `${dateStr}T${start}:00`;
    const endDateTimeStr = `${dateStr}T${end}:00`;

    const startUTC = new Date(startDateTimeStr).toISOString();
    const endUTC = new Date(endDateTimeStr).toISOString();

    try {
        // Validate range
        const minTimeParts = sessionData.dynamic_config.min_time.split(':').map(Number);
        const maxTimeParts = sessionData.dynamic_config.max_time.split(':').map(Number);
        const startParts = start.split(':').map(Number);
        const endParts = end.split(':').map(Number);

        const minMins = minTimeParts[0] * 60 + minTimeParts[1];
        const maxMins = maxTimeParts[0] * 60 + maxTimeParts[1];
        const startMins = startParts[0] * 60 + startParts[1];
        const endMins = endParts[0] * 60 + endParts[1];

        if (startMins < minMins || endMins > maxMins) {
            throw new Error(`زمان انتخابی باید بین ${sessionData.dynamic_config.min_time} و ${sessionData.dynamic_config.max_time} باشد`);
        }
        if (startMins >= endMins) {
            throw new Error('زمان شروع باید قبل از زمان پایان باشد');
        }

        const res = await fetch(`${API_BASE}/sessions/${sessionData.id}/timeslots`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                start_utc: startUTC,
                end_utc: endUTC
            })
        });

        if (!res.ok) {
            const err = await res.json();
            throw new Error(err.error || 'خطا در افزودن زمان');
        }

        showToast('زمان جدید با موفقیت اضافه شد', 'success');
        fetchSession(sessionData.id);
    } catch (err) {
        showToast(err.message, 'error');
    }
};

function renderCreateSessionForm() {
    app.innerHTML = `
        <div class="max-w-2xl mx-auto bg-white p-6 rounded-lg shadow">
            <h2 class="text-2xl font-bold mb-6 text-gray-800 text-center">ایجاد جلسه جدید</h2>
            
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">عنوان جلسه</label>
                    <input type="text" id="sessionTitle" class="w-full p-2 border rounded dark:bg-gray-700 dark:border-gray-600 dark:text-white" placeholder="مثلاً: جلسه هفتگی تیم">
                </div>
                
                <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">نام ایجاد کننده</label>
                    <input type="text" id="creatorName" class="w-full p-2 border rounded dark:bg-gray-700 dark:border-gray-600 dark:text-white" placeholder="نام شما">
                </div>

                    <label class="flex items-center gap-2 cursor-pointer flex-1 justify-center bg-white dark:bg-gray-700 p-2 rounded border dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors">
                        <input type="radio" name="sessionType" value="fixed" checked onchange="toggleSessionType('fixed')">
                        <span class="text-sm font-medium dark:text-white">زمان‌های مشخص</span>
                    </label>
                    <label class="flex items-center gap-2 cursor-pointer flex-1 justify-center bg-white dark:bg-gray-700 p-2 rounded border dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors">
                        <input type="radio" name="sessionType" value="weekly" onchange="toggleSessionType('weekly')">
                        <span class="text-sm font-medium dark:text-white">الگوی هفتگی</span>
                    </label>                   
                    <label class="flex items-center gap-2 cursor-pointer flex-1 justify-center bg-white dark:bg-gray-700 p-2 rounded border dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors">
                        <input type="radio" name="sessionType" value="dynamic" onchange="toggleSessionType('dynamic')">
                        <span class="text-sm font-medium dark:text-white">بازه زمانی</span>
                    </label>
                </div>

                <!-- Type Descriptions -->
                <div id="typeDescription" class="mb-4 text-sm text-gray-600 bg-blue-50 p-3 rounded border border-blue-100">
                    در این حالت شما چند زمان مشخص را پیشنهاد می‌دهید و شرکت‌کنندگان می‌توانند به آن‌ها رای دهند.
                </div>

                <!-- Fixed Times Section -->
                <div id="fixedTimeSection">
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">زمان‌های پیشنهادی</label>
                    <div id="timeslotsContainer" class="space-y-4">
                        <!-- Timeslot inputs will go here -->
                    </div>
                    <button onclick="addTimeslotInput()" class="mt-4 w-full border-2 border-dashed border-blue-300 text-blue-600 dark:border-blue-500 dark:text-blue-400 py-2 rounded hover:bg-blue-50 dark:hover:bg-gray-700 transition-colors">
                        + افزودن زمان جدید
                    </button>
                </div>

                <!-- Weekly Section -->
                <div id="weeklyTimeSection" class="hidden space-y-4 border p-4 rounded bg-purple-50 dark:bg-purple-900/20 dark:border-purple-800">
                    <div>
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">روزهای هفته</label>
                        <div class="grid grid-cols-2 gap-2" id="weekDaysContainer">
                            ${['شنبه', 'یکشنبه', 'دوشنبه', 'سه‌شنبه', 'چهارشنبه', 'پنج‌شنبه', 'جمعه'].map((day, idx) => `
                                <label class="flex items-center gap-2 bg-white dark:bg-gray-700 p-2 rounded border dark:border-gray-600 cursor-pointer">
                                    <input type="checkbox" name="weekDay" value="${(idx + 6) % 7}">
                                    <span class="text-sm dark:text-white">${day}</span>
                                </label>
                            `).join('')}
                        </div>
                    </div>
                    
                    <div class="flex gap-4 items-center justify-center">
                        <div class="flex flex-col items-center">
                            <span class="text-xs text-gray-500 dark:text-gray-400 mb-1">شروع</span>
                            ${renderTimePicker('weekly_start', 9, 0)}
                        </div>
                        <div class="text-gray-400 mt-4">←</div>
                        <div class="flex flex-col items-center">
                            <span class="text-xs text-gray-500 dark:text-gray-400 mb-1">پایان</span>
                            ${renderTimePicker('weekly_end', 17, 0)}
                        </div>
                    </div>

                    <div>
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">مدت اعتبار لینک (هفته)</label>
                        <input type="number" id="linkDurationWeeks" class="w-full p-2 border rounded dark:bg-gray-700 dark:border-gray-600 dark:text-white" value="4" min="1" max="52">
                        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">پس از این مدت، لینک غیرفعال خواهد شد.</p>
                    </div>
                </div>

                <!-- Dynamic Time Section -->
                <div id="dynamicTimeSection" class="hidden space-y-4 border p-4 rounded bg-blue-50 dark:bg-blue-900/20 dark:border-blue-800">
                    <div>
                        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">تاریخ جلسه</label>
                        ${renderJalaliDatePicker('dyn_date')}
                    </div>
                    <div class="flex gap-4 items-center justify-center">
                        <div class="flex flex-col items-center">
                            <span class="text-xs text-gray-500 dark:text-gray-400 mb-1">از ساعت</span>
                            ${renderTimePicker('dyn_min', 9, 0)}
                        </div>
                        <div class="text-gray-400 mt-4">←</div>
                        <div class="flex flex-col items-center">
                            <span class="text-xs text-gray-500 dark:text-gray-400 mb-1">تا ساعت</span>
                            ${renderTimePicker('dyn_max', 17, 0)}
                        </div>
                    </div>
                </div>

                <button onclick="submitCreateSession()" class="w-full bg-green-600 text-white py-3 rounded-lg hover:bg-green-700 font-bold shadow-lg mt-6">
                    ایجاد جلسه
                </button>
            </div>
        </div>
    `;

    addTimeslotInput();
}

function renderAdminDashboard(stats) {
    app.innerHTML = `
        <div class="max-w-4xl mx-auto bg-white p-8 rounded-lg shadow-lg">
            <h2 class="text-3xl font-bold mb-8 text-gray-800 text-center border-b pb-4">داشبورد مدیریت</h2>
            
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div class="bg-blue-50 p-6 rounded-xl border border-blue-100 text-center transform hover:scale-105 transition-transform">
                    <div class="text-4xl font-bold text-blue-600 mb-2">${stats.total_sessions}</div>
                    <div class="text-gray-600 font-medium">کل جلسات</div>
                </div>
                
                <div class="bg-green-50 p-6 rounded-xl border border-green-100 text-center transform hover:scale-105 transition-transform">
                    <div class="text-4xl font-bold text-green-600 mb-2">${stats.total_timeslots}</div>
                    <div class="text-gray-600 font-medium">زمان‌های پیشنهادی</div>
                </div>
                
                <div class="bg-purple-50 p-6 rounded-xl border border-purple-100 text-center transform hover:scale-105 transition-transform">
                    <div class="text-4xl font-bold text-purple-600 mb-2">${stats.total_votes}</div>
                    <div class="text-gray-600 font-medium">آرای ثبت شده</div>
                </div>
            </div>

            <div class="mt-12 text-center">
                <a href="/" class="inline-block px-6 py-3 bg-gray-800 text-white rounded-lg hover:bg-gray-700 transition-colors">
                    بازگشت به صفحه اصلی
                </a>
            </div>
        </div>
    `;
}

window.addTimeslotInput = function () {
    const container = document.getElementById('timeslotsContainer');
    const id = Date.now(); // Unique ID for this row
    const div = document.createElement('div');
    div.className = 'bg-gray-50 p-3 rounded border relative';
    div.dataset.id = id;

    div.innerHTML = `
        <button onclick="this.parentElement.remove()" class="absolute top-2 left-2 text-red-400 hover:text-red-600">
            &times;
        </button>
        <div class="mb-3">
            <label class="text-xs text-gray-500 block mb-1">تاریخ</label>
            ${renderJalaliDatePicker(`ts_${id}_date`)}
        </div>
        <div class="flex gap-4 items-center justify-center">
            <div class="flex flex-col items-center">
                <span class="text-xs text-gray-500 mb-1">شروع</span>
                ${renderTimePicker(`ts_${id}_start`, 10, 0)}
            </div>
            <div class="text-gray-400 mt-4">←</div>
            <div class="flex flex-col items-center">
                <span class="text-xs text-gray-500 mb-1">پایان</span>
                ${renderTimePicker(`ts_${id}_end`, 11, 0)}
            </div>
        </div>
    `;
    container.appendChild(div);
};

window.toggleSessionType = function (type) {
    const fixedSection = document.getElementById('fixedTimeSection');
    const weeklySection = document.getElementById('weeklyTimeSection');
    const dynamicSection = document.getElementById('dynamicTimeSection');
    const typeDesc = document.getElementById('typeDescription');

    fixedSection.classList.add('hidden');
    weeklySection.classList.add('hidden');
    dynamicSection.classList.add('hidden');

    if (type === 'fixed') {
        fixedSection.classList.remove('hidden');
        typeDesc.innerText = 'در این حالت شما چند زمان مشخص را پیشنهاد می‌دهید و شرکت‌کنندگان می‌توانند به آن‌ها رای دهند.';
    } else if (type === 'weekly') {
        weeklySection.classList.remove('hidden');
        typeDesc.innerText = 'در این حالت شما روزهای مجاز هفته و بازه زمانی را مشخص می‌کنید. شرکت‌کنندگان می‌توانند در این روزها زمان پیشنهاد دهند.';
    } else {
        dynamicSection.classList.remove('hidden');
        typeDesc.innerText = 'در این حالت شما یک تاریخ و بازه زمانی کلی مشخص می‌کنید و شرکت‌کنندگان می‌توانند هر زمانی در این بازه را پیشنهاد دهند.';
    }
};

window.submitCreateSession = async function () {
    const title = document.getElementById('sessionTitle').value;
    const creatorName = document.getElementById('creatorName').value;
    const type = document.querySelector('input[name="sessionType"]:checked').value;

    if (!title || !creatorName) {
        showToast('لطفاً عنوان و نام خود را وارد کنید', 'warning');
        return;
    }

    let payload = {
        title,
        creator_name: creatorName,
        type,
        timeslots: []
    };

    if (type === 'fixed') {
        const rows = document.getElementById('timeslotsContainer').children;
        for (let row of rows) {
            const id = row.dataset.id;
            const jDate = getJalaliDateFromPicker(`ts_${id}_date`);
            const startTime = getTimeFromPicker(`ts_${id}_start`);
            const endTime = getTimeFromPicker(`ts_${id}_end`);

            const startISO = jalaliToISO(jDate, startTime);
            const endISO = jalaliToISO(jDate, endTime);

            payload.timeslots.push({
                start_utc: startISO,
                end_utc: endISO
            });
        }

        if (payload.timeslots.length < 2) {
            showToast('لطفاً حداقل دو زمان را مشخص کنید تا کاربران حق انتخاب داشته باشند', 'warning');
            return;
        }
    } else if (type === 'weekly') {
        // Weekly Pattern
        const checkboxes = document.querySelectorAll('input[name="weekDay"]:checked');
        if (checkboxes.length === 0) {
            showToast('لطفاً حداقل یک روز هفته را انتخاب کنید', 'warning');
            return;
        }

        const startTime = getTimeFromPicker('weekly_start');
        const endTime = getTimeFromPicker('weekly_end');
        const durationWeeks = parseInt(document.getElementById('linkDurationWeeks').value) || 4;

        // Calculate expiration
        const expirationDate = new Date();
        expirationDate.setDate(expirationDate.getDate() + (durationWeeks * 7));
        payload.expires_at_utc = expirationDate.toISOString();

        const selectedDays = Array.from(checkboxes).map(cb => parseInt(cb.value));

        payload.type = 'weekly';
        payload.dynamic_config = {
            min_time: startTime,
            max_time: endTime,
            allowed_days: selectedDays
        };
    } else {
        // Dynamic
        const jDate = getJalaliDateFromPicker('dyn_date');
        const minTime = getTimeFromPicker('dyn_min');
        const maxTime = getTimeFromPicker('dyn_max');

        // Convert to UTC midnight for date_utc
        // We use the same logic as before: Gregorian date at 00:00 UTC
        const g = jalaali.toGregorian(jDate.y, jDate.m, jDate.d);
        const dateUTC = new Date(Date.UTC(g.gy, g.gm - 1, g.gd)).toISOString();

        payload.dynamic_config = {
            date_utc: dateUTC,
            min_time: minTime,
            max_time: maxTime
        };
    }

    try {
        const res = await fetch(`${API_BASE}/sessions`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (!res.ok) {
            const err = await res.json();
            throw new Error(err.error || 'خطا در ایجاد جلسه');
        }

        const data = await res.json();
        window.location.href = `/${data.id}`;
    } catch (err) {
        showToast(err.message, 'error');
    }
};

// Init
const id = getSessionIDFromURL();
if (id === 'admin') {
    fetchAdminStats();
} else if (id) {
    fetchSession(id);
} else {
    renderCreateSessionForm();
}
