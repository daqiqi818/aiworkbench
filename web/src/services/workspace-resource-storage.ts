import { isLocalRuntimeMode, isNativeDesktopRuntime } from "@/lib/runtime-mode";
import { resourceIdFromStorageKey } from "@/services/api/resources";

/**
 * Resource storage policy for the two local runtimes:
 * - native desktop: Go resource service is the durable local store;
 * - browser local: IndexedDB is the durable local store.
 * Hosted mode always attempts the remote resource API first.
 */
export function usesBrowserLocalResourceStore() {
    return isLocalRuntimeMode() && !isNativeDesktopRuntime();
}

export function usesNativeLocalResourceStore() {
    return isLocalRuntimeMode() && isNativeDesktopRuntime();
}

const WORKSPACE_MEDIA_KINDS = new Set(["image", "video", "audio", "model"]);

export function isWorkspaceMediaAssetKind(kind: string) {
    return WORKSPACE_MEDIA_KINDS.has(kind);
}

/**
 * Video used by the short-drama workbench is always a BeefTV-owned source.
 * Browser-local storage can retain a recoverable copy, but cannot be the only
 * durable location because the backend and OpenMontage cannot read a Blob URL.
 */
export function isCanonicalWorkspaceMediaPersistSource(input: { kind?: string; storageKey?: string; pendingRemoteUpload?: boolean }) {
    if (input.kind && !isWorkspaceMediaAssetKind(input.kind)) return true;
    if (input.pendingRemoteUpload) return false;
    if (input.kind === "video") return Boolean(resourceIdFromStorageKey(input.storageKey));
    if (usesBrowserLocalResourceStore()) return Boolean(input.storageKey?.trim());
    return Boolean(resourceIdFromStorageKey(input.storageKey));
}

export function workspaceAssetMediaStorageKey(asset: { kind: string; data: object }) {
    const storageKey = "storageKey" in asset.data ? asset.data.storageKey : undefined;
    return typeof storageKey === "string" ? storageKey : undefined;
}

export function workspaceAssetHasCanonicalMediaPersist(asset: { kind: string; data: object; pendingRemoteUpload?: boolean }) {
    return isCanonicalWorkspaceMediaPersistSource({
        kind: asset.kind,
        storageKey: workspaceAssetMediaStorageKey(asset),
        pendingRemoteUpload: asset.pendingRemoteUpload,
    });
}
