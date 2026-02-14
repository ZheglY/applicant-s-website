// ===== УПРАВЛЕНИЕ ТЁМНОЙ ТЕМОЙ =====
document.addEventListener('DOMContentLoaded', function() {
    const themeToggle = document.getElementById('themeToggle');
    const themeText = document.getElementById('themeText');
    const body = document.body;
    
    // Тема
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
    
    // ===== УПРАВЛЕНИЕ МЕНЮ НАСТРОЕК =====
    const settingsToggle = document.getElementById('settingsToggle');
    const settingsMenu = document.getElementById('settingsMenu');
    
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
    
    // ===== УПРАВЛЕНИЕ НАПРАВЛЕНИЯМИ (АККОРДЕОН) =====
    window.toggleDirection = function(header) {
        const directionItem = header.closest('.direction_item');
        const content = directionItem.querySelector('.direction_content');
        const icon = header.querySelector('.direction_icon');
        
        content.classList.toggle('open');
        
        if (content.classList.contains('open')) {
            icon.textContent = '▼';
            directionItem.classList.add('active');
        } else {
            icon.textContent = '▶';
            directionItem.classList.remove('active');
        }
    };
    
    // Открываем первое направление по умолчанию
    const firstDirection = document.querySelector('.direction_item');
    if (firstDirection) {
        const header = firstDirection.querySelector('.direction_header');
        const content = firstDirection.querySelector('.direction_content');
        const icon = firstDirection.querySelector('.direction_icon');
        
        content.classList.add('open');
        icon.textContent = '▼';
        firstDirection.classList.add('active');
    }
    
    // ===== ЗАГРУЗКА СОХРАНЁННЫХ СТАТУСОВ =====
    loadSavedStatuses();
});

// Функция для загрузки сохранённых статусов
function loadSavedStatuses() {
    const savedStatuses = JSON.parse(localStorage.getItem('applicantStatuses')) || {};
    
    document.querySelectorAll('.abiturients_table tbody tr').forEach((row, index) => {
        const statusCell = row.querySelector('td:nth-child(7)');
        const button = row.querySelector('.detail_btn');
        if (!button) return;
        
        const onclickAttr = button.getAttribute('onclick');
        const match = onclickAttr.match(/viewProfile\((\d+)\)/);
        if (!match) return;
        
        const applicantId = match[1];
        
        if (savedStatuses[applicantId]) {
            // Берем первый статус из сохраненных
            const firstDirectionStatus = Object.values(savedStatuses[applicantId])[0];
            
            if (firstDirectionStatus) {
                let statusClass = '';
                let statusText = '';
                
                switch(firstDirectionStatus) {
                    case 'accepted':
                        statusClass = 'status_accepted';
                        statusText = 'Зачислен';
                        break;
                    case 'pending':
                        statusClass = 'status_pending';
                        statusText = 'На рассмотрении';
                        break;
                    case 'rejected':
                        statusClass = 'status_rejected';
                        statusText = 'Отклонён';
                        break;
                }
                
                statusCell.innerHTML = `<span class="status_badge ${statusClass}">${statusText}</span>`;
            }
        }
    });
}

// Функция для перехода в профиль (lk.html)
function viewProfile(id) {
    window.location.href = `lk.html?id=${id}`;
}