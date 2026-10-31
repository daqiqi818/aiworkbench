package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"infinite-canvas/backend/internal/dramaedit"
	"infinite-canvas/backend/internal/kernel"
)

// ImportOpenMontageDelivery streams only final artifacts back into BeefTV. Source
// resources are never copied into, or registered by, the OpenMontage workspace.
func (s *Service) ImportOpenMontageDelivery(userID, folderID, title, productionID string, runtime dramaedit.NativeRuntime) (string, string, error) {
	video, err := s.importOpenMontageArtifact(userID, folderID, title+" 成片", productionID, "video", "video", runtime)
	if err != nil {
		return "", "", err
	}
	report, err := s.importOpenMontageArtifact(userID, folderID, title+" 渲染报告", productionID, "report", "text", runtime)
	if err != nil {
		return "", "", err
	}
	return video, report, nil
}

func (s *Service) importOpenMontageArtifact(userID, folderID, title, productionID, artifact, kind string, runtime dramaedit.NativeRuntime) (string, error) {
	body, err := runtime.Artifact(userID, productionID, artifact)
	if err != nil {
		return "", err
	}
	defer body.Close()
	temporary, err := os.CreateTemp("", "beeftv-openmontage-*")
	if err != nil {
		return "", err
	}
	path := temporary.Name()
	defer os.Remove(path)
	size, err := io.Copy(temporary, body)
	if err != nil {
		temporary.Close()
		return "", err
	}
	if _, err := temporary.Seek(0, io.SeekStart); err != nil {
		temporary.Close()
		return "", err
	}
	name := "openmontage-final.mp4"
	if artifact == "report" {
		name = "openmontage-render-report.json"
	}
	resource, err := s.UploadLocalResourceFile(userID, name, size, kind, 0, 0, 0, temporary, "openmontage:"+productionID+":"+artifact)
	temporary.Close()
	if err != nil {
		return "", err
	}
	if artifact == "report" {
		return resource.ID, nil
	}
	payload, _ := json.Marshal(map[string]any{"id": kernel.NewID(), "folderId": folderID, "kind": kind, "category": "material", "status": "confirmed", "title": title, "coverUrl": "", "tags": []string{"OpenMontage"}, "data": map[string]any{"storageKey": "resource:" + resource.ID, "url": "", "width": 0, "height": 0, "bytes": resource.Size, "mimeType": resource.MimeType, "source": "OpenMontage", "productionId": productionID}, "createdAt": time.Now().UTC().Format(time.RFC3339), "updatedAt": time.Now().UTC().Format(time.RFC3339)})
	if _, err := s.UpsertUserAsset(userID, payload); err != nil {
		return "", fmt.Errorf("register BeefTV asset: %w", err)
	}
	return resource.ID, nil
}
