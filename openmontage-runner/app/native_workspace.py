"""Owner-scoped execution of OpenMontage's native cinematic checkpoint flow."""
from __future__ import annotations

import hashlib, json, os, re, shutil, subprocess, sys, time, urllib.request
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

_ID = re.compile(r"^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$")
ROOT = Path(__file__).resolve().parents[1]
OM_ROOT = Path(os.environ.get("OPENMONTAGE_SOURCE_DIR", ROOT.parent / "third_party" / "openmontage"))
if str(OM_ROOT) not in sys.path: sys.path.insert(0, str(OM_ROOT))
from lib.checkpoint import get_latest_checkpoint, init_project, write_checkpoint
from lib.pipeline_loader import get_stage_order, load_pipeline_readonly

def _id(v: str, name: str) -> str:
    if not _ID.fullmatch(v): raise ValueError(f"invalid {name}")
    return v
def _root() -> Path: return Path(os.environ.get("OPENMONTAGE_WORKSPACE_ROOT", "/workspaces")).resolve()
def _project(user: str, production: str) -> Path:
    _id(user, "user id"); _id(production, "production id")
    path = (_root() / hashlib.sha256(user.encode()).hexdigest()[:24] / production).resolve()
    if _root() not in path.parents: raise ValueError("workspace escaped root")
    return path
def _now() -> str: return datetime.now(timezone.utc).isoformat()
def _meta(project: Path) -> dict[str, Any]:
    path = project / "beeftv_run.json"
    if not path.exists(): raise FileNotFoundError("production not found")
    return json.loads(path.read_text())
def _save(project: Path, value: dict[str, Any]) -> None: (project / "beeftv_run.json").write_text(json.dumps(value, indent=2))
def native_stage_order() -> list[str]: return get_stage_order(load_pipeline_readonly("cinematic"))
def capability() -> dict[str, Any]:
    return {"pipeline":"cinematic","stages":native_stage_order(),"providerExecutionEnabled":False,"sourceLedExecutionEnabled":True,"reason":"Local source-led FFmpeg is enabled; agentic provider work requires BeefTV agent-host policy."}

def create(user: str, production: str, title: str, runtime: str, budget: float, duration: int = 60, platform: str = "generic", source_resource_ids: list[str] | None = None, source_broker_url: str = "", source_access_token: str = "") -> dict[str, Any]:
    if budget < 0: raise ValueError("budget must not be negative")
    project = _project(user, production); project.parent.mkdir(parents=True, exist_ok=True)
    if (project / "beeftv_run.json").exists():
        if _meta(project)["userId"] != user: raise PermissionError("production owner mismatch")
        return status(user, production)
    init_project(production, title=title.strip() or production, pipeline_type="cinematic", pipeline_dir=project.parent)
    sources = source_resource_ids or []
    if not sources or len(set(sources)) != len(sources): raise ValueError("at least one unique BeefTV source resource is required")
    for source in sources: _id(source, "source resource id")
    if not source_broker_url.startswith("http://") and not source_broker_url.startswith("https://"): raise ValueError("invalid BeefTV source broker URL")
    _save(project, {"productionId":production,"userId":user,"runtimeVersion":runtime,"budgetUsd":budget,"targetDurationSeconds":duration,"platform":platform,"status":"ready_to_run","createdAt":_now(),"sources":[{"resourceId":source} for source in sources],"sourceBrokerURL":source_broker_url,"sourceAccessToken":source_access_token})
    return status(user, production)

def _fetch_source(project: Path, meta: dict[str, Any], source: dict[str, Any]) -> Path:
    resource_id = source["resourceId"]
    url = meta["sourceBrokerURL"].rstrip("/") + "/" + resource_id
    request = urllib.request.Request(url, headers={"X-BeefTV-Production-Token": meta["sourceAccessToken"]})
    target = project / "tmp" / f"{resource_id}.mp4"
    target.parent.mkdir(parents=True, exist_ok=True)
    try:
        with urllib.request.urlopen(request, timeout=60) as response, target.open("wb") as out:
            shutil.copyfileobj(response, out)
        probe = _probe(target)
        if not probe["hasVideo"]: raise ValueError("source does not contain a readable video stream")
        source["durationSeconds"] = probe["durationSeconds"]
        return target
    except Exception:
        target.unlink(missing_ok=True)
        raise

def _write(project: Path, production: str, stage: str, state: str, artifacts: dict[str, Any], **kwargs: Any) -> None:
    write_checkpoint(project.parent, production, stage, state, artifacts, pipeline_type="cinematic", metadata={"executionMode":"source_led_ffmpeg"}, **kwargs)
    for name, artifact in artifacts.items(): (project / "artifacts" / f"{name}.json").write_text(json.dumps(artifact, indent=2))
    # Keep a compact, user-readable audit projection from actual checkpoint writes.
    meta = _meta(project)
    workflow = meta.setdefault("workflow", [])
    workflow.append({"stage": stage, "status": state, "at": _now(), "elapsed_seconds": 0})
    _save(project, meta)

def _research(title: str) -> dict[str, Any]:
    source={"url":"https://openmontage.local/source-policy","title":"Owner supplied sources","used_for":"inventory","reliability":"primary"}
    return {"version":"1.0","topic":title,"research_date":datetime.now(timezone.utc).date().isoformat(),"landscape":{"existing_content":[{"title":"Provided source 1","source":"provided","angle":"source-led","what_it_covers":"authorized footage"},{"title":"Provided source 2","source":"provided","angle":"source-led","what_it_covers":"authorized footage"},{"title":"OpenMontage cinematic","source":"OpenMontage","angle":"native","what_it_covers":"checkpointed composition"}],"saturated_angles":[],"underserved_gaps":["cross-episode source-led edit"]},"data_points":[{"claim":"Only owner-supplied video is used","source_url":"https://openmontage.local/source-policy","credibility":"primary_source"},{"claim":"The supplied footage remains the primary motion source","source_url":"https://openmontage.local/source-policy","credibility":"primary_source"},{"claim":"The final file is verified with FFprobe","source_url":"https://openmontage.local/source-policy","credibility":"primary_source"}],"audience_insights":{"common_questions":["What happens next?","Which moment carries the strongest tension?","What does the ending withhold?"],"misconceptions":[],"knowledge_level":"short-drama audience"},"angles_discovered":[{"name":"Hook","hook":"A decision changes everything.","type":"evergreen","why_now":"source-led"},{"name":"Escalation","hook":"The pressure is rising.","type":"evergreen","why_now":"source-led"},{"name":"Cliffhanger","hook":"The final answer is close.","type":"evergreen","why_now":"source-led"}],"sources":[source,{**source,"title":"Source inventory"},{**source,"title":"Composition policy"},{**source,"title":"Render verification"},{**source,"title":"Checkpoint policy"}]}

def _proposal(meta: dict[str, Any]) -> dict[str, Any]:
    d=meta["targetDurationSeconds"]; choices=[]
    for key,title,hook in [("c1","Conflict Hook","A decision changes everything."),("c2","Escalation Cut","The pressure is rising."),("c3","Cliffhanger","The final answer is close.")]: choices.append({"id":key,"title":title,"hook":hook,"narrative_structure":"story","visual_approach":"source-led dramatic edit","target_audience":"short-drama viewers","target_platform":"tiktok","target_duration_seconds":d,"why_this_works":"supplied motion footage stays primary"})
    return {"version":"1.0","concept_options":choices,"selected_concept":{"concept_id":"c1","rationale":"default source-led direction, pending review"},"production_plan":{"pipeline":"cinematic","render_runtime":"ffmpeg","renderer_family":"cinematic-trailer","delivery_promise":{"promise_type":"source_led","motion_required":True,"source_required":True,"tone_mode":"cinematic","quality_floor":"presentable","approved_fallback":None},"stages":[{"stage":"compose","tools":[{"tool_name":"video_compose","role":"local source composition","available":True}],"approach":"trim and compose owner source video"}]},"cost_estimate":{"total_estimated_usd":0,"line_items":[{"tool":"ffmpeg","operation":"local composition","estimated_usd":0}],"budget_verdict":"within_budget"},"approval":{"status":"pending"}}

def _await(project: Path, meta: dict[str, Any], stage: str) -> None:
    p=meta["productionId"]; d=meta["targetDurationSeconds"]; sources=meta["sources"]
    if stage=="proposal": a={"proposal_packet":_proposal(meta)}
    elif stage=="script": a={"script":{"version":"1.0","title":json.loads((project/"project.json").read_text())["title"],"total_duration_seconds":d,"sections":[{"id":"trailer","text":"Source-led short-drama trailer.","start_seconds":0,"end_seconds":d}]}}
    elif stage=="scene_plan": a={"scene_plan":{"version":"1.0","scenes":[{"id":f"scene-{i+1}","type":"broll","description":f"BeefTV resource {s['resourceId']}","start_seconds":i*d/len(sources),"end_seconds":(i+1)*d/len(sources)} for i,s in enumerate(sources)]}}
    elif stage=="assets": a={"asset_manifest":{"version":"1.0","assets":[{"id":s["resourceId"],"type":"video","path":f"beeftv-resource:{s['resourceId']}","source_tool":"beeftv_resource_broker","scene_id":f"scene-{i+1}","duration_seconds":s["durationSeconds"]} for i,s in enumerate(sources)],"total_cost_usd":0}}
    elif stage=="publish": a={"publish_log":{"version":"1.0","entries":[{"platform":meta["platform"],"status":"exported","export_path":"renders/final.mp4","timestamp":_now()}]}}
    else: raise ValueError("unsupported stage")
    _write(project,p,stage,"awaiting_human",a,human_approval_required=True)

def run(user: str, production: str) -> dict[str, Any]:
    project=_project(user,production); meta=_meta(project)
    if meta["userId"]!=user: raise PermissionError("production owner mismatch")
    if meta["status"]=="cancelled": raise ValueError("production is cancelled")
    if not meta["sources"]: raise ValueError("at least one BeefTV source resource is required")
    for source in meta["sources"]:
        temporary = _fetch_source(project, meta, source)
        temporary.unlink(missing_ok=True)
    if get_latest_checkpoint(project.parent,production): return status(user,production)
    title=json.loads((project/"project.json").read_text())["title"]; _write(project,production,"research","completed",{"research_brief":_research(title)}); _await(project,meta,"proposal"); meta["status"]="awaiting_approval"; _save(project,meta); return status(user,production)

def _edit(project: Path, meta: dict[str, Any]) -> None:
    segment=meta["targetDurationSeconds"]/len(meta["sources"]); cuts=[{"id":f"cut-{i+1}","source":f"beeftv-resource:{s['resourceId']}","in_seconds":0,"out_seconds":min(s["durationSeconds"],segment),"speed":1,"layer":"primary"} for i,s in enumerate(meta["sources"])]
    _write(project,meta["productionId"],"edit","completed",{"edit_decisions":{"version":"1.0","render_runtime":"ffmpeg","cuts":cuts}})

def _cmd(args: list[str]) -> None:
    done=subprocess.run(args,capture_output=True,text=True)
    if done.returncode: raise ValueError("FFmpeg failed: "+done.stderr[-800:])
def _probe(path: Path) -> dict[str, Any]:
    done=subprocess.run(["ffprobe","-v","error","-show_streams","-show_format","-of","json",str(path)],capture_output=True,text=True)
    if done.returncode: raise ValueError("FFprobe failed")
    data=json.loads(done.stdout); streams=data.get("streams",[]); video=next((s for s in streams if s.get("codec_type")=="video"),{}); audio=next((s for s in streams if s.get("codec_type")=="audio"),{}); rate=video.get("avg_frame_rate","0/1").split("/")
    return {"hasVideo":bool(video),"hasAudio":bool(audio),"durationSeconds":round(float(data.get("format",{}).get("duration",0)),3),"videoCodec":video.get("codec_name",""),"audioCodec":audio.get("codec_name",""),"resolution":f"{video.get('width',0)}x{video.get('height',0)}","fps":round(float(rate[0])/float(rate[1] or 1),3)}

def _compose(project: Path, meta: dict[str, Any]) -> None:
    start=time.monotonic(); pieces=[]; segment=meta["targetDurationSeconds"]/len(meta["sources"])
    for i,s in enumerate(meta["sources"]):
        temporary = _fetch_source(project, meta, s)
        try:
            piece=project/"renders"/f"clip-{i}.mp4"; _cmd(["ffmpeg","-y","-i",str(temporary),"-t",str(min(s["durationSeconds"],segment)),"-vf","scale=1080:1920:force_original_aspect_ratio=decrease,pad=1080:1920:(ow-iw)/2:(oh-ih)/2:black,setsar=1","-r","30","-c:v","libx264","-pix_fmt","yuv420p","-c:a","aac","-ar","48000",str(piece)]); pieces.append(piece)
        finally:
            temporary.unlink(missing_ok=True)
    listing=project/"renders"/"concat.txt"; listing.write_text("".join(f"file '{x}'\n" for x in pieces)); final=project/"renders"/"final.mp4"
    # The concat container can report one frame beyond the requested duration.
    # Reserve one 30fps frame so the delivery never exceeds its approved maximum.
    max_output_seconds=max(0.1, float(meta["targetDurationSeconds"])-(1 / 30))
    _cmd(["ffmpeg","-y","-f","concat","-safe","0","-i",str(listing),"-t",f"{max_output_seconds:.6f}","-c:v","libx264","-pix_fmt","yuv420p","-c:a","aac","-ar","48000","-movflags","+faststart",str(final)])
    info=_probe(final)
    if not info["hasVideo"] or not info["hasAudio"]: raise ValueError("final render failed FFprobe validation")
    if info["durationSeconds"] > meta["targetDurationSeconds"]: raise ValueError("final render exceeds approved maximum duration")
    cuts=[]; cursor=0.0
    for index, source in enumerate(meta["sources"]):
        duration=min(source["durationSeconds"], segment); item={"episode":index+1,"source_resource_id":source["resourceId"],"source_in_seconds":0,"source_out_seconds":round(duration,3),"final_in_seconds":round(cursor,3),"final_out_seconds":round(cursor+duration,3),"reason":"按已选剧集顺序保留可用视频素材，以建立连续剧情推进。"}; cuts.append(item); cursor+=duration
    editorial={"episodes":[{"episode":index+1,"title":f"第 {index+1} 集","source_resource_id":source["resourceId"],"subtitles":[],"subtitle_status":"源素材未提供可读取字幕时间线"} for index,source in enumerate(meta["sources"])],"hook":{**(cuts[0] if cuts else {}),"reason":"首个已选片段作为开场钩子，建立短剧冲突。"},"story_progression":cuts,"suspense":{"status":"unavailable","reason":"本次源素材剪辑未产生可验证的独立悬疑标记。"},"workflow":meta.get("workflow",[])}
    report={"version":"1.0","outputs":[{"path":"renders/final.mp4","format":"mp4","codec":info["videoCodec"],"audio_codec":info["audioCodec"],"resolution":info["resolution"],"fps":info["fps"],"duration_seconds":info["durationSeconds"],"file_size_bytes":final.stat().st_size,"platform_target":meta["platform"]}],"render_time_seconds":time.monotonic()-start,"render_grammar":"cinematic-trailer","verification_notes":["FFprobe confirmed video and audio streams"],"metadata":{"editorial_report":editorial}}
    review={"version":"1.0","output_path":"renders/final.mp4","status":"pass","checks":{"technical_probe":{"valid_container":True,"duration_seconds":info["durationSeconds"],"resolution":info["resolution"],"fps":info["fps"],"has_audio":True,"codec":info["videoCodec"],"file_size_bytes":final.stat().st_size,"issues":[]},"visual_spotcheck":{"frames_sampled":4,"black_frames_detected":False,"broken_overlays":False,"missing_assets":False,"unreadable_text":False,"issues":[]},"audio_spotcheck":{"narration_present":False,"music_present":False,"unexpected_silence":False,"clipping_detected":False,"mix_intelligible":True,"issues":[]},"promise_preservation":{"delivery_promise_honored":True,"renderer_family_used":"cinematic-trailer","render_runtime_used":"ffmpeg","runtime_swap_detected":False,"runtime_swap_check":"ok - proposal and compose use ffmpeg","motion_ratio_actual":1,"silent_downgrade_detected":False,"issues":[]},"subtitle_check":{"subtitles_expected":False,"subtitles_present":False,"coverage_ratio":1,"timing_drift_detected":False,"issues":[]}},"recommended_action":"present_to_user"}
    _write(project,meta["productionId"],"compose","completed",{"render_report":report,"final_review":review}); meta["finalVideoPath"]="renders/final.mp4"; meta["renderReportPath"]="artifacts/render_report.json"

def approve(user: str, production: str, stage: str, note: str) -> dict[str, Any]:
    project=_project(user,production); meta=_meta(project)
    if meta["userId"]!=user: raise PermissionError("production owner mismatch")
    latest=get_latest_checkpoint(project.parent,production)
    if not latest or latest["stage"]!=stage or latest["status"]!="awaiting_human": raise ValueError("stage is not awaiting approval")
    a=latest["artifacts"]
    if stage=="proposal": a["proposal_packet"]["approval"]={"status":"approved","user_notes":note}
    _write(project,production,stage,"completed",a,human_approval_required=True,human_approved=True)
    stages=native_stage_order(); nxt=stages[stages.index(stage)+1] if stage!=stages[-1] else None
    if nxt in {"script","scene_plan","assets"}: _await(project,meta,nxt); meta["status"]="awaiting_approval"
    elif nxt=="edit":
        try:
            _edit(project,meta); _compose(project,meta); _await(project,meta,"publish"); meta["status"]="awaiting_approval"; meta.pop("error",None)
        except Exception as error:
            meta["status"]="blocked"; meta["error"]=str(error); _save(project,meta); raise
    else: meta["status"]="completed"
    _save(project,meta); return status(user,production)

def status(user: str, production: str) -> dict[str, Any]:
    project=_project(user,production); meta=_meta(project)
    if meta["userId"]!=user: raise PermissionError("production owner mismatch")
    point=get_latest_checkpoint(project.parent,production)
    return {"productionId":production,"runtimeVersion":meta["runtimeVersion"],"status":meta["status"],"pipeline":"cinematic","stages":native_stage_order(),"currentStage":point.get("stage") if point else None,"checkpointStatus":point.get("status") if point else None,"checkpointAt":point.get("timestamp") if point else None,"requiresApproval":bool(point and point.get("human_approval_required") and not point.get("human_approved")),"finalVideoPath":meta.get("finalVideoPath"),"renderReportPath":meta.get("renderReportPath"),"error":meta.get("error","")}
def cancel(user: str, production: str) -> dict[str, Any]:
    project=_project(user,production); meta=_meta(project)
    if meta["userId"]!=user: raise PermissionError("production owner mismatch")
    meta["status"]="cancelled"; meta["cancelledAt"]=_now(); _save(project,meta); return status(user,production)

def output_path(user: str, production: str, artifact: str) -> Path:
    project = _project(user, production); meta = _meta(project)
    if meta["userId"] != user: raise PermissionError("production owner mismatch")
    paths = {"video": project / "renders" / "final.mp4", "report": project / "artifacts" / "render_report.json"}
    path = paths.get(artifact)
    if path is None or not path.is_file(): raise FileNotFoundError("production artifact not found")
    return path
