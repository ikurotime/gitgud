package git

import (
	"net/http"
	"net/http/cgi"
	"os/exec"
	"path/filepath"
	"strings"
)

type Backend struct {
	reposDir    string
	httpBackend string
}

func NewBackend(reposDir string) (*Backend, error) {
	out, err := exec.Command("git", "--exec-path").Output()
	if err != nil {
		return nil, err
	}
	bin := filepath.Join(strings.TrimSpace(string(out)), "git-http-backend")
	return &Backend{reposDir: reposDir, httpBackend: bin}, nil
}

func (b *Backend) Handler(remoteUser string) http.Handler {
	return &cgi.Handler{
		Path: b.httpBackend,
		Env: []string{
			"GIT_PROJECT_ROOT=" + b.reposDir,
			"GIT_HTTP_EXPORT_ALL=1",
			"REMOTE_USER=" + remoteUser,
		},
	}
}
