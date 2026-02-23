// ===== ПОДКЛЮЧАЕМ ProfileManager =====
// Этот файл теперь только для управления новостями

// ===== УПРАВЛЕНИЕ МОДАЛЬНЫМ ОКНОМ НОВОСТЕЙ =====
const modalOverlay = document.getElementById('modalOverlay');
const addNewsBtn = document.getElementById('addNewsBtn');
const closeModalBtn = document.getElementById('closeModalBtn');
const cancelModalBtn = document.getElementById('cancelModalBtn');
const submitNewsBtn = document.getElementById('submitNewsBtn');

const newsTitle = document.getElementById('newsTitle');
const newsSubtitle = document.getElementById('newsSubtitle');
const newsText = document.getElementById('newsText');
const newsImage = document.getElementById('newsImage');

function openModal() {
    modalOverlay.classList.add('active');
    document.body.style.overflow = 'hidden';
}

function closeModal() {
    modalOverlay.classList.remove('active');
    document.body.style.overflow = '';
    
    newsTitle.value = '';
    newsSubtitle.value = '';
    newsText.value = '';
    newsImage.value = '';
}

if (addNewsBtn) {
    addNewsBtn.addEventListener('click', openModal);
}
if (closeModalBtn) {
    closeModalBtn.addEventListener('click', closeModal);
}
if (cancelModalBtn) {
    cancelModalBtn.addEventListener('click', closeModal);
}

if (modalOverlay) {
    modalOverlay.addEventListener('click', function(e) {
        if (e.target === modalOverlay) {
            closeModal();
        }
    });
}

document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape' && modalOverlay && modalOverlay.classList.contains('active')) {
        closeModal();
    }
});

// ===== УПРАВЛЕНИЕ НОВОСТЯМИ =====
const newsContainer = document.getElementById('newsContainer');
let news = JSON.parse(localStorage.getItem('news')) || [];

function getDefaultImage() {
    return 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="800" height="400" viewBox="0 0 800 400"%3E%3Crect width="800" height="400" fill="%232b7a9c"/%3E%3Ctext x="400" y="200" font-family="Arial" font-size="32" fill="white" text-anchor="middle" dy=".3em"%3EUnik University%3C/text%3E%3C/svg%3E';
}

function renderNews() {
    if (!newsContainer) return;
    
    if (news.length === 0) {
        newsContainer.innerHTML = '<div class="empty_news">📰 Пока нет новостей. Нажмите "Добавить новость", чтобы создать первую запись!</div>';
        return;
    }
    
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

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

window.deleteNews = function(id) {
    if (confirm('Вы уверены, что хотите удалить эту новость?')) {
        news = news.filter(item => item.id !== id);
        localStorage.setItem('news', JSON.stringify(news));
        renderNews();
        if (window.ProfileManager) {
            ProfileManager.showNotification('Новость удалена', 'info');
        }
    }
};

function addNews() {
    const title = newsTitle.value.trim();
    const subtitle = newsSubtitle.value.trim();
    const text = newsText.value.trim();
    const image = newsImage.value.trim();
    
    if (!title || !subtitle || !text) {
        if (window.ProfileManager) {
            ProfileManager.showNotification('Пожалуйста, заполните все поля', 'warning');
        }
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
    
    closeModal();
    renderNews();
    
    if (window.ProfileManager) {
        ProfileManager.showNotification('Новость опубликована!', 'success');
    }
}

if (submitNewsBtn) {
    submitNewsBtn.addEventListener('click', addNews);
}

function initializeDemoNews() {
    if (news.length === 0) {
        const demoNews = [
            {
                id: 1,
                title: 'Начало приёмной комиссии 2026 года',
                subtitle: 'принято более 1500+ заявок на поступление',
                text: 'Прием на обучение по программам бакалавриата и программам специалитета проводится на основании результатов единого государственного экзамена...',
                image: null,
                date: '12.02.2026'
            },
            {
                id: 2,
                title: 'День открытых дверей',
                subtitle: 'приглашаем абитуриентов и их родителей',
                text: '25 февраля 2026 года в 15:00 состоится день открытых дверей. Вы сможете познакомиться с факультетами...',
                image: null,
                date: '10.02.2026'
            }
        ];
        news = demoNews;
        localStorage.setItem('news', JSON.stringify(demoNews));
    }
}

// Инициализация новостей
initializeDemoNews();
renderNews();