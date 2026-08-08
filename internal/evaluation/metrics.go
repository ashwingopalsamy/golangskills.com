package evaluation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
)

// Report summarizes deterministic scores, routing behavior, collection-level
// results, and client-reported token use for one arm.
type Report struct {
	SchemaVersion int                `json:"schema_version"`
	Arm           string             `json:"arm"`
	Cases         int                `json:"cases"`
	Passed        int                `json:"passed"`
	PassRate      float64            `json:"pass_rate"`
	MeanScore     float64            `json:"mean_score"`
	CILower95     float64            `json:"ci_lower_95"`
	CIUpper95     float64            `json:"ci_upper_95"`
	CriticalFails int                `json:"critical_failures"`
	Collections   []CollectionReport `json:"collections"`
	Routing       RoutingReport      `json:"routing"`
	Tokens        TokenReport        `json:"tokens"`
	FailureCases  []string           `json:"failure_cases,omitempty"`
	Input         string             `json:"input"`
}

// CollectionReport keeps release-domain non-inferiority machine-checkable.
type CollectionReport struct {
	Collection string  `json:"collection"`
	Cases      int     `json:"cases"`
	Passed     int     `json:"passed"`
	PassRate   float64 `json:"pass_rate"`
	MeanScore  float64 `json:"mean_score"`
}

// RoutingReport separates multiclass routing quality from unrelated-prompt
// false activation. A zero unrelated case count is evidence missing, not 0%.
type RoutingReport struct {
	Cases               int     `json:"cases"`
	MacroF1             float64 `json:"macro_f1"`
	UnrelatedCases      int     `json:"unrelated_cases"`
	FalseActivations    int     `json:"false_activations"`
	FalseActivationRate float64 `json:"false_activation_rate"`
}

// TokenReport uses total client-reported input tokens for the primary
// efficiency metric and exposes cached tokens separately.
type TokenReport struct {
	InputTokens           int64   `json:"input_tokens"`
	CachedInputTokens     int64   `json:"cached_input_tokens"`
	CacheWriteInputTokens int64   `json:"cache_write_input_tokens"`
	OutputTokens          int64   `json:"output_tokens"`
	ReasoningOutputTokens int64   `json:"reasoning_output_tokens"`
	ScorePerKInputTokens  float64 `json:"score_per_k_input_tokens"`
}

// ComparisonReport is a paired deterministic comparison over identical case
// keys. Ties contribute one half to paired_win_rate; the bootstrap resamples
// complete pairs, never independent arms.
type ComparisonReport struct {
	SchemaVersion           int                    `json:"schema_version"`
	Candidate               string                 `json:"candidate"`
	Competitor              string                 `json:"competitor"`
	CandidateInput          string                 `json:"candidate_input"`
	CompetitorInput         string                 `json:"competitor_input"`
	CompletePairs           bool                   `json:"complete_pairs"`
	PairedCases             int                    `json:"paired_cases"`
	CandidateOnlyCases      []string               `json:"candidate_only_cases,omitempty"`
	CompetitorOnlyCases     []string               `json:"competitor_only_cases,omitempty"`
	CandidateWins           int                    `json:"candidate_wins"`
	CompetitorWins          int                    `json:"competitor_wins"`
	Ties                    int                    `json:"ties"`
	PairedWinRate           float64                `json:"paired_win_rate"`
	PairedWinCILower95      float64                `json:"paired_win_ci_lower_95"`
	PairedWinCIUpper95      float64                `json:"paired_win_ci_upper_95"`
	MeanScoreDifference     float64                `json:"mean_score_difference"`
	MeanDifferenceCILower95 float64                `json:"mean_difference_ci_lower_95"`
	MeanDifferenceCIUpper95 float64                `json:"mean_difference_ci_upper_95"`
	CandidateCriticalFails  int                    `json:"candidate_critical_failures"`
	CompetitorCriticalFails int                    `json:"competitor_critical_failures"`
	CandidateTokenRate      float64                `json:"candidate_score_per_k_input_tokens"`
	CompetitorTokenRate     float64                `json:"competitor_score_per_k_input_tokens"`
	ParetoRelation          string                 `json:"pareto_relation"`
	Collections             []CollectionComparison `json:"collections"`
	CandidateLossCases      []string               `json:"candidate_loss_cases,omitempty"`
}

// CollectionComparison reports paired mean scores by canonical collection.
type CollectionComparison struct {
	Collection     string  `json:"collection"`
	Cases          int     `json:"cases"`
	CandidateMean  float64 `json:"candidate_mean"`
	CompetitorMean float64 `json:"competitor_mean"`
	Difference     float64 `json:"difference"`
}

type reportAccumulator struct {
	cases  int
	passed int
	score  float64
}

// ReportFile summarizes a scored JSONL artifact.
func ReportFile(inputPath string) (Report, error) {
	scores, err := readScores(inputPath)
	if err != nil {
		return Report{}, err
	}
	return reportScores(scores, inputPath, "")
}

// ReportFileForArm selects one arm from a mixed matrix score artifact.
func ReportFileForArm(inputPath, arm string) (Report, error) {
	scores, err := readScores(inputPath)
	if err != nil {
		return Report{}, err
	}
	return reportScores(scores, inputPath, arm)
}

func reportScores(scores []Score, inputPath, selectedArm string) (Report, error) {
	report := Report{SchemaVersion: 2, Input: inputPath}
	if selectedArm != "" {
		report.Arm = selectedArm
	}
	collections := make(map[string]*reportAccumulator)
	var expectedLabels, predictedLabels []string
	for _, score := range scores {
		if selectedArm != "" && score.Arm != selectedArm {
			continue
		}
		if report.Arm == "" {
			report.Arm = score.Arm
		} else if score.Arm != report.Arm {
			return Report{}, fmt.Errorf("%s contains multiple arms: %s and %s", inputPath, report.Arm, score.Arm)
		}
		report.Cases++
		report.MeanScore += score.Score
		if score.Passed {
			report.Passed++
		} else {
			report.FailureCases = append(report.FailureCases, caseKey(score))
			if isCritical(score) {
				report.CriticalFails++
			}
		}
		collection := score.Collection
		if collection == "" {
			collection = collectionForSkill(score.Skill)
		}
		bucket := collections[collection]
		if bucket == nil {
			bucket = &reportAccumulator{}
			collections[collection] = bucket
		}
		bucket.cases++
		bucket.score += score.Score
		if score.Passed {
			bucket.passed++
		}
		usage := score.Usage
		if usage.InputTokens == 0 {
			usage = parseUsage(score.RunnerEvents)
		}
		report.Tokens.InputTokens += usage.InputTokens
		report.Tokens.CachedInputTokens += usage.CachedInputTokens
		report.Tokens.CacheWriteInputTokens += usage.CacheWriteInputTokens
		report.Tokens.OutputTokens += usage.OutputTokens
		report.Tokens.ReasoningOutputTokens += usage.ReasoningOutputTokens
		if score.Kind == "routing" {
			report.Routing.Cases++
			expected, predicted := routingLabels(score)
			expectedLabels = append(expectedLabels, expected)
			predictedLabels = append(predictedLabels, predicted)
			if expected == "none" {
				report.Routing.UnrelatedCases++
				if predicted != "none" {
					report.Routing.FalseActivations++
				}
			}
		}
	}
	if report.Cases == 0 {
		return Report{}, fmt.Errorf("%s contains no scores for arm %q", inputPath, selectedArm)
	}
	if report.Cases > 0 {
		report.PassRate = float64(report.Passed) / float64(report.Cases)
		report.MeanScore /= float64(report.Cases)
		report.CILower95, report.CIUpper95 = wilson(report.Passed, report.Cases)
	}
	if report.Tokens.InputTokens > 0 {
		report.Tokens.ScorePerKInputTokens = report.MeanScore * 1000 / (float64(report.Tokens.InputTokens) / float64(max(1, report.Cases)))
	}
	report.Routing.MacroF1 = macroF1(expectedLabels, predictedLabels)
	if report.Routing.UnrelatedCases > 0 {
		report.Routing.FalseActivationRate = float64(report.Routing.FalseActivations) / float64(report.Routing.UnrelatedCases)
	}
	for collection, bucket := range collections {
		report.Collections = append(report.Collections, CollectionReport{
			Collection: collection, Cases: bucket.cases, Passed: bucket.passed,
			PassRate: float64(bucket.passed) / float64(bucket.cases), MeanScore: bucket.score / float64(bucket.cases),
		})
	}
	sort.Slice(report.Collections, func(i, j int) bool { return report.Collections[i].Collection < report.Collections[j].Collection })
	sort.Strings(report.FailureCases)
	return report, nil
}

// CompareFiles constructs a paired comparison from two scored artifacts.
func CompareFiles(candidatePath, competitorPath string) (ComparisonReport, error) {
	return CompareArms(candidatePath, "", competitorPath, "")
}

// CompareArms selects named arms from separate or shared mixed score files.
func CompareArms(candidatePath, candidateArm, competitorPath, competitorArm string) (ComparisonReport, error) {
	candidateScores, err := readScores(candidatePath)
	if err != nil {
		return ComparisonReport{}, err
	}
	competitorScores, err := readScores(competitorPath)
	if err != nil {
		return ComparisonReport{}, err
	}
	if candidateArm != "" {
		candidateScores = scoresForArm(candidateScores, candidateArm)
	}
	if competitorArm != "" {
		competitorScores = scoresForArm(competitorScores, competitorArm)
	}
	candidateReport, err := reportScores(candidateScores, candidatePath, candidateArm)
	if err != nil {
		return ComparisonReport{}, err
	}
	competitorReport, err := reportScores(competitorScores, competitorPath, competitorArm)
	if err != nil {
		return ComparisonReport{}, err
	}
	report := ComparisonReport{
		SchemaVersion: 2, Candidate: candidateReport.Arm, Competitor: competitorReport.Arm,
		CandidateInput: candidatePath, CompetitorInput: competitorPath,
		CandidateCriticalFails: candidateReport.CriticalFails, CompetitorCriticalFails: competitorReport.CriticalFails,
		CandidateTokenRate: candidateReport.Tokens.ScorePerKInputTokens, CompetitorTokenRate: competitorReport.Tokens.ScorePerKInputTokens,
	}
	candidate := indexScores(candidateScores)
	competitor := indexScores(competitorScores)
	var utilities, differences []float64
	collections := make(map[string][][2]float64)
	for key, left := range candidate {
		right, exists := competitor[key]
		if !exists {
			report.CandidateOnlyCases = append(report.CandidateOnlyCases, key)
			continue
		}
		report.PairedCases++
		difference := left.Score - right.Score
		differences = append(differences, difference)
		utility := 0.5
		switch {
		case difference > 0:
			report.CandidateWins++
			utility = 1
		case difference < 0:
			report.CompetitorWins++
			report.CandidateLossCases = append(report.CandidateLossCases, key)
			utility = 0
		default:
			report.Ties++
		}
		utilities = append(utilities, utility)
		collection := left.Collection
		if collection == "" {
			collection = collectionForSkill(left.Skill)
		}
		collections[collection] = append(collections[collection], [2]float64{left.Score, right.Score})
	}
	for key := range competitor {
		if _, exists := candidate[key]; !exists {
			report.CompetitorOnlyCases = append(report.CompetitorOnlyCases, key)
		}
	}
	report.CompletePairs = report.PairedCases > 0 && len(report.CandidateOnlyCases) == 0 && len(report.CompetitorOnlyCases) == 0
	if report.PairedCases > 0 {
		report.PairedWinRate = mean(utilities)
		report.MeanScoreDifference = mean(differences)
		report.PairedWinCILower95, report.PairedWinCIUpper95 = pairedBootstrapCI(utilities, 10_000, 1)
		report.MeanDifferenceCILower95, report.MeanDifferenceCIUpper95 = pairedBootstrapCI(differences, 10_000, 2)
	}
	report.ParetoRelation = paretoRelation(candidateReport, competitorReport)
	for collection, pairs := range collections {
		var left, right []float64
		for _, pair := range pairs {
			left = append(left, pair[0])
			right = append(right, pair[1])
		}
		report.Collections = append(report.Collections, CollectionComparison{
			Collection: collection, Cases: len(pairs), CandidateMean: mean(left), CompetitorMean: mean(right), Difference: mean(left) - mean(right),
		})
	}
	sort.Slice(report.Collections, func(i, j int) bool { return report.Collections[i].Collection < report.Collections[j].Collection })
	sort.Strings(report.CandidateOnlyCases)
	sort.Strings(report.CompetitorOnlyCases)
	sort.Strings(report.CandidateLossCases)
	return report, nil
}

func scoresForArm(scores []Score, arm string) []Score {
	result := make([]Score, 0, len(scores))
	for _, score := range scores {
		if score.Arm == arm {
			result = append(result, score)
		}
	}
	return result
}

func readScores(path string) ([]Score, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var scores []Score
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		var score Score
		if err := json.Unmarshal(scanner.Bytes(), &score); err != nil {
			return nil, fmt.Errorf("decode %s: %w", path, err)
		}
		scores = append(scores, score)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(scores) == 0 {
		return nil, fmt.Errorf("%s contains no scores", path)
	}
	return scores, nil
}

func indexScores(scores []Score) map[string]Score {
	result := make(map[string]Score, len(scores))
	for _, score := range scores {
		result[caseKey(score)] = score
	}
	return result
}

func caseKey(score Score) string {
	key := score.Skill + "/" + score.CaseID
	if score.Mode == "explicit" {
		key += "@explicit"
	}
	if score.Repetition > 0 {
		key += fmt.Sprintf("#r%d", score.Repetition)
	}
	return key
}

func isCritical(score Score) bool {
	collection := score.Collection
	if collection == "" {
		collection = collectionForSkill(score.Skill)
	}
	return collection == "fintech-skills-for-go"
}

func collectionForSkill(skill string) string {
	switch skill {
	case "go-concurrency-lifecycle", "go-data-consistency", "go-message-processing", "go-service-resilience", "go-distributed-coordination", "review-go-distributed-change":
		return "distributed-systems-skills-for-go"
	case "go-money-and-ledgers", "go-payment-lifecycles", "go-financial-idempotency", "go-clearing-settlement-reconciliation", "go-fintech-security-compliance", "review-go-fintech-change":
		return "fintech-skills-for-go"
	default:
		return "engineering-skills-for-go"
	}
}

func routingLabels(score Score) (string, string) {
	expectedRoutes := score.ExpectedRoutes
	if len(expectedRoutes) == 0 && score.ExpectedRoute != "" {
		expectedRoutes = []string{score.ExpectedRoute}
	}
	predicted := normalizeRoute(score.Response)
	for _, route := range expectedRoutes {
		if routeMatches(score.Response, route) {
			return normalizeRoute(route), normalizeRoute(route)
		}
	}
	if len(expectedRoutes) == 0 {
		return "none", predicted
	}
	return normalizeRoute(expectedRoutes[0]), predicted
}

func normalizeRoute(route string) string {
	route = strings.ToLower(strings.TrimSpace(route))
	if index := strings.LastIndex(route, ":"); index >= 0 {
		route = route[index+1:]
	}
	if route == "" {
		return "none"
	}
	return route
}

func macroF1(expected, predicted []string) float64 {
	if len(expected) == 0 || len(expected) != len(predicted) {
		return 0
	}
	labels := make(map[string]struct{})
	for index := range expected {
		labels[expected[index]] = struct{}{}
		labels[predicted[index]] = struct{}{}
	}
	var total float64
	for label := range labels {
		tp, fp, fn := 0, 0, 0
		for index := range expected {
			switch {
			case expected[index] == label && predicted[index] == label:
				tp++
			case expected[index] != label && predicted[index] == label:
				fp++
			case expected[index] == label && predicted[index] != label:
				fn++
			}
		}
		denominator := 2*tp + fp + fn
		if denominator > 0 {
			total += float64(2*tp) / float64(denominator)
		}
	}
	return total / float64(len(labels))
}

func wilson(successes, trials int) (float64, float64) {
	if trials == 0 {
		return 0, 0
	}
	z := 1.959963984540054
	n := float64(trials)
	p := float64(successes) / n
	denominator := 1 + z*z/n
	center := (p + z*z/(2*n)) / denominator
	margin := z * math.Sqrt((p*(1-p)+z*z/(4*n))/n) / denominator
	return math.Max(0, center-margin), math.Min(1, center+margin)
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var total float64
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func pairedBootstrapCI(values []float64, samples int, seed int64) (float64, float64) {
	if len(values) == 0 || samples <= 0 {
		return 0, 0
	}
	random := rand.New(rand.NewSource(seed))
	means := make([]float64, samples)
	for sample := range means {
		var total float64
		for range values {
			total += values[random.Intn(len(values))]
		}
		means[sample] = total / float64(len(values))
	}
	sort.Float64s(means)
	return percentile(means, 0.025), percentile(means, 0.975)
}

func percentile(sortedValues []float64, probability float64) float64 {
	if len(sortedValues) == 0 {
		return 0
	}
	position := probability * float64(len(sortedValues)-1)
	lower := int(math.Floor(position))
	upper := int(math.Ceil(position))
	if lower == upper {
		return sortedValues[lower]
	}
	weight := position - float64(lower)
	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

func paretoRelation(candidate, competitor Report) string {
	if candidate.Tokens.InputTokens == 0 || competitor.Tokens.InputTokens == 0 {
		return "unknown-missing-token-data"
	}
	candidateBetterScore := candidate.MeanScore >= competitor.MeanScore
	candidateBetterTokens := candidate.Tokens.InputTokens <= competitor.Tokens.InputTokens
	competitorBetterScore := competitor.MeanScore >= candidate.MeanScore
	competitorBetterTokens := competitor.Tokens.InputTokens <= candidate.Tokens.InputTokens
	switch {
	case candidateBetterScore && candidateBetterTokens && (candidate.MeanScore > competitor.MeanScore || candidate.Tokens.InputTokens < competitor.Tokens.InputTokens):
		return "candidate-dominates"
	case competitorBetterScore && competitorBetterTokens && (competitor.MeanScore > candidate.MeanScore || competitor.Tokens.InputTokens < candidate.Tokens.InputTokens):
		return "competitor-dominates"
	case candidate.MeanScore == competitor.MeanScore && candidate.Tokens.InputTokens == competitor.Tokens.InputTokens:
		return "equal"
	default:
		return "tradeoff"
	}
}
