package prompts

import "testing"

func TestSpinnerStaticFrame(t *testing.T) {
	t.Setenv("COLUMNS", "80")
	t.Setenv("LINES", "24")
	s := &Spinner{Prompt: newPrompt(), message: "Fetching Data..."}
	s.Prompt.newLinesWritten = 1
	// First render includes the leading blank line.
	s.renderFrame()
	got := s.prevFrame
	want := "\n " + cyan("⠂") + " Fetching Data...\n"
	if got != want {
		t.Errorf("spinner frame = %q, want %q", got, want)
	}
}

func TestSpinnerAnimationFrames(t *testing.T) {
	s := &Spinner{Prompt: newPrompt(), message: "Loading"}
	s.prevFrame = "x\n"
	for i := range spinnerFrames {
		s.count = i
		s.renderFrame()
		want := " " + cyan(spinnerFrames[i]) + " Loading\n"
		if got := s.prevFrame; got != want {
			t.Errorf("frame %d = %q, want %q", i, got, want)
		}
	}
}
