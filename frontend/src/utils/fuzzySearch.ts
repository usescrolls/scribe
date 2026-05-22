import type { SkillInfo } from "../types/skill"

interface SearchField {
  value: string
  weight: number
}

interface SearchScore {
  skill: SkillInfo
  score: number
}

export function fuzzyFilterSkills(
  skills: SkillInfo[],
  query: string,
): SkillInfo[] {
  const normalized = normalizeSearchQuery(query)
  if (!normalized) return [...skills]

  const matches = skills
    .map((skill) => {
      const score = scoreSkillInfo(skill, normalized)
      return score === null ? null : { skill, score }
    })
    .filter((result) => result !== null) as SearchScore[]

  return matches
    .sort((a, b) => {
      if (a.score !== b.score) return b.score - a.score
      return a.skill.name.localeCompare(b.skill.name)
    })
    .map((result) => result.skill)
}

export function scoreSkillInfo(skill: SkillInfo, query: string): number | null {
  const tokens = normalizeSearchQuery(query).split(/\s+/).filter(Boolean)
  if (tokens.length === 0) return 0

  const fields = skillSearchFields(skill)
  let total = 0
  for (const token of tokens) {
    let best = 0
    for (const field of fields) {
      const score = fuzzyTokenScore(token, field.value)
      if (score === null) continue
      best = Math.max(best, score * field.weight)
    }
    if (best === 0) return null
    total += best
  }
  return total
}

function skillSearchFields(skill: SkillInfo): SearchField[] {
  const fields: SearchField[] = [
    { value: skill.name, weight: 8 },
    { value: skill.description, weight: 3 },
    { value: skill.source, weight: 4 },
    { value: skill.sourceType, weight: 2 },
    { value: skill.agents.join(" "), weight: 2 },
    { value: skill.content ?? "", weight: 1 },
  ]
  if (
    skill.displayName &&
    skill.displayName.toLowerCase() !== skill.name.toLowerCase()
  ) {
    fields.push({ value: skill.displayName, weight: 8 })
  }
  return fields
}

function normalizeSearchQuery(query: string): string {
  return query.trim().toLowerCase()
}

function fuzzyTokenScore(token: string, candidate: string): number | null {
  const normalizedToken = normalizeSearchQuery(token)
  const normalizedCandidate = normalizeSearchQuery(candidate)
  if (!normalizedToken) return 0
  if (!normalizedCandidate) return null

  if (normalizedCandidate === normalizedToken) return 1000
  if (normalizedCandidate.startsWith(normalizedToken)) {
    return 900 - lengthPenalty(normalizedCandidate, normalizedToken)
  }

  const index = normalizedCandidate.indexOf(normalizedToken)
  if (index >= 0) {
    let score =
      750 - index - lengthPenalty(normalizedCandidate, normalizedToken)
    if (isSearchBoundary(normalizedCandidate, index)) score += 80
    return score
  }

  return fuzzySubsequenceScore(normalizedToken, normalizedCandidate)
}

function fuzzySubsequenceScore(
  token: string,
  candidate: string,
): number | null {
  const positions: number[] = []
  let start = 0

  for (const char of token) {
    const found = candidate.indexOf(char, start)
    if (found === -1) return null
    positions.push(found)
    start = found + 1
  }

  const gaps =
    positions[positions.length - 1] - positions[0] + 1 - positions.length
  let consecutive = 0
  let boundaryBonus = 0
  for (const [i, position] of positions.entries()) {
    if (isSearchBoundary(candidate, position)) boundaryBonus += 12
    if (i > 0 && position === positions[i - 1] + 1) consecutive += 1
  }
  if (gaps > token.length * 2 + 1 && boundaryBonus < token.length * 12) {
    return null
  }

  return Math.max(
    80,
    300 +
      token.length * 18 +
      consecutive * 12 +
      boundaryBonus -
      gaps * 6 -
      positions[0] * 3,
  )
}

function lengthPenalty(candidate: string, token: string): number {
  return Math.min(Math.max(candidate.length - token.length, 0), 120)
}

function isSearchBoundary(value: string, index: number): boolean {
  if (index <= 0) return true
  const previous = value[index - 1]
  return !/[a-z0-9]/.test(previous)
}
