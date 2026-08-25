package controller

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/QuantumNous/new-api/model"
	perfmetrics "github.com/QuantumNous/new-api/pkg/perf_metrics"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/console_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

func GetPerfMetricsGroups(c *gin.Context) {
	hours := 24
	if rawHours := c.Query("hours"); rawHours != "" {
		if parsed, err := strconv.Atoi(rawHours); err == nil {
			hours = parsed
		}
	}

	userId := c.GetInt("id")
	userGroup, _ := model.GetUserGroup(userId, false)
	userUsableGroups := service.GetUserUsableGroups(userGroup)
	ratioGroups := ratio_setting.GetGroupRatioCopy()
	configured := console_setting.GetGroupAvailabilityGroups()

	groupNames := make([]string, 0, len(configured))
	descriptions := make(map[string]string, len(configured))
	order := make(map[string]int, len(configured))
	for _, groupName := range configured {
		if groupName != "auto" {
			if _, ok := ratioGroups[groupName]; !ok {
				continue
			}
		}
		order[groupName] = len(groupNames)
		groupNames = append(groupNames, groupName)
		descriptions[groupName] = userUsableGroups[groupName]
	}

	result, err := perfmetrics.QueryGroupAvailability(hours, groupNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	for i := range result.Groups {
		result.Groups[i].Description = descriptions[result.Groups[i].Group]
	}
	sort.SliceStable(result.Groups, func(i, j int) bool {
		return order[result.Groups[i].Group] < order[result.Groups[j].Group]
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

func GetPerfMetricsSummary(c *gin.Context) {
	hours := 24
	if rawHours := c.Query("hours"); rawHours != "" {
		if parsed, err := strconv.Atoi(rawHours); err == nil {
			hours = parsed
		}
	}

	activeGroups := append(lo.Keys(ratio_setting.GetGroupRatioCopy()), "auto")
	result, err := perfmetrics.QuerySummaryAll(hours, activeGroups)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

func GetPerfMetrics(c *gin.Context) {
	modelName := c.Query("model")
	if modelName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "model is required",
		})
		return
	}

	hours := 24
	if rawHours := c.Query("hours"); rawHours != "" {
		if parsed, err := strconv.Atoi(rawHours); err == nil {
			hours = parsed
		}
	}

	result, err := perfmetrics.Query(perfmetrics.QueryParams{
		Model: modelName,
		Group: c.Query("group"),
		Hours: hours,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	result.Groups = filterActiveGroups(result.Groups)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

func filterActiveGroups(groups []perfmetrics.GroupResult) []perfmetrics.GroupResult {
	activeRatios := ratio_setting.GetGroupRatioCopy()
	return lo.Filter(groups, func(g perfmetrics.GroupResult, _ int) bool {
		_, ok := activeRatios[g.Group]
		return ok || g.Group == "auto"
	})
}
