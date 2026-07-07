package skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"fkteams/internal/app/appdata"
	"fkteams/internal/runtime/atomicfile"

	"github.com/goccy/go-yaml"
)

var localSkillSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

// LocalSkillInfo 表示本地已安装技能。
type LocalSkillInfo struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// LocalSkillSpec 表示本地技能创建请求。
type LocalSkillSpec struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

// ListLocalSkills 列出本地已安装技能。
func ListLocalSkills() ([]LocalSkillInfo, error) {
	skillsDir := appdata.SkillsDir()

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read skills dir: %w", err)
	}

	var skills []LocalSkillInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillFile := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		data, err := os.ReadFile(skillFile)
		if err != nil {
			continue
		}

		content := string(data)
		parts := strings.SplitN(content, "---", 3)
		if len(parts) < 3 {
			continue
		}

		var info struct {
			Name        string `yaml:"name"`
			Description string `yaml:"description"`
		}
		if err := yaml.Unmarshal([]byte(parts[1]), &info); err != nil {
			continue
		}

		name := info.Name
		if name == "" {
			name = entry.Name()
		}

		skills = append(skills, LocalSkillInfo{
			Slug:        entry.Name(),
			Name:        name,
			Description: strings.Join(strings.Fields(info.Description), " "),
		})
	}

	return skills, nil
}

// SkillFileEntry 表示技能目录中的文件。
type SkillFileEntry struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

// ListSkillFiles 列出技能目录下指定路径的文件。
func ListSkillFiles(slug, subPath string) ([]SkillFileEntry, error) {
	targetDir, cleanSub, err := resolveSkillPath(slug, subPath, true)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("skill %s is not installed", slug)
		}
		return nil, err
	}

	type sortableSkillFileEntry struct {
		entry       SkillFileEntry
		modUnixNano int64
	}
	var sortableEntries []sortableSkillFileEntry
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		sortableEntries = append(sortableEntries, sortableSkillFileEntry{
			entry: SkillFileEntry{
				Name:  e.Name(),
				Path:  filepath.ToSlash(filepath.Join(cleanSub, e.Name())),
				IsDir: e.IsDir(),
				Size:  info.Size(),
			},
			modUnixNano: info.ModTime().UnixNano(),
		})
	}
	sort.SliceStable(sortableEntries, func(i, j int) bool {
		left := sortableEntries[i]
		right := sortableEntries[j]
		if left.entry.IsDir != right.entry.IsDir {
			return left.entry.IsDir
		}
		if left.modUnixNano != right.modUnixNano {
			return left.modUnixNano > right.modUnixNano
		}
		if left.entry.Size != right.entry.Size {
			return left.entry.Size > right.entry.Size
		}
		return left.entry.Name < right.entry.Name
	})

	result := make([]SkillFileEntry, 0, len(sortableEntries))
	for _, item := range sortableEntries {
		result = append(result, item.entry)
	}
	return result, nil
}

// ReadSkillFile 读取技能目录中的文件。
func ReadSkillFile(slug, filePath string) (string, error) {
	cleanPath, _, err := resolveSkillPath(slug, filePath, false)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CreateLocalSkill 创建用户自定义本地技能。
func CreateLocalSkill(spec LocalSkillSpec) (LocalSkillInfo, error) {
	spec.Slug = strings.TrimSpace(spec.Slug)
	spec.Name = strings.TrimSpace(spec.Name)
	spec.Description = strings.TrimSpace(spec.Description)
	if err := validateSkillSlug(spec.Slug); err != nil {
		return LocalSkillInfo{}, err
	}
	if spec.Name == "" {
		spec.Name = spec.Slug
	}

	skillDir := filepath.Join(appdata.SkillsDir(), spec.Slug)
	if _, err := os.Stat(skillDir); err == nil {
		return LocalSkillInfo{}, fmt.Errorf("skill already exists")
	} else if err != nil && !os.IsNotExist(err) {
		return LocalSkillInfo{}, fmt.Errorf("stat skill dir: %w", err)
	}
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return LocalSkillInfo{}, fmt.Errorf("create skill dir: %w", err)
	}
	content := strings.TrimSpace(spec.Content)
	if content == "" {
		content = defaultSkillContent(spec.Name, spec.Description)
	}
	if err := atomicfile.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content+"\n"), 0644); err != nil {
		return LocalSkillInfo{}, err
	}
	return LocalSkillInfo{Slug: spec.Slug, Name: spec.Name, Description: spec.Description}, nil
}

// SaveSkillFile 保存技能文件内容。
func SaveSkillFile(slug, filePath, content string) error {
	targetPath, _, err := resolveSkillPath(slug, filePath, false)
	if err != nil {
		return err
	}
	info, err := os.Stat(targetPath)
	if err == nil && info.IsDir() {
		return fmt.Errorf("path is a directory")
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat skill file: %w", err)
	}
	return atomicfile.WriteFile(targetPath, []byte(content), 0644)
}

// CreateSkillFile 创建技能文件或目录。
func CreateSkillFile(slug, filePath, content string, isDir bool) error {
	targetPath, _, err := resolveSkillPath(slug, filePath, false)
	if err != nil {
		return err
	}
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("path already exists")
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat skill path: %w", err)
	}
	if isDir {
		return os.MkdirAll(targetPath, 0755)
	}
	return atomicfile.WriteFile(targetPath, []byte(content), 0644)
}

// DeleteSkillFile 删除技能中的文件或目录。
func DeleteSkillFile(slug, filePath string) error {
	targetPath, cleanSub, err := resolveSkillPath(slug, filePath, false)
	if err != nil {
		return err
	}
	if cleanSub == "SKILL.md" {
		return fmt.Errorf("SKILL.md cannot be deleted")
	}
	if _, err := os.Stat(targetPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path not found")
		}
		return fmt.Errorf("stat skill path: %w", err)
	}
	return os.RemoveAll(targetPath)
}

// InstallSkillFromProvider 从指定 provider 安装技能。
func InstallSkillFromProvider(ctx context.Context, slug, version string, provider Provider) error {
	return installSkill(ctx, slug, version, provider)
}

// RemoveLocalSkill 删除已安装技能。
func RemoveLocalSkill(slug string) error {
	if err := validateSkillSlug(slug); err != nil {
		return err
	}
	targetDir := filepath.Join(appdata.SkillsDir(), slug)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("skill %s is not installed", slug)
	}
	return os.RemoveAll(targetDir)
}

func validateSkillSlug(slug string) error {
	if !localSkillSlugPattern.MatchString(slug) {
		return fmt.Errorf("invalid skill slug")
	}
	return nil
}

func resolveSkillPath(slug, subPath string, allowRoot bool) (string, string, error) {
	if err := validateSkillSlug(slug); err != nil {
		return "", "", err
	}
	skillsDir := filepath.Join(appdata.SkillsDir(), slug)
	cleanSub := filepath.Clean(filepath.ToSlash(strings.TrimSpace(subPath)))
	if cleanSub == "." {
		cleanSub = ""
	}
	if cleanSub == "" && !allowRoot {
		return "", "", fmt.Errorf("path is required")
	}
	if strings.HasPrefix(cleanSub, "..") || strings.HasPrefix(cleanSub, "/") || filepath.IsAbs(cleanSub) {
		return "", "", fmt.Errorf("invalid path")
	}

	targetPath := filepath.Join(skillsDir, filepath.FromSlash(cleanSub))
	cleanTarget := filepath.Clean(targetPath)
	cleanRoot := filepath.Clean(skillsDir)
	if cleanTarget != cleanRoot && !strings.HasPrefix(cleanTarget, cleanRoot+string(os.PathSeparator)) {
		return "", "", fmt.Errorf("invalid path")
	}
	return cleanTarget, cleanSub, nil
}

func defaultSkillContent(name, description string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "---\nname: %s\ndescription: %s\n---\n\n", yamlQuote(name), yamlQuote(description))
	fmt.Fprintf(&sb, "# %s\n\n", name)
	if description != "" {
		fmt.Fprintf(&sb, "%s\n\n", description)
	}
	sb.WriteString("## Use when\n\n- Describe when this skill should be used.\n\n")
	sb.WriteString("## Instructions\n\n- Add the reusable workflow, constraints, and examples here.\n")
	return sb.String()
}

func yamlQuote(value string) string {
	data, err := yaml.Marshal(value)
	if err != nil {
		return `""`
	}
	return strings.TrimSpace(string(data))
}
