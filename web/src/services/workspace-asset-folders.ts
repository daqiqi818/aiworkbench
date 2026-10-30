import { assertUserScope, captureUserScope, type CapturedUserScope } from "@/lib/user-scope-guard";
import { localForageStorageForScope } from "@/lib/localforage-storage";
import {
    createAssetFolder,
    deleteAssetFolder,
    listAssetFolders,
    moveAssetsToFolder,
    putWorkspaceAsset,
    updateAssetFolder,
    type AssetFolder,
} from "@/services/api/workspace-data";
import { usesBrowserLocalResourceStore } from "@/services/workspace-resource-storage";
import { flushAssetStorePersistence, runAssetStoreProjection, useAssetStore } from "@/stores/use-asset-store";

export const WORKSPACE_ASSET_FOLDERS_KEY = "infinite-canvas:asset-folders";

/**
 * Asset folders always use the local backend identity. Browser-local media remains in IndexedDB,
 * but its metadata has to be visible to owner-scoped workflows such as drama editing.
 */
export function usesWorkspaceAssetFolderApi() {
    return true;
}

function parseLocalFolders(raw: string | null): AssetFolder[] {
    if (!raw) return [];
    try {
        const parsed = JSON.parse(raw) as unknown;
        if (!Array.isArray(parsed)) return [];
        return parsed.filter((folder): folder is AssetFolder => Boolean(folder && typeof folder === "object" && typeof (folder as AssetFolder).id === "string" && typeof (folder as AssetFolder).name === "string"));
    } catch {
        return [];
    }
}

async function readLocalFolders(userScope: string) {
    return parseLocalFolders(await localForageStorageForScope(userScope).getItem(WORKSPACE_ASSET_FOLDERS_KEY));
}

async function writeLocalFolders(userScope: string, folders: AssetFolder[]) {
    await localForageStorageForScope(userScope).setItem(WORKSPACE_ASSET_FOLDERS_KEY, JSON.stringify(folders));
    return folders;
}

export async function listWorkspaceAssetFolders(expectedScope?: CapturedUserScope) {
    const expected = expectedScope ?? captureUserScope();
    assertUserScope(expected);
    const folders = (await listAssetFolders({ expectedScope: expected })).folders;
    assertUserScope(expected);
    return folders;
}

export async function createWorkspaceAssetFolder(name: string, expectedScope?: CapturedUserScope) {
    const expected = expectedScope ?? captureUserScope();
    assertUserScope(expected);
    const trimmed = name.trim();
    if (!trimmed) throw new Error("请输入分类名称");
    return (await createAssetFolder(trimmed, { expectedScope: expected })).folder;
}

export async function renameWorkspaceAssetFolder(id: string, name: string, expectedScope?: CapturedUserScope) {
    const expected = expectedScope ?? captureUserScope();
    assertUserScope(expected);
    const folderId = id.trim();
    const trimmed = name.trim();
    if (!folderId) throw new Error("素材分类 ID 不能为空");
    if (!trimmed) throw new Error("请输入分类名称");
    return (await updateAssetFolder(folderId, trimmed, { expectedScope: expected })).folder;
}

export async function deleteWorkspaceAssetFolder(id: string, expectedScope?: CapturedUserScope) {
    const expected = expectedScope ?? captureUserScope();
    assertUserScope(expected);
    const folderId = id.trim();
    if (!folderId) throw new Error("素材分类 ID 不能为空");
    await deleteAssetFolder(folderId, { expectedScope: expected });
    assertUserScope(expected);
    runAssetStoreProjection(() => {
        const store = useAssetStore.getState();
        for (const asset of store.assets) {
            if (asset.folderId === folderId) store.updateAsset(asset.id, { folderId: undefined });
        }
    });
}

export async function assignWorkspaceAssetsFolder(assetIds: string[], folderId = "", expectedScope?: CapturedUserScope) {
    const expected = expectedScope ?? captureUserScope();
    assertUserScope(expected);
    const ids = assetIds.map((id) => id.trim()).filter(Boolean);
    if (!ids.length) return;
    const nextFolderId = folderId.trim();
    if (usesBrowserLocalResourceStore()) {
        await migrateBrowserLocalAssetFolders(expected);
        assertUserScope(expected);
        for (const id of ids) {
            const asset = useAssetStore.getState().assets.find((item) => item.id === id);
            if (!asset || asset.kind === "entity") continue;
            // Browser-local files already live in IndexedDB; mirror only their metadata before
            // asking the backend to change the folder relation.
            await putWorkspaceAsset(id, asset, { expectedScope: expected });
            assertUserScope(expected);
        }
    }
    await moveAssetsToFolder(ids, nextFolderId, { expectedScope: expected });
    assertUserScope(expected);
    runAssetStoreProjection(() => {
        const store = useAssetStore.getState();
        for (const id of ids) store.updateAsset(id, { folderId: nextFolderId || undefined });
    });
    await flushAssetStorePersistence(expected);
}

function folderNameKey(name: string) {
    return name.trim().toLocaleLowerCase();
}

/**
 * Older browser-local sessions created folder IDs only in IndexedDB. Migrate those
 * IDs to backend folder IDs by name, then rebind in-memory assets before their
 * metadata is mirrored. Media blobs are intentionally not read or uploaded here.
 */
export async function migrateBrowserLocalAssetFolders(expectedScope?: CapturedUserScope) {
    const expected = expectedScope ?? captureUserScope();
    assertUserScope(expected);
    if (!usesBrowserLocalResourceStore()) return new Map<string, string>();

    const legacyFolders = await readLocalFolders(expected.userScope);
    assertUserScope(expected);
    if (!legacyFolders.length) return new Map<string, string>();

    const remote = (await listAssetFolders({ expectedScope: expected })).folders;
    assertUserScope(expected);
    const byName = new Map(remote.map((folder) => [folderNameKey(folder.name), folder]));
    const rebound = new Map<string, string>();
    const canonical: AssetFolder[] = [];

    for (const legacy of legacyFolders) {
        const key = folderNameKey(legacy.name);
        if (!key) continue;
        let folder = byName.get(key);
        if (!folder) {
            folder = (await createAssetFolder(legacy.name, { expectedScope: expected })).folder;
            assertUserScope(expected);
            byName.set(key, folder);
        }
        rebound.set(legacy.id, folder.id);
        canonical.push(folder);
    }

    await writeLocalFolders(expected.userScope, canonical);
    assertUserScope(expected);
    if (rebound.size) {
        runAssetStoreProjection(() => {
            const store = useAssetStore.getState();
            for (const asset of store.assets) {
                const folderId = asset.folderId ? rebound.get(asset.folderId) : undefined;
                if (folderId && folderId !== asset.folderId) store.updateAsset(asset.id, { folderId });
            }
        });
        await flushAssetStorePersistence(expected);
    }
    return rebound;
}
