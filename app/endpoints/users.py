from fastapi import APIRouter, Depends
from fastapi import FastAPI, Form, UploadFile
from app.db.models import User


router = APIRouter(
    prefix="/users",
    tags=["users"],
    # dependencies=[Depends(get_current_user)]  # Все эндпоинты требуют авторизации
)

@router.get("/")
async def news():
    return {"message": "Тут заглушка для отображения новостной ленты"}


@router.get('/students/{student_id}')
async def get_student(student_id: int):
    return {"user_id": student_id, "message": "здесь должен отображаться каибнет студента"}


@router.get('/students/list')
async def show_students_list(skip: int = 0, limit: int = 10):
    # return dict(list(fake_users.items())[:limit])
    return {"message": "Тут заглушка для отображения списка студентов"}


@router.post("/register/")
async def register_user(
    username: str = Form(...),
    email: str = Form(...),
    age: int = Form(...),
    password: str = Form(...)
):
    return {
        "username": username, 
        "email": email, 
        "age": age, 
        "password_length": len(password)
    }


@router.patch('/users/{user_id}')
async def change_student_status(user: User):
    return {"message": "Изменение статуса студента"}


@router.post("/uploadfile/")
async def create_upload_file(file: UploadFile):
    """Можно сделать возможность добавлять учеников в бд с помощью 
    загрузки файла пользователя"""

    if file.content_type not in ["text/txt", "csv/csv"]:
        return {"error": "Только .txt и .csv файлы разрешены"}
    content = await file.read()  # Читаем весь файл в память
    return {"filename": file.filename, "size": len(content)}