import json
import subprocess
from threading import Thread
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

from fastapi.testclient import TestClient

from app.main import app


def client(monkeypatch, tmp_path):
    monkeypatch.setenv("OPENMONTAGE_WORKSPACE_ROOT", str(tmp_path / "workspaces"))
    monkeypatch.setenv("OPENMONTAGE_RUNNER_TOKEN", "test-token")
    return TestClient(app), {"X-Openmontage-Runner-Token": "test-token"}


def source_broker(source, token):
    class Handler(BaseHTTPRequestHandler):
        def do_GET(self):
            if self.headers.get("X-BeefTV-Production-Token") != token:
                self.send_response(403); self.end_headers(); return
            self.send_response(200); self.send_header("Content-Type", "video/mp4"); self.end_headers()
            self.wfile.write(source.read_bytes())
        def log_message(self, *_): pass
    server = ThreadingHTTPServer(("127.0.0.1", 0), Handler)
    Thread(target=server.serve_forever, daemon=True).start()
    return server, f"http://127.0.0.1:{server.server_port}/sources"


def test_capability_uses_native_cinematic_manifest(monkeypatch, tmp_path):
    api, headers = client(monkeypatch, tmp_path)
    response = api.get("/internal/openmontage/capabilities", headers=headers)
    assert response.status_code == 200
    assert response.json()["stages"] == ["research", "proposal", "script", "scene_plan", "assets", "edit", "compose", "publish"]
    assert response.json()["providerExecutionEnabled"] is False


def test_productions_are_owner_scoped_and_checkpointed(monkeypatch, tmp_path):
    api, headers = client(monkeypatch, tmp_path)
    payload = {"productionId": "prod-a", "userId": "user-a", "title": "Trailer", "runtimeVersion": "openmontage-9590f56", "budgetUsd": 0, "sourceResourceIds": ["source-a"], "sourceBrokerURL": "http://127.0.0.1:9/sources", "sourceAccessToken": "a" * 24}
    created = api.post("/internal/openmontage/productions", json=payload, headers=headers)
    assert created.status_code == 200
    assert created.json()["currentStage"] is None
    assert created.json()["status"] == "ready_to_run"

    denied = api.get("/internal/openmontage/productions/prod-a?user_id=user-b", headers=headers)
    assert denied.status_code == 404
    cancelled = api.post("/internal/openmontage/productions/prod-a/cancel?user_id=user-a", headers=headers)
    assert cancelled.status_code == 200
    assert cancelled.json()["status"] == "cancelled"


def test_source_led_cinematic_flow_renders_real_video(monkeypatch, tmp_path):
    api, headers = client(monkeypatch, tmp_path)
    source = tmp_path / "source.mp4"
    subprocess.run([
        "ffmpeg", "-y", "-f", "lavfi", "-i", "color=c=blue:s=640x360:d=2",
        "-f", "lavfi", "-i", "sine=frequency=440:duration=2", "-c:v", "libx264",
        "-pix_fmt", "yuv420p", "-c:a", "aac", str(source),
    ], check=True, capture_output=True)
    server, url = source_broker(source, "a" * 24)
    payload = {"productionId": "prod-flow", "userId": "user-a", "title": "Trailer", "runtimeVersion": "openmontage-9590f56", "budgetUsd": 0, "targetDurationSeconds": 15, "sourceResourceIds": ["episode-1"], "sourceBrokerURL": url, "sourceAccessToken": "a" * 24}
    assert api.post("/internal/openmontage/productions", json=payload, headers=headers).status_code == 200
    started = api.post("/internal/openmontage/productions/prod-flow/run", headers=headers, json={"userId": "user-a"})
    assert started.status_code == 200, started.text
    assert started.json()["currentStage"] == "proposal"
    for stage in ["proposal", "script", "scene_plan", "assets", "publish"]:
        advanced = api.post(f"/internal/openmontage/productions/prod-flow/stages/{stage}/approve", headers=headers, json={"userId": "user-a"})
        assert advanced.status_code == 200, advanced.text
    result = api.get("/internal/openmontage/productions/prod-flow?user_id=user-a", headers=headers)
    assert result.json()["status"] == "completed"
    assert result.json()["finalVideoPath"] == "renders/final.mp4"
    output = next((tmp_path / "workspaces").rglob("final.mp4"))
    duration = float(subprocess.check_output(["ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", str(output)], text=True))
    assert duration <= 15
    report = json.loads(next((tmp_path / "workspaces").rglob("render_report.json")).read_text())
    assert report["version"] == "1.0"
    assert report["metadata"]["editorial_report"]["story_progression"]
    assert not list((tmp_path / "workspaces").rglob("episode-1.mp4"))
    server.shutdown()


def test_runner_requires_internal_token(monkeypatch, tmp_path):
    api, _ = client(monkeypatch, tmp_path)
    assert api.get("/internal/openmontage/capabilities").status_code == 401
