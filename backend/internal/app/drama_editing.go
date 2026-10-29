package app

import "infinite-canvas/backend/internal/dramaedit"

func (s *Service) DramaEditingService() *dramaedit.Service { return s.dramaEditing }
