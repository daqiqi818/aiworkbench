import { Alert, App, Button, Checkbox, Drawer, Empty, Form, Input, InputNumber, Modal, Progress, Select, Tag } from "antd";
import { Check, Download, FilePenLine, FileText, FolderOpen, Play, Plus, RefreshCw, Trash2, X } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { PageHeader, WorkspacePage } from "@/components/layout/workspace-page";
import {
    approveDramaEditOpenMontageStage,
    approveDramaEditTask,
    cancelDramaEditTask,
    createDramaEditTask,
    deleteDramaEditTasks,
    getDramaEditTask,
    listDramaEditEpisodeFolders,
    listDramaEditTasks,
    requestDramaEditRevision,
    startDramaEditProduction,
    type CreateDramaEditTask,
    type DramaEditStatus,
    type DramaEditTask,
} from "@/services/api/drama-editing";
import { persistWorkspaceAssetChanges } from "@/services/workspace-asset-repository";
import "./drama-editing.css";

const preferenceKey = "beeftv.drama-editing.create-preferences";
const fixedSpoilerPolicy = { policy: "no-major-spoilers" as const, restrictedItems: [] };
const defaults = {
    deliveryPreferences: { aspectRatio: "9:16" as const, minDurationSeconds: 15, maxDurationSeconds: 60, language: "zh" as const, platform: "douyin" as const },
    spoilerPolicy: fixedSpoilerPolicy,
    episodeAssetIds: [],
};
const statusText: Record<DramaEditStatus, string> = { awaiting_approval: "待审核", revision_requested: "待修改", approved: "已批准", processing: "制作中", completed: "已完成", blocked: "已受阻", cancelled: "已取消" };
const languages = [
    { value: "zh", label: "中文" },
    { value: "en", label: "英文" },
];
const platforms = [
    { value: "tiktok", label: "TikTok" },
    { value: "facebook", label: "Facebook" },
    { value: "wechat_channels", label: "视频号" },
    { value: "douyin", label: "抖音" },
];
const progress = (t: DramaEditTask) => t.productionProgress ?? (t.status === "completed" ? 100 : t.status === "processing" ? 50 : 0);
const taskStatusText = (task: DramaEditTask) => (task.status === "processing" ? `${statusText.processing} ${progress(task)}%` : statusText[task.status]);

export default function DramaEditingPage() {
    const { message, modal } = App.useApp();
    const [form] = Form.useForm<CreateDramaEditTask>();
    const [tasks, setTasks] = useState<DramaEditTask[]>([]);
    const [selected, setSelected] = useState<DramaEditTask | null>(null);
    const [createOpen, setCreateOpen] = useState(false);
    const [revisionOpen, setRevisionOpen] = useState(false);
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState("");
    const [filter, setFilter] = useState<DramaEditStatus>();
    const [folderID, setFolderID] = useState("");
    const [checked, setChecked] = useState<string[]>([]);
    const [marked, setMarked] = useState<string[]>([]);
    const [video, setVideo] = useState<DramaEditTask | null>(null);
    const videoRef = useRef<HTMLVideoElement>(null);
    const [report, setReport] = useState<any>(null);
    const [reportOpen, setReportOpen] = useState(false);
    const [page, setPage] = useState(0);
    const foldersQuery = useQuery({
        queryKey: ["drama-editing", "episode-folders"],
        enabled: createOpen,
        queryFn: async () => {
            await persistWorkspaceAssetChanges();
            return listDramaEditEpisodeFolders();
        },
    });
    const folders = foldersQuery.data?.folders || [];
    const folder = folders.find((f) => f.folderId === folderID);
    const load = useCallback(async () => {
        setBusy(true);
        try {
            setTasks((await listDramaEditTasks(filter)).tasks);
        } catch (e) {
            message.error(e instanceof Error ? e.message : "加载短剧剪辑任务失败");
        } finally {
            setBusy(false);
        }
    }, [filter, message]);
    useEffect(() => {
        void load();
    }, [load]);
    useEffect(() => {
        if (!selected || selected.status !== "processing") return;
        const timer = window.setInterval(
            () =>
                void getDramaEditTask(selected.id).then((t) => {
                    setSelected(t);
                    void load();
                }),
            2500,
        );
        return () => window.clearInterval(timer);
    }, [selected?.id, selected?.status, load]);
    useEffect(() => {
        if (folder) setChecked(folder.videos.map((v) => v.assetId));
        else setChecked([]);
    }, [folderID, folder?.folderId]);
    useEffect(() => {
        if (!video?.deliveryUrl) return;
        const player = videoRef.current;
        if (!player) return;
        player.src = video.deliveryUrl;
        player.load();
        void player.play().catch(() => undefined);
    }, [video]);
    const openCreate = () => {
        let saved: any = {};
        try {
            saved = JSON.parse(localStorage.getItem(preferenceKey) || "{}");
        } catch {}
        form.setFieldsValue({ ...defaults, ...saved, spoilerPolicy: fixedSpoilerPolicy, dramaTitle: "" });
        setFolderID(saved.folderId || "");
        setError("");
        setCreateOpen(true);
    };
    const save = async () => {
        setError("");
        if (foldersQuery.isFetching || foldersQuery.isLoading) return setError("剧集素材仍在同步，请稍后再试。");
        let values: CreateDramaEditTask;
        try {
            values = await form.validateFields();
        } catch {
            return setError("请完成短剧名称和交付设置后再创建任务。");
        }
        if (checked.length < 2) return setError(`当前仅选中 ${checked.length} 集，至少需要两集视频。`);
        setBusy(true);
        try {
            const taskRequest = { ...values, spoilerPolicy: fixedSpoilerPolicy, episodeAssetIds: checked };
            localStorage.setItem(preferenceKey, JSON.stringify({ ...taskRequest, dramaTitle: undefined, folderId: folderID }));
            const task = await createDramaEditTask(taskRequest);
            setCreateOpen(false);
            setSelected(task);
            message.success("短剧剪辑任务已创建");
            await load();
        } catch (e) {
            setError(e instanceof Error ? e.message : "创建任务失败");
        } finally {
            setBusy(false);
        }
    };
    const apply = async (op: "approve" | "start" | "cancel" | "revision" | "native", note = "") => {
        if (!selected) return;
        setBusy(true);
        setError("");
        try {
            const t =
                op === "approve"
                    ? await approveDramaEditTask(selected.id)
                    : op === "start"
                      ? await startDramaEditProduction(selected.id)
                      : op === "cancel"
                        ? await cancelDramaEditTask(selected.id, note)
                        : op === "native"
                          ? await approveDramaEditOpenMontageStage(selected.id, selected.openMontageStage || "", note)
                          : await requestDramaEditRevision(selected.id, note);
            setSelected(t);
            setRevisionOpen(false);
            await load();
        } catch (e) {
            setError(e instanceof Error ? e.message : "更新任务失败");
        } finally {
            setBusy(false);
        }
    };
    const remove = () => {
        if (!marked.length) return;
        let deleteEpisodeFiles = false;
        const activeCount = tasks.filter((task) => marked.includes(task.id) && (task.status === "approved" || task.status === "processing")).length;
        modal.confirm({
            title: `删除 ${marked.length} 个任务？`,
            content: (
                <div>
                    <p>默认仅删除任务记录，成片与渲染报告会保留。</p>
                    {activeCount > 0 && <p>其中 {activeCount} 个任务正在制作，将先停止 OpenMontage 制作后再删除。</p>}
                    <Checkbox
                        onChange={(event) => {
                            deleteEpisodeFiles = event.target.checked;
                        }}
                    >
                        同时删除剧集文件
                    </Checkbox>
                    <p className="drama-editing-muted">仅删除本次任务纳入且不再被其他业务引用的源视频。</p>
                </div>
            ),
            okButtonProps: { danger: true },
            onOk: async () => {
                const { results } = await deleteDramaEditTasks(marked, deleteEpisodeFiles);
                const failed = results.filter((r) => !r.deleted);
                const fileIssues = results.flatMap((r) => r.episodeFiles || []).filter((r) => !r.deleted);
                setMarked([]);
                if (selected && results.some((r) => r.id === selected.id && r.deleted)) setSelected(null);
                await load();
                if (failed.length || fileIssues.length) message.warning([...failed.map((r) => r.reason), ...fileIssues.map((r) => r.reason)].filter(Boolean).join("；"));
                else message.success(deleteEpisodeFiles ? "任务及可删除剧集文件已处理" : "任务已删除，剧集文件已保留");
            },
        });
    };
    const closeVideo = () => {
        const player = videoRef.current;
        if (player) {
            player.pause();
            player.removeAttribute("src");
            player.load();
        }
        setVideo(null);
    };
    const openReport = async (t: DramaEditTask) => {
        if (!t.renderReportUrl) return;
        try {
            const res = await fetch(t.renderReportUrl, { credentials: "include" });
            if (!res.ok) throw new Error("无法读取渲染报告");
            setReport(await res.json());
            setPage(0);
            setReportOpen(true);
        } catch (e) {
            message.error(e instanceof Error ? e.message : "无法读取渲染报告");
        }
    };
    const editorial = report?.editorial_report || report?.metadata?.editorial_report;
    const episodes = editorial?.episodes || [];
    const episode = episodes[page];
    return (
        <WorkspacePage className="drama-editing-page">
            <PageHeader
                title="短剧剪辑"
                description="从剧集素材文件夹创建、审核并跟踪剪辑任务"
                actions={
                    <>
                        <Button icon={<RefreshCw className="size-4" />} loading={busy} onClick={() => void load()}>
                            刷新
                        </Button>
                        <Button type="primary" icon={<Plus className="size-4" />} onClick={openCreate}>
                            新建任务
                        </Button>
                    </>
                }
            />
            <div className="drama-editing-toolbar">
                <Select value={filter} allowClear placeholder="全部状态" onChange={setFilter} options={Object.entries(statusText).map(([value, label]) => ({ value, label }))} />
                <Button danger disabled={!marked.length} icon={<Trash2 className="size-4" />} onClick={remove}>
                    删除所选
                </Button>
            </div>
            <section className="drama-editing-list">
                {tasks.length ? (
                    tasks.map((t) => (
                        <div className="drama-editing-task" key={t.id}>
                            <Checkbox checked={marked.includes(t.id)} onChange={(e) => setMarked((v) => (e.target.checked ? [...v, t.id] : v.filter((id) => id !== t.id)))} />
                            <button type="button" className="drama-editing-task-main" onClick={() => setSelected(t)}>
                                <strong>{t.dramaTitle}</strong>
                                <small>
                                    {t.episodes.length} 集 · {t.deliveryPreferences.minDurationSeconds}-{t.deliveryPreferences.maxDurationSeconds} 秒
                                </small>
                            </button>
                            <span className="drama-editing-task-meta">
                                <Tag color={t.status === "completed" ? "success" : t.status === "blocked" ? "error" : "processing"}>{taskStatusText(t)}</Tag>
                                {t.deliveryUrl && (
                                    <>
                                        <Button type="text" title="打开成片" icon={<Play className="size-4" />} onClick={() => setVideo({ ...t })} />
                                        <Button
                                            type="text"
                                            title="下载成片"
                                            icon={<Download className="size-4" />}
                                            onClick={() => {
                                                const a = document.createElement("a");
                                                a.href = t.deliveryUrl!;
                                                a.download = "";
                                                a.click();
                                            }}
                                        />
                                    </>
                                )}
                                {t.renderReportUrl && <Button type="text" title="渲染报告" icon={<FileText className="size-4" />} onClick={() => void openReport(t)} />}
                            </span>
                        </div>
                    ))
                ) : (
                    <Empty description="尚未创建短剧剪辑任务" image={Empty.PRESENTED_IMAGE_SIMPLE} />
                )}
            </section>
            <Drawer title="短剧剪辑任务" open={!!selected} onClose={() => setSelected(null)} width={560}>
                {selected && (
                    <div className="drama-editing-detail">
                        <h2>{selected.dramaTitle}</h2>
                        <p>{selected.nextAction}</p>
                        <Progress percent={progress(selected)} />
                        <section>
                            <h3>交付范围</h3>
                            <p>
                                {selected.deliveryPreferences.minDurationSeconds}-{selected.deliveryPreferences.maxDurationSeconds} 秒
                            </p>
                        </section>
                        <section>
                            <h3>剪辑方案</h3>
                            <p>{selected.proposal.hook}</p>
                        </section>
                        {error && <Alert type="error" showIcon message="操作失败" description={error} />} {selected.productionError && <Alert type="error" showIcon message="制作受阻" description={selected.productionError} />}{" "}
                        {(selected.status === "awaiting_approval" || selected.status === "revision_requested") && (
                            <div className="drama-editing-actions">
                                <Button icon={<FilePenLine className="size-4" />} onClick={() => setRevisionOpen(true)}>
                                    请求修改
                                </Button>
                                <Button type="primary" icon={<Check className="size-4" />} loading={busy} onClick={() => void apply("approve")}>
                                    批准并开始制作
                                </Button>
                                <Button danger icon={<X className="size-4" />} onClick={() => void apply("cancel")}>
                                    取消任务
                                </Button>
                            </div>
                        )}
                        {selected.openMontageRequiresApproval && (
                            <div className="drama-editing-actions">
                                <Button type="primary" loading={busy} onClick={() => void apply("native")}>
                                    批准 OpenMontage 阶段
                                </Button>
                            </div>
                        )}
                    </div>
                )}
            </Drawer>
            <Modal title="新建任务" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={() => void save()} confirmLoading={busy} width={680}>
                <Form form={form} layout="vertical" className="drama-editing-compact-form">
                    {error && <Alert type="error" showIcon message={error} />}
                    <div className="drama-editing-form-grid">
                        <Form.Item className="drama-editing-title-field" name="dramaTitle" label="短剧名称" rules={[{ required: true, message: "请输入短剧名称" }]}>
                            <Input />
                        </Form.Item>
                        <Form.Item name={["deliveryPreferences", "aspectRatio"]} label="画幅">
                            <Select options={["9:16", "16:9", "1:1"].map((value) => ({ value }))} />
                        </Form.Item>
                        <Form.Item name={["deliveryPreferences", "language"]} label="语言">
                            <Select options={languages} />
                        </Form.Item>
                        <Form.Item name={["deliveryPreferences", "platform"]} label="投放平台">
                            <Select options={platforms} />
                        </Form.Item>
                        <Form.Item name={["deliveryPreferences", "minDurationSeconds"]} label="最小时长（秒）">
                            <InputNumber min={15} max={300} className="w-full" />
                        </Form.Item>
                        <Form.Item name={["deliveryPreferences", "maxDurationSeconds"]} label="最大时长（秒）">
                            <InputNumber min={15} max={300} className="w-full" />
                        </Form.Item>
                    </div>
                    <section className="drama-editing-folder-picker">
                        <div className="drama-editing-section-heading">
                            <span>剧集素材</span>
                            <Tag>
                                {checked.length}/{folder?.videos.length || 0} 集已选
                            </Tag>
                        </div>
                        <Select value={folderID || undefined} options={folders.map((f) => ({ value: f.folderId, label: `${f.folderName}（${f.videoCount}集）` }))} onChange={setFolderID} placeholder="选择资产剧集文件夹" className="w-full" />
                        {folder ? (
                            <Checkbox.Group value={checked} onChange={(v) => setChecked(v as string[])} className="drama-editing-folder-options">
                                {folder.videos.map((v) => (
                                    <Checkbox key={v.assetId} value={v.assetId}>
                                        <FolderOpen className="size-3.5 shrink-0" />
                                        <span className="drama-editing-episode-title" title={v.assetTitle}>
                                            {v.assetTitle}
                                        </span>
                                    </Checkbox>
                                ))}
                            </Checkbox.Group>
                        ) : (
                            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="请选择含视频的剧集文件夹" />
                        )}
                    </section>
                </Form>
            </Modal>
            <Modal title="请求修改" open={revisionOpen} onCancel={() => setRevisionOpen(false)} onOk={() => void apply("revision", (document.getElementById("drama-edit-revision-note") as HTMLTextAreaElement)?.value || "")}>
                <Input.TextArea id="drama-edit-revision-note" />
            </Modal>
            <Modal title={video?.dramaTitle || "成片"} open={!!video} footer={null} onCancel={closeVideo} destroyOnHidden width={560} styles={{ body: { paddingBlock: 12 } }}>
                {video?.deliveryUrl && <video ref={videoRef} controls autoPlay className="drama-editing-video-player" src={video.deliveryUrl} />}
            </Modal>
            <Modal title="制作报告" open={reportOpen} footer={null} onCancel={() => setReportOpen(false)} width={820}>
                <section>
                    <h3>剧集字幕时间线</h3>
                    {episode ? (
                        <>
                            <p>{episode.title || `第 ${page + 1} 集`}</p>
                            {episode.subtitles?.length ? (
                                episode.subtitles.map((s: any, i: number) => (
                                    <p key={i}>
                                        {s.start_seconds}-{s.end_seconds} 秒：{s.text}
                                    </p>
                                ))
                            ) : (
                                <p>该剧集未产生字幕时间线。</p>
                            )}
                            <div className="drama-editing-actions">
                                <Button disabled={!page} onClick={() => setPage((p) => p - 1)}>
                                    上一集
                                </Button>
                                <Button disabled={page >= episodes.length - 1} onClick={() => setPage((p) => p + 1)}>
                                    下一集
                                </Button>
                            </div>
                        </>
                    ) : (
                        <p>没有可用的剧集时间线。</p>
                    )}
                </section>
                <section>
                    <h3>钩子</h3>
                    <p>{editorial?.hook?.reason || "未产生钩子选段说明。"}</p>
                </section>
                <section>
                    <h3>剧情推进</h3>
                    {editorial?.story_progression?.length ? (
                        editorial.story_progression.map((x: any, i: number) => (
                            <p key={i}>
                                成片 {x.final_in_seconds}-{x.final_out_seconds} 秒；原片 {x.source_in_seconds}-{x.source_out_seconds} 秒；{x.reason}
                            </p>
                        ))
                    ) : (
                        <p>未产生剧情推进片段说明。</p>
                    )}
                </section>
                <section>
                    <h3>悬疑</h3>
                    <p>{editorial?.suspense?.reason || "未产生悬疑片段说明。"}</p>
                </section>
                <section>
                    <h3>制作流程</h3>
                    {editorial?.workflow?.length ? (
                        editorial.workflow.map((x: any, i: number) => (
                            <p key={x.stage}>
                                {x.stage}：{x.status}，{x.elapsed_seconds} 秒
                            </p>
                        ))
                    ) : (
                        <p>没有可用的阶段耗时记录。</p>
                    )}
                </section>
            </Modal>
        </WorkspacePage>
    );
}
