// ===== УПРАВЛЕНИЕ ТЁМНОЙ ТЕМОЙ =====
const themeToggle = document.getElementById('themeToggle');
const themeText = document.getElementById('themeText');
const body = document.body;

// Проверяем сохранённую тему
const savedTheme = localStorage.getItem('theme') || 'light';
body.setAttribute('data-theme', savedTheme);
updateThemeText(savedTheme);

themeToggle.addEventListener('click', function(e) {
    e.stopPropagation();
    const currentTheme = body.getAttribute('data-theme');
    const newTheme = currentTheme === 'light' ? 'dark' : 'light';
    
    body.setAttribute('data-theme', newTheme);
    localStorage.setItem('theme', newTheme);
    updateThemeText(newTheme);
});

function updateThemeText(theme) {
    if (theme === 'dark') {
        themeText.textContent = 'Светлая тема';
    } else {
        themeText.textContent = 'Тёмная тема';
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

// ===== УПРАВЛЕНИЕ МОДАЛЬНЫМ ОКНОМ =====
const modalOverlay = document.getElementById('modalOverlay');
const addNewsBtn = document.getElementById('addNewsBtn');
const closeModalBtn = document.getElementById('closeModalBtn');
const cancelModalBtn = document.getElementById('cancelModalBtn');
const submitNewsBtn = document.getElementById('submitNewsBtn');

// Поля формы
const newsTitle = document.getElementById('newsTitle');
const newsSubtitle = document.getElementById('newsSubtitle');
const newsText = document.getElementById('newsText');
const newsImage = document.getElementById('newsImage');

// Функция открытия модального окна
function openModal() {
    modalOverlay.classList.add('active');
    body.style.overflow = 'hidden'; // Блокируем прокрутку страницы
}

// Функция закрытия модального окна
function closeModal() {
    modalOverlay.classList.remove('active');
    body.style.overflow = ''; // Возвращаем прокрутку
    
    // Очищаем форму
    newsTitle.value = '';
    newsSubtitle.value = '';
    newsText.value = '';
    newsImage.value = '';
}

// Открытие модального окна при клике на кнопку
addNewsBtn.addEventListener('click', openModal);

// Закрытие по крестику
closeModalBtn.addEventListener('click', closeModal);

// Закрытие по кнопке "Отмена"
cancelModalBtn.addEventListener('click', closeModal);

// Закрытие при клике на оверлей
modalOverlay.addEventListener('click', function(e) {
    if (e.target === modalOverlay) {
        closeModal();
    }
});

// Закрытие по ESC
document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape' && modalOverlay.classList.contains('active')) {
        closeModal();
    }
});

// ===== УПРАВЛЕНИЕ НОВОСТЯМИ =====
const newsContainer = document.getElementById('newsContainer');

// Загружаем новости из localStorage
let news = JSON.parse(localStorage.getItem('news')) || [];

// Функция для создания стандартного изображения
function getDefaultImage() {
    return 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="800" height="400" viewBox="0 0 800 400"%3E%3Crect width="800" height="400" fill="%232b7a9c"/%3E%3Ctext x="400" y="200" font-family="Arial" font-size="32" fill="white" text-anchor="middle" dy=".3em"%3EUnik University%3C/text%3E%3C/svg%3E';
}

// Функция для отображения новостей
function renderNews() {
    if (news.length === 0) {
        newsContainer.innerHTML = '<div class="empty_news">📰 Пока нет новостей. Нажмите "Добавить новость", чтобы создать первую запись!</div>';
        return;
    }
    
    // Сортируем новости: сначала новые
    const sortedNews = [...news].sort((a, b) => b.id - a.id);
    
    let html = '';
    sortedNews.forEach(item => {
        html += `
            <div class="card_news" data-id="${item.id}">
                <div class="card_img">
                    <img src="${item.image || getDefaultImage()}" class="img_news" alt="Новость">
                </div>
                <div class="card_head">
                    ${escapeHtml(item.title)}
                </div>
                <div class="card_subtitle">
                    ${escapeHtml(item.subtitle)}
                </div>
                <div class="card_txt">
                    ${escapeHtml(item.text).replace(/\n/g, '<br>')}
                </div>
                <a class="delete_btn" onclick="deleteNews(${item.id})"></a>
            </div>
        `;
    });
    
    newsContainer.innerHTML = html;
}

// Защита от XSS
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Функция удаления новости
window.deleteNews = function(id) {
    if (confirm('Вы уверены, что хотите удалить эту новость?')) {
        news = news.filter(item => item.id !== id);
        localStorage.setItem('news', JSON.stringify(news));
        renderNews();
    }
};

// Функция добавления новости
function addNews() {
    const title = newsTitle.value.trim();
    const subtitle = newsSubtitle.value.trim();
    const text = newsText.value.trim();
    const image = newsImage.value.trim();
    
    if (!title || !subtitle || !text) {
        alert('Пожалуйста, заполните заголовок, подзаголовок и текст новости');
        return;
    }
    
    const newNews = {
        id: Date.now(),
        title: title,
        subtitle: subtitle,
        text: text,
        image: image || null,
        date: new Date().toLocaleDateString('ru-RU')
    };
    
    news.unshift(newNews);
    localStorage.setItem('news', JSON.stringify(news));
    
    // Закрываем модальное окно
    closeModal();
    
    // Обновляем отображение
    renderNews();
}

// Обработчик отправки новости
submitNewsBtn.addEventListener('click', addNews);

// Инициализация демо-новостей при первом запуске
function initializeDemoNews() {
    if (news.length === 0) {
        const demoNews = [
            {
                id: 1,
                title: 'Начало приёмной комиссии 2026 года',
                subtitle: 'принято более 1500+ заявок на поступление',
                text: 'Прием на обучение по программам бакалавриата и программам специалитета проводится на основании результатов единого государственного экзамена, если иное не предусмотрено настоящим Федеральным законом.\n\nРезультаты единого государственного экзамена при приеме на обучение по программам бакалавриата и программам специалитета действительны четыре года, следующих за годом получения таких результатов.\n\nМинимальное количество баллов единого государственного экзамена по общеобразовательным предметам устанавливается образовательной организацией высшего образования.',
                image: null,
                date: '12.02.2026'
            },
            {
                id: 2,
                title: 'День открытых дверей',
                subtitle: 'приглашаем абитуриентов и их родителей',
                text: '25 февраля 2026 года в 15:00 состоится день открытых дверей. Вы сможете познакомиться с факультетами, преподавателями и задать все интересующие вопросы о поступлении.\n\nВ программе: презентация направлений подготовки, экскурсия по кампусу, мастер-классы от ведущих преподавателей.',
                image: null,
                date: '10.02.2026'
            }
        ];
        news = demoNews;
        localStorage.setItem('news', JSON.stringify(demoNews));
    }
}

// Инициализация при загрузке
initializeDemoNews();
renderNews();