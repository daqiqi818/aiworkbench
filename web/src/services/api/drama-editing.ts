import { http, type HttpRequestConfig } from "@/services/api/request";

export type DramaEditStatus = "awaiting_approval" | "revision_requested" | "approved" | "blocked" | "cancelled";

export type DramaEditSource = { reference: string; rightsConfirmed: boolean; notes?: string };
export type DramaEditEpisode = { episodeId: string; sequenceNumber: number; summary: string; inclusion: "included" | "excluded"; exclusionReason?: string; restrictions?: string[]; sources: DramaEditSource[] };
export type DramaEditDelivery = { aspectRatio: "9:16" | "16:9" | "1:1"; targetDurationSeconds: number; language: string; destinationLabel: string };
export type DramaEditSpoilerPolicy = { policy: "no-major-spoilers" | "limited-spoilers" | "approved-spoilers"; restrictedItems: string[] };
export type DramaEditProposal = { hook: string; emotionalArc: string; episodeCoverage: string[]; excludedEpisodes: string[]; deliveryDefinition: string; costEstimate: string; sourceLimitations: string[] };
export type DramaEditDecision = { id: string; action: string; note?: string; createdAt: string };
export type DramaEditTask = { id: string; dramaTitle: string; campaignObjective: string; targetAudience: string; deliveryPreferences: DramaEditDelivery; spoilerPolicy: DramaEditSpoilerPolicy; episodes: DramaEditEpisode[]; proposal: DramaEditProposal; status: DramaEditStatus; stage: string; nextAction: string; decisions?: DramaEditDecision[]; createdAt: string; updatedAt: string };
export type CreateDramaEditTask = Omit<DramaEditTask, "id" | "proposal" | "status" | "stage" | "nextAction" | "decisions" | "createdAt" | "updatedAt">;

export const listDramaEditTasks = (status?: DramaEditStatus, config?: HttpRequestConfig) => http.get<{ tasks: DramaEditTask[] }>("/drama-editing/tasks", { ...config, params: status ? { status } : undefined });
export const getDramaEditTask = (id: string, config?: HttpRequestConfig) => http.get<DramaEditTask>(`/drama-editing/tasks/${encodeURIComponent(id)}`, config);
export const createDramaEditTask = (input: CreateDramaEditTask, config?: HttpRequestConfig) => http.post<DramaEditTask>("/drama-editing/tasks", input, config);
export const approveDramaEditTask = (id: string, config?: HttpRequestConfig) => http.post<DramaEditTask>(`/drama-editing/tasks/${encodeURIComponent(id)}/approve`, {}, config);
export const requestDramaEditRevision = (id: string, note: string, config?: HttpRequestConfig) => http.post<DramaEditTask>(`/drama-editing/tasks/${encodeURIComponent(id)}/revision`, { note }, config);
export const cancelDramaEditTask = (id: string, note?: string, config?: HttpRequestConfig) => http.post<DramaEditTask>(`/drama-editing/tasks/${encodeURIComponent(id)}/cancel`, { note: note || "" }, config);
