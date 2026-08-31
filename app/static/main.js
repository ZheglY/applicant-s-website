const modalOverlay = document.getElementById('modalOverlay');
const addNewsBtn = document.getElementById('addNewsBtn');
const closeModalBtn = document.getElementById('closeModalBtn');
const cancelModalBtn = document.getElementById('cancelModalBtn');
const submitNewsBtn = document.getElementById('submitNewsBtn');

const newsTitle = document.getElementById('newsTitle');
const newsSubtitle = document.getElementById('newsSubtitle');
const newsText = document.getElementById('newsText');
const newsImage = document.getElementById('newsImage');
const newsContainer = document.getElementById('newsContainer');

const canManageNews = !!addNewsBtn;
let news = [];

function openModal() {
    if (!modalOverlay) return;
    modalOverlay.classList.add('active');
    document.body.style.overflow = 'hidden';
}

function closeModal() {
    if (!modalOverlay) return;
    modalOverlay.classList.remove('active');
    document.body.style.overflow = '';
    if (newsTitle) newsTitle.value = '';
    if (newsSubtitle) newsSubtitle.value = '';
    if (newsText) newsText.value = '';
    if (newsImage) newsImage.value = '';
}

if (addNewsBtn) addNewsBtn.addEventListener('click', openModal);
if (closeModalBtn) closeModalBtn.addEventListener('click', closeModal);
if (cancelModalBtn) cancelModalBtn.addEventListener('click', closeModal);

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

function getDefaultImage() {
    const svg = '<svg xmlns="http://www.w3.org/2000/svg" width="800" height="400" viewBox="0 0 800 400">' +
        '<defs><linearGradient id="bg" x1="0" x2="1"><stop stop-color="#00466f"/><stop offset="1" stop-color="#2b7a9c"/></linearGradient></defs>' +
        '<rect width="800" height="400" fill="url(#bg)"/>' +
        '<circle cx="675" cy="70" r="150" fill="rgba(255,255,255,.08)"/>' +
        '<text x="400" y="185" font-family="Arial" font-size="54" font-weight="700" fill="white" text-anchor="middle">Unik</text>' +
        '<text x="400" y="245" font-family="Arial" font-size="24" fill="#d7eef8" text-anchor="middle">Приёмная кампания 2026</text>' +
        '</svg>';
    return 'data:image/svg+xml;charset=UTF-8,' + encodeURIComponent(svg);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function renderNews() {
    if (!newsContainer) return;

    if (news.length === 0) {
        newsContainer.innerHTML = '<div class="empty_news">📰 Пока нет новостей. Здесь появятся публикации приёмной комиссии.</div>';
        return;
    }

    let html = '';
    news.forEach(item => {
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
                ${canManageNews ? '<a class="delete_btn" onclick="deleteNews(' + item.id + ')"></a>' : ''}
            </div>
        `;
    });

    newsContainer.innerHTML = html;
}

async function fetchNews() {
    try {
        const response = await fetch('/users/news/data');
        const result = await response.json();
        news = result.items || [];
        renderNews();
    } catch (err) {
        console.error(err);
        if (newsContainer) {
            newsContainer.innerHTML = '<div class="empty_news">Не удалось загрузить новости.</div>';
        }
    }
}

async function addNews() {
    if (!canManageNews) return;

    const title = newsTitle.value.trim();
    const subtitle = newsSubtitle.value.trim();
    const text = newsText.value.trim();
    const image = newsImage.value.trim();

    if (!title || !subtitle || !text) {
        alert('Пожалуйста, заполните все поля');
        return;
    }

    try {
        const response = await fetch('/users/news', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ title, subtitle, text, image })
        });
        const result = await response.json();
        if (!response.ok) {
            alert(result.detail || 'Ошибка публикации');
            return;
        }
        news.unshift(result);
        closeModal();
        renderNews();
    } catch (err) {
        console.error(err);
        alert('Ошибка сети. Попробуйте позже.');
    }
}

if (submitNewsBtn) submitNewsBtn.addEventListener('click', addNews);

window.deleteNews = async function(id) {
    if (!canManageNews) return;
    if (!confirm('Вы уверены, что хотите удалить эту новость?')) return;

    try {
        const response = await fetch(`/users/news/${id}`, { method: 'DELETE' });
        if (!response.ok) {
            const result = await response.json();
            alert(result.detail || 'Ошибка удаления');
            return;
        }
        news = news.filter(item => item.id !== id);
        renderNews();
    } catch (err) {
        console.error(err);
        alert('Ошибка сети. Попробуйте позже.');
    }
};

fetchNews();
