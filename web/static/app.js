// API Base URL - will be set based on environment
const API_BASE = window.location.origin;

// Global state
let currentMembers = [];
let currentInventory = [];
let wheelSpinning = false;

// Initialize the application
document.addEventListener('DOMContentLoaded', function () {
    loadDashboard();
    setupEventListeners();
});

// Event listeners
function setupEventListeners() {
    // Add member form
    document.getElementById('add-member-form').addEventListener('submit', handleAddMember);

    // Add inventory form
    document.getElementById('add-inventory-form').addEventListener('submit', handleAddInventory);

    // Close modals when clicking outside
    window.addEventListener('click', function (event) {
        if (event.target.classList.contains('modal')) {
            event.target.style.display = 'none';
        }
    });
}

// Tab management
function showTab(tabName) {
    // Hide all tab contents
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
    });

    // Remove active class from all tab buttons
    document.querySelectorAll('.tab-button').forEach(button => {
        button.classList.remove('active');
    });

    // Show selected tab
    document.getElementById(tabName).classList.add('active');
    event.target.classList.add('active');

    // Load tab-specific data
    switch (tabName) {
        case 'dashboard':
            loadDashboard();
            break;
        case 'members':
            loadMembers();
            break;
        case 'inventory':
            loadInventory();
            break;
        case 'picker':
            loadPickerWheel();
            break;
        case 'history':
            loadHistory();
            break;
    }
}

// Dashboard functions
async function loadDashboard() {
    try {
        // Load members for stats
        const membersResponse = await fetch(`${API_BASE}/api/members`);
        const membersData = await membersResponse.json();

        if (membersData.success) {
            const members = membersData.data;
            const silverEligible = members.filter(m => m.silver_eligible).length;
            const goldEligible = members.filter(m => m.gold_eligible).length;

            document.getElementById('total-members').textContent = members.length;
            document.getElementById('silver-eligible').textContent = silverEligible;
            document.getElementById('gold-eligible').textContent = goldEligible;
        }

        // Load inventory summary
        const inventoryResponse = await fetch(`${API_BASE}/api/inventory/summary`);
        const inventoryData = await inventoryResponse.json();

        if (inventoryData.success) {
            const summary = inventoryData.data;
            let totalLinks = 0;
            let summaryHTML = '<div class="inventory-summary-grid">';

            for (const [linkType, qualities] of Object.entries(summary)) {
                let linkTotal = 0;
                let qualityText = '';

                for (const [quality, count] of Object.entries(qualities)) {
                    linkTotal += count;
                    totalLinks += count;
                    const emoji = getQualityEmoji(quality);
                    qualityText += `${emoji} ${count} `;
                }

                summaryHTML += `
                    <div class="summary-item">
                        <div class="link-type">${linkType}</div>
                        <div class="link-counts">${qualityText}</div>
                        <div class="link-total">Total: ${linkTotal}</div>
                    </div>
                `;
            }

            summaryHTML += '</div>';
            document.getElementById('inventory-summary-content').innerHTML = summaryHTML;
            document.getElementById('total-links').textContent = totalLinks;
        }

    } catch (error) {
        console.error('Error loading dashboard:', error);
    }
}

// Members functions
async function loadMembers() {
    try {
        const response = await fetch(`${API_BASE}/api/members`);
        const data = await response.json();

        if (data.success) {
            currentMembers = data.data;
            displayMembers(currentMembers);
        } else {
            document.getElementById('members-list').innerHTML = `<p class="error">Error: ${data.error}</p>`;
        }
    } catch (error) {
        console.error('Error loading members:', error);
        document.getElementById('members-list').innerHTML = '<p class="error">Failed to load members</p>';
    }
}

function displayMembers(members) {
    const container = document.getElementById('members-list');

    if (members.length === 0) {
        container.innerHTML = '<p>No members found</p>';
        return;
    }

    const membersHTML = members.map(member => `
        <div class="member-card">
            <div class="member-header">
                <h4>${member.username}</h4>
                <span class="rank-badge rank-${member.rank.toLowerCase().replace(' ', '-')}">${member.rank}</span>
            </div>
            <div class="member-details">
                <p><strong>Days in Guild:</strong> ${member.days_in_guild}</p>
                <p><strong>Silver Eligible:</strong> ${member.silver_eligible ? '✅' : '❌'}</p>
                <p><strong>Gold Eligible:</strong> ${member.gold_eligible ? '✅' : '❌'}</p>
                <p><strong>Added:</strong> ${new Date(member.added_date).toLocaleDateString()}</p>
            </div>
            ${member.is_officer ? '' : `
                <button class="btn btn-secondary btn-sm" onclick="promoteMember('${member.discord_id}')">
                    Promote to Maester
                </button>
            `}
        </div>
    `).join('');

    container.innerHTML = membersHTML;
}

// Inventory functions
async function loadInventory() {
    try {
        const response = await fetch(`${API_BASE}/api/inventory`);
        const data = await response.json();

        if (data.success) {
            currentInventory = data.data;
            displayInventory(currentInventory);
        } else {
            document.getElementById('inventory-list').innerHTML = `<p class="error">Error: ${data.error}</p>`;
        }
    } catch (error) {
        console.error('Error loading inventory:', error);
        document.getElementById('inventory-list').innerHTML = '<p class="error">Failed to load inventory</p>';
    }
}

function displayInventory(inventory) {
    const container = document.getElementById('inventory-list');

    if (inventory.length === 0) {
        container.innerHTML = '<p>No links in inventory</p>';
        return;
    }

    // Group by link type and quality
    const grouped = {};
    inventory.forEach(link => {
        const key = `${link.link_type}_${link.quality}`;
        if (!grouped[key]) {
            grouped[key] = {
                link_type: link.link_type,
                quality: link.quality,
                bonus: link.bonus,
                category: link.category,
                count: 0
            };
        }
        grouped[key].count++;
    });

    const inventoryHTML = Object.values(grouped).map(item => `
        <div class="inventory-card">
            <div class="inventory-header">
                <h4>${item.link_type}</h4>
                <span class="quality-badge quality-${item.quality}">${item.quality.toUpperCase()}</span>
            </div>
            <div class="inventory-details">
                <p><strong>Bonus:</strong> ${item.bonus}</p>
                <p><strong>Category:</strong> ${item.category}</p>
                <p><strong>Available:</strong> ${item.count}</p>
            </div>
        </div>
    `).join('');

    container.innerHTML = inventoryHTML;
}

// Picker wheel functions
async function loadPickerWheel() {
    drawWheel([]);
}

async function spinWheel() {
    if (wheelSpinning) return;

    const quality = document.getElementById('picker-quality').value;

    try {
        // Get eligible members
        const response = await fetch(`${API_BASE}/api/distribution/eligible?quality=${quality}`);
        const data = await response.json();

        if (!data.success) {
            alert(`Error: ${data.error}`);
            return;
        }

        const eligibleMembers = data.data;

        if (eligibleMembers.length === 0) {
            alert(`No members eligible for ${quality} links`);
            return;
        }

        // Pick random winner
        const winnerResponse = await fetch(`${API_BASE}/api/distribution/pick-winner?list_id=temp`, {
            method: 'POST'
        });
        const winnerData = await winnerResponse.json();

        if (winnerData.success) {
            const winner = winnerData.data.winner;
            animateWheel(eligibleMembers, winner);
        } else {
            alert(`Error picking winner: ${winnerData.error}`);
        }

    } catch (error) {
        console.error('Error spinning wheel:', error);
        alert('Failed to spin wheel');
    }
}

function drawWheel(members) {
    const canvas = document.getElementById('wheel');
    const ctx = canvas.getContext('2d');
    const centerX = canvas.width / 2;
    const centerY = canvas.height / 2;
    const radius = 180;

    // Clear canvas
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    if (members.length === 0) {
        // Draw empty wheel
        ctx.beginPath();
        ctx.arc(centerX, centerY, radius, 0, 2 * Math.PI);
        ctx.fillStyle = '#f0f0f0';
        ctx.fill();
        ctx.strokeStyle = '#ccc';
        ctx.lineWidth = 2;
        ctx.stroke();

        ctx.fillStyle = '#666';
        ctx.font = '16px Arial';
        ctx.textAlign = 'center';
        ctx.fillText('Select quality and spin!', centerX, centerY);
        return;
    }

    const anglePerSegment = (2 * Math.PI) / members.length;
    const colors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FFEAA7', '#DDA0DD', '#98D8C8', '#F7DC6F'];

    // Draw segments
    members.forEach((member, index) => {
        const startAngle = index * anglePerSegment;
        const endAngle = startAngle + anglePerSegment;

        ctx.beginPath();
        ctx.moveTo(centerX, centerY);
        ctx.arc(centerX, centerY, radius, startAngle, endAngle);
        ctx.closePath();

        ctx.fillStyle = colors[index % colors.length];
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
        ctx.fillStyle = '#000';
        ctx.font = '12px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(member.username, 0, 0);
        ctx.restore();
    });

    // Draw center circle
    ctx.beginPath();
    ctx.arc(centerX, centerY, 20, 0, 2 * Math.PI);
    ctx.fillStyle = '#333';
    ctx.fill();
}

function animateWheel(members, winner) {
    wheelSpinning = true;
    const canvas = document.getElementById('wheel');
    const ctx = canvas.getContext('2d');

    // Find winner index
    const winnerIndex = members.findIndex(m => m.discord_id === winner.discord_id);
    const anglePerSegment = (2 * Math.PI) / members.length;
    const targetAngle = winnerIndex * anglePerSegment + anglePerSegment / 2;

    let currentAngle = 0;
    const spinDuration = 3000; // 3 seconds
    const startTime = Date.now();
    const totalRotation = Math.PI * 8 + targetAngle; // Multiple spins + target

    function animate() {
        const elapsed = Date.now() - startTime;
        const progress = Math.min(elapsed / spinDuration, 1);

        // Easing function for smooth deceleration
        const easeOut = 1 - Math.pow(1 - progress, 3);
        currentAngle = totalRotation * easeOut;

        // Redraw wheel with rotation
        ctx.save();
        ctx.translate(canvas.width / 2, canvas.height / 2);
        ctx.rotate(-currentAngle);
        ctx.translate(-canvas.width / 2, -canvas.height / 2);

        drawWheel(members);

        ctx.restore();

        if (progress < 1) {
            requestAnimationFrame(animate);
        } else {
            // Show winner
            showWinner(winner);
            wheelSpinning = false;
        }
    }

    animate();
}

function showWinner(winner) {
    const resultDiv = document.getElementById('winner-result');
    const infoDiv = document.getElementById('winner-info');

    infoDiv.innerHTML = `
        <div class="winner-card">
            <h4>${winner.username}</h4>
            <p><strong>Rank:</strong> ${winner.rank}</p>
            <p><strong>Days in Guild:</strong> ${winner.days_in_guild}</p>
        </div>
    `;

    resultDiv.style.display = 'block';
}

// History functions
async function loadHistory() {
    try {
        const response = await fetch(`${API_BASE}/api/distribution/history`);
        const data = await response.json();

        if (data.success) {
            displayHistory(data.data);
        } else {
            document.getElementById('history-list').innerHTML = `<p class="error">Error: ${data.error}</p>`;
        }
    } catch (error) {
        console.error('Error loading history:', error);
        document.getElementById('history-list').innerHTML = '<p class="error">Failed to load history</p>';
    }
}

function displayHistory(history) {
    const container = document.getElementById('history-list');

    if (history.length === 0) {
        container.innerHTML = '<p>No distribution history found</p>';
        return;
    }

    // Sort by date (newest first)
    history.sort((a, b) => new Date(b.distributed_at) - new Date(a.distributed_at));

    const historyHTML = history.map(item => `
        <div class="history-item">
            <div class="history-header">
                <h4>${item.member_username}</h4>
                <span class="quality-badge quality-${item.quality}">${item.quality.toUpperCase()}</span>
            </div>
            <div class="history-details">
                <p><strong>Link:</strong> ${item.link_type} (${item.bonus})</p>
                <p><strong>Date:</strong> ${new Date(item.distributed_at).toLocaleString()}</p>
                <p><strong>Method:</strong> ${item.method}</p>
                <p><strong>Distributed by:</strong> ${item.distributed_by}</p>
            </div>
        </div>
    `).join('');

    container.innerHTML = historyHTML;
}

// Modal functions
function showAddMemberModal() {
    document.getElementById('add-member-modal').style.display = 'block';
}

function showAddInventoryModal() {
    document.getElementById('add-inventory-modal').style.display = 'block';
    populateLinkTypeDropdown(); // Populate dropdown when modal opens
}

// Populate link type dropdown with all available link types
function populateLinkTypeDropdown() {
    const linkTypes = [
        "Aegis Keep Damage",
        "Alchemy/Healing/Veterinary",
        "Backstab Damage",
        "Bard Reset/Break Ignore Chance",
        "Barding Effect Durations",
        "Boss Damage Resistance",
        "Cavernam Damage",
        "Chance on Stealth for 5 Extra Steps",
        "Chest Success Chances/Progress",
        "Chivalry Skill",
        "Crewmember Damage",
        "Crewmember Damage Resistance",
        "Damage Dealt By Player",
        "Damage on Ships",
        "Damage Resistance",
        "Damage Resistance on Ships",
        "Damage to Barded Creatures",
        "Damage to Bestial Creatures",
        "Damage to Bleeding Creatures",
        "Damage to Bosses",
        "Damage to Construct Creatures",
        "Damage to Creatures Above 66% HP",
        "Damage to Creatures Below 33% HP",
        "Damage to Daemonic Creatures",
        "Damage to Diseased Creatures",
        "Damage to Elemental Creatures",
        "Damage to Humanoid Creatures",
        "Damage to Monstrous Creatures",
        "Damage to Nature Creatures",
        "Damage to Poisoned Creatures",
        "Damage to Undead Creatures",
        "Darkmire Temple Damage",
        "Effective Alchemy Skill",
        "Effective Arms Lore",
        "Effective Barding Skill",
        "Effective Camping Skill",
        "Effective Harvest Skill",
        "Effective Magic Resist Skill",
        "Effective Parrying Skill",
        "Effective Poisoning Skill",
        "Effective Skill on Chests",
        "Exceptional Quality Chance",
        "Follower Accuracy/Defense",
        "Follower Attack Speed",
        "Follower Damage",
        "Follower Damage Resistance",
        "Follower Healing Received",
        "Gold/Doubloon Drop Increase",
        "Healing Received",
        "Inferno Damage",
        "Kraul Hive Damage",
        "Mausoleum Damage",
        "Meditation Rate",
        "Meditation Rate/Disrupt Avoid Chance",
        "Melee Accuracy",
        "Melee Accuracy/Defense",
        "Melee Aspect Effect Chance",
        "Melee Aspect Effect Modifier",
        "Melee Damage",
        "Melee Damage/Ignore Armor Chance",
        "Melee Defense",
        "Melee Ignore Armor Chance",
        "Melee Special Chance",
        "Melee Special Chance/Special Damage",
        "Melee Swing Speed",
        "Mount Petram Damage",
        "Necromancy Skill",
        "Netherzone Damage",
        "Nusero Damage",
        "Ossuary Damage",
        "Physical Damage Resistance",
        "Poison Damage",
        "Poison Damage/Resist Ignore",
        "Pulma Damage",
        "Rare Loot Chance",
        "Shadowspire Cathedral Damage",
        "Ship Cannon Damage",
        "Special Loot Chance",
        "Special/Rare Loot Chance",
        "Spell Aspect Effect Modifier",
        "Spell Aspect Special Chance",
        "Spell Charged Chance",
        "Spell Charged Chance/Charged Damage",
        "Spell Charged Damage",
        "Spell Damage",
        "Spell Damage Resistance",
        "Spell Damage When No Followers",
        "Spell Damage/Ignore Resist Chance",
        "Spell Disrupt Avoid Chance",
        "Spell Ignore Resist Chance",
        "Spirit Speak/Inscription",
        "Summon Duration and Dispel Resist",
        "Time Dungeon Damage",
        "Trap Damage",
        "Wilderness Damage"
    ];

    const selectElement = document.getElementById('link-type');

    // Clear existing options except the first placeholder
    selectElement.innerHTML = '<option value="">Select a link type...</option>';

    // Add all link types as options
    linkTypes.forEach(linkType => {
        const option = document.createElement('option');
        option.value = linkType;
        option.textContent = linkType;
        selectElement.appendChild(option);
    });
}

function closeModal(modalId) {
    document.getElementById(modalId).style.display = 'none';
}

// Form handlers
async function handleAddMember(event) {
    event.preventDefault();

    const formData = {
        discord_id: document.getElementById('discord-id').value,
        username: document.getElementById('username').value,
        join_date: document.getElementById('join-date').value + 'T00:00:00Z'
    };

    try {
        const response = await fetch(`${API_BASE}/api/member/create`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(formData)
        });

        const data = await response.json();

        if (data.success) {
            alert('Member added successfully!');
            closeModal('add-member-modal');
            document.getElementById('add-member-form').reset();
            loadMembers(); // Refresh members list
            loadDashboard(); // Refresh dashboard stats
        } else {
            alert(`Error: ${data.error}`);
        }
    } catch (error) {
        console.error('Error adding member:', error);
        alert('Failed to add member');
    }
}

async function handleAddInventory(event) {
    event.preventDefault();

    const formData = {
        link_type: document.getElementById('link-type').value,
        quality: document.getElementById('quality').value,
        count: parseInt(document.getElementById('count').value)
    };

    try {
        const response = await fetch(`${API_BASE}/api/inventory/add`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(formData)
        });

        const data = await response.json();

        if (data.success) {
            alert(`Added ${formData.count} ${formData.quality} ${formData.link_type} links!`);
            closeModal('add-inventory-modal');
            document.getElementById('add-inventory-form').reset();
            loadInventory(); // Refresh inventory list
            loadDashboard(); // Refresh dashboard stats
        } else {
            alert(`Error: ${data.error}`);
        }
    } catch (error) {
        console.error('Error adding inventory:', error);
        alert('Failed to add inventory');
    }
}

// Member actions
async function promoteMember(discordId) {
    if (!confirm('Are you sure you want to promote this member to Maester?')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/api/member/promote?discord_id=${discordId}`, {
            method: 'POST'
        });

        const data = await response.json();

        if (data.success) {
            alert('Member promoted successfully!');
            loadMembers(); // Refresh members list
        } else {
            alert(`Error: ${data.error}`);
        }
    } catch (error) {
        console.error('Error promoting member:', error);
        alert('Failed to promote member');
    }
}

// Utility functions
function getQualityEmoji(quality) {
    switch (quality) {
        case 'bronze': return '🥉';
        case 'silver': return '🥈';
        case 'gold': return '🥇';
        default: return '⚪';
    }
}
