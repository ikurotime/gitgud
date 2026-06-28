package templates

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitgud/internal/domain"
	"gitgud/internal/interface/web/presenter"
)

type ctxKey int

const (
	flashKey ctxKey = iota
	csrfKey
)

func WithFlash(ctx context.Context, msg string) context.Context {
	return context.WithValue(ctx, flashKey, msg)
}

func flashOf(ctx context.Context) string {
	s, _ := ctx.Value(flashKey).(string)
	return s
}

func WithCSRF(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, csrfKey, token)
}

func csrfToken(ctx context.Context) string {
	s, _ := ctx.Value(csrfKey).(string)
	return s
}

func markdown(s string) string {
	return presenter.RenderMarkdown([]byte(s))
}

func fmtTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

func stateAction(state domain.IssueState) string {
	if state == domain.IssueOpen {
		return "/close"
	}
	return "/reopen"
}

func stateTabClass(current, want string) string {
	if current == want {
		return "font-semibold text-ink"
	}
	return "text-muted hover:text-ink"
}

func cloneInstructions(repo *domain.Repository) string {
	url := "http://localhost:8080/" + repo.OwnerName + "/" + repo.Name + ".git"
	return fmt.Sprintf(`git clone %s
cd %s
echo "# %s" > README.md
git add README.md
git commit -m "first commit"
git push -u origin %s`, url, repo.Name, repo.Name, repo.DefaultBranch)
}

type Crumb struct {
	Name string
	Href string
}

func treeCrumbs(repo *domain.Repository, ref, p string) []Crumb {
	base := "/" + repo.OwnerName + "/" + repo.Name + "/tree/" + ref
	crumbs := []Crumb{{Name: repo.Name, Href: base}}
	if p = strings.Trim(p, "/"); p != "" {
		acc := base
		for _, seg := range strings.Split(p, "/") {
			acc += "/" + seg
			crumbs = append(crumbs, Crumb{Name: seg, Href: acc})
		}
	}
	return crumbs
}

func entryHref(repo *domain.Repository, ref string, e domain.TreeEntry) string {
	kind := "blob"
	if e.IsDir {
		kind = "tree"
	}
	return "/" + repo.OwnerName + "/" + repo.Name + "/" + kind + "/" + ref + "/" + e.Path
}

func repoPath(repo *domain.Repository, sub string) string {
	return "/" + repo.OwnerName + "/" + repo.Name + sub
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func countVisibility(repos []*domain.Repository, private bool) int {
	n := 0
	for _, r := range repos {
		if r.IsPrivate == private {
			n++
		}
	}
	return n
}

func initial(s string) string {
	if s == "" {
		return "?"
	}
	return strings.ToUpper(s[:1])
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func diffLineClass(line string) string {
	switch {
	case strings.HasPrefix(line, "@@"):
		return "diff-hunk"
	case strings.HasPrefix(line, "+"):
		return "diff-add"
	case strings.HasPrefix(line, "-"):
		return "diff-del"
	default:
		return "text-muted"
	}
}

// statBlocks renders a 5-square GitHub-style diffstat: green for additions,
// red for deletions, proportional to the change, padded with neutral squares.
func statBlocks(added, deleted int) []string {
	blocks := make([]string, 5)
	total := added + deleted
	if total == 0 {
		for i := range blocks {
			blocks[i] = "none"
		}
		return blocks
	}
	greens := added * 5 / total
	if added > 0 && greens == 0 {
		greens = 1
	}
	reds := 5 - greens
	if deleted > 0 && reds == 0 && greens > 0 {
		greens--
		reds = 1
	}
	for i := range blocks {
		switch {
		case i < greens:
			blocks[i] = "add"
		case i < greens+reds:
			blocks[i] = "del"
		default:
			blocks[i] = "none"
		}
	}
	return blocks
}

// langLabel maps a file path to a human language label for the blob header.
func langLabel(path string) string {
	ext := strings.ToLower(path)
	if i := strings.LastIndexByte(ext, '.'); i >= 0 {
		ext = ext[i+1:]
	}
	switch ext {
	case "go":
		return "Go"
	case "js", "mjs", "cjs":
		return "JavaScript"
	case "ts", "tsx":
		return "TypeScript"
	case "py":
		return "Python"
	case "rs":
		return "Rust"
	case "rb":
		return "Ruby"
	case "java":
		return "Java"
	case "c", "h":
		return "C"
	case "cpp", "cc", "hpp":
		return "C++"
	case "sh", "bash", "zsh":
		return "Shell"
	case "html", "htm":
		return "HTML"
	case "css":
		return "CSS"
	case "json":
		return "JSON"
	case "yml", "yaml":
		return "YAML"
	case "md", "markdown":
		return "Markdown"
	case "sql":
		return "SQL"
	case "templ":
		return "Templ"
	case "":
		return "Text"
	default:
		return strings.ToUpper(ext)
	}
}

func diffLines(patch string) []string {
	return strings.Split(strings.TrimRight(patch, "\n"), "\n")
}
