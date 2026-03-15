function setupThemeAndMenu() {
    const themeToggle = document.getElementById('themeToggle');
    const themeText = document.getElementById('themeText');
    const body = document.body;
    const settingsToggle = document.getElementById('settingsToggle');
    const settingsMenu = document.getElementById('settingsMenu');

    const savedTheme = localStorage.getItem('theme') || 'light';
    body.setAttribute('data-theme', savedTheme);
    updateThemeText(savedTheme);

    if (themeToggle) {
        themeToggle.addEventListener('click', function(e) {
            e.stopPropagation();
            const currentTheme = body.getAttribute('data-theme');
            const newTheme = currentTheme === 'light' ? 'dark' : 'light';

            body.setAttribute('data-theme', newTheme);
            localStorage.setItem('theme', newTheme);
            updateThemeText(newTheme);
        });
    }

    function updateThemeText(theme) {
        if (themeText) {
            themeText.textContent = theme === 'dark' ? 'Светлая тема' : 'Тёмная тема';
        }
    }

    if (settingsToggle && settingsMenu) {
        settingsToggle.addEventListener('click', function(e) {
            e.stopPropagation();
            settingsMenu.classList.toggle('show');
        });

        document.addEventListener('click', function(e) {
            if (!settingsMenu.contains(e.target) && !settingsToggle.contains(e.target)) {
                settingsMenu.classList.remove('show');
            }
        });

        settingsMenu.addEventListener('click', function(e) {
            e.stopPropagation();
        });
    }
}

async function updateStatus(status) {
    const applicantId = document.body.dataset.applicantId;
    const directionId = document.body.dataset.directionId || document.querySelector('.direction_card')?.dataset?.directionId;
    if (!applicantId || !directionId) return;

    try {
        const response = await fetch(`/users/applicants/${applicantId}/status`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ direction_id: directionId, status })
        });
        const result = await response.json();
        if (!response.ok) {
            alert(result.detail || 'Ошибка обновления статуса');
            return;
        }

        const statusLabel = {
            accepted: 'Зачислен',
            pending: 'На рассмотрении',
            rejected: 'Отклонён'
        }[status];

        const statusClass = {
            accepted: 'accepted',
            pending: 'pending',
            rejected: 'rejected'
        }[status];

        const card = document.querySelector('.direction_card');
        if (card) {
            const badge = card.querySelector('.direction_status');
            if (badge) {
                badge.textContent = statusLabel;
                badge.className = `direction_status ${statusClass}`;
            }
        }

        alert('Статус обновлён');
    } catch (err) {
        console.error(err);
        alert('Ошибка сети. Попробуйте позже.');
    }
}

document.addEventListener('DOMContentLoaded', function() {
    setupThemeAndMenu();

    document.querySelectorAll('.status_btn').forEach(btn => {
        btn.addEventListener('click', function() {
            const status = btn.dataset.status;
            if (status) updateStatus(status);
        });
    });
});