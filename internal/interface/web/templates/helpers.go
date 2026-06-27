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
		return "font-semibold"
	}
	return "text-gray-600"
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

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func diffLineClass(line string) string {
	switch {
	case strings.HasPrefix(line, "+"):
		return "bg-green-100"
	case strings.HasPrefix(line, "-"):
		return "bg-red-100"
	default:
		return ""
	}
}

func diffLines(patch string) []string {
	return strings.Split(strings.TrimRight(patch, "\n"), "\n")
}
