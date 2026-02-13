from fastapi import FastAPI
from models import User


app = FastAPI()


@app.get("/")
async def news():
    return {"message": "Тут заглушка для отображения новостной ленты"}


@app.get('/students/{student_id}')
async def get_student(user_id: int):
    return {"user_id": user_id, "message": "здесь должен отображаться каибнет студента"}


@app.get('/students/list')
async def show_students_list(limit: int = 10):
    # return dict(list(fake_users.items())[:limit])
    return {"message": "Тут заглушка для отображения списка студентов"}


@app.post('/register/')
async def register(user: User):
    return {"message": "Тут заглушка для регистрации пользователей"}


