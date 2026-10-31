"""Internal-only OpenMontage runner foundation."""

from __future__ import annotations

import os
import secrets
from fastapi import FastAPI, Header, HTTPException, status
from fastapi.responses import FileResponse
from pydantic import BaseModel, Field

from . import native_workspace

app = FastAPI(title="BeefTV OpenMontage Runner", docs_url=None, redoc_url=None)


class ProductionRequest(BaseModel):
    productionId: str = Field(pattern=r"^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$")
    userId: str = Field(pattern=r"^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$")
    title: str = Field(min_length=1, max_length=200)
    runtimeVersion: str = Field(min_length=1, max_length=128)
    budgetUsd: float = Field(ge=0)
    targetDurationSeconds: int = Field(default=60, ge=15, le=300)
    platform: str = Field(default="generic", max_length=32)
    sourceResourceIds: list[str] = Field(min_length=1, max_length=12)
    sourceBrokerURL: str = Field(min_length=1, max_length=2000)
    sourceAccessToken: str = Field(min_length=24, max_length=256)


class StageRequest(BaseModel):
    userId: str = Field(pattern=r"^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$")
    note: str = Field(default="", max_length=1000)


def _authenticate(token: str | None) -> None:
    configured = os.environ.get("OPENMONTAGE_RUNNER_TOKEN", "")
    if not configured or token is None or not secrets.compare_digest(configured, token):
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="unauthorized runner request")


def _result(call):
    try:
        return call()
    except FileNotFoundError:
        raise HTTPException(status_code=404, detail="production not found") from None
    except PermissionError:
        raise HTTPException(status_code=403, detail="production owner mismatch") from None
    except ValueError as error:
        raise HTTPException(status_code=400, detail=str(error)) from error


@app.get("/health")
def health():
    return {"status": "ok"}


@app.get("/internal/openmontage/capabilities")
def capabilities(x_openmontage_runner_token: str | None = Header(default=None)):
    _authenticate(x_openmontage_runner_token)
    return native_workspace.capability()


@app.post("/internal/openmontage/productions")
def create_production(request: ProductionRequest, x_openmontage_runner_token: str | None = Header(default=None)):
    _authenticate(x_openmontage_runner_token)
    return _result(lambda: native_workspace.create(request.userId, request.productionId, request.title, request.runtimeVersion, request.budgetUsd, request.targetDurationSeconds, request.platform, request.sourceResourceIds, request.sourceBrokerURL, request.sourceAccessToken))


@app.post("/internal/openmontage/productions/{production_id}/run")
def run_production(production_id: str, request: StageRequest, x_openmontage_runner_token: str | None = Header(default=None)):
    _authenticate(x_openmontage_runner_token)
    return _result(lambda: native_workspace.run(request.userId, production_id))


@app.post("/internal/openmontage/productions/{production_id}/stages/{stage}/approve")
def approve_stage(production_id: str, stage: str, request: StageRequest, x_openmontage_runner_token: str | None = Header(default=None)):
    _authenticate(x_openmontage_runner_token)
    return _result(lambda: native_workspace.approve(request.userId, production_id, stage, request.note))


@app.get("/internal/openmontage/productions/{production_id}")
def get_production(production_id: str, user_id: str, x_openmontage_runner_token: str | None = Header(default=None)):
    _authenticate(x_openmontage_runner_token)
    return _result(lambda: native_workspace.status(user_id, production_id))

@app.get("/internal/openmontage/productions/{production_id}/artifacts/{artifact}")
def get_artifact(production_id: str, artifact: str, user_id: str, x_openmontage_runner_token: str | None = Header(default=None)):
    _authenticate(x_openmontage_runner_token)
    path = _result(lambda: native_workspace.output_path(user_id, production_id, artifact))
    media = "video/mp4" if artifact == "video" else "application/json"
    return FileResponse(path, media_type=media, filename=path.name)


@app.post("/internal/openmontage/productions/{production_id}/cancel")
def cancel_production(production_id: str, user_id: str, x_openmontage_runner_token: str | None = Header(default=None)):
    _authenticate(x_openmontage_runner_token)
    return _result(lambda: native_workspace.cancel(user_id, production_id))
