// stats.js - только для аналитики, профиль управляется ProfileManager

document.addEventListener('DOMContentLoaded', function() {
    // ===== АНИМАЦИЯ ШКАЛ =====
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