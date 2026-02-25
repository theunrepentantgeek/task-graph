package cmd

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/theunrepentantgeek/task-graph/internal/config"
)

func TestCreateConfig_DefaultsWhenNoFileProvided(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg).To(Equal(config.New()))
}

func TestCreateConfig_LoadsYAMLConfig(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{
		Config: filepath.Join("testdata", "config.yaml"),
	}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.Graphviz.Font).To(Equal("Fira Code"))
	g.Expect(cfg.Graphviz.FontSize).To(Equal(12))
	g.Expect(cfg.Graphviz.CallEdges.Color).To(Equal("red"))
	g.Expect(cfg.Graphviz.TaskNodes.Color).To(Equal("black"))
}

func TestCreateConfig_LoadsJSONConfig(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{
		Config: filepath.Join("testdata", "config.json"),
	}

	cfg, err := cli.CreateConfig()

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cfg.Graphviz.Font).To(Equal("JetBrains Mono"))
	g.Expect(cfg.Graphviz.DependencyEdges.Color).To(Equal("green"))
	// Ensure defaults remain when not overridden.
	g.Expect(cfg.Graphviz.CallEdges.Style).To(Equal("dashed"))
}

func TestCreateConfig_MissingFileReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cli := CLI{Config: "does-not-exist.yml"}

	cfg, err := cli.CreateConfig()

	g.Expect(cfg).To(BeNil())
	g.Expect(err).To(MatchError(ContainSubstring("failed to read config file")))
}

// TestRenderImage_WithFakeDot_CreatesImageFile tests that renderImage calls dot correctly
// using a fake dot executable that creates the expected output file.
func TestRenderImage_WithFakeDot_CreatesImageFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	dir := t.TempDir()
	dotFile := filepath.Join(dir, "graph.dot")
	imageFile := filepath.Join(dir, "graph.png")

	// Create a fake dot file as input
	err := os.WriteFile(dotFile, []byte("digraph {}"), 0o600)
	g.Expect(err).NotTo(HaveOccurred())

	// Create a fake dot executable that creates the output file.
	// The real dot command is invoked as: dot -T<type> <dotFile> -o <imageFile>
	// so the output file path is the 4th positional argument ($4).
	fakeDot := filepath.Join(dir, "dot")
	fakeDotScript := "#!/bin/sh\ntouch \"$4\"\n"
	err = os.WriteFile(fakeDot, []byte(fakeDotScript), 0o755)
	g.Expect(err).NotTo(HaveOccurred())

	cli := CLI{
		Output:      dotFile,
		RenderImage: "png",
	}
	flags := &Flags{
		Log:    slog.New(slog.NewTextHandler(os.Stderr, nil)),
		Config: &config.Config{DotPath: fakeDot},
	}

	// Act
	err = cli.renderImage(t.Context(), flags)

	// Assert
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(imageFile).To(BeAnExistingFile())
}

// TestRenderImage_DotNotFound_ReturnsError tests that renderImage returns an error
// when the dot executable cannot be found.
func TestRenderImage_DotNotFound_ReturnsError(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	dir := t.TempDir()
	dotFile := filepath.Join(dir, "graph.dot")

	cli := CLI{
		Output:      dotFile,
		RenderImage: "png",
	}
	flags := &Flags{
		Log:    slog.New(slog.NewTextHandler(os.Stderr, nil)),
		Config: &config.Config{DotPath: filepath.Join(dir, "nonexistent-dot")},
	}

	// Act
	err := cli.renderImage(t.Context(), flags)

	// Assert
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError(ContainSubstring("failed to find dot executable")))
}

// TestRenderImage_ImagePathDerivedFromOutput tests that the image file path
// is correctly derived from the output dot file path.
func TestRenderImage_ImagePathDerivedFromOutput(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Arrange
	dir := t.TempDir()
	dotFile := filepath.Join(dir, "mygraph.dot")
	expectedImageFile := filepath.Join(dir, "mygraph.svg")

	err := os.WriteFile(dotFile, []byte("digraph {}"), 0o600)
	g.Expect(err).NotTo(HaveOccurred())

	// Create a fake dot executable that creates the output file.
	// The real dot command is invoked as: dot -T<type> <dotFile> -o <imageFile>
	// so the output file path is the 4th positional argument ($4).
	fakeDot := filepath.Join(dir, "dot")
	fakeDotScript := "#!/bin/sh\ntouch \"$4\"\n"
	err = os.WriteFile(fakeDot, []byte(fakeDotScript), 0o755)
	g.Expect(err).NotTo(HaveOccurred())

	cli := CLI{
		Output:      dotFile,
		RenderImage: "svg",
	}
	flags := &Flags{
		Log:    slog.New(slog.NewTextHandler(os.Stderr, nil)),
		Config: &config.Config{DotPath: fakeDot},
	}

	// Act
	err = cli.renderImage(t.Context(), flags)

	// Assert
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(expectedImageFile).To(BeAnExistingFile())
}
