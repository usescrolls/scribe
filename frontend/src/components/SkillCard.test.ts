import { describe, it, expect } from "vitest"
import { mount } from "@vue/test-utils"
import SkillCard from "./SkillCard.vue"
import type { SkillInfo } from "../types/skill"

describe("SkillCard", () => {
  const mockSkill: SkillInfo = {
    name: "react-patterns",
    description: "React best practices and patterns",
    source: "vercel-labs/skills",
    sourceType: "github",
    installedAt: "2025-01-29T10:00:00Z",
    agents: ["claude-code", "cursor"],
  }

  function mountSkillCard(skill: SkillInfo = mockSkill) {
    return mount(SkillCard, {
      props: { skill },
    })
  }

  describe("rendering", () => {
    it("renders skill name", () => {
      const wrapper = mountSkillCard()

      expect(wrapper.find(".name").text()).toBe("react-patterns")
    })

    it("renders skill description", () => {
      const wrapper = mountSkillCard()

      expect(wrapper.find(".description").text()).toBe(
        "React best practices and patterns",
      )
    })

    it("renders source type badge", () => {
      const wrapper = mountSkillCard()

      expect(wrapper.find(".source-type").text()).toBe("github")
    })

    it("renders source", () => {
      const wrapper = mountSkillCard()

      expect(wrapper.find(".source").text()).toBe("vercel-labs/skills")
    })

    it("renders agent badges", () => {
      const wrapper = mountSkillCard()

      const badges = wrapper.findAll(".agent-badge")
      expect(badges).toHaveLength(2)
      expect(badges[0].text()).toBe("Claude Code")
      expect(badges[1].text()).toBe("Cursor")
    })

    it("handles skill with no description", () => {
      const skillWithoutDesc: SkillInfo = {
        ...mockSkill,
        description: "",
      }
      const wrapper = mountSkillCard(skillWithoutDesc)

      expect(wrapper.find(".description").exists()).toBe(false)
    })

    it("handles skill with no agents", () => {
      const skillWithoutAgents: SkillInfo = {
        ...mockSkill,
        agents: [],
      }
      const wrapper = mountSkillCard(skillWithoutAgents)

      expect(wrapper.find(".agents").exists()).toBe(false)
    })
  })

  describe("agent name formatting", () => {
    it("formats agent IDs to display names", () => {
      const skillWithVariousAgents: SkillInfo = {
        ...mockSkill,
        agents: ["claude-code", "github-copilot", "windsurf"],
      }
      const wrapper = mountSkillCard(skillWithVariousAgents)

      const badges = wrapper.findAll(".agent-badge")
      expect(badges[0].text()).toBe("Claude Code")
      expect(badges[1].text()).toBe("Github Copilot")
      expect(badges[2].text()).toBe("Windsurf")
    })
  })

  describe("uninstall", () => {
    it("emits uninstall event with skill name", async () => {
      const wrapper = mountSkillCard()

      await wrapper.find(".btn-danger").trigger("click")

      expect(wrapper.emitted("uninstall")).toBeTruthy()
      expect(wrapper.emitted("uninstall")![0]).toEqual(["react-patterns"])
    })
  })
})
