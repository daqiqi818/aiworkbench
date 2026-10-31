import { http, type HttpRequestConfig } from "@/services/api/request";

export type DramaEditStatus = "awaiting_approval" | "revision_requested" | "approved" | "processing" | "completed" | "blocked" | "cancelled";

export type DramaEditLanguage = "zh" | "en";
export type DramaEditPlatform = "tiktok" | "facebook" | "wechat_channels" | "douyin";
export type DramaEditEpisode = { folderId: string; folderName: string; assetId: string; assetTitle: string; sequenceNumber: number };
export type DramaEditEpisodeAsset = { assetId: string; assetTitle: string };
export type DramaEditEpisodeFolder = { folderId: string; folderName: string; videoCount: number; videos: DramaEditEpisodeAsset[] };
export type DramaEditDelivery = { aspectRatio: "9:16" | "16:9" | "1:1"; minDurationSeconds: number; maxDurationSeconds: number; language: DramaEditLanguage; platform: DramaEditPlatform };
export type DramaEditSpoilerPolicy = { policy: "no-major-spoilers" | "limited-spoilers" | "approved-spoilers"; restrictedItems: string[] };
export type DramaEditProposal = { hook: string; emotionalArc: string; episodeCoverage: string[]; excludedEpisodes: string[]; deliveryDefinition: string; costEstimate: string; sourceLimitations: string[] };
export type DramaEditDecision = { id: string; action: string; note?: string; createdAt: string };
export type DramaEditEpisodeFileDeleteResult = { assetId: string; deleted: boolean; reason?: string };
export type DramaEditDeleteResult = { id: string; deleted: boolean; reason?: string; episodeFiles?: DramaEditEpisodeFileDeleteResult[] };
export type DramaEditTask = { id: string; dramaTitle: string; deliveryPreferences: DramaEditDelivery; spoilerPolicy: DramaEditSpoilerPolicy; episodes: DramaEditEpisode[]; proposal: DramaEditProposal; status: DramaEditStatus; stage: string; nextAction: string; productionTaskId?: string; productionStatus?: string; productionProgress?: number; deliveryResourceId?: string; deliveryUrl?: string; productionError?: string; openMontageProductionId?: string; openMontageStage?: string; openMontageCheckpointStatus?: string; openMontageRequiresApproval?: boolean; renderReportResourceId?: string; renderReportUrl?: string; decisions?: DramaEditDecision[]; createdAt: string; updatedAt: string };
export type CreateDramaEditTask = { dramaTitle: string; deliveryPreferences: DramaEditDelivery; spoilerPolicy: DramaEditSpoilerPolicy; episodeAssetIds: string[] };

// These commands synchronously advance the native runner and the compose/publish
// transitions can include FFmpeg work. They must not inherit the UI's short JSON timeout.
const productionControlConfig = (config?: HttpRequestConfig): HttpRequestConfig => ({ timeout: 0, ...config });

export const listDramaEditEpisodeFolders = (config?: HttpRequestConfig) => http.get<{ folders: DramaEditEpisodeFolder[] }>("/drama-editing/episode-folders", config);
export const listDramaEditTasks = (status?: DramaEditStatus, config?: HttpRequestConfig) => http.get<{ tasks: DramaEditTask[] }>("/drama-editing/tasks", { ...config, params: status ? { status } : undefined });
export const getDramaEditTask = (id: string, config?: HttpRequestConfig) => http.get<DramaEditTask>(`/drama-editing/tasks/${encodeURIComponent(id)}`, config);
export const createDramaEditTask = (input: CreateDramaEditTask, config?: HttpRequestConfig) => http.post<DramaEditTask>("/drama-editing/tasks", input, config);
export const deleteDramaEditTasks = (ids: string[], deleteEpisodeFiles = false, config?: HttpRequestConfig) => http.deleteWithBody<{ results: DramaEditDeleteResult[] }>("/drama-editing/tasks", { ids, deleteEpisodeFiles }, productionControlConfig(config));
export const approveDramaEditTask = (id: string, config?: HttpRequestConfig) => http.post<DramaEditTask>(`/drama-editing/tasks/${encodeURIComponent(id)}/approve`, {}, productionControlConfig(config));
export const startDramaEditProduction = (id: string, config?: HttpRequestConfig) => http.post<DramaEditTask>(`/drama-editing/tasks/${encodeURIComponent(id)}/start`, {}, productionControlConfig(config));
export const approveDramaEditOpenMontageStage = (id: string, stage: string, note = "", config?: HttpRequestConfig) => http.post<DramaEditTask>(`/drama-editing/tasks/${encodeURIComponent(id)}/openmontage/stages/${encodeURIComponent(stage)}/approve`, { note }, productionControlConfig(config));
export const requestDramaEditRevision = (id: string, note: string, config?: HttpRequestConfig) => http.post<DramaEditTask>(`/drama-editing/tasks/${encodeURIComponent(id)}/revision`, { note }, config);
export const cancelDramaEditTask = (id: string, note?: string, config?: HttpRequestConfig) => http.post<DramaEditTask>(`/drama-editing/tasks/${encodeURIComponent(id)}/cancel`, { note: note || "" }, config);
