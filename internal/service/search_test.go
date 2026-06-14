package service

import "testing"

func TestSearch_AcrossTypes(t *testing.T) {
	svc, root := testService(t)
	writeSkill(t, root, "video-summary", "name: video-summary\ndescription: summarize videos\n")
	writePrompt(t, root, "video-prompt", "id: video-prompt\ndescription: prompt for videos\n")

	hits, err := svc.Search("video")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("got %d hits, want 2 (skill + prompt)", len(hits))
	}

	gotKinds := map[string]bool{}
	for _, h := range hits {
		gotKinds[h.Kind] = true
	}
	if !gotKinds["skill"] || !gotKinds["prompt"] {
		t.Errorf("expected hits in skill and prompt; got %v", gotKinds)
	}
}

func TestSearch_EmptyQueryReturnsNothing(t *testing.T) {
	svc, root := testService(t)
	writeSkill(t, root, "alpha", "name: alpha\n")

	hits, err := svc.Search("")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Fatalf("empty query should return no hits, got %d", len(hits))
	}
}
