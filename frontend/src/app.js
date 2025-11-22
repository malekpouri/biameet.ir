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
    // Remove leading slash
    const id = path.substring(1);
    // Basic validation: 5 chars alphanumeric
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
        note: 'Yes' // Simple Yes vote for now
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

    // Calculate votes per timeslot
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
                <div class="flex gap-2 items-end bg-gray-50 p-3 rounded border">
                    <div class="flex-1">
                        <label class="text-xs text-gray-500 block mb-1">از ساعت</label>
                        <input type="time" id="dynamicStart" class="w-full p-2 border rounded text-sm">
                    </div>
                    <div class="flex-1">
                        <label class="text-xs text-gray-500 block mb-1">تا ساعت</label>
                        <input type="time" id="dynamicEnd" class="w-full p-2 border rounded text-sm">
                    </div>
                    <button onclick="submitDynamicTimeslot()" class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 text-sm">
                        افزودن
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

    // Re-attach event listener for input since we re-rendered
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
    const start = document.getElementById('dynamicStart').value;
    const end = document.getElementById('dynamicEnd').value;

    if (!start || !end) {
        alert('لطفاً ساعت شروع و پایان را وارد کنید');
        return;
    }

    // Construct dates
    // We use the session date (UTC) to get the YYYY-MM-DD
    const sessionDate = new Date(sessionData.dynamic_config.date_utc);
    // Format to YYYY-MM-DD
    const yyyy = sessionDate.getFullYear();
    const mm = String(sessionDate.getMonth() + 1).padStart(2, '0');
    const dd = String(sessionDate.getDate()).padStart(2, '0');
    const dateStr = `${yyyy}-${mm}-${dd}`;

    // Combine with time
    const startDateTimeStr = `${dateStr}T${start}:00`;
    const endDateTimeStr = `${dateStr}T${end}:00`;

    // Convert to UTC
    // We assume the user input is in local time (browser time), which should match the intended timezone
    // Ideally we should use a specific timezone (Asia/Tehran) but for now browser time is the best proxy
    const startUTC = new Date(startDateTimeStr).toISOString();
    const endUTC = new Date(endDateTimeStr).toISOString();

    // Basic validation against min/max time
    // This is tricky because of timezone conversion.
    // Let's skip strict client-side validation for now and let backend/logic handle it,
    // or just trust the user input for this MVP.

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

        // Reload session
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

                <div class="flex gap-4 mb-4">
                    <label class="flex items-center gap-2 cursor-pointer">
                        <input type="radio" name="sessionType" value="fixed" checked onchange="toggleSessionType('fixed')">
                        <span>زمان‌های مشخص (Fixed)</span>
                    </label>
                    <label class="flex items-center gap-2 cursor-pointer">
                        <input type="radio" name="sessionType" value="dynamic" onchange="toggleSessionType('dynamic')">
                        <span>انتخاب روز خاص (Dynamic)</span>
                    </label>
                </div>

                <!-- Fixed Times Section -->
                <div id="fixedTimeSection">
                    <label class="block text-sm font-medium text-gray-700 mb-2">زمان‌های پیشنهادی</label>
                    <div id="timeslotsContainer" class="space-y-3">
                        <!-- Timeslot inputs will go here -->
                    </div>
                    <button onclick="addTimeslotInput()" class="mt-2 text-blue-600 text-sm hover:underline">+ افزودن زمان</button>
                </div>

                <!-- Dynamic Time Section -->
                <div id="dynamicTimeSection" class="hidden space-y-4 border p-4 rounded bg-blue-50">
                    <div>
                        <label class="block text-sm font-medium text-gray-700 mb-2">تاریخ جلسه (شمسی)</label>
                        <div class="flex gap-2">
                            <select id="jDay" class="p-2 border rounded flex-1"></select>
                            <select id="jMonth" class="p-2 border rounded flex-1"></select>
                            <select id="jYear" class="p-2 border rounded flex-1"></select>
                        </div>
                    </div>
                    <div class="flex gap-4">
                        <div class="flex-1">
                            <label class="block text-sm font-medium text-gray-700 mb-1">از ساعت</label>
                            <input type="time" id="minTime" class="w-full p-2 border rounded" value="09:00">
                        </div>
                        <div class="flex-1">
                            <label class="block text-sm font-medium text-gray-700 mb-1">تا ساعت</label>
                            <input type="time" id="maxTime" class="w-full p-2 border rounded" value="17:00">
                        </div>
                    </div>
                </div>

                <button onclick="submitCreateSession()" class="w-full bg-green-600 text-white py-3 rounded-lg hover:bg-green-700 font-bold shadow-lg mt-6">
                    ایجاد جلسه
                </button>
            </div>
        </div>
    `;

    // Add initial timeslot input
    addTimeslotInput();
    // Init Jalali options
    generateJalaliOptions();
}

window.addTimeslotInput = function () {
    const container = document.getElementById('timeslotsContainer');
    const id = Date.now();
    const div = document.createElement('div');
    div.className = 'flex gap-2 items-end bg-gray-50 p-3 rounded border';
    div.innerHTML = `
        <div class="flex-1">
            <label class="text-xs text-gray-500 block mb-1">شروع</label>
            <input type="datetime-local" class="ts-start w-full p-1 border rounded text-sm" required>
        </div>
        <div class="flex-1">
            <label class="text-xs text-gray-500 block mb-1">پایان</label>
            <input type="datetime-local" class="ts-end w-full p-1 border rounded text-sm" required>
        </div>
        <button onclick="this.parentElement.remove()" class="text-red-500 hover:text-red-700 p-1">
            &times;
        </button>
    `;
    container.appendChild(div);
};

function toggleSessionType(type) {
    const fixedSection = document.getElementById('fixedTimeSection');
    const dynamicSection = document.getElementById('dynamicTimeSection');

    if (type === 'fixed') {
        fixedSection.classList.remove('hidden');
        dynamicSection.classList.add('hidden');
    } else {
        fixedSection.classList.add('hidden');
        dynamicSection.classList.remove('hidden');
    }
}

function generateJalaliOptions() {
    const yearSelect = document.getElementById('jYear');
    const monthSelect = document.getElementById('jMonth');
    const daySelect = document.getElementById('jDay');

    // Current Jalali Year
    const current = jalaali.toJalaali(new Date());

    // Years: Current + 1
    for (let y = current.jy; y <= current.jy + 1; y++) {
        const opt = document.createElement('option');
        opt.value = y;
        opt.text = y;
        yearSelect.add(opt);
    }

    // Months
    const months = ['فروردین', 'اردیبهشت', 'خرداد', 'تیر', 'مرداد', 'شهریور', 'مهر', 'آبان', 'آذر', 'دی', 'بهمن', 'اسفند'];
    months.forEach((m, i) => {
        const opt = document.createElement('option');
        opt.value = i + 1;
        opt.text = m;
        monthSelect.add(opt);
    });

    // Days
    for (let d = 1; d <= 31; d++) {
        const opt = document.createElement('option');
        opt.value = d;
        opt.text = d;
        daySelect.add(opt);
    }

    // Set defaults
    yearSelect.value = current.jy;
    monthSelect.value = current.jm;
    daySelect.value = current.jd;
}

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
        const startInputs = document.querySelectorAll('.ts-start');
        const endInputs = document.querySelectorAll('.ts-end');

        for (let i = 0; i < startInputs.length; i++) {
            const start = startInputs[i].value;
            const end = endInputs[i].value;

            if (start && end) {
                payload.timeslots.push({
                    start_utc: new Date(start).toISOString(),
                    end_utc: new Date(end).toISOString()
                });
            }
        }

        if (payload.timeslots.length === 0) {
            alert('لطفاً حداقل یک زمان را مشخص کنید');
            return;
        }
    } else {
        // Dynamic
        const jy = parseInt(document.getElementById('jYear').value);
        const jm = parseInt(document.getElementById('jMonth').value);
        const jd = parseInt(document.getElementById('jDay').value);
        const minTime = document.getElementById('minTime').value;
        const maxTime = document.getElementById('maxTime').value;

        // Convert Jalali to Gregorian for DateUTC
        const g = jalaali.toGregorian(jy, jm, jd);
        const date = new Date(g.gy, g.gm - 1, g.gd); // Month is 0-indexed in JS Date
        // We want midnight UTC
        // Actually, let's store it as ISO string of that date at 00:00 UTC?
        // Or just YYYY-MM-DD? The backend expects string.
        // Let's send ISO string of midnight UTC.
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
        // Redirect to session page
        window.location.href = `/${data.id}`;
    } catch (err) {
        alert(err.message);
    }
};

// Init
const id = getSessionIDFromURL();
if (id) {
    fetchSession(id);
} else {
    // Show Create Session Form
    renderCreateSessionForm();
}
