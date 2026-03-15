function nextStep(step) {
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

function validateStep2() {
    const school = document.getElementById('school').value;
    if (!school) {
        alert('Пожалуйста, укажите образовательную организацию');
        return false;
    }
    return true;
}

function updatePriorities() {
    const selects = document.querySelectorAll('.priority-select');
    const selectedValues = [];

    selects.forEach(select => {
        if (select.value) {
            selectedValues.push(select.value);
        }
    });

    const hasDuplicates = new Set(selectedValues).size !== selectedValues.length;

    if (hasDuplicates) {
        document.getElementById('priorityHint').textContent = '⚠️ Направления не должны повторяться';
        document.getElementById('priorityHint').style.color = '#bc4036';
    } else {
        document.getElementById('priorityHint').textContent = selectedValues.length === 0
            ? 'Выберите хотя бы одно направление'
            : `✓ Выбрано: ${selectedValues.join(' → ')}`;
        document.getElementById('priorityHint').style.color = '#0f7b5c';
    }

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

const form = document.getElementById('registerForm');
if (form) {
    form.addEventListener('submit', async function(e) {
        e.preventDefault();

        if (!document.getElementById('agreement').checked) {
            alert('Необходимо подтвердить согласие на обработку персональных данных');
            return;
        }

        const btn = document.querySelector('.btn-submit');
        if (btn) {
            btn.style.transform = 'scale(0.98)';
            setTimeout(() => {
                btn.style.transform = '';
            }, 200);
        }

        const formData = {
            fullname: document.getElementById('fullname').value.trim(),
            password: document.getElementById('password').value,
            password_confirm: document.getElementById('password_confirm').value,
            birthdate: document.getElementById('birthdate').value,
            phone: document.getElementById('phone').value,
            email: document.getElementById('email').value.trim(),
            telegram: document.getElementById('telegram')?.value?.trim() || null,
            school: document.getElementById('school').value.trim(),
            achievements: document.getElementById('achievements')?.value?.trim() || null,
            priorities: [],
            agreement: document.getElementById('agreement').checked
        };

        const egeScores = {};
        document.querySelectorAll('.ege-item').forEach(item => {
            const subject = item.querySelector('.ege-subject').textContent;
            const score = item.querySelector('.ege-score').value;
            if (score) {
                egeScores[subject] = score;
            }
        });
        formData.ege_scores = egeScores;

        document.querySelectorAll('.priority-select').forEach(select => {
            if (select.value) {
                formData.priorities.push(select.value);
            }
        });

        if (formData.priorities.length === 0) {
            alert('Выберите хотя бы одно направление');
            return;
        }

        if (formData.password !== formData.password_confirm) {
            alert('Пароли не совпадают');
            return;
        }

        if (formData.password.length < 6) {
            alert('Пароль должен быть не менее 6 символов');
            return;
        }

        try {
            const response = await fetch('/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(formData)
            });
            const result = await response.json();

            if (!response.ok) {
                const msg = result.detail || (Array.isArray(result.detail)
                    ? result.detail.map(e => e.msg).join(', ')
                    : 'Ошибка при регистрации');
                alert(msg);
                return;
            }

            alert('Заявление успешно отправлено!');
            window.location.href = '/auth/enter';
        } catch (err) {
            console.error(err);
            alert('Ошибка сети. Попробуйте позже.');
        }
    });
} else {
    console.error('Форма registerForm не найдена.');
}

function init() {
    updatePriorities();

    setTimeout(() => {
        const container = document.querySelector('.register-form-container');
        if (container) {
            container.style.opacity = '1';
            container.style.transform = 'translateY(0)';
        }
    }, 100);

    const phoneInput = document.getElementById('phone');
    if (phoneInput) phoneInput.addEventListener('input', function(e) {
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
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}