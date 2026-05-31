from fastapi import APIRouter, HTTPException, Request
from fastapi.responses import HTMLResponse, JSONResponse, RedirectResponse

from app.core.config import auth_config, security
from app.core.templates import templates
from app.db.engine import SessionDep
from app.schemas.user import ApplicantRegisterSchema, UserLoginSchema
from app.repositories.user_repository import list_directions
from app.services.auth_service import authenticate_user, register_user


router = APIRouter(
    prefix="/auth",
    tags=["authentication"],
    responses={404: {"description": "Not found"}},
)


@router.get("/register", response_class=HTMLResponse)
async def register_page(request: Request, session: SessionDep):
    directions = await list_directions(session)
    subjects = sorted({subject for direction in directions for subject in (direction.subjects or [])})
    return templates.TemplateResponse(
        "regist.html",
        {
            "request": request,
            "directions": directions,
            "subjects": subjects,
        },
    )


@router.get("/enter", response_class=HTMLResponse)
async def login_page(request: Request):
    return templates.TemplateResponse("enter.html", {"request": request})


@router.get("/login", response_class=HTMLResponse)
async def login_redirect():
    return RedirectResponse(url="/auth/enter", status_code=302)


@router.post("/register")
async def register(data: ApplicantRegisterSchema, session: SessionDep):
    applicant = await register_user(session, data)
    return {
        "message": "Регистрация успешна",
        "id": applicant.id,
        "email": applicant.email,
    }


@router.post("/login")
async def login(data: UserLoginSchema, session: SessionDep):
    user = await authenticate_user(session, data.username, data.password)
    if not user:
        raise HTTPException(status_code=401, detail="Неверный логин или пароль")

    token = security.create_access_token(
        uid=str(user.id),
        data={"role": user.role, "user_id": user.id},
    )

    response = JSONResponse(
        {
            "message": "Вход выполнен",
            "role": user.role,
            "user_id": user.id,
            "full_name": user.full_name,
        }
    )
    response.set_cookie(
        auth_config.JWT_ACCESS_COOKIE_NAME,
        token,
        httponly=True,
        samesite="lax",
        path="/",
    )
    return response


@router.get("/logout")
async def logout():
    response = RedirectResponse(url="/auth/enter", status_code=302)
    response.delete_cookie(auth_config.JWT_ACCESS_COOKIE_NAME, path="/")
    return response
