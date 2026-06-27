package templates

import (
	"fmt"

	"gitgud/internal/domain"
)

func cloneInstructions(repo *domain.Repository) string {
	url := "http://localhost:8080/" + repo.OwnerName + "/" + repo.Name + ".git"
	return fmt.Sprintf(`git clone %s
cd %s
echo "# %s" > README.md
git add README.md
git commit -m "first commit"
git push -u origin %s`, url, repo.Name, repo.Name, repo.DefaultBranch)
}
