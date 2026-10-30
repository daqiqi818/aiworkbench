import { Alert, App, Button, Checkbox, Drawer, Empty, Form, Input, InputNumber, Modal, Select, Tag } from "antd";
import { Check, FilePenLine, FolderOpen, Plus, RefreshCw, X } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { PageHeader, WorkspacePage } from "@/components/layout/workspace-page";
import { approveDramaEditTask, cancelDramaEditTask, createDramaEditTask, getDramaEditTask, listDramaEditEpisodeFolders, listDramaEditTasks, requestDramaEditRevision, startDramaEditProduction, type CreateDramaEditTask, type DramaEditStatus, type DramaEditTask } from "@/services/api/drama-editing";
import { resourceFileUrl, resourceIdFromStorageKey, resourceStorageKey, uploadResourceFile } from "@/services/api/resources";
import { getMediaBlob } from "@/services/file-storage";
import { getWorkspaceAsset, putWorkspaceAsset } from "@/services/api/workspace-data";
import { usesBrowserLocalResourceStore } from "@/services/workspace-resource-storage";
import { persistWorkspaceAssetChanges } from "@/services/workspace-asset-repository";
import { useAssetStore, type Asset, type VideoAsset } from "@/stores/use-asset-store";
import "./drama-editing.css";

const statusText: Record<DramaEditStatus, string> = { awaiting_approval: "待审核", revision_requested: "待修改", approved: "已批准", processing: "制作中", completed: "已完成", blocked: "已受阻", cancelled: "已取消" };
const languageOptions = [{ value: "zh", label: "中文" }, { value: "en", label: "英文" }];
const platformOptions = [{ value: "tiktok", label: "TikTok" }, { value: "facebook", label: "Facebook" }, { value: "wechat_channels", label: "视频号" }, { value: "douyin", label: "抖音" }];

async function handoffProductionSources(task: DramaEditTask) {
    if (!usesBrowserLocalResourceStore()) return;
    const assets = new Map(useAssetStore.getState().assets.map((asset) => [asset.id, asset]));
    for (const episode of task.episodes) {
        let asset = assets.get(episode.assetId);
        if (!asset) {
            try {
                asset = (await getWorkspaceAsset(episode.assetId)).asset;
            } catch {
                throw new Error(`${episode.assetTitle} 的素材记录不存在，请在资产菜单重新导入该集后重试`);
            }
        }
        if (asset.kind !== "video") throw new Error(`${episode.assetTitle} 不是可制作的视频素材，请重新选择剧集后重试`);
        const video = asset as VideoAsset;
        const storageKey = video.data.storageKey || "";
        if (resourceIdFromStorageKey(storageKey)) continue;
        if (!storageKey) throw new Error(`${episode.assetTitle} 没有本地视频引用，请在资产菜单重新导入该集后重试`);
        const blob = await getMediaBlob(storageKey);
        if (!blob || !blob.size) throw new Error(`${episode.assetTitle} 的本地视频无法读取，请在资产菜单重新导入该集后重试`);
        const resource = await uploadResourceFile(blob, "video", {
            fileName: `${episode.assetTitle}.mp4`,
            width: video.data.width,
            height: video.data.height,
            durationMs: video.data.durationMs,
            idempotencyKey: `drama-edit-source:${asset.id}:${storageKey}`,
        });
        const updated: Asset = {
            ...video,
            data: { ...video.data, storageKey: resourceStorageKey(resource.id), url: resource.publicUrl || resourceFileUrl(resource.id), bytes: resource.size || blob.size, mimeType: resource.mimeType || blob.type || video.data.mimeType },
        };
        await putWorkspaceAsset(updated.id, updated);
        useAssetStore.getState().updateAsset(updated.id, updated);
    }
}

export default function DramaEditingPage() {
    const { message } = App.useApp();
    const [form] = Form.useForm<CreateDramaEditTask>();
    const [tasks, setTasks] = useState<DramaEditTask[]>([]);
    const [selected, setSelected] = useState<DramaEditTask | null>(null);
    const [createOpen, setCreateOpen] = useState(false);
    const [revisionOpen, setRevisionOpen] = useState(false);
    const [busy, setBusy] = useState(false);
    const [submissionError, setSubmissionError] = useState("");
    const [operationError, setOperationError] = useState("");
    const [filter, setFilter] = useState<DramaEditStatus | undefined>();
    const [selectedFolderIDs, setSelectedFolderIDs] = useState<string[]>([]);
    const [selectedEpisodeAssetIDs, setSelectedEpisodeAssetIDs] = useState<string[]>([]);
    const folderQuery = useQuery({
        queryKey: ["drama-editing", "episode-folders"],
        queryFn: async () => {
            await persistWorkspaceAssetChanges();
            return await listDramaEditEpisodeFolders();
        },
        enabled: createOpen,
    });
    const folders = folderQuery.data?.folders || [];
    const visibleEpisodes = folders.filter((folder) => selectedFolderIDs.includes(folder.folderId)).flatMap((folder) => folder.videos.map((video) => ({ ...video, folderId: folder.folderId, folderName: folder.folderName })));

    const load = useCallback(async () => {
        setBusy(true);
        try { const result = await listDramaEditTasks(filter); setTasks(result.tasks); }
        catch (error) { message.error(error instanceof Error ? error.message : "加载剪辑任务失败"); }
        finally { setBusy(false); }
    }, [filter, message]);

    useEffect(() => { void load(); }, [load]);
    useEffect(() => {
        if (!selected || selected.status !== "processing") return;
        const timer = window.setInterval(() => {
            void getDramaEditTask(selected.id).then((task) => { setSelected(task); void load(); }).catch(() => undefined);
        }, 2500);
        return () => window.clearInterval(timer);
    }, [selected?.id, selected?.status, load]);
    useEffect(() => {
        if (!createOpen || !folderQuery.data) return;
        setSelectedFolderIDs(folderQuery.data.folders.map((folder) => folder.folderId));
        setSelectedEpisodeAssetIDs(folderQuery.data.folders.flatMap((folder) => folder.videos.map((video) => video.assetId)));
    }, [createOpen, folderQuery.data]);

    const openCreate = () => {
        form.setFieldsValue({ deliveryPreferences: { aspectRatio: "9:16", minDurationSeconds: 15, maxDurationSeconds: 60, language: "zh", platform: "douyin" }, spoilerPolicy: { policy: "no-major-spoilers", restrictedItems: [] }, episodeAssetIds: [] });
        setSelectedFolderIDs([]);
        setSelectedEpisodeAssetIDs([]);
        setSubmissionError("");
        setCreateOpen(true);
    };
    const saveTask = async () => {
        setSubmissionError("");
        if (folderQuery.isLoading || folderQuery.isFetching) {
            setSubmissionError("剧集素材仍在同步，请完成后再创建任务。");
            return;
        }
        if (folderQuery.isError) {
            setSubmissionError("剧集素材同步失败，请重试后再创建任务。");
            return;
        }
        let values: CreateDramaEditTask;
        try {
            values = await form.validateFields();
        } catch {
            const first = form.getFieldsError().find((field) => field.errors.length);
            if (first?.name.length) void form.scrollToField(first.name);
            setSubmissionError("请完成短剧名称和交付设置后再创建任务。");
            return;
        }
        if (selectedEpisodeAssetIDs.length < 2) {
            setSubmissionError(`当前仅选中 ${selectedEpisodeAssetIDs.length} 集。多剧集剪辑任务至少需要选择两个剧集视频。`);
            return;
        }
        setBusy(true);
        try {
            const task = await createDramaEditTask({ ...values, episodeAssetIds: selectedEpisodeAssetIDs });
            setCreateOpen(false); form.resetFields(); setSelected(task); message.success("剪辑任务已创建，等待审核"); await load();
        } catch (error) {
            const detail = error instanceof Error ? error.message : "创建剪辑任务失败";
            setSubmissionError(detail);
            message.error(detail);
        }
        finally { setBusy(false); }
    };
    const apply = async (operation: "approve" | "start" | "cancel" | "revision", note = "") => {
        if (!selected) return;
        setOperationError("");
        setBusy(true);
        try {
            if (operation === "approve" || operation === "start") await handoffProductionSources(selected);
            const task = operation === "approve" ? await approveDramaEditTask(selected.id) : operation === "start" ? await startDramaEditProduction(selected.id) : operation === "cancel" ? await cancelDramaEditTask(selected.id, note) : await requestDramaEditRevision(selected.id, note);
            setSelected(task); setRevisionOpen(false); message.success("剪辑任务已更新"); await load();
        } catch (error) { const detail = error instanceof Error ? error.message : "更新剪辑任务失败"; setOperationError(detail); message.error(detail); }
        finally { setBusy(false); }
    };

    return <WorkspacePage className="drama-editing-page">
        <PageHeader title="多剧集剪辑" description="从剧集素材文件夹创建、审核并跟踪剪辑任务" actions={<><Button icon={<RefreshCw className="size-4" />} loading={busy} onClick={() => void load()}>刷新</Button><Button type="primary" icon={<Plus className="size-4" />} onClick={openCreate}>新建剪辑任务</Button></>} />
        <div className="drama-editing-toolbar"><Select aria-label="按状态筛选" value={filter} allowClear placeholder="全部状态" onChange={setFilter} options={Object.entries(statusText).map(([value, label]) => ({ value, label }))} /></div>
        <section className="drama-editing-list" aria-label="剪辑任务列表">
            {tasks.length ? tasks.map((task) => <button type="button" className="drama-editing-task" key={task.id} onClick={() => setSelected(task)}><span><strong>{task.dramaTitle}</strong><small>{task.episodes.length} 集 · {task.deliveryPreferences.minDurationSeconds}-{task.deliveryPreferences.maxDurationSeconds} 秒</small></span><span className="drama-editing-task-meta"><Tag color={task.status === "approved" ? "success" : task.status === "cancelled" ? "default" : "processing"}>{statusText[task.status]}</Tag><small>{task.stage}</small></span></button>) : <Empty description="尚未创建多剧集剪辑任务" image={Empty.PRESENTED_IMAGE_SIMPLE} />}
        </section>
        <Drawer title="剪辑任务" open={Boolean(selected)} onClose={() => setSelected(null)} width={560} extra={selected ? <Tag color="processing">{statusText[selected.status]}</Tag> : null}>
            {selected ? <div className="drama-editing-detail"><h2>{selected.dramaTitle}</h2><p>{selected.nextAction}</p><section><h3>交付范围</h3><p>{selected.deliveryPreferences.minDurationSeconds}-{selected.deliveryPreferences.maxDurationSeconds} 秒 · {languageOptions.find((item) => item.value === selected.deliveryPreferences.language)?.label} · {platformOptions.find((item) => item.value === selected.deliveryPreferences.platform)?.label}</p></section><section><h3>剧集素材</h3><div className="drama-editing-episode-summary">{selected.episodes.map((episode) => <span key={episode.assetId}>{episode.sequenceNumber}. {episode.folderName} / {episode.assetTitle}</span>)}</div></section><section><h3>剪辑方案</h3><dl><dt>叙事钩子</dt><dd>{selected.proposal.hook}</dd><dt>情绪走向</dt><dd>{selected.proposal.emotionalArc}</dd><dt>覆盖剧集</dt><dd>{selected.proposal.episodeCoverage.join("、")}</dd><dt>交付定义</dt><dd>{selected.proposal.deliveryDefinition}</dd><dt>成本</dt><dd>{selected.proposal.costEstimate}</dd></dl></section>{selected.productionTaskId ? <section><h3>制作进度</h3><p>{selected.stage}{typeof selected.productionProgress === "number" ? ` · ${selected.productionProgress}%` : ""}</p></section> : null}{operationError ? <Alert type="error" showIcon message="素材交接失败" description={operationError} /> : null}{selected.productionError ? <Alert type="error" showIcon message="制作受阻" description={selected.productionError} /> : null}{selected.deliveryUrl ? <section><h3>成片</h3><Button type="primary" href={selected.deliveryUrl} target="_blank">打开成片</Button></section> : null}<section><h3>决策记录</h3>{selected.decisions?.map((item) => <p key={item.id}><strong>{item.action}</strong>{item.note ? `：${item.note}` : ""}</p>)}</section>{(selected.status === "awaiting_approval" || selected.status === "revision_requested") ? <div className="drama-editing-actions"><Button icon={<FilePenLine className="size-4" />} onClick={() => setRevisionOpen(true)}>请求修改</Button><Button type="primary" icon={<Check className="size-4" />} loading={busy} onClick={() => void apply("approve")}>批准并开始制作</Button><Button danger icon={<X className="size-4" />} onClick={() => void apply("cancel")}>取消任务</Button></div> : selected.status === "approved" || selected.status === "blocked" ? <div className="drama-editing-actions"><Button type="primary" loading={busy} onClick={() => void apply("start")}>开始制作</Button></div> : null}</div> : null}
        </Drawer>
        <Modal title="新建多剧集剪辑任务" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={() => void saveTask()} confirmLoading={busy} width={680}>
            <Form form={form} layout="vertical" className="drama-editing-compact-form">{submissionError ? <Alert type="error" showIcon message={submissionError} className="mb-3" /> : null}<div className="drama-editing-form-grid"><Form.Item name="dramaTitle" label="短剧名称" rules={[{ required: true, message: "请输入短剧名称" }]}><Input /></Form.Item><Form.Item name={["deliveryPreferences", "aspectRatio"]} label="画幅" rules={[{ required: true }]}><Select options={["9:16", "16:9", "1:1"].map((value) => ({ value }))} /></Form.Item><Form.Item name={["deliveryPreferences", "minDurationSeconds"]} label="最小时长（秒）" rules={[{ required: true }]}><InputNumber min={15} max={300} className="w-full" /></Form.Item><Form.Item name={["deliveryPreferences", "maxDurationSeconds"]} label="最大时长（秒）" dependencies={[["deliveryPreferences", "minDurationSeconds"]]} rules={[{ required: true }, ({ getFieldValue }) => ({ validator(_, value) { const min = getFieldValue(["deliveryPreferences", "minDurationSeconds"]); return typeof value === "number" && value >= min ? Promise.resolve() : Promise.reject(new Error("最大时长不能小于最小时长")); } })]}><InputNumber min={15} max={300} className="w-full" /></Form.Item><Form.Item name={["deliveryPreferences", "language"]} label="语言" rules={[{ required: true }]}><Select options={languageOptions} /></Form.Item><Form.Item name={["deliveryPreferences", "platform"]} label="投放平台" rules={[{ required: true }]}><Select options={platformOptions} /></Form.Item></div><Form.Item name={["spoilerPolicy", "policy"]} label="剧透限制" rules={[{ required: true }]}><Select options={[{ value: "no-major-spoilers", label: "避免重大剧透" }, { value: "limited-spoilers", label: "允许有限剧透" }, { value: "approved-spoilers", label: "允许已批准剧透" }]} /></Form.Item><section className="drama-editing-folder-picker"><div className="drama-editing-section-heading"><span>剧集素材</span><Tag>{selectedEpisodeAssetIDs.length}/{visibleEpisodes.length} 集已选</Tag></div>{folderQuery.isLoading ? <p className="drama-editing-muted">正在同步并识别剧集文件夹…</p> : folderQuery.isError ? <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={folderQuery.error instanceof Error ? folderQuery.error.message : "剧集素材同步失败，请重试。"}><Button onClick={() => void folderQuery.refetch()}>重试</Button></Empty> : folders.length ? <><Select mode="multiple" value={selectedFolderIDs} options={folders.map((folder) => ({ value: folder.folderId, label: `${folder.folderName}（${folder.videoCount} 集）` }))} onChange={(values) => { setSubmissionError(""); setSelectedFolderIDs(values); const available = new Set(folders.filter((folder) => values.includes(folder.folderId)).flatMap((folder) => folder.videos.map((video) => video.assetId))); setSelectedEpisodeAssetIDs((current) => current.filter((id) => available.has(id))); }} placeholder="选择资产剧集文件夹" className="w-full" />{visibleEpisodes.length ? <Checkbox.Group value={selectedEpisodeAssetIDs} onChange={(values) => { setSubmissionError(""); setSelectedEpisodeAssetIDs(values as string[]); }} className="drama-editing-folder-options">{visibleEpisodes.map((episode) => <Checkbox key={episode.assetId} value={episode.assetId}><FolderOpen className="size-3.5" /> {episode.assetTitle}<small>{episode.folderName}</small></Checkbox>)}</Checkbox.Group> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="请选择至少一个含视频的剧集文件夹。" />}</> : <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="没有含视频的剧集文件夹，请先在资产菜单创建文件夹并归入视频素材。" />}</section></Form>
        </Modal>
        <Modal title="请求修改" open={revisionOpen} onCancel={() => setRevisionOpen(false)} onOk={() => { const note = (document.getElementById("drama-edit-revision-note") as HTMLTextAreaElement | null)?.value || ""; void apply("revision", note); }} confirmLoading={busy}><Input.TextArea id="drama-edit-revision-note" autoSize={{ minRows: 3 }} placeholder="说明需要调整的剪辑方向" /></Modal>
    </WorkspacePage>;
}
