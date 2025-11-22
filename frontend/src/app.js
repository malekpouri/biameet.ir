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
    // Using jalaali-js from CDN (window.jalaali) or dayjs
    // Let's use dayjs with jalali plugin if available, or standard Intl if not.
    // The index.html has jalaali-js and dayjs.
    // Let's assume dayjs is loaded.
    // We need to extend dayjs with jalali if we loaded the plugin.
    // If we just have jalaali.js, we can use toJalaali.
    
    const date = new Date(utcDateString);
    // Simple fallback if libs fail
    return date.toLocaleString('fa-IR'); 
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

    const { title, creator_name, timeslots } = sessionData;

    // Calculate votes per timeslot
    const voteCounts = {};
    timeslots.forEach(ts => {
        voteCounts[ts.id] = (ts.votes || []).length;
    });

    app.innerHTML = `
        <div class="max-w-2xl mx-auto bg-white p-6 rounded-lg shadow">
            <h2 class="text-2xl font-bold mb-2 text-gray-800">${title}</h2>
            <p class="text-gray-600 mb-6">ایجاد شده توسط: ${creator_name}</p>

            <div class="mb-6">
                <label class="block text-sm font-medium text-gray-700 mb-2">نام شما:</label>
                <input type="text" id="voterNameInput" class="w-full p-2 border rounded text-right" placeholder="نام خود را وارد کنید...">
            </div>

            <div class="space-y-3">
                <h3 class="font-semibold text-gray-700">زمان‌های پیشنهادی:</h3>
                ${timeslots.map(ts => {
                    const isSelected = selectedTimeslots.has(ts.id);
                    const count = voteCounts[ts.id] || 0;
                    const voters = (ts.votes || []).map(v => v.voter_name).join(', ');
                    
                    return `
                    <div class="timeslot-card border rounded p-3 cursor-pointer transition-colors ${isSelected ? 'bg-blue-50 border-blue-500' : 'hover:bg-gray-50'}"
                         onclick="toggleTimeslot('${ts.id}')">
                        <div class="flex justify-between items-center">
                            <div>
                                <div class="font-bold text-gray-800">${formatJalali(ts.start_utc)}</div>
                                <div class="text-sm text-gray-500">تا ${formatJalali(ts.end_utc)}</div>
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

            <button onclick="submitVote()" class="mt-8 w-full bg-blue-600 text-white py-3 rounded-lg hover:bg-blue-700 font-bold shadow-lg transition-transform transform hover:scale-105">
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
window.toggleTimeslot = function(id) {
    if (selectedTimeslots.has(id)) {
        selectedTimeslots.delete(id);
    } else {
        selectedTimeslots.add(id);
    }
    renderSession();
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

                <div>
                    <label class="block text-sm font-medium text-gray-700 mb-2">زمان‌های پیشنهادی</label>
                    <div id="timeslotsContainer" class="space-y-3">
                        <!-- Timeslot inputs will go here -->
                    </div>
                    <button onclick="addTimeslotInput()" class="mt-2 text-blue-600 text-sm hover:underline">+ افزودن زمان</button>
                </div>

                <button onclick="submitCreateSession()" class="w-full bg-green-600 text-white py-3 rounded-lg hover:bg-green-700 font-bold shadow-lg mt-6">
                    ایجاد جلسه
                </button>
            </div>
        </div>
    `;
    
    // Add initial timeslot input
    addTimeslotInput();
}

window.addTimeslotInput = function() {
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

window.submitCreateSession = async function() {
    const title = document.getElementById('sessionTitle').value;
    const creatorName = document.getElementById('creatorName').value;
    
    if (!title || !creatorName) {
        alert('لطفاً عنوان و نام خود را وارد کنید');
        return;
    }

    const timeslots = [];
    const startInputs = document.querySelectorAll('.ts-start');
    const endInputs = document.querySelectorAll('.ts-end');

    for (let i = 0; i < startInputs.length; i++) {
        const start = startInputs[i].value;
        const end = endInputs[i].value;
        
        if (start && end) {
            timeslots.push({
                start_utc: new Date(start).toISOString(),
                end_utc: new Date(end).toISOString()
            });
        }
    }

    if (timeslots.length === 0) {
        alert('لطفاً حداقل یک زمان را مشخص کنید');
        return;
    }

    try {
        const res = await fetch(`${API_BASE}/sessions`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                title,
                creator_name: creatorName,
                timeslots
            })
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
