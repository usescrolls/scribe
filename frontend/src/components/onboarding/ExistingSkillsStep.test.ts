import { describe, it, expect } from "vitest"
import { mount } from "@vue/test-utils"
import ExistingSkillsStep from "./ExistingSkillsStep.vue"
import type { ExistingSkillInfo, SkillConflict } from "../../types/skill"

describe("ExistingSkillsStep", () => {
  const mockSkills: ExistingSkillInfo[] = [
    {
      name: "react-patterns",
      path: "/path/a",
      agentId: "claude-code",
      agentName: "Claude Code",
      isGitRepo: true,
    },
    {
      name: "vue-utils",
      path: "/path/b",
      agentId: "cursor",
      agentName: "Cursor",
      isGitRepo: false,
    },
  ]

  function mountStep(
    props: Partial<{
      skills: ExistingSkillInfo[]
      conflicts: SkillConflict[]
      loading: boolean
    }> = {},
  ) {
    return mount(ExistingSkillsStep, {
      props: {
        skills: mockSkills,
        conflicts: [],
        loading: false,
        ...props,
      },
    })
  }

  it("renders title", () => {
    const wrapper = mountStep()

    expect(wrapper.find("h1").text()).toBe("Existing Skills Found")
  })

  it("shows loading state", () => {
    const wrapper = mountStep({ loading: true })

    expect(wrapper.find(".loading").exists()).toBe(true)
  })

  it("shows skill count in description", () => {
    const wrapper = mountStep()

    expect(wrapper.find(".description").text()).toContain("2 skills")
  })

  it("renders skill items", () => {
    const wrapper = mountStep()

    const items = wrapper.findAll(".skill-item")
    expect(items).toHaveLength(2)
  })

  it("shows git badge for git repos", () => {
    const wrapper = mountStep()

    const badges = wrapper.findAll(".git-badge")
    expect(badges).toHaveLength(1)
    expect(badges[0].text()).toBe("git")
  })

  it("shows no skills message when empty", () => {
    const wrapper = mountStep({ skills: [] })

    expect(wrapper.find(".no-skills").exists()).toBe(true)
  })

  it("renders import and delete buttons", () => {
    const wrapper = mountStep()

    expect(wrapper.find(".primary-button").text()).toBe("Import All")
    expect(wrapper.find(".secondary-button").text()).toContain("Delete All")
  })

  it("emits import-all on import click", async () => {
    const wrapper = mountStep()

    await wrapper.find(".primary-button").trigger("click")

    expect(wrapper.emitted("import-all")).toBeTruthy()
  })

  it("emits delete-all on delete click", async () => {
    const wrapper = mountStep()

    await wrapper.find(".secondary-button").trigger("click")

    expect(wrapper.emitted("delete-all")).toBeTruthy()
  })

  it("shows conflicts section when conflicts exist", () => {
    const conflicts: SkillConflict[] = [
      {
        name: "dupe-skill",
        sources: [
          {
            name: "dupe-skill",
            path: "/a",
            agentId: "claude-code",
            agentName: "Claude Code",
            isGitRepo: false,
          },
          {
            name: "dupe-skill",
            path: "/b",
            agentId: "cursor",
            agentName: "Cursor",
            isGitRepo: false,
          },
        ],
      },
    ]
    const wrapper = mountStep({ conflicts })

    expect(wrapper.find(".conflicts-section").exists()).toBe(true)
    expect(wrapper.find(".conflict-name").text()).toBe("dupe-skill")
  })

  it("emits resolve-conflict with path on conflict option click", async () => {
    const conflicts: SkillConflict[] = [
      {
        name: "dupe-skill",
        sources: [
          {
            name: "dupe-skill",
            path: "/path/a",
            agentId: "claude-code",
            agentName: "Claude Code",
            isGitRepo: false,
          },
        ],
      },
    ]
    const wrapper = mountStep({ conflicts })

    await wrapper.find(".conflict-option").trigger("click")

    expect(wrapper.emitted("resolve-conflict")).toBeTruthy()
    expect(wrapper.emitted("resolve-conflict")![0]).toEqual(["/path/a"])
  })
})
