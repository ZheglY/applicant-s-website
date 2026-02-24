// Управление темой и выпадающим меню в шапке (лента, список, аналитика)
document.addEventListener('DOMContentLoaded', function() {
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

    const logout = document.getElementById('logout');
    if (logout) {
        logout.addEventListener('click', function(e) {
            e.preventDefault();
            window.location.href = '/auth/enter';
        });
    }
});
