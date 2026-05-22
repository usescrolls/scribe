package scribe

import (
	"sort"
	"strings"
	"unicode"
)

type skillSearchScore struct {
	skill SkillInfo
	score int
}

type skillSearchField struct {
	value  string
	weight int
}

// SearchSkillInfo returns installed skill info matching query, sorted by match quality.
// Empty queries return the input order unchanged.
func SearchSkillInfo(skills []SkillInfo, query string) []SkillInfo {
	query = normalizeSearchQuery(query)
	if query == "" {
		return append([]SkillInfo(nil), skills...)
	}

	matches := make([]skillSearchScore, 0, len(skills))
	for _, skill := range skills {
		score, ok := ScoreSkillInfo(skill, query)
		if ok {
			matches = append(matches, skillSearchScore{skill: skill, score: score})
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return strings.ToLower(matches[i].skill.Name) < strings.ToLower(matches[j].skill.Name)
	})

	results := make([]SkillInfo, 0, len(matches))
	for _, match := range matches {
		results = append(results, match.skill)
	}
	return results
}

// ScoreSkillInfo reports a fuzzy-search score for a skill. Every query token
// must match at least one searchable skill field.
func ScoreSkillInfo(skill SkillInfo, query string) (int, bool) {
	tokens := strings.Fields(normalizeSearchQuery(query))
	if len(tokens) == 0 {
		return 0, true
	}

	fields := skillSearchFields(skill)
	total := 0
	for _, token := range tokens {
		best := 0
		for _, field := range fields {
			score, ok := fuzzyTokenScore(token, field.value)
			if !ok {
				continue
			}
			score *= field.weight
			if score > best {
				best = score
			}
		}
		if best == 0 {
			return 0, false
		}
		total += best
	}

	return total, true
}

func skillSearchFields(skill SkillInfo) []skillSearchField {
	fields := []skillSearchField{
		{value: skill.Name, weight: 8},
		{value: skill.Description, weight: 3},
		{value: skill.Source, weight: 4},
		{value: skill.SourceType, weight: 2},
		{value: strings.Join(skill.Agents, " "), weight: 2},
		{value: skill.Content, weight: 1},
	}
	if skill.DisplayName != "" && !strings.EqualFold(skill.DisplayName, skill.Name) {
		fields = append(fields, skillSearchField{value: skill.DisplayName, weight: 8})
	}
	return fields
}

func normalizeSearchQuery(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}

func fuzzyTokenScore(token, candidate string) (int, bool) {
	token = normalizeSearchQuery(token)
	candidate = normalizeSearchQuery(candidate)
	if token == "" {
		return 0, true
	}
	if candidate == "" {
		return 0, false
	}

	if candidate == token {
		return 1000, true
	}
	if strings.HasPrefix(candidate, token) {
		return 900 - lengthPenalty(candidate, token), true
	}
	if index := strings.Index(candidate, token); index >= 0 {
		score := 750 - index - lengthPenalty(candidate, token)
		if isSearchBoundary(candidate, index) {
			score += 80
		}
		return score, true
	}

	return fuzzySubsequenceScore(token, candidate)
}

func fuzzySubsequenceScore(token, candidate string) (int, bool) {
	tokenRunes := []rune(token)
	candidateRunes := []rune(candidate)
	positions := make([]int, 0, len(tokenRunes))
	start := 0

	for _, queryRune := range tokenRunes {
		found := -1
		for i := start; i < len(candidateRunes); i++ {
			if candidateRunes[i] == queryRune {
				found = i
				break
			}
		}
		if found == -1 {
			return 0, false
		}
		positions = append(positions, found)
		start = found + 1
	}

	gaps := positions[len(positions)-1] - positions[0] + 1 - len(positions)
	consecutive := 0
	boundaryBonus := 0
	for i, position := range positions {
		if isSearchBoundaryRunes(candidateRunes, position) {
			boundaryBonus += 12
		}
		if i > 0 && position == positions[i-1]+1 {
			consecutive++
		}
	}
	if gaps > len(tokenRunes)*2+1 && boundaryBonus < len(tokenRunes)*12 {
		return 0, false
	}

	score := 300 + len(tokenRunes)*18 + consecutive*12 + boundaryBonus - gaps*6 - positions[0]*3
	if score < 80 {
		score = 80
	}
	return score, true
}

func lengthPenalty(candidate, token string) int {
	diff := len([]rune(candidate)) - len([]rune(token))
	if diff < 0 {
		return 0
	}
	if diff > 120 {
		return 120
	}
	return diff
}

func isSearchBoundary(value string, index int) bool {
	return isSearchBoundaryRunes([]rune(value), len([]rune(value[:index])))
}

func isSearchBoundaryRunes(value []rune, index int) bool {
	if index <= 0 {
		return true
	}
	return !unicode.IsLetter(value[index-1]) && !unicode.IsDigit(value[index-1])
}
