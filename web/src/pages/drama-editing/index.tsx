import { App, Button, Checkbox, Drawer, Empty, Form, Input, InputNumber, Modal, Select, Tag } from "antd";
import { Check, ClipboardList, FilePenLine, Plus, RefreshCw, X } from "lucide-react";
import { useCallback, useEffect, useState } from "react";

import { PageHeader, WorkspacePage } from "@/components/layout/workspace-page";
import { approveDramaEditTask, cancelDramaEditTask, createDramaEditTask, listDramaEditTasks, requestDramaEditRevision, type CreateDramaEditTask, type DramaEditEpisode, type DramaEditStatus, type DramaEditTask } from "@/services/api/drama-editing";
import "./drama-editing.css";

const defaultEpisodes = (): DramaEditEpisode[] => [1, 2].map((sequenceNumber) => ({ episodeId: `EP${sequenceNumber}`, sequenceNumber, summary: "", inclusion: "included", sources: [{ reference: "", rightsConfirmed: true }] }));
const statusText: Record<DramaEditStatus, string> = { awaiting_approval: "待审核", revision_requested: "待修改", approved: "已批准", blocked: "已受阻", cancelled: "已取消" };

export default function DramaEditingPage() {
    const { message } = App.useApp();
    const [form] = Form.useForm<CreateDramaEditTask>();
    const [tasks, setTasks] = useState<DramaEditTask[]>([]);
    const [selected, setSelected] = useState<DramaEditTask | null>(null);
    const [createOpen, setCreateOpen] = useState(false);
    const [revisionOpen, setRevisionOpen] = useState(false);
    const [busy, setBusy] = useState(false);
    const [filter, setFilter] = useState<DramaEditStatus | undefined>();

    const load = useCallback(async () => {
        setBusy(true);
        try { const result = await listDramaEditTasks(filter); setTasks(result.tasks); }
        catch (error) { message.error(error instanceof Error ? error.message : "加载剪辑任务失败"); }
        finally { setBusy(false); }
    }, [filter, message]);

    useEffect(() => { void load(); }, [load]);

    const saveTask = async () => {
        const values = await form.validateFields();
        setBusy(true);
        try {
            const task = await createDramaEditTask(values);
            setCreateOpen(false); form.resetFields(); setSelected(task); message.success("剪辑任务已创建，等待审核"); await load();
        } catch (error) { message.error(error instanceof Error ? error.message : "创建剪辑任务失败"); }
        finally { setBusy(false); }
    };
    const apply = async (operation: "approve" | "cancel" | "revision", note = "") => {
        if (!selected) return;
        setBusy(true);
        try {
            const task = operation === "approve" ? await approveDramaEditTask(selected.id) : operation === "cancel" ? await cancelDramaEditTask(selected.id, note) : await requestDramaEditRevision(selected.id, note);
            setSelected(task); setRevisionOpen(false); message.success("剪辑任务已更新"); await load();
        } catch (error) { message.error(error instanceof Error ? error.message : "更新剪辑任务失败"); }
        finally { setBusy(false); }
    };

    return <WorkspacePage className="drama-editing-page">
        <PageHeader title="多剧集剪辑" description="跨剧集组织素材，审核剪辑方案并跟踪任务状态" actions={<><Button icon={<RefreshCw className="size-4" />} loading={busy} onClick={() => void load()}>刷新</Button><Button type="primary" icon={<Plus className="size-4" />} onClick={() => { form.setFieldsValue({ deliveryPreferences: { aspectRatio: "9:16", targetDurationSeconds: 60, language: "中文", destinationLabel: "短视频平台" }, spoilerPolicy: { policy: "no-major-spoilers", restrictedItems: [] }, episodes: defaultEpisodes() }); setCreateOpen(true); }}>新建剪辑任务</Button></>} />
        <div className="drama-editing-toolbar"><Select aria-label="按状态筛选" value={filter} allowClear placeholder="全部状态" onChange={setFilter} options={Object.entries(statusText).map(([value, label]) => ({ value, label }))} /></div>
        <section className="drama-editing-list" aria-label="剪辑任务列表">
            {tasks.length ? tasks.map((task) => <button type="button" className="drama-editing-task" key={task.id} onClick={() => setSelected(task)}><span><strong>{task.dramaTitle}</strong><small>{task.campaignObjective}</small></span><span className="drama-editing-task-meta"><Tag color={task.status === "approved" ? "success" : task.status === "cancelled" ? "default" : "processing"}>{statusText[task.status]}</Tag><small>{task.stage}</small></span></button>) : <Empty description="尚未创建多剧集剪辑任务" image={Empty.PRESENTED_IMAGE_SIMPLE} />}
        </section>
        <Drawer title="剪辑任务" open={Boolean(selected)} onClose={() => setSelected(null)} width={560} extra={selected ? <Tag color="processing">{statusText[selected.status]}</Tag> : null}>
            {selected ? <div className="drama-editing-detail"><h2>{selected.dramaTitle}</h2><p>{selected.nextAction}</p><section><h3>剪辑方案</h3><dl><dt>叙事钩子</dt><dd>{selected.proposal.hook}</dd><dt>情绪走向</dt><dd>{selected.proposal.emotionalArc}</dd><dt>覆盖剧集</dt><dd>{selected.proposal.episodeCoverage.join("、")}</dd><dt>交付定义</dt><dd>{selected.proposal.deliveryDefinition}</dd><dt>成本</dt><dd>{selected.proposal.costEstimate}</dd></dl></section><section><h3>决策记录</h3>{selected.decisions?.map((item) => <p key={item.id}><strong>{item.action}</strong>{item.note ? `：${item.note}` : ""}</p>)}</section>{(selected.status === "awaiting_approval" || selected.status === "revision_requested") ? <div className="drama-editing-actions"><Button icon={<FilePenLine className="size-4" />} onClick={() => setRevisionOpen(true)}>请求修改</Button><Button type="primary" icon={<Check className="size-4" />} loading={busy} onClick={() => void apply("approve")}>批准方案</Button><Button danger icon={<X className="size-4" />} onClick={() => void apply("cancel")}>取消任务</Button></div> : null}</div> : null}
        </Drawer>
        <Modal title="新建多剧集剪辑任务" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={() => void saveTask()} confirmLoading={busy} width={760}>
            <Form form={form} layout="vertical"><div className="drama-editing-form-grid"><Form.Item name="dramaTitle" label="短剧名称" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name="targetAudience" label="目标受众" rules={[{ required: true }]}><Input /></Form.Item></div><Form.Item name="campaignObjective" label="剪辑目标" rules={[{ required: true }]}><Input.TextArea autoSize={{ minRows: 2, maxRows: 4 }} /></Form.Item><div className="drama-editing-form-grid"><Form.Item name={["deliveryPreferences", "aspectRatio"]} label="画幅" rules={[{ required: true }]}><Select options={["9:16", "16:9", "1:1"].map((value) => ({ value }))} /></Form.Item><Form.Item name={["deliveryPreferences", "targetDurationSeconds"]} label="目标时长（秒）" rules={[{ required: true }]}><InputNumber min={10} max={180} className="w-full" /></Form.Item><Form.Item name={["deliveryPreferences", "language"]} label="语言" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name={["deliveryPreferences", "destinationLabel"]} label="投放位置" rules={[{ required: true }]}><Input /></Form.Item></div><Form.List name="episodes">{(fields, { add, remove }) => <section><div className="drama-editing-section-heading"><h3>剧集素材</h3><Button size="small" icon={<Plus className="size-3.5" />} onClick={() => add({ episodeId: `EP${fields.length + 1}`, sequenceNumber: fields.length + 1, inclusion: "included", sources: [{ reference: "", rightsConfirmed: true }] })}>添加剧集</Button></div>{fields.map((field) => <div className="drama-editing-episode" key={field.key}><Form.Item name={[field.name, "episodeId"]} label="剧集编号" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name={[field.name, "sequenceNumber"]} label="顺序" rules={[{ required: true }]}><InputNumber min={1} /></Form.Item><Form.Item name={[field.name, "summary"]} label="剧情摘要" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name={[field.name, "sources", 0, "reference"]} label="素材引用" rules={[{ required: true }]}><Input /></Form.Item><Form.Item name={[field.name, "sources", 0, "rightsConfirmed"]} valuePropName="checked" hidden initialValue><Checkbox /></Form.Item><Button type="text" danger aria-label="移除剧集" icon={<X className="size-4" />} disabled={fields.length <= 2} onClick={() => remove(field.name)} /></div>)}</section>}</Form.List></Form>
        </Modal>
        <Modal title="请求修改" open={revisionOpen} onCancel={() => setRevisionOpen(false)} onOk={() => { const note = (document.getElementById("drama-edit-revision-note") as HTMLTextAreaElement | null)?.value || ""; void apply("revision", note); }} confirmLoading={busy}><Input.TextArea id="drama-edit-revision-note" autoSize={{ minRows: 3 }} placeholder="说明需要调整的剪辑方向" /></Modal>
    </WorkspacePage>;
}
