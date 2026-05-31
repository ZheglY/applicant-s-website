const passwordInput = document.getElementById('password');

function togglePassword() {
    const toggleBtn = document.querySelector('.password-toggle');
    if (!passwordInput || !toggleBtn) return;
    if (passwordInput.type === 'password') {
        passwordInput.type = 'text';
        toggleBtn.textContent = '🔒';
    } else {
        passwordInput.type = 'password';
        toggleBtn.textContent = '👁️';
    }
}

window.togglePassword = togglePassword;

function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification-${type}`;
    notification.textContent = message;
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 15px 25px;
        border-radius: 10px;
        color: white;
        font-weight: 500;
        z-index: 10000;
        animation: slideIn 0.3s ease;
        box-shadow: 0 5px 15px rgba(0,0,0,0.2);
    `;

    const colors = {
        success: '#0f7b5c',
        error: '#bc4036',
        warning: '#e67e22',
        info: '#2b7a9c'
    };

    notification.style.backgroundColor = colors[type] || colors.info;
    document.body.appendChild(notification);

    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from { transform: translateX(100%); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
    }

    @keyframes slideOut {
        from { transform: translateX(0); opacity: 1; }
        to { transform: translateX(100%); opacity: 0; }
    }

    @keyframes shake {
        0%, 100% { transform: translateX(0); }
        10%, 30%, 50%, 70%, 90% { transform: translateX(-5px); }
        20%, 40%, 60%, 80% { transform: translateX(5px); }
    }
`;

document.head.appendChild(style);

const loginForm = document.getElementById('loginForm');
if (loginForm) {
    loginForm.addEventListener('submit', async function(e) {
        e.preventDefault();

        const btn = document.querySelector('.login-btn');
        if (btn) {
            btn.style.transform = 'scale(0.98)';
            setTimeout(() => {
                btn.style.transform = '';
            }, 200);
        }

        const username = document.getElementById('username')?.value || '';
        const password = document.getElementById('password')?.value || '';

        if (!username || !password) {
            showNotification('Введите логин и пароль', 'error');
            return;
        }

        try {
            const response = await fetch('/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password })
            });
            const result = await response.json();

            if (!response.ok) {
                const msg = result.detail || 'Неверный логин или пароль';
                showNotification(msg, 'error');
                loginForm.style.animation = 'shake 0.5s ease';
                setTimeout(() => {
                    loginForm.style.animation = '';
                }, 500);
                return;
            }

            showNotification('Вход выполнен успешно', 'success');
            setTimeout(() => {
                window.location.href = '/users/news';
            }, 600);
        } catch (err) {
            console.error(err);
            showNotification('Ошибка сети. Попробуйте позже.', 'error');
        }
    });
}
