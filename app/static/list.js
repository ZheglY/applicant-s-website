window.toggleDirection = function(header) {
    const directionItem = header.closest('.direction_item');
    const content = directionItem.querySelector('.direction_content');
    const icon = header.querySelector('.direction_icon');

    content.classList.toggle('open');

    if (content.classList.contains('open')) {
        icon.textContent = '▼';
        directionItem.classList.add('active');
    } else {
        icon.textContent = '▲';
        directionItem.classList.remove('active');
    }
};

document.addEventListener('DOMContentLoaded', function() {
    const firstDirection = document.querySelector('.direction_item');
    if (firstDirection) {
        const header = firstDirection.querySelector('.direction_header');
        const content = firstDirection.querySelector('.direction_content');
        const icon = firstDirection.querySelector('.direction_icon');

        content.classList.add('open');
        icon.textContent = '▼';
        firstDirection.classList.add('active');
    }
});

function viewProfile(id) {
    window.location.href = `/users/applicants/${id}`;
}

window.viewProfile = viewProfile;