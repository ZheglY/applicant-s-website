function togglePassword() {
            const passwordInput = document.getElementById('password');
            const toggleBtn = document.querySelector('.password-toggle');
            
            if (passwordInput.type === 'password') {
                passwordInput.type = 'text';
                toggleBtn.textContent = '🔓';
            } else {
                passwordInput.type = 'password';
                toggleBtn.textContent = '👁️';
            }
        }
        
        document.getElementById('loginForm').addEventListener('submit', function(e) {
            e.preventDefault();
            
            // Эффект нажатия на кнопку
            const btn = document.querySelector('.login-btn');
            btn.style.transform = 'scale(0.98)';
            setTimeout(() => {
                btn.style.transform = '';
            }, 200);
            
            // Здесь будет ваша логика входа
            window.location.href = 'main.html';
        });
        
        // Анимация появления формы
        document.addEventListener('DOMContentLoaded', function() {
            setTimeout(() => {
                document.querySelector('.login-form-container').style.opacity = '1';
                document.querySelector('.login-form-container').style.transform = 'translateY(0)';
            }, 100);
        });