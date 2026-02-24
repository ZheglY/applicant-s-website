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
document.addEventListener('DOMContentLoaded', function() {
    const firstDirection = document.querySelector('.direction_item');
    if (firstDirection) {
        const header = firstDirection.querySelector('.direction_header');
        const content = firstDirection.querySelector('.direction_content');
        const icon = firstDirection.querySelector('.direction_icon');
        
        content.classList.add('open');
        icon.textContent = '▼';
        firstDirection.classList.add('active');
    }
    
    // Загружаем сохранённые статусы
    loadSavedStatuses();
});

// Функция для загрузки сохранённых статусов
function loadSavedStatuses() {
    const savedStatuses = JSON.parse(localStorage.getItem('applicantStatuses')) || {};
    
    document.querySelectorAll('.abiturients_table tbody tr').forEach(row => {
        const statusCell = row.querySelector('td:nth-child(7)');
        const button = row.querySelector('.detail_btn');
        if (!button) return;
        
        const onclickAttr = button.getAttribute('onclick');
        const match = onclickAttr.match(/viewProfile\((\d+)\)/);
        if (!match) return;
        
        const applicantId = match[1];
        
        if (savedStatuses[applicantId]) {
            // Берём первый статус из сохранённых
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

// Функция для перехода в профиль
function viewProfile(id) {
    window.location.href = `/users/applicants/${id}`;
}

// Сохранение статуса (будет вызываться из lk.html)
window.saveApplicantStatus = function(applicantId, directionId, status) {
    const savedStatuses = JSON.parse(localStorage.getItem('applicantStatuses')) || {};
    
    if (!savedStatuses[applicantId]) {
        savedStatuses[applicantId] = {};
    }
    
    savedStatuses[applicantId][directionId] = status;
    localStorage.setItem('applicantStatuses', JSON.stringify(savedStatuses));
    
    // Обновляем статус в таблице
    loadSavedStatuses();
    
    if (window.ProfileManager) {
        ProfileManager.showNotification('Статус обновлён', 'success');
    }
};