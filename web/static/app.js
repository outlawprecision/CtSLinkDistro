// Global variables
let currentMembers = [];
let currentEligibleMembers = [];
let isSpinning = false;

// API base URL
const API_BASE = '/api';

// Initialize the application
document.addEventListener('DOMContentLoaded', function () {
    showTab('dashboard');
    refreshStatus();
    loadMembers();
});

// Tab management
function showTab(tabName) {
    // Hide all tab contents
    const tabContents = document.querySelectorAll('.tab-content');
    tabContents.forEach(tab => tab.classList.remove('active'));

    // Remove active class from all tab buttons
    const tabButtons = document.querySelectorAll('.tab-button');
    tabButtons.forEach(button => button.classList.remove('active'));

    // Show selected tab content
    document.getElementById(tabName).classList.add('active');

    // Add active class to clicked button
    event.target.classList.add('active');

    // Load data for specific tabs
    if (tabName === 'members') {
        loadMembers();
    } else if (tabName === 'wheel') {
        loadEligibleMembersForWheel();
    } else if (tabName === 'history') {
        loadHistory();
    }
}

// API helper functions
async function apiCall(endpoint, options = {}) {
    showLoading();
    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });

        const data = await response.json();

        if (!data.success) {
            throw new Error(data.error || 'API call failed');
        }

        return data.data;
    } catch (error) {
        showNotification(error.message, 'error');
        throw error;
    } finally {
        hideLoading();
    }
}

// Dashboard functions
async function refreshStatus() {
    try {
        const status = await apiCall('/distribution/status');
        updateDashboard(status);
    } catch (error) {
        console.error('Failed to refresh status:', error);
    }
}

function updateDashboard(status) {
    // Update silver status
    document.getElementById('silver-eligible').textContent = status.silver.eligible_count;
    document.getElementById('silver-completed').textContent = status.silver.completed_count;
    document.getElementById('silver-compensation').textContent = status.silver.compensation_count;
    document.getElementById('silver-progress').textContent = `${status.silver.completion_percentage.toFixed(1)}%`;

    // Update gold status
    document.getElementById('gold-eligible').textContent = status.gold.eligible_count;
    document.getElementById('gold-completed').textContent = status.gold.completed_count;
    document.getElementById('gold-compensation').textContent = status.gold.compensation_count;
    document.getElementById('gold-progress').textContent = `${status.gold.completion_percentage.toFixed(1)}%`;
}

async function updateLists() {
    try {
        await apiCall('/utility/update-lists', { method: 'POST' });
        showNotification('Distribution lists updated successfully', 'success');
        refreshStatus();
    } catch (error) {
        console.error('Failed to update lists:', error);
    }
}

async function resetWeekly() {
    if (!confirm('Are you sure you want to reset weekly participation for all members?')) {
        return;
    }

    try {
        await apiCall('/utility/reset-weekly', { method: 'POST' });
        showNotification('Weekly participation reset successfully', 'success');
        refreshStatus();
    } catch (error) {
        console.error('Failed to reset weekly participation:', error);
    }
}

// Wheel functions
async function loadEligibleMembersForWheel() {
    const linkType = document.querySelector('input[name="linkType"]:checked').value;
    try {
        currentEligibleMembers = await apiCall(`/distribution/eligible?type=${linkType}&active_only=true`);
        drawWheel();
    } catch (error) {
        console.error('Failed to load eligible members:', error);
        currentEligibleMembers = [];
        drawWheel();
    }
}

function drawWheel() {
    const canvas = document.getElementById('wheelCanvas');
    const ctx = canvas.getContext('2d');
    const centerX = canvas.width / 2;
    const centerY = canvas.height / 2;
    const radius = 180;

    // Clear canvas
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    if (currentEligibleMembers.length === 0) {
        // Draw empty wheel
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius, 0, 2 * Math.PI);
        ctx.fillStyle = '#f0f0f0';
        ctx.fill();
        ctx.strokeStyle = '#ccc';
        ctx.lineWidth = 3;
        ctx.stroke();

        // Draw "No eligible members" text
        ctx.fillStyle = '#666';
        ctx.font = '20px Arial';
        ctx.textAlign = 'center';
        ctx.fillText('No Eligible Members', centerX, centerY);
        return;
    }

    const anglePerSegment = (2 * Math.PI) / currentEligibleMembers.length;
    const colors = generateColors(currentEligibleMembers.length);

    // Draw segments
    for (let i = 0; i < currentEligibleMembers.length; i++) {
        const startAngle = i * anglePerSegment - Math.PI / 2;
        const endAngle = (i + 1) * anglePerSegment - Math.PI / 2;

        // Draw segment
        ctx.beginPath();
        ctx.moveTo(centerX, centerY);
        ctx.arc(centerX, centerY, radius, startAngle, endAngle);
        ctx.closePath();
        ctx.fillStyle = colors[i];
        ctx.fill();
        ctx.strokeStyle = '#fff';
        ctx.lineWidth = 2;
        ctx.stroke();

        // Draw text
        const textAngle = startAngle + anglePerSegment / 2;
        const textX = centerX + Math.cos(textAngle) * (radius * 0.7);
        const textY = centerY + Math.sin(textAngle) * (radius * 0.7);

        ctx.save();
        ctx.translate(textX, textY);
        ctx.rotate(textAngle + Math.PI / 2);
        ctx.fillStyle = '#333';
        ctx.font = '12px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(currentEligibleMembers[i].discord_username, 0, 0);
        ctx.restore();
    }
}

function generateColors(count) {
    const colors = [];
    for (let i = 0; i < count; i++) {
        const hue = (i * 360 / count) % 360;
        colors.push(`hsl(${hue}, 70%, 80%)`);
    }
    return colors;
}

async function spinWheel() {
    if (isSpinning) return;
    if (currentEligibleMembers.length === 0) {
        showNotification('No eligible members available', 'warning');
        return;
    }

    isSpinning = true;
    const spinButton = document.getElementById('spin-button');
    const canvas = document.getElementById('wheelCanvas');
    const winnerDisplay = document.getElementById('winner-display');

    spinButton.disabled = true;
    spinButton.textContent = 'SPINNING...';
    winnerDisplay.style.display = 'none';

    // Add spinning animation
    canvas.classList.add('wheel-spinning');

    try {
        const linkType = document.querySelector('input[name="linkType"]:checked').value;
        const result = await apiCall(`/distribution/spin?type=${linkType}`, { method: 'POST' });

        // Wait for animation to complete
        setTimeout(() => {
            displayWinner(result);
            canvas.classList.remove('wheel-spinning');
            refreshStatus();
            loadEligibleMembersForWheel(); // Reload wheel with updated members
        }, 3000);

    } catch (error) {
        canvas.classList.remove('wheel-spinning');
        console.error('Failed to spin wheel:', error);
    } finally {
        setTimeout(() => {
            isSpinning = false;
            spinButton.disabled = false;
            spinButton.textContent = '🎯 SPIN THE WHEEL';
        }, 3000);
    }
}

function displayWinner(result) {
    const winnerDisplay = document.getElementById('winner-display');
    const winnerInfo = document.getElementById('winner-info');

    const compensationText = result.is_compensation ? ' (Compensation)' : '';
    const linkTypeText = result.link_history.link_type.charAt(0).toUpperCase() + result.link_history.link_type.slice(1);

    winnerInfo.innerHTML = `
        <div class="winner-card">
            <h4>${result.winner.discord_username}${compensationText}</h4>
            <p><strong>Link Type:</strong> ${linkTypeText}</p>
            <p><strong>Date:</strong> ${new Date(result.link_history.date_received).toLocaleDateString()}</p>
            <p><strong>Days in Guild:</strong> ${Math.floor((new Date() - new Date(result.winner.guild_join_date)) / (1000 * 60 * 60 * 24))}</p>
            ${result.winner.character_names.length > 0 ? `<p><strong>Characters:</strong> ${result.winner.character_names.join(', ')}</p>` : ''}
        </div>
    `;

    winnerDisplay.style.display = 'block';
    showNotification(`${result.winner.discord_username} won a ${linkTypeText} link!`, 'success');
}

// Update wheel when link type changes
document.addEventListener('change', function (e) {
    if (e.target.name === 'linkType') {
        loadEligibleMembersForWheel();
    }
});

// Members functions
async function loadMembers() {
    try {
        currentMembers = await apiCall('/members');
        displayMembers(currentMembers);
    } catch (error) {
        console.error('Failed to load members:', error);
        currentMembers = [];
        displayMembers([]);
    }
}

function displayMembers(members) {
    const membersList = document.getElementById('members-list');

    if (members.length === 0) {
        membersList.innerHTML = '<p>No members found.</p>';
        return;
    }

    membersList.innerHTML = members.map(member => `
        <div class="member-card">
            <div class="member-header">
                <div class="member-name">${member.discord_username}</div>
                <div class="member-role role-${member.role}">${member.role}</div>
            </div>
            <div class="member-details">
                <div class="member-detail">
                    <span class="detail-label">Discord ID:</span>
                    <span>${member.discord_id}</span>
                </div>
                <div class="member-detail">
                    <span class="detail-label">Guild Join Date:</span>
                    <span>${new Date(member.guild_join_date).toLocaleDateString()}</span>
                </div>
                <div class="member-detail">
                    <span class="detail-label">Days in Guild:</span>
                    <span>${Math.floor((new Date() - new Date(member.guild_join_date)) / (1000 * 60 * 60 * 24))}</span>
                </div>
                <div class="member-detail">
                    <span class="detail-label">Weekly Boss:</span>
                    <span>${member.weekly_boss_participation ? '✅' : '❌'}</span>
                </div>
                <div class="member-detail">
                    <span class="detail-label">Omni Absences:</span>
                    <span>${member.omni_absence_count}</span>
                </div>
                <div class="member-detail">
                    <span class="detail-label">Characters:</span>
                    <span>${member.character_names.join(', ') || 'None'}</span>
                </div>
            </div>
        </div>
    `).join('');
}

function filterMembers() {
    const searchTerm = document.getElementById('member-search').value.toLowerCase();
    const filteredMembers = currentMembers.filter(member =>
        member.discord_username.toLowerCase().includes(searchTerm) ||
        member.character_names.some(name => name.toLowerCase().includes(searchTerm))
    );
    displayMembers(filteredMembers);
}

function showAddMemberForm() {
    document.getElementById('add-member-form').style.display = 'block';
}

function hideAddMemberForm() {
    document.getElementById('add-member-form').style.display = 'none';
    // Reset form
    document.querySelector('#add-member-form form').reset();
}

async function addMember(event) {
    event.preventDefault();

    const formData = new FormData(event.target);
    const characterNames = formData.get('character-names')
        ? formData.get('character-names').split(',').map(name => name.trim()).filter(name => name)
        : [];

    const memberData = {
        discord_id: document.getElementById('discord-id').value,
        discord_username: document.getElementById('discord-username').value,
        character_names: characterNames,
        guild_join_date: new Date(document.getElementById('guild-join-date').value).toISOString(),
        role: document.getElementById('role').value
    };

    try {
        await apiCall('/member/create', {
            method: 'POST',
            body: JSON.stringify(memberData)
        });

        showNotification('Member added successfully', 'success');
        hideAddMemberForm();
        loadMembers();
        refreshStatus();
    } catch (error) {
        console.error('Failed to add member:', error);
    }
}

// History functions
async function loadHistory() {
    // This would need to be implemented with a proper history endpoint
    // For now, show a placeholder
    const historyList = document.getElementById('history-list');
    historyList.innerHTML = '<p>History functionality will be implemented with proper backend support.</p>';
}

// Utility functions
function showLoading() {
    document.getElementById('loading').style.display = 'flex';
}

function hideLoading() {
    document.getElementById('loading').style.display = 'none';
}

function showNotification(message, type = 'success') {
    const notification = document.getElementById('notification');
    notification.textContent = message;
    notification.className = `notification ${type}`;
    notification.style.display = 'block';

    setTimeout(() => {
        notification.style.display = 'none';
    }, 5000);
}

// Error handling for uncaught errors
window.addEventListener('error', function (e) {
    console.error('Uncaught error:', e.error);
    showNotification('An unexpected error occurred', 'error');
});

window.addEventListener('unhandledrejection', function (e) {
    console.error('Unhandled promise rejection:', e.reason);
    showNotification('An unexpected error occurred', 'error');
});
