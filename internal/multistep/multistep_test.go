package multistep

import (
	"testing"
)

func TestParse_NumericSteps(t *testing.T) {
	content := `# Step 1
First step content.

# Step 2
Second step content.
`
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Steps()) != 2 {
		t.Errorf("expected 2 steps, got %d", len(p.Steps()))
	}

	c1, ok := p.GetStep("1")
	if !ok {
		t.Error("step 1 not found")
	}
	if c1 != "First step content." {
		t.Errorf("step 1 content = %q, want %q", c1, "First step content.")
	}

	c2, ok := p.GetStep("2")
	if !ok {
		t.Error("step 2 not found")
	}
	if c2 != "Second step content." {
		t.Errorf("step 2 content = %q, want %q", c2, "Second step content.")
	}
}

func TestParse_NamedSteps(t *testing.T) {
	content := `# Step Find Tests
Locate all test files in the project.

# Step Migrate Dependencies
Run rinku convert to migrate dependencies.
`
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	steps := p.Steps()
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	if steps[0] != "Find Tests" {
		t.Errorf("first step = %q, want %q", steps[0], "Find Tests")
	}
	if steps[1] != "Migrate Dependencies" {
		t.Errorf("second step = %q, want %q", steps[1], "Migrate Dependencies")
	}

	c, ok := p.GetStep("Find Tests")
	if !ok {
		t.Error("step 'Find Tests' not found")
	}
	if c != "Locate all test files in the project." {
		t.Errorf("content = %q", c)
	}
}

func TestParse_CaseInsensitive(t *testing.T) {
	content := `# step 1
Content for step one.

# STEP 2
Content for step two.
`
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(p.Steps()) != 2 {
		t.Errorf("expected 2 steps, got %d", len(p.Steps()))
	}

	if _, ok := p.GetStep("1"); !ok {
		t.Error("step 1 not found")
	}
	if _, ok := p.GetStep("2"); !ok {
		t.Error("step 2 not found")
	}
}

func TestParse_PreservesOrder(t *testing.T) {
	content := `# Step 3
Third.

# Step 1
First.

# Step 2
Second.
`
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	steps := p.Steps()
	expected := []string{"3", "1", "2"}
	if len(steps) != len(expected) {
		t.Fatalf("expected %d steps, got %d", len(expected), len(steps))
	}
	for i, s := range expected {
		if steps[i] != s {
			t.Errorf("step[%d] = %q, want %q", i, steps[i], s)
		}
	}
}

func TestParse_EmptyContent(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("expected error for empty content")
	}
}

func TestParse_NoSteps(t *testing.T) {
	content := `Just some text without any steps.

More text here.
`
	_, err := Parse(content)
	if err == nil {
		t.Error("expected error for content without steps")
	}
}

func TestGetStep_Missing(t *testing.T) {
	content := `# Step 1
Content.
`
	p, _ := Parse(content)

	_, ok := p.GetStep("nonexistent")
	if ok {
		t.Error("expected false for missing step")
	}
}

func TestFirstStep(t *testing.T) {
	content := `# Step Setup
Setup content.

# Step Build
Build content.
`
	p, _ := Parse(content)

	if p.FirstStep() != "Setup" {
		t.Errorf("FirstStep() = %q, want %q", p.FirstStep(), "Setup")
	}
}

func TestBootstrap(t *testing.T) {
	content := `# Step 1
Do something.
`
	p, _ := Parse(content)

	got := p.Bootstrap("rinku migrate")
	want := "Execute 'rinku migrate 1'. This will return instructions. Execute those instructions."
	if got != want {
		t.Errorf("Bootstrap() = %q, want %q", got, want)
	}
}

func TestBootstrap_NamedStep(t *testing.T) {
	content := `# Step Analyze
Analyze the project.
`
	p, _ := Parse(content)

	got := p.Bootstrap("hashi")
	want := "Execute 'hashi Analyze'. This will return instructions. Execute those instructions."
	if got != want {
		t.Errorf("Bootstrap() = %q, want %q", got, want)
	}
}

func TestSteps_ReturnsCopy(t *testing.T) {
	content := `# Step 1
One.

# Step 2
Two.
`
	p, _ := Parse(content)

	steps := p.Steps()
	steps[0] = "modified"

	// Original should be unchanged
	if p.Steps()[0] != "1" {
		t.Error("Steps() should return a copy")
	}
}

func TestParse_MultilineContent(t *testing.T) {
	content := `# Step 1
Line one.
Line two.
Line three.

# Step 2
Next step.
`
	p, _ := Parse(content)

	c, _ := p.GetStep("1")
	want := "Line one.\nLine two.\nLine three."
	if c != want {
		t.Errorf("content = %q, want %q", c, want)
	}
}

func TestParse_WrongHeaderLevel(t *testing.T) {
	// ## is not a valid step header, only # is
	content := `## Step 1
This should not be parsed as a step.
`
	_, err := Parse(content)
	if err == nil {
		t.Error("expected error: ## should not be recognized as step header")
	}
}

func TestParse_StepsNotStep(t *testing.T) {
	// "Steps" is not the same as "Step"
	content := `# Steps
This is not a valid step header.
`
	_, err := Parse(content)
	if err == nil {
		t.Error("expected error: 'Steps' should not be recognized as 'Step'")
	}
}

func TestParse_StepWithoutID(t *testing.T) {
	// "# Step" without an ID should not be recognized
	content := `# Step
This has no step ID.
`
	_, err := Parse(content)
	if err == nil {
		t.Error("expected error: '# Step' without ID should not be recognized")
	}
}

func TestParse_LeadingWhitespace(t *testing.T) {
	content := `  # Step 1
Content with leading whitespace on header.
`
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Steps()) != 1 {
		t.Errorf("expected 1 step, got %d", len(p.Steps()))
	}
}

func TestParse_MixedValidInvalid(t *testing.T) {
	// Only valid step headers should be parsed
	content := `# Step 1
Valid step.

## Step 2
Invalid - wrong header level.

# Steps 3
Invalid - "Steps" not "Step".

# Step 4
Another valid step.
`
	p, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	steps := p.Steps()
	if len(steps) != 2 {
		t.Errorf("expected 2 valid steps, got %d: %v", len(steps), steps)
	}
	if steps[0] != "1" || steps[1] != "4" {
		t.Errorf("expected steps [1, 4], got %v", steps)
	}
}
