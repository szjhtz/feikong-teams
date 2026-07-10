// Package skills 负责组装技能市场适配器。
package skills

import (
	"net/http"
	"time"

	skillhub "fkteams/internal/adapters/skill/skillhub"
	appskill "fkteams/internal/app/skill"
)

const defaultSkillHubURL = "https://lightmake.site/api/skills"

func NewDefaultProviderRegistry() *appskill.ProviderRegistry {
	client := &http.Client{Timeout: 120 * time.Second}
	return appskill.NewProviderRegistry(skillhub.New(defaultSkillHubURL, client))
}
