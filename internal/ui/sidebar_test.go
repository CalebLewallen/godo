package ui

import (
	"testing"
	"github.com/CalebLewallen/godo/internal/model"
)

func BenchmarkGrandparentID(b *testing.B) {
	// Create a model with many folders
	numFolders := 1000
	m := NewSidebarModel()
	folders := make([]model.Folder, numFolders)
	for i := 0; i < numFolders; i++ {
		folders[i] = model.Folder{
			ID: int64(i + 1),
			Name: "Folder",
		}
		if i > 0 {
			pid := int64(i)
			folders[i].ParentID = &pid
		}
	}
	// Init m
	m.folders = folders
	m.folderMap = make(map[int64]*model.Folder, len(m.folders))
	for i := range m.folders {
		m.folderMap[m.folders[i].ID] = &m.folders[i]
	}
	// Let's also say there's a folderMap if we were testing the optimized version

	// The target folder to check grandparent for
	pid := int64(numFolders - 1)
	f := &model.Folder{
		ID: int64(numFolders + 1),
		ParentID: &pid,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.grandparentID(f)
	}
}
