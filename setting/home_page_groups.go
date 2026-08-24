package setting

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
)

type HomePageGroupRatioItem struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Ratio       float64 `json:"ratio"`
}

func ParseHomePageDisplayedGroups(jsonStr string) []string {
	if strings.TrimSpace(jsonStr) == "" {
		return nil
	}
	var names []string
	if err := common.UnmarshalJsonStr(jsonStr, &names); err != nil {
		return nil
	}
	out := make([]string, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func BuildHomePageGroupRatios(displayedJSON string) []HomePageGroupRatioItem {
	names := ParseHomePageDisplayedGroups(displayedJSON)
	if len(names) == 0 {
		return []HomePageGroupRatioItem{}
	}
	ratios := ratio_setting.GetGroupRatioCopy()
	usable := GetUserUsableGroupsCopy()
	items := make([]HomePageGroupRatioItem, 0, len(names))
	for _, name := range names {
		ratio, ok := ratios[name]
		if !ok {
			continue
		}
		desc := usable[name]
		if desc == "" {
			desc = name
		}
		items = append(items, HomePageGroupRatioItem{
			Name:        name,
			Description: desc,
			Ratio:       ratio,
		})
	}
	return items
}
