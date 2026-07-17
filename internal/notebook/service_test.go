package notebook

import "testing"

func TestDefaultNotebookName(t *testing.T) {
	if DefaultNotebookName != "默认" {
		t.Errorf("expected 默认, got %s", DefaultNotebookName)
	}
}

func TestCreateNotebookInput_Defaults(t *testing.T) {
	in := CreateNotebookInput{}
	if in.Name != "" || in.Color != "" {
		t.Errorf("expected empty defaults")
	}
}

func TestNotebook_TableName(t *testing.T) {
	if (Notebook{}).TableName() != "notebooks" {
		t.Errorf("expected table name notebooks")
	}
}
