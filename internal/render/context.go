package render

// SourceType represents the source of issues (project or group).
type SourceType string

const (
	// SourceTypeProject indicates issues from a single project.
	SourceTypeProject SourceType = "project"
	// SourceTypeGroup indicates issues from a group (potentially multiple projects).
	SourceTypeGroup SourceType = "group"
)

// Context provides contextual information for rendering issues.
type Context struct {
	Source      SourceType       // "project" or "group"
	ProjectPath string           // For single project, e.g., "namespace/project"
	GroupPath   string           // For group queries, e.g., "namespace/group"
	ProjectMap  map[int64]string // Maps ProjectID -> PathWithNamespace for multi-project scenarios
}

// NewProjectContext creates context for single-project rendering.
func NewProjectContext(projectPath string) *Context {
	return &Context{
		Source:      SourceTypeProject,
		ProjectPath: projectPath,
	}
}

// NewGroupContext creates context for group rendering.
func NewGroupContext(groupPath string, projectMap map[int64]string) *Context {
	return &Context{
		Source:     SourceTypeGroup,
		GroupPath:  groupPath,
		ProjectMap: projectMap,
	}
}
