package analysis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// PerformanceAnalyzer 性能分析器
type PerformanceAnalyzer struct {
	mu sync.RWMutex

	// 性能数据收集器
	collectors map[string]PerformanceCollector

	// 分析规则
	rules []AnalysisRule

	// 分析结果
	results []AnalysisResult

	logger log.Logger
}

// PerformanceCollector 性能数据收集器接口
type PerformanceCollector interface {
	Collect(ctx context.Context) (*PerformanceData, error)
	GetName() string
}

// PerformanceData 性能数据
type PerformanceData struct {
	Timestamp time.Time
	Collector string
	Metrics   map[string]float64
	RawData   interface{}
}

// AnalysisRule 分析规则
type AnalysisRule struct {
	Name        string
	Description string
	Threshold   float64
	Severity    Severity
	Condition   func(*PerformanceData) bool
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	Timestamp   time.Time
	Rule        string
	Severity    Severity
	Message     string
	Data        *PerformanceData
	Suggestions []string
}

// Severity 严重程度
type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// NewPerformanceAnalyzer 创建新的性能分析器
func NewPerformanceAnalyzer(logger log.Logger) *PerformanceAnalyzer {
	analyzer := &PerformanceAnalyzer{
		collectors: make(map[string]PerformanceCollector),
		rules:      make([]AnalysisRule, 0),
		results:    make([]AnalysisResult, 0),
		logger:     logger,
	}

	// 添加默认分析规则
	analyzer.addDefaultRules()

	return analyzer
}

// AddCollector 添加性能数据收集器
func (pa *PerformanceAnalyzer) AddCollector(collector PerformanceCollector) error {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	name := collector.GetName()
	if _, exists := pa.collectors[name]; exists {
		return fmt.Errorf("collector %s already exists", name)
	}

	pa.collectors[name] = collector
	pa.logger.Log(log.LevelInfo, "Added performance collector", "name", name)

	return nil
}

// AddRule 添加分析规则
func (pa *PerformanceAnalyzer) AddRule(rule AnalysisRule) error {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.rules = append(pa.rules, rule)
	pa.logger.Log(log.LevelInfo, "Added analysis rule", "name", rule.Name)

	return nil
}

// Analyze 执行性能分析
func (pa *PerformanceAnalyzer) Analyze(ctx context.Context) ([]AnalysisResult, error) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	var allResults []AnalysisResult

	// 收集所有性能数据
	for name, collector := range pa.collectors {
		data, err := collector.Collect(ctx)
		if err != nil {
			pa.logger.Log(log.LevelError, "Failed to collect data from collector", "name", name, "error", err)
			continue
		}

		// 应用分析规则
		results := pa.applyRules(data)
		allResults = append(allResults, results...)
	}

	// 保存分析结果
	pa.results = append(pa.results, allResults...)

	// 限制结果数量
	if len(pa.results) > 1000 {
		pa.results = pa.results[len(pa.results)-1000:]
	}

	pa.logger.Log(log.LevelInfo, "Performance analysis completed", "results", len(allResults))

	return allResults, nil
}

// applyRules 应用分析规则
func (pa *PerformanceAnalyzer) applyRules(data *PerformanceData) []AnalysisResult {
	var results []AnalysisResult

	for _, rule := range pa.rules {
		if rule.Condition(data) {
			result := AnalysisResult{
				Timestamp:   time.Now(),
				Rule:        rule.Name,
				Severity:    rule.Severity,
				Message:     rule.Description,
				Data:        data,
				Suggestions: pa.generateSuggestions(rule, data),
			}

			results = append(results, result)
		}
	}

	return results
}

// generateSuggestions 生成优化建议
func (pa *PerformanceAnalyzer) generateSuggestions(rule AnalysisRule, data *PerformanceData) []string {
	var suggestions []string

	switch rule.Name {
	case "cache_hit_rate_low":
		suggestions = append(suggestions, "增加缓存大小", "优化缓存策略", "检查缓存键设计")
	case "response_time_high":
		suggestions = append(suggestions, "优化算法实现", "增加缓存", "检查依赖服务性能")
	case "error_rate_high":
		suggestions = append(suggestions, "检查错误日志", "优化错误处理", "增加重试机制")
	case "memory_usage_high":
		suggestions = append(suggestions, "检查内存泄漏", "优化数据结构", "增加垃圾回收")
	default:
		suggestions = append(suggestions, "检查系统配置", "优化代码实现", "增加监控指标")
	}

	return suggestions
}

// addDefaultRules 添加默认分析规则
func (pa *PerformanceAnalyzer) addDefaultRules() {
	// 缓存命中率规则
	pa.AddRule(AnalysisRule{
		Name:        "cache_hit_rate_low",
		Description: "缓存命中率低于阈值",
		Threshold:   0.8,
		Severity:    SeverityMedium,
		Condition: func(data *PerformanceData) bool {
			if hitRate, exists := data.Metrics["cache_hit_rate"]; exists {
				return hitRate < 0.8
			}
			return false
		},
	})

	// 响应时间规则
	pa.AddRule(AnalysisRule{
		Name:        "response_time_high",
		Description: "响应时间过高",
		Threshold:   100, // 毫秒
		Severity:    SeverityHigh,
		Condition: func(data *PerformanceData) bool {
			if responseTime, exists := data.Metrics["response_time"]; exists {
				return responseTime > 100
			}
			return false
		},
	})

	// 错误率规则
	pa.AddRule(AnalysisRule{
		Name:        "error_rate_high",
		Description: "错误率过高",
		Threshold:   0.05, // 5%
		Severity:    SeverityCritical,
		Condition: func(data *PerformanceData) bool {
			if errorRate, exists := data.Metrics["error_rate"]; exists {
				return errorRate > 0.05
			}
			return false
		},
	})
}

// GetResults 获取分析结果
func (pa *PerformanceAnalyzer) GetResults(ctx context.Context, severity Severity) []AnalysisResult {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	var filteredResults []AnalysisResult

	for _, result := range pa.results {
		if result.Severity >= severity {
			filteredResults = append(filteredResults, result)
		}
	}

	return filteredResults
}

// GenerateReport 生成性能分析报告
func (pa *PerformanceAnalyzer) GenerateReport(ctx context.Context) (*PerformanceReport, error) {
	results, err := pa.Analyze(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze performance: %w", err)
	}

	report := &PerformanceReport{
		Timestamp:      time.Now(),
		Summary:        pa.generateSummary(results),
		CriticalIssues: pa.filterResultsBySeverity(results, SeverityCritical),
		HighIssues:     pa.filterResultsBySeverity(results, SeverityHigh),
		MediumIssues:   pa.filterResultsBySeverity(results, SeverityMedium),
		LowIssues:      pa.filterResultsBySeverity(results, SeverityLow),
		Suggestions:    pa.generateOverallSuggestions(results),
	}

	return report, nil
}

// PerformanceReport 性能分析报告
type PerformanceReport struct {
	Timestamp      time.Time
	Summary        string
	CriticalIssues []AnalysisResult
	HighIssues     []AnalysisResult
	MediumIssues   []AnalysisResult
	LowIssues      []AnalysisResult
	Suggestions    []string
}

// generateSummary 生成摘要
func (pa *PerformanceAnalyzer) generateSummary(results []AnalysisResult) string {
	total := len(results)
	critical := len(pa.filterResultsBySeverity(results, SeverityCritical))
	high := len(pa.filterResultsBySeverity(results, SeverityHigh))

	if critical > 0 {
		return fmt.Sprintf("发现 %d 个严重问题，%d 个高级问题，建议立即处理", critical, high)
	} else if high > 0 {
		return fmt.Sprintf("发现 %d 个高级问题，建议优先处理", high)
	} else {
		return fmt.Sprintf("系统性能正常，共分析 %d 个指标", total)
	}
}

// filterResultsBySeverity 按严重程度过滤结果
func (pa *PerformanceAnalyzer) filterResultsBySeverity(results []AnalysisResult, severity Severity) []AnalysisResult {
	var filtered []AnalysisResult
	for _, result := range results {
		if result.Severity == severity {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// generateOverallSuggestions 生成整体优化建议
func (pa *PerformanceAnalyzer) generateOverallSuggestions(results []AnalysisResult) []string {
	var suggestions []string

	// 根据问题类型生成建议
	hasCacheIssues := false
	hasPerformanceIssues := false
	hasErrorIssues := false

	for _, result := range results {
		switch result.Rule {
		case "cache_hit_rate_low":
			hasCacheIssues = true
		case "response_time_high":
			hasPerformanceIssues = true
		case "error_rate_high":
			hasErrorIssues = true
		}
	}

	if hasCacheIssues {
		suggestions = append(suggestions, "优化缓存策略和配置")
	}

	if hasPerformanceIssues {
		suggestions = append(suggestions, "优化算法实现和数据结构")
	}

	if hasErrorIssues {
		suggestions = append(suggestions, "加强错误处理和监控")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "系统运行良好，继续保持")
	}

	return suggestions
}

// GetCollectors 获取所有收集器
func (pa *PerformanceAnalyzer) GetCollectors() map[string]PerformanceCollector {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	result := make(map[string]PerformanceCollector)
	for k, v := range pa.collectors {
		result[k] = v
	}
	return result
}

// GetRules 获取所有分析规则
func (pa *PerformanceAnalyzer) GetRules() []AnalysisRule {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	rules := make([]AnalysisRule, len(pa.rules))
	copy(rules, pa.rules)
	return rules
}

// ClearResults 清空分析结果
func (pa *PerformanceAnalyzer) ClearResults() {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.results = make([]AnalysisResult, 0)
	pa.logger.Log(log.LevelInfo, "Analysis results cleared")
}
