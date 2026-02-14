 // Переключение между шагами формы
        function nextStep(step) {
            // Валидация текущего шага
            if (step === 2) {
                if (!validateStep1()) return;
            }
            if (step === 3) {
                if (!validateStep2()) return;
            }
            
            document.querySelectorAll('.form-section').forEach(section => {
                section.classList.remove('active');
            });
            document.getElementById(`step${step}`).classList.add('active');
            
            // Обновление прогресс-бара
            document.querySelectorAll('.step').forEach((stepEl, index) => {
                if (index < step) {
                    stepEl.classList.add('active');
                } else {
                    stepEl.classList.remove('active');
                }
            });
        }
        
        function prevStep(step) {
            document.querySelectorAll('.form-section').forEach(section => {
                section.classList.remove('active');
            });
            document.getElementById(`step${step}`).classList.add('active');
            
            document.querySelectorAll('.step').forEach((stepEl, index) => {
                if (index < step) {
                    stepEl.classList.add('active');
                } else {
                    stepEl.classList.remove('active');
                }
            });
        }
        
        // Валидация шага 1
        function validateStep1() {
            const fullname = document.getElementById('fullname').value;
            const birthdate = document.getElementById('birthdate').value;
            const phone = document.getElementById('phone').value;
            const email = document.getElementById('email').value;
            
            if (!fullname || !birthdate || !phone || !email) {
                alert('Пожалуйста, заполните все обязательные поля');
                return false;
            }
            return true;
        }
        
        // Валидация шага 2
        function validateStep2() {
            const school = document.getElementById('school').value;
            if (!school) {
                alert('Пожалуйста, укажите образовательную организацию');
                return false;
            }
            return true;
        }
        
        // Обновление доступных опций в селектах
        function updatePriorities() {
            const selects = document.querySelectorAll('.priority-select');
            const selectedValues = [];
            
            selects.forEach(select => {
                if (select.value) {
                    selectedValues.push(select.value);
                }
            });
            
            // Проверка на дубликаты
            const hasDuplicates = new Set(selectedValues).size !== selectedValues.length;
            
            if (hasDuplicates) {
                document.getElementById('priorityHint').textContent = '⚠️ Направления не должны повторяться';
                document.getElementById('priorityHint').style.color = '#bc4036';
            } else {
                document.getElementById('priorityHint').textContent = selectedValues.length === 0 ? 'Выберите хотя бы одно направление' : `✓ Выбрано: ${selectedValues.join(' → ')}`;
                document.getElementById('priorityHint').style.color = '#0f7b5c';
            }
            
            // Блокировка уже выбранных опций в других селектах
            selects.forEach(select => {
                Array.from(select.options).forEach(option => {
                    if (option.value && selectedValues.includes(option.value) && option.value !== select.value) {
                        option.disabled = true;
                    } else {
                        option.disabled = false;
                    }
                });
            });
        }
        
        // Отправка формы
        document.getElementById('registerForm').addEventListener('submit', function(e) {
            e.preventDefault();
            
            if (!document.getElementById('agreement').checked) {
                alert('Необходимо подтвердить согласие на обработку персональных данных');
                return;
            }
            
            // Эффект нажатия
            const btn = document.querySelector('.btn-submit');
            btn.style.transform = 'scale(0.98)';
            setTimeout(() => {
                btn.style.transform = '';
            }, 200);
            
            // Сбор данных формы
            const formData = {
                fullname: document.getElementById('fullname').value,
                birthdate: document.getElementById('birthdate').value,
                phone: document.getElementById('phone').value,
                email: document.getElementById('email').value,
                telegram: document.getElementById('telegram').value,
                school: document.getElementById('school').value,
                achievements: document.getElementById('achievements').value,
                priorities: []
            };
            
            // Собираем баллы ЕГЭ
            const egeScores = {};
            document.querySelectorAll('.ege-item').forEach(item => {
                const subject = item.querySelector('.ege-subject').textContent;
                const score = item.querySelector('.ege-score').value;
                if (score) {
                    egeScores[subject] = score;
                }
            });
            formData.egeScores = egeScores;
            
            // Собираем выбранные направления
            document.querySelectorAll('.priority-select').forEach(select => {
                if (select.value) {
                    formData.priorities.push(select.value);
                }
            });
            
            if (formData.priorities.length === 0) {
                alert('Выберите хотя бы одно направление');
                return;
            }
            
            // Здесь будет отправка на сервер
            console.log('Данные формы:', formData);
            
            // Показываем успешную отправку
            alert('Заявление успешно отправлено!\nПриёмная комиссия свяжется с вами в ближайшее время.');
            
            // Перенаправление на главную
            setTimeout(() => {
                window.location.href = 'main.html';
            }, 1500);
        });
        
        // Инициализация подсказки приоритетов
        document.addEventListener('DOMContentLoaded', function() {
            updatePriorities();
            
            // Анимация появления формы
            setTimeout(() => {
                document.querySelector('.register-form-container').style.opacity = '1';
                document.querySelector('.register-form-container').style.transform = 'translateY(0)';
            }, 100);
            
            // Автоформатирование телефона
            const phoneInput = document.getElementById('phone');
            phoneInput.addEventListener('input', function(e) {
                let value = e.target.value.replace(/\D/g, '');
                if (value.length > 0) {
                    if (value.startsWith('7')) value = value.substring(1);
                    if (value.length > 0) {
                        let formatted = '+7';
                        if (value.length > 0) formatted += ' (' + value.substring(0, 3);
                        if (value.length >= 4) formatted += ') ' + value.substring(3, 6);
                        if (value.length >= 7) formatted += '-' + value.substring(6, 8);
                        if (value.length >= 9) formatted += '-' + value.substring(8, 10);
                        e.target.value = formatted;
                    }
                }
            });
        });