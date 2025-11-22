const API_BASE = '/api/v1';

// State
let sessionData = null;
let selectedTimeslots = new Set();
let voterName = '';

// DOM Elements
const app = document.getElementById('app');

// Utils
function getSessionIDFromURL() {
    const path = window.location.pathname;
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
            <select id="${prefix}_day" class="p-2 border rounded bg-white text-center text-sm flex-1">${daysOpts}</select>
            <select id="${prefix}_month" class="p-2 border rounded bg-white text-center text-sm flex-1">${monthsOpts}</select>
            <select id="${prefix}_year" class="p-2 border rounded bg-white text-center text-sm flex-1">${yearsOpts}</select>
        </div>
    `;
}

function renderTimePicker(prefix, defaultHour = 9, defaultMinute = 0) {
    let hourOpts = '';
    for (let i = 0; i < 24; i++) {
        const val = String(i).padStart(2, '0');
        hourOpts += `<option value="${val}" ${i === defaultHour ? 'selected' : ''}>${val}</option>`;
    }

    let minOpts = '';
    for (let i = 0; i < 60; i += 15) {
        const val = String(i).padStart(2, '0');
        minOpts += `<option value="${val}" ${i === defaultMinute ? 'selected' : ''}>${val}</option>`;
    }

    return `
        <div class="flex gap-1 items-center justify-center bg-white border rounded p-1" dir="ltr">
            <select id="${prefix}_hour" class="p-1 bg-transparent text-center font-mono w-12 outline-none">${hourOpts}</select>
            <span class="text-gray-500">:</span>
            <select id="${prefix}_minute" class="p-1 bg-transparent text-center font-mono w-12 outline-none">${minOpts}</select>
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

async function submitVote() {
    if (!voterName) {
        alert('لطفاً نام خود را وارد کنید');
        return;
    }
    if (selectedTimeslots.size === 0) {
        alert('لطفاً حداقل یک زمان را انتخاب کنید');
        return;
    }

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
                votes: votes
            })
        });

        if (!res.ok) {
            const err = await res.json();
            throw new Error(err.error || 'خطا در ثبت رای');
        }

        alert('رای شما با موفقیت ثبت شد');
        window.location.reload();
    } catch (err) {
        alert(err.message);
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
            <div class="bg-blue-50 p-4 rounded-lg mb-6 border border-blue-100">
                <div class="text-center">
                    <span class="block text-gray-500 text-sm mb-1">تاریخ انتخابی</span>
                    <span class="text-xl font-bold text-blue-800">${j.jy}/${j.jm}/${j.jd}</span>
                </div>
                <div class="mt-2 text-center text-sm text-gray-600">
                    بازه مجاز: ${dynamic_config.min_time} تا ${dynamic_config.max_time}
                </div>
            </div>
        `;

        dynamicInput = `
            <div class="mt-8 border-t pt-6">
                <h3 class="font-semibold text-gray-700 mb-4">پیشنهاد زمان جدید</h3>
                <div class="bg-gray-50 p-4 rounded border">
                    <div class="flex gap-4 items-center justify-center mb-4">
                        <div class="flex flex-col items-center">
                            <span class="text-xs text-gray-500 mb-1">از ساعت</span>
                            ${renderTimePicker('dynamic_start', 10, 0)}
                        </div>
                        <div class="text-gray-400 mt-4">←</div>
                        <div class="flex flex-col items-center">
                            <span class="text-xs text-gray-500 mb-1">تا ساعت</span>
                            ${renderTimePicker('dynamic_end', 11, 0)}
                        </div>
                    </div>
                    <button onclick="submitDynamicTimeslot()" class="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 text-sm font-bold">
                        افزودن زمان
                    </button>
                </div>
            </div>
        `;
    }

    app.innerHTML = `
        <div class="max-w-2xl mx-auto bg-white p-6 rounded-lg shadow">
            <h2 class="text-2xl font-bold mb-2 text-gray-800 text-center">${title}</h2>
            <p class="text-gray-600 mb-6 text-center">ایجاد شده توسط: ${creator_name}</p>

            ${dynamicHeader}

            <div class="mb-6">
                <label class="block text-sm font-medium text-gray-700 mb-2">نام شما:</label>
                <input type="text" id="voterNameInput" class="w-full p-2 border rounded text-right" placeholder="نام خود را وارد کنید...">
            </div>

            <div class="space-y-3">
                <h3 class="font-semibold text-gray-700">زمان‌های موجود:</h3>
                ${timeslots.length === 0 ? '<p class="text-gray-400 text-sm text-center italic">هنوز زمانی ثبت نشده است</p>' : ''}
                ${timeslots.map(ts => {
        const isSelected = selectedTimeslots.has(ts.id);
        const count = voteCounts[ts.id] || 0;
        const voters = (ts.votes || []).map(v => v.voter_name).join(', ');

        return `
                    <div class="timeslot-card border rounded p-3 cursor-pointer transition-colors ${isSelected ? 'bg-blue-50 border-blue-500' : 'hover:bg-gray-50'}"
                         onclick="toggleTimeslot('${ts.id}')">
                        <div class="flex justify-between items-center">
                            <div>
                                <div class="font-bold text-gray-800">${formatTime(ts.start_utc)} - ${formatTime(ts.end_utc)}</div>
                            </div>
                            <div class="flex items-center space-x-2 space-x-reverse">
                                <span class="bg-gray-200 text-gray-700 px-2 py-1 rounded text-xs" title="${voters}">
                                    ${count} رای
                                </span>
                                ${isSelected ? '<span class="text-blue-600">✓</span>' : ''}
                            </div>
                        </div>
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

window.submitDynamicTimeslot = async function () {
    const start = getTimeFromPicker('dynamic_start');
    const end = getTimeFromPicker('dynamic_end');

    // Construct dates
    const sessionDate = new Date(sessionData.dynamic_config.date_utc);
    const yyyy = sessionDate.getFullYear();
    const mm = String(sessionDate.getMonth() + 1).padStart(2, '0');
    const dd = String(sessionDate.getDate()).padStart(2, '0');
    const dateStr = `${yyyy}-${mm}-${dd}`;

    const startDateTimeStr = `${dateStr}T${start}:00`;
    const endDateTimeStr = `${dateStr}T${end}:00`;

    const startUTC = new Date(startDateTimeStr).toISOString();
    const endUTC = new Date(endDateTimeStr).toISOString();

    try {
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

        fetchSession(sessionData.id);
    } catch (err) {
        alert(err.message);
    }
};

function renderCreateSessionForm() {
    app.innerHTML = `
        <div class="max-w-2xl mx-auto bg-white p-6 rounded-lg shadow">
            <h2 class="text-2xl font-bold mb-6 text-gray-800 text-center">ایجاد جلسه جدید</h2>
            
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">عنوان جلسه</label>
                    <input type="text" id="sessionTitle" class="w-full p-2 border rounded" placeholder="مثلاً: جلسه هفتگی تیم">
                </div>
                
                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-1">نام ایجاد کننده</label>
                    <input type="text" id="creatorName" class="w-full p-2 border rounded" placeholder="نام شما">
                </div>

                <div class="flex gap-4 mb-4 bg-gray-50 p-2 rounded">
                    <label class="flex items-center gap-2 cursor-pointer flex-1 justify-center">
                        <input type="radio" name="sessionType" value="fixed" checked onchange="toggleSessionType('fixed')">
                        <span class="text-sm font-medium">چند زمانه</span>
                    </label>
                    <label class="flex items-center gap-2 cursor-pointer flex-1 justify-center">
                        <input type="radio" name="sessionType" value="dynamic" onchange="toggleSessionType('dynamic')">
                        <span class="text-sm font-medium">تک روز </span>
                    </label>
                </div>

                <!-- Fixed Times Section -->
                <div id="fixedTimeSection">
                    <label class="block text-sm font-medium text-gray-700 mb-2">زمان‌های پیشنهادی</label>
                    <div id="timeslotsContainer" class="space-y-4">
                        <!-- Timeslot inputs will go here -->
                    </div>
                    <button onclick="addTimeslotInput()" class="mt-4 w-full border-2 border-dashed border-blue-300 text-blue-600 py-2 rounded hover:bg-blue-50 transition-colors">
                        + افزودن زمان جدید
                    </button>
                </div>

                <!-- Dynamic Time Section -->
                <div id="dynamicTimeSection" class="hidden space-y-4 border p-4 rounded bg-blue-50">
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">تاریخ جلسه</label>
                        ${renderJalaliDatePicker('dyn_date')}
                    </div>
                    <div class="flex gap-4 items-center justify-center">
                        <div class="flex flex-col items-center">
                            <span class="text-xs text-gray-500 mb-1">از ساعت</span>
                            ${renderTimePicker('dyn_min', 9, 0)}
                        </div>
                        <div class="text-gray-400 mt-4">←</div>
                        <div class="flex flex-col items-center">
                            <span class="text-xs text-gray-500 mb-1">تا ساعت</span>
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
    const dynamicSection = document.getElementById('dynamicTimeSection');

    if (type === 'fixed') {
        fixedSection.classList.remove('hidden');
        dynamicSection.classList.add('hidden');
    } else {
        fixedSection.classList.add('hidden');
        dynamicSection.classList.remove('hidden');
    }
};

window.submitCreateSession = async function () {
    const title = document.getElementById('sessionTitle').value;
    const creatorName = document.getElementById('creatorName').value;
    const type = document.querySelector('input[name="sessionType"]:checked').value;

    if (!title || !creatorName) {
        alert('لطفاً عنوان و نام خود را وارد کنید');
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

        if (payload.timeslots.length === 0) {
            alert('لطفاً حداقل یک زمان را مشخص کنید');
            return;
        }
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
        alert(err.message);
    }
};

// Init
const id = getSessionIDFromURL();
if (window.location.pathname === '/admin') {
    renderAdminDashboard();
} else if (id) {
    fetchSession(id);
} else {
    renderCreateSessionForm();
}

async function renderAdminDashboard() {
    try {
        const res = await fetch(`${API_BASE}/admin/stats`);
        if (!res.ok) throw new Error('Failed to fetch stats');
        const stats = await res.json();

        app.innerHTML = `
            <div class="max-w-4xl mx-auto bg-white p-8 rounded-lg shadow">
                <h2 class="text-3xl font-bold mb-8 text-gray-800 text-center">پنل مدیریت</h2>
                
                <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                    <div class="bg-blue-50 p-6 rounded-lg border border-blue-100 text-center">
                        <div class="text-4xl font-bold text-blue-600 mb-2">${stats.total_sessions}</div>
                        <div class="text-gray-600">کل جلسات</div>
                    </div>
                    <div class="bg-green-50 p-6 rounded-lg border border-green-100 text-center">
                        <div class="text-4xl font-bold text-green-600 mb-2">${stats.total_timeslots}</div>
                        <div class="text-gray-600">کل زمان‌ها</div>
                    </div>
                    <div class="bg-purple-50 p-6 rounded-lg border border-purple-100 text-center">
                        <div class="text-4xl font-bold text-purple-600 mb-2">${stats.total_votes}</div>
                        <div class="text-gray-600">کل آرا</div>
                    </div>
                </div>

                <div class="text-center">
                    <a href="/" class="text-blue-600 hover:underline">بازگشت به صفحه اصلی</a>
                </div>
            </div>
        `;
    } catch (err) {
        app.innerHTML = `<div class="text-red-500 text-center mt-10">خطا در دریافت اطلاعات: ${err.message}</div>`;
    }
}
