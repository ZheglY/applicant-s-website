
        // Показать/скрыть пароль
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

        // Функция показа уведомлений
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

        // Добавляем стили для анимаций
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

        // Обработка входа
        document.getElementById('loginForm').addEventListener('submit', function(e) {
            e.preventDefault();
            
            // Эффект нажатия на кнопку
            const btn = document.querySelector('.login-btn');
            btn.style.transform = 'scale(0.98)';
            setTimeout(() => {
                btn.style.transform = '';
            }, 200);
            
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            const rememberMe = document.getElementById('rememberMe').checked;
            
            // Валидация
            if (!username || !password) {
                showNotification('Введите логин и пароль', 'error');
                return;
            }
            
            // Демо-вход
            if ((username === 'препод' || username === 'prepod@unik.edu') && password === '123456') {
                showNotification('Вход выполнен успешно!', 'success');
                
                // Создаём объект пользователя
                const user = {
                    name: 'Препод Преподов',
                    email: username.includes('@') ? username : 'prepod@unik.edu',
                    role: 'Преподаватель',
                    loginTime: new Date().toISOString()
                };
                
                // Сохраняем в localStorage или sessionStorage
                if (rememberMe) {
                    localStorage.setItem('user', JSON.stringify(user));
                } else {
                    sessionStorage.setItem('user', JSON.stringify(user));
                }
                
                // Сохраняем профиль
                const profile = {
                    name: 'Препод Преподов',
                    avatar: 'data:image/svg+xml,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' width=\'40\' height=\'40\' viewBox=\'0 0 40 40\'%3E%3Ccircle cx=\'20\' cy=\'20\' r=\'20\' fill=\'%23004b7c\'/%3E%3Ccircle cx=\'20\' cy=\'13\' r=\'6\' fill=\'white\'/%3E%3Cpath d=\'M6 30 C6 23, 14 22, 20 22 C26 22, 34 23, 34 30 L34 34 L6 34 Z\' fill=\'white\'/%3E%3C/svg%3E',
                    phone: '+7 (999) 123-45-67'
                };
                localStorage.setItem('userProfile', JSON.stringify(profile));
                
                // Перенаправляем на главную
                setTimeout(() => {
                    window.location.href = 'main.html';
                }, 1000);
            } 
            else if (username === 'admin' && password === 'admin') {
                showNotification('Добро пожаловать, администратор!', 'success');
                
                const user = {
                    name: 'Администратор',
                    email: 'admin@unik.edu',
                    role: 'Администратор',
                    loginTime: new Date().toISOString()
                };
                
                if (rememberMe) {
                    localStorage.setItem('user', JSON.stringify(user));
                } else {
                    sessionStorage.setItem('user', JSON.stringify(user));
                }
                
                const profile = {
                    name: 'Администратор',
                    avatar: 'data:image/svg+xml,%3Csvg xmlns=\'http://www.w3.org/2000/svg\' width=\'40\' height=\'40\' viewBox=\'0 0 40 40\'%3E%3Ccircle cx=\'20\' cy=\'20\' r=\'20\' fill=\'%23004b7c\'/%3E%3Ccircle cx=\'20\' cy=\'13\' r=\'6\' fill=\'white\'/%3E%3Cpath d=\'M6 30 C6 23, 14 22, 20 22 C26 22, 34 23, 34 30 L34 34 L6 34 Z\' fill=\'white\'/%3E%3C/svg%3E',
                    phone: '+7 (999) 888-88-88'
                };
                localStorage.setItem('userProfile', JSON.stringify(profile));
                
                setTimeout(() => {
                    window.location.href = 'main.html';
                }, 1000);
            }
            else {
                showNotification('Неверный логин или пароль', 'error');
                
                // Тряска формы при ошибке
                const form = document.getElementById('loginForm');
                form.style.animation = 'shake 0.5s ease';
                setTimeout(() => {
                    form.style.animation = '';
                }, 500);
            }
        });

        // Анимация появления формы
        document.addEventListener('DOMContentLoaded', function() {
            setTimeout(() => {
                document.querySelector('.login-form-container').style.opacity = '1';
                document.querySelector('.login-form-container').style.transform = 'translateY(0)';
            }, 100);
            
            // Проверяем сохранённую сессию
            const savedUser = localStorage.getItem('user');
            if (savedUser) {
                window.location.href = 'main.html';
            }
        });

        // Обработка "Забыли пароль?"
        document.querySelector('.forgot-link').addEventListener('click', function(e) {
            e.preventDefault();
            showNotification('Обратитесь к администратору для восстановления пароля', 'info');
        });