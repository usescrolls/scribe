import { describe, expect, it } from "vitest"
import { fuzzyFilterSkills } from "./fuzzySearch"
import type { SkillInfo } from "../types/skill"

describe("fuzzyFilterSkills", () => {
  const skills: SkillInfo[] = [
    {
      name: "typescript-patterns",
      description: "TypeScript tips",
      source: "local/path",
      sourceType: "local",
      installedAt: "2025-01-28T10:00:00Z",
      agents: ["claude-code"],
    },
    {
      name: "react-best-practices",
      displayName: "react-best-practices",
      description: "React component patterns",
      source: "vercel-labs/skills",
      sourceType: "github",
      installedAt: "2025-01-29T10:00:00Z",
      agents: ["cursor"],
    },
    {
      name: "go-idioms",
      description: "Go patterns",
      source: "example/go-skills",
      sourceType: "github",
      installedAt: "2025-01-27T10:00:00Z",
      agents: ["claude-code"],
    },
  ]

  it("matches fuzzy subsequences in names", () => {
    expect(fuzzyFilterSkills(skills, "rct").map((skill) => skill.name)).toEqual(
      ["react-best-practices"],
    )
  })

  it("requires every query token to match", () => {
    expect(
      fuzzyFilterSkills(skills, "react comp").map((skill) => skill.name),
    ).toEqual(["react-best-practices"])
  })

  it("matches source and agent fields", () => {
    expect(
      fuzzyFilterSkills(skills, "vercel").map((skill) => skill.name),
    ).toEqual(["react-best-practices"])
    expect(
      fuzzyFilterSkills(skills, "cursor").map((skill) => skill.name),
    ).toEqual(["react-best-practices"])
  })

  it("ranks skill names above descriptions", () => {
    const ranked = fuzzyFilterSkills(
      [
        { ...skills[2], description: "React migration notes" },
        {
          ...skills[1],
          name: "react-patterns",
          description: "Component patterns",
        },
      ],
      "react",
    )

    expect(ranked.map((skill) => skill.name)).toEqual([
      "react-patterns",
      "go-idioms",
    ])
  })

  it("rejects distant subsequence matches", () => {
    const filtered = fuzzyFilterSkills(
      [
        { ...skills[0], source: "owner/typescript" },
        { ...skills[1], source: "owner/react" },
      ],
      "rct",
    )

    expect(filtered.map((skill) => skill.name)).toEqual([
      "react-best-practices",
    ])
  })

  it("matches SKILL.md content", () => {
    const filtered = fuzzyFilterSkills(
      [
        {
          ...skills[0],
          description: "No body marker here",
          content:
            "---\nname: typescript-patterns\ndescription: No body marker here\n---\nUse browser automation for visual checks.",
        },
        {
          ...skills[1],
          description: "Other metadata",
          content:
            "---\nname: react-best-practices\ndescription: Other metadata\n---\nUse database migrations.",
        },
      ],
      "browser automation",
    )

    expect(filtered.map((skill) => skill.name)).toEqual(["typescript-patterns"])
  })

  it("preserves input order for empty queries", () => {
    expect(fuzzyFilterSkills(skills, "   ")).toEqual(skills)
  })
})
