from fastapi import APIRouter, Depends, Request
from fastapi.responses import HTMLResponse
from fastapi import (
    Form, 
    UploadFile,
    HTTPException,
    )

from schemas.user import ApplicantSchema
from core.config import security
from core.roles import require_analyst, require_admissions


router = APIRouter(
    prefix="/users",
    tags=["users"],
    # dependencies=[Depends(get_current_user)]  # Все эндпоинты требуют авторизации
)

@router.get("/")
async def news():
    return {"message": "Тут заглушка для отображения новостной ленты"}


@router.get('/applicants/{student_id}', response_class=HTMLResponse)
async def get_applicant_account(request: Request, student_id: int):
    """
    Отображение личного кабинета абитуриента по ID пользователя
    Возвращает HTML-страницу
    """
    pass
#     try:
#         applicant_data = get_applicant_by_id(student_id)
        
#         if applicant_data:
#             return templates.TemplateResponse(
#                 "applicant_cabinet.html",  # имя шаблона
#                 {
#                     "request": request,  # обязательно!
#                     "student_id": student_id,
#                     "applicant": applicant_data,
#                     "title": f"Личный кабинет абитуриента {student_id}"
#                     }
#                 )
#         raise HTTPException(status_code=404, detail="Абитуриент на найден")

#     except Exception as e:
#         logger.error(f"Ошибка при получении абитуриента {student_id}: {e}")
#         return templates.TemplateResponse(
#             "error.html",
#             {"request": request, "error": "Внутренняя ошибка сервера"},
#             status_code=500
#         )
    

@router.get(
        '/applicants/list',
        tags=["Пользователи"],
        summary="Получить список абитуриентов"
        )
async def show_applicants_list(skip: int = 0, limit: int = 10):
    """
    Редерит сайт со списоком абитуриентов
    
    :param skip: Description
    :type skip: int
    :param limit: Description
    :type limit: int
    """
    # return dict(list(fake_users.items())[:limit])
    return {"message": "Тут заглушка для отображения списка студентов"}


@router.post(
        "/register/",
        summary="Регистрация нового пользователя",
        description="Создает нового пользователя в базе данных и переводит на страницу новостей"
        )
async def applicant_registration(
    user_data: ApplicantSchema,
):
    """
    Регистрирует нового пользователя и переводит на страницу новостной ленты
    """
    return {
        "username": "username", 
        "email": "email", 
        "age": "age", 
    }


# @router.patch('/applicant/{user_id}')
# async def change_applicant_status(user: User):
#     return {"message": "Изменение статуса студента"}


@router.post("/uploadfile/")
async def create_upload_file(file: UploadFile):
    """Можно сделать возможность добавлять учеников в бд с помощью 
    загрузки файла пользователя"""

    if file.content_type not in ["text/txt", "csv/csv"]:
        return {"error": "Только .txt и .csv файлы разрешены"}
    content = await file.read()  # Читаем весь файл в память
    return {"filename": file.filename, "size": len(content)}



@router.get("/analitics", dependencies=[Depends(require_analyst)])
async def show_analics():
    """
    Просмотр аналитики. Доступ только для роли analyst.
    Студент и приёмная комиссия получат 403.
    """
    return {"data": "TOP_SECRET"}


@router.delete("/applicants/{student_id}", dependencies=[Depends(require_admissions)])
async def delete_applicant(student_id: int):
    """
    Удаление студента. Доступ только для приёмной комиссии (admissions).
    """
    # TODO: реализовать удаление из БД
    return {"message": f"Студент {student_id} удалён (заглушка)"}