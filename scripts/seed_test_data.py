# -*- coding: utf-8 -*-
from __future__ import annotations

import asyncio
from datetime import date, timedelta
from pathlib import Path
from random import Random
import re
import sys

from sqlalchemy import select, text

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))

from app.db.engine import new_session, seed_directions, setup_database
from app.db.models import Applicant, ApplicantPriority, FacultyDirection, News
from app.services.auth_service import ensure_staff_users
from app.utils.hash_password import hash_password


RNG = Random(20260530)
STUDENT_COUNT = 300
STUDENT_PASSWORD = "student123"

MALE_NAMES = [
    ("Ярослав", "Олегович"),
    ("Александр", "Сергеевич"),
    ("Дмитрий", "Андреевич"),
    ("Максим", "Ильич"),
    ("Михаил", "Алексеевич"),
    ("Иван", "Романович"),
    ("Артём", "Владимирович"),
    ("Никита", "Денисович"),
    ("Кирилл", "Евгеньевич"),
    ("Матвей", "Павлович"),
    ("Егор", "Михайлович"),
    ("Илья", "Антонович"),
    ("Даниил", "Игоревич"),
    ("Андрей", "Викторович"),
    ("Роман", "Константинович"),
    ("Лев", "Николаевич"),
    ("Георгий", "Петрович"),
    ("Тимофей", "Васильевич"),
    ("Владислав", "Аркадьевич"),
    ("Степан", "Григорьевич"),
]

FEMALE_NAMES = [
    ("Анна", "Сергеевна"),
    ("Мария", "Алексеевна"),
    ("София", "Андреевна"),
    ("Екатерина", "Ильинична"),
    ("Алиса", "Романовна"),
    ("Полина", "Дмитриевна"),
    ("Виктория", "Олеговна"),
    ("Елизавета", "Максимовна"),
    ("Дарья", "Павловна"),
    ("Ксения", "Владимировна"),
    ("Анастасия", "Игоревна"),
    ("Валерия", "Михайловна"),
    ("Арина", "Денисовна"),
    ("Вероника", "Евгеньевна"),
    ("Ульяна", "Антоновна"),
    ("Александра", "Николаевна"),
    ("Варвара", "Константиновна"),
    ("Диана", "Викторовна"),
    ("Карина", "Артёмовна"),
    ("Милана", "Григорьевна"),
]

SURNAMES = [
    ("Смирнов", "Смирнова"),
    ("Иванов", "Иванова"),
    ("Кузнецов", "Кузнецова"),
    ("Соколов", "Соколова"),
    ("Попов", "Попова"),
    ("Лебедев", "Лебедева"),
    ("Козлов", "Козлова"),
    ("Новиков", "Новикова"),
    ("Морозов", "Морозова"),
    ("Волков", "Волкова"),
    ("Соловьёв", "Соловьёва"),
    ("Васильев", "Васильева"),
    ("Зайцев", "Зайцева"),
    ("Павлов", "Павлова"),
    ("Семёнов", "Семёнова"),
    ("Голубев", "Голубева"),
    ("Виноградов", "Виноградова"),
    ("Богданов", "Богданова"),
    ("Воробьёв", "Воробьёва"),
    ("Фёдоров", "Фёдорова"),
    ("Михайлов", "Михайлова"),
    ("Беляев", "Беляева"),
    ("Тарасов", "Тарасова"),
    ("Орлов", "Орлова"),
    ("Макаров", "Макарова"),
    ("Андреев", "Андреева"),
    ("Ковалёв", "Ковалёва"),
    ("Никитин", "Никитина"),
    ("Громов", "Громова"),
    ("Комаров", "Комарова"),
]

LOCATIONS = [
    ("Москва", "Москва", ["ГБОУ школа № 1234", "Лицей НИУ ВШЭ", "Школа № 1535", "Гимназия № 1518"]),
    ("Московская область", "Подольск", ["Лицей № 26", "Гимназия № 4", "Школа № 32"]),
    ("Московская область", "Мытищи", ["Лицей № 15", "Гимназия № 16", "Школа № 27"]),
    ("Санкт-Петербург", "Санкт-Петербург", ["Академическая гимназия № 56", "Лицей № 239", "Школа № 619"]),
    ("Республика Татарстан", "Казань", ["Лицей № 131", "Гимназия № 19", "Школа № 39"]),
    ("Нижегородская область", "Нижний Новгород", ["Лицей № 40", "Школа № 33", "Гимназия № 13"]),
    ("Свердловская область", "Екатеринбург", ["СУНЦ УрФУ", "Гимназия № 9", "Лицей № 110"]),
    ("Новосибирская область", "Новосибирск", ["СУНЦ НГУ", "Гимназия № 1", "Лицей № 130"]),
    ("Самарская область", "Самара", ["Лицей авиационного профиля № 135", "Гимназия № 1", "Школа № 41"]),
    ("Краснодарский край", "Краснодар", ["Гимназия № 23", "Лицей № 48", "Школа № 71"]),
    ("Ростовская область", "Ростов-на-Дону", ["Лицей № 11", "Гимназия № 36", "Школа № 80"]),
    ("Пермский край", "Пермь", ["Лицей № 10", "Гимназия № 17", "Школа № 9"]),
    ("Республика Башкортостан", "Уфа", ["Лицей № 153", "Гимназия № 39", "Школа № 45"]),
    ("Воронежская область", "Воронеж", ["Лицей № 1", "Гимназия им. Басова", "Школа № 102"]),
    ("Тюменская область", "Тюмень", ["Физико-математическая школа", "Гимназия № 16", "Лицей № 34"]),
]

EMAIL_DOMAINS = ["mail.ru", "yandex.ru", "gmail.com", "inbox.ru"]

ACHIEVEMENTS = [
    {"text": "Золотая медаль", "points": 5},
    {"text": "Серебряная медаль", "points": 3},
    {"text": "Призёр региональной олимпиады", "points": 4},
    {"text": "Победитель муниципальной олимпиады", "points": 2},
    {"text": "Значок ГТО", "points": 2},
    {"text": "Волонтёрская деятельность", "points": 1},
    {"text": "Спортивный разряд", "points": 2},
]

DIRECTION_GROUPS = [
    ["ПИ", "ИИ", "ПМ", "ИБ", "РОБ"],
    ["ПИ", "ВЕБ", "СА"],
    ["БА", "ЭК"],
    ["СА", "ИБ", "ПИ"],
    ["ДИЗ"],
]

CYRILLIC_TO_LATIN = str.maketrans(
    {
        "а": "a", "б": "b", "в": "v", "г": "g", "д": "d", "е": "e",
        "ё": "e", "ж": "zh", "з": "z", "и": "i", "й": "y", "к": "k",
        "л": "l", "м": "m", "н": "n", "о": "o", "п": "p", "р": "r",
        "с": "s", "т": "t", "у": "u", "ф": "f", "х": "h", "ц": "c",
        "ч": "ch", "ш": "sh", "щ": "sch", "ъ": "", "ы": "y", "ь": "",
        "э": "e", "ю": "yu", "я": "ya",
    }
)


def _slug(value: str) -> str:
    transliterated = value.lower().translate(CYRILLIC_TO_LATIN)
    return re.sub(r"[^a-z0-9]+", "", transliterated)


def _random_birth_date() -> date:
    start = date(2007, 1, 1)
    return start + timedelta(days=RNG.randint(0, 950))


def _format_phone(index: int) -> str:
    number = 9_000_000_000 + ((index * 7919) % 100_000_000)
    digits = str(number)
    return f"+7 ({digits[:3]}) {digits[3:6]}-{digits[6:8]}-{digits[8:]}"


def _pick_achievements() -> list[dict] | None:
    if RNG.random() < 0.58:
        return None
    items = RNG.sample(ACHIEVEMENTS, k=1 if RNG.random() < 0.78 else 2)
    total = 0
    result = []
    for item in items:
        if total + item["points"] <= 10:
            result.append(dict(item))
            total += item["points"]
    return result or None


def _pick_directions(directions_by_code: dict[str, FacultyDirection]) -> list[FacultyDirection]:
    group = RNG.choices(DIRECTION_GROUPS, weights=[42, 20, 20, 12, 6], k=1)[0]
    count = RNG.randint(1, min(3, len(group)))
    codes = RNG.sample(group, k=count)
    return [directions_by_code[code] for code in codes]


def _make_scores(directions: list[FacultyDirection]) -> dict[str, int]:
    subjects = sorted({subject for direction in directions for subject in direction.subjects})
    return {
        subject: round(RNG.triangular(52, 100, 78))
        for subject in subjects
    }


def _make_statuses(count: int) -> list[str]:
    statuses = []
    accepted_added = False
    for _ in range(count):
        status = RNG.choices(["pending", "accepted", "rejected"], weights=[68, 18, 14], k=1)[0]
        if status == "accepted" and accepted_added:
            status = "pending"
        if status == "accepted":
            accepted_added = True
        statuses.append(status)
    return statuses


def _competitive_score(applicant_scores: dict[str, int], direction: FacultyDirection) -> int:
    return sum(applicant_scores.get(subject, 0) for subject in direction.subjects)


async def _reset_database() -> None:
    await setup_database()
    async with new_session() as session:
        await session.execute(
            text(
                "TRUNCATE TABLE news, applicant_priorities, applicants, "
                "faculty_directions RESTART IDENTITY CASCADE"
            )
        )
        await session.commit()

    await seed_directions()
    async with new_session() as session:
        await ensure_staff_users(session)


async def _seed_students() -> None:
    student_password_hash = hash_password(STUDENT_PASSWORD)

    async with new_session() as session:
        directions = list(
            (
                await session.execute(
                    select(FacultyDirection).order_by(FacultyDirection.id)
                )
            )
            .scalars()
            .all()
        )
        directions_by_code = {direction.direction_code: direction for direction in directions}

        for index in range(1, STUDENT_COUNT + 1):
            is_male = RNG.random() < 0.48
            first_name, middle_name = RNG.choice(MALE_NAMES if is_male else FEMALE_NAMES)
            surname_pair = RNG.choice(SURNAMES)
            last_name = surname_pair[0] if is_male else surname_pair[1]
            if index == 1:
                is_male = True
                first_name, middle_name = "Ярослав", "Олегович"
                last_name = "Смирнов"
            region, city, schools = RNG.choice(LOCATIONS)
            school = f"{RNG.choice(schools)}, г. {city}"
            selected_directions = _pick_directions(directions_by_code)
            ege_scores = _make_scores(selected_directions)
            achievements = _pick_achievements()
            bonus_points = sum(item["points"] for item in achievements or [])
            total_score = max(
                _competitive_score(ege_scores, direction)
                for direction in selected_directions
            ) + bonus_points
            email = (
                f"{_slug(first_name)}.{_slug(last_name)}{index:03d}"
                f"@{RNG.choice(EMAIL_DOMAINS)}"
            )

            applicant = Applicant(
                last_name=last_name,
                first_name=first_name,
                middle_name=middle_name,
                email=email,
                login=email,
                password_hash=student_password_hash,
                phone=_format_phone(index),
                telegram=f"@{_slug(first_name)}_{_slug(last_name)}_{index:03d}",
                birth_date=_random_birth_date(),
                total_score=total_score,
                role="student",
                sex=is_male,
                achievements=achievements,
                school=school,
                region=region,
                ege_scores=ege_scores,
            )
            session.add(applicant)
            await session.flush()

            statuses = _make_statuses(len(selected_directions))
            for priority, (direction, status) in enumerate(
                zip(selected_directions, statuses),
                start=1,
            ):
                session.add(
                    ApplicantPriority(
                        applicant_id=applicant.id,
                        direction_id=direction.id,
                        priority=priority,
                        status=status,
                    )
                )

        admin = await session.scalar(
            select(Applicant).where(Applicant.role == "admissions")
        )
        session.add_all(
            [
                News(
                    title="Старт приёмной кампании 2026",
                    subtitle="Открыт приём заявлений на программы бакалавриата",
                    text="Подать заявление можно онлайн. Проверьте контактные данные и расставьте направления в порядке приоритета.",
                    author_id=admin.id if admin else None,
                ),
                News(
                    title="Опубликован график консультаций",
                    subtitle="Приёмная комиссия отвечает на вопросы абитуриентов",
                    text="Консультации проходят по будням. Подробности можно уточнить у сотрудников приёмной комиссии.",
                    author_id=admin.id if admin else None,
                ),
                News(
                    title="Проверьте выбранные направления",
                    subtitle="До завершения кампании можно уточнить приоритеты",
                    text="Убедитесь, что выбранные направления расположены в желаемом порядке и баллы ЕГЭ указаны верно.",
                    author_id=admin.id if admin else None,
                ),
            ]
        )
        await session.commit()


async def _print_summary() -> None:
    async with new_session() as session:
        students = await session.scalar(
            select(text("count(*)")).select_from(Applicant).where(Applicant.role == "student")
        )
        staff = await session.scalar(
            select(text("count(*)")).select_from(Applicant).where(Applicant.role != "student")
        )
        priorities = await session.scalar(
            select(text("count(*)")).select_from(ApplicantPriority)
        )
        news = await session.scalar(select(text("count(*)")).select_from(News))

    print(f"OK: students={students}, staff={staff}, priorities={priorities}, news={news}")
    print(f"Student demo password: {STUDENT_PASSWORD}")


async def main() -> None:
    await _reset_database()
    await _seed_students()
    await _print_summary()


if __name__ == "__main__":
    asyncio.run(main())
