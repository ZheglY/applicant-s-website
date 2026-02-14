// ===== УПРАВЛЕНИЕ ТЁМНОЙ ТЕМОЙ =====
document.addEventListener('DOMContentLoaded', function() {
    const themeToggle = document.getElementById('themeToggle');
    const themeText = document.getElementById('themeText');
    const body = document.body;
    
    // Проверяем сохранённую тему
    const savedTheme = localStorage.getItem('theme') || 'light';
    body.setAttribute('data-theme', savedTheme);
    updateThemeText(savedTheme);
    
    // Переключение темы
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
            if (theme === 'dark') {
                themeText.textContent = 'Светлая тема';
            } else {
                themeText.textContent = 'Тёмная тема';
            }
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
    
    // ===== АКТИВНАЯ СТРАНИЦА В ХЕДЕРЕ =====
    // Определяем текущую страницу по URL
    const currentPage = window.location.pathname.split('/').pop() || 'stats.html';
    
    // Удаляем класс active у всех кнопок
    document.querySelectorAll('.header_btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Добавляем класс active кнопке текущей страницы
    document.querySelectorAll('.header_btn').forEach(btn => {
        const href = btn.getAttribute('href');
        if (href === currentPage) {
            btn.classList.add('active');
        }
    });
    
    // ===== АНИМАЦИЯ ШКАЛ =====
    // Запускаем анимацию заполнения шкал при загрузке
    setTimeout(() => {
        document.querySelectorAll('.scale-fill').forEach(fill => {
            const width = fill.style.width;
            fill.style.width = '0%';
            setTimeout(() => {
                fill.style.width = width;
                fill.style.transition = 'width 1s ease-in-out';
            }, 100);
        });
    }, 200);
    
    // ===== ИНТЕРАКТИВНЫЕ ЭЛЕМЕНТЫ =====
    // Подсветка карточек при наведении
    document.querySelectorAll('.card, .stats_card').forEach(card => {
        card.addEventListener('mouseenter', function() {
            this.style.transition = 'all 0.3s ease';
        });
    });
    
    // ===== ФОРМАТИРОВАНИЕ ЧИСЕЛ =====
    function formatNumber(num) {
        return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
    }
    
    // Форматируем все числа на странице
    document.querySelectorAll('.big-number, .light-number, .stats_value, .hundred_number, .label-value').forEach(el => {
        const text = el.textContent;
        const num = parseFloat(text.replace(/[^0-9.-]/g, ''));
        if (!isNaN(num) && text.indexOf('.') === -1) {
            el.textContent = formatNumber(num);
        }
    });
});