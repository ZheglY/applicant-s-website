// ===== УПРАВЛЕНИЕ ТЁМНОЙ ТЕМОЙ И МЕНЮ =====
document.addEventListener('DOMContentLoaded', function() {
    const themeToggle = document.getElementById('themeToggle');
    const themeText = document.getElementById('themeText');
    const body = document.body;
    const settingsToggle = document.getElementById('settingsToggle');
    const settingsMenu = document.getElementById('settingsMenu');
    
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
    
    // Меню настроек
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
    
    // ===== ЗАГРУЗКА ДАННЫХ АБИТУРИЕНТА =====
    loadApplicantData();
});

// ===== ПОЛНАЯ БАЗА ДАННЫХ АБИТУРИЕНТОВ =====
const applicantsData = {
    1: {
        name: 'Иванов Иван Иванович',
        age: '19 лет',
        region: 'г. Москва',
        totalScore: 273,
        bonusPoints: 5,
        ege: [
            { subject: 'Русский язык', score: 91 },
            { subject: 'Математика (проф.)', score: 94 },
            { subject: 'Информатика', score: 88 }
        ],
        achievements: [
            { text: 'Победитель олимпиады по математике', points: 3 },
            { text: 'Волонтёрская деятельность', points: 2 }
        ],
        documents: [
            { name: 'Паспорт', status: 'verified' },
            { name: 'Аттестат', status: 'pending' },
            { name: 'Диплом олимпиады', status: 'verified' }
        ],
        directions: [
            { 
                priority: 1, 
                name: 'Программная инженерия', 
                status: 'accepted',
                score: 273,
                position: '12 из 45',
                type: 'Бюджет'
            },
            { 
                priority: 2, 
                name: 'Искусственный интеллект', 
                status: 'pending',
                score: 273,
                position: '8 из 40',
                type: 'Бюджет'
            },
            { 
                priority: 3, 
                name: 'Бизнес-аналитика', 
                status: 'rejected',
                score: 273,
                reason: 'Высокий конкурс'
            }
        ],
        contacts: {
            email: 'ivanov.ivan@example.com',
            phone: '+7 (999) 123-45-67',
            telegram: '@ivanov_ivan'
        }
    },
    // ... остальные данные (2-18) как в вашем коде ...
};

// Функция загрузки данных абитуриента
function loadApplicantData() {
    const urlParams = new URLSearchParams(window.location.search);
    const applicantId = urlParams.get('id') || '1';
    
    // Загружаем сохранённые статусы
    const savedStatuses = JSON.parse(localStorage.getItem('applicantStatuses')) || {};
    
    // Получаем данные абитуриента
    const data = applicantsData[applicantId];
    
    if (!data) {
        console.error('Абитуриент не найден');
        return;
    }
    
    // Применяем сохранённые статусы к направлениям
    if (savedStatuses[applicantId]) {
        data.directions.forEach((dir, index) => {
            if (savedStatuses[applicantId][dir.name]) {
                dir.status = savedStatuses[applicantId][dir.name];
            }
        });
    }
    
    // Заполняем основную информацию
    document.getElementById('profileName').textContent = data.name;
    document.getElementById('profileAge').textContent = data.age;
    document.getElementById('profileRegion').textContent = data.region;
    document.getElementById('totalEgeScore').textContent = data.totalScore;
    document.getElementById('bonusPoints').textContent = `+${data.bonusPoints}`;
    
    // Заполняем ЕГЭ
    const egeContainer = document.getElementById('egeSubjects');
    egeContainer.innerHTML = '';
    data.ege.forEach(item => {
        egeContainer.innerHTML += `
            <div class="subject_ege">
                <span class="subject_name">${item.subject}</span>
                <span class="subject_score">${item.score}</span>
            </div>
        `;
    });
    
    // Заполняем достижения
    const achievementsContainer = document.getElementById('achievementsList');
    achievementsContainer.innerHTML = '';
    data.achievements.forEach(item => {
        achievementsContainer.innerHTML += `
            <div class="achievement_item">
                <span>${item.text}</span>
                <span class="achievement_points">+${item.points}</span>
            </div>
        `;
    });
    
    // Заполняем документы
    const documentsContainer = document.getElementById('documentsList');
    documentsContainer.innerHTML = '';
    data.documents.forEach(item => {
        let statusClass = '';
        let statusText = '';
        
        switch(item.status) {
            case 'verified':
                statusClass = 'verified';
                statusText = 'Подтверждён';
                break;
            case 'pending':
                statusClass = 'pending';
                statusText = 'На проверке';
                break;
        }
        
        documentsContainer.innerHTML += `
            <div class="document_item">
                <div class="document_info">
                    <span class="document_name">${item.name}</span>
                </div>
                <span class="document_status ${statusClass}">${statusText}</span>
            </div>
        `;
    });
    
    document.querySelector('.btn_show_all').textContent = `Показать все документы (${data.documents.length})`;
    
    // Заполняем направления
    const directionsContainer = document.getElementById('directionsList');
    directionsContainer.innerHTML = '';
    data.directions.forEach(dir => {
        let statusClass = '';
        let statusText = '';
        
        switch(dir.status) {
            case 'accepted':
                statusClass = 'accepted';
                statusText = 'Зачислен';
                break;
            case 'pending':
                statusClass = 'pending';
                statusText = 'Рассматривается';
                break;
            case 'rejected':
                statusClass = 'rejected';
                statusText = 'Отклонён';
                break;
        }
        
        let details = '';
        if (dir.status === 'rejected') {
            details = `
                <div class="detail_row">
                    <span>Конкурсный балл:</span>
                    <span class="detail_value">${dir.score}</span>
                </div>
                <div class="detail_row">
                    <span>Причина:</span>
                    <span class="detail_value">${dir.reason}</span>
                </div>
            `;
        } else {
            details = `
                <div class="detail_row">
                    <span>Конкурсный балл:</span>
                    <span class="detail_value">${dir.score}</span>
                </div>
                <div class="detail_row">
                    <span>Место в рейтинге:</span>
                    <span class="detail_value">${dir.position}</span>
                </div>
                <div class="detail_row">
                    <span>Тип места:</span>
                    <span class="detail_value budget">${dir.type}</span>
                </div>
            `;
        }
        
        directionsContainer.innerHTML += `
            <div class="direction_card" data-direction="${dir.name}">
                <div class="direction_header">
                    <div>
                        <span class="direction_priority priority_${dir.priority}">Приоритет ${dir.priority}</span>
                        <h4 class="direction_name">${dir.name}</h4>
                    </div>
                    <span class="direction_status ${statusClass}">${statusText}</span>
                </div>
                <div class="direction_details">
                    ${details}
                </div>
            </div>
        `;
    });
    
    // Заполняем контакты
    const contactsContainer = document.querySelector('.contacts_list');
    contactsContainer.innerHTML = `
        <div class="contact_item">
            <span class="contact_value">${data.contacts.email}</span>
        </div>
        <div class="contact_item">
            <span class="contact_value">${data.contacts.phone}</span>
        </div>
        <div class="contact_item">
            <span class="contact_value">${data.contacts.telegram}</span>
        </div>
    `;
}

// ИСПРАВЛЕННАЯ ФУНКЦИЯ ДЛЯ КНОПКИ "НАЗАД"
function goBack() {
    // Возвращаемся на предыдущую страницу в истории браузера
    window.history.back();
    
    // ИЛИ если хотите всегда возвращаться на список:
    // window.location.href = 'list.html';
}

// Функция смены статуса
function changeStatus(status) {
    const urlParams = new URLSearchParams(window.location.search);
    const applicantId = urlParams.get('id') || '1';
    
    // Получаем название направления из карточки
    const firstDirection = document.querySelector('.direction_card');
    if (!firstDirection) return;
    
    const directionName = firstDirection.querySelector('.direction_name').textContent;
    
    // Обновляем статус в UI
    const statusBadge = firstDirection.querySelector('.direction_status');
    
    let statusClass = '';
    let statusText = '';
    
    switch(status) {
        case 'accepted':
            statusClass = 'accepted';
            statusText = 'Зачислен';
            break;
        case 'pending':
            statusClass = 'pending';
            statusText = 'Рассматривается';
            break;
        case 'rejected':
            statusClass = 'rejected';
            statusText = 'Отклонён';
            break;
    }
    
    statusBadge.className = `direction_status ${statusClass}`;
    statusBadge.textContent = statusText;
    
    // Сохраняем статус в localStorage
    const savedStatuses = JSON.parse(localStorage.getItem('applicantStatuses')) || {};
    
    if (!savedStatuses[applicantId]) {
        savedStatuses[applicantId] = {};
    }
    
    savedStatuses[applicantId][directionName] = status;
    localStorage.setItem('applicantStatuses', JSON.stringify(savedStatuses));
    
    // Визуальная обратная связь
    const statusButtons = document.querySelectorAll('.status_btn');
    statusButtons.forEach(btn => {
        btn.style.opacity = '0.7';
        btn.style.transform = 'scale(0.98)';
    });
    
    setTimeout(() => {
        statusButtons.forEach(btn => {
            btn.style.opacity = '1';
            btn.style.transform = '';
        });
    }, 200);
    
    // Показываем сообщение
    let message = '';
    switch(status) {
        case 'accepted':
            message = 'Абитуриент зачислен';
            break;
        case 'pending':
            message = 'Статус изменён на "На рассмотрении"';
            break;
        case 'rejected':
            message = 'Абитуриент отклонён';
            break;
    }
    
    alert(message);
}