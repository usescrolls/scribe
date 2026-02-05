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

  function mountSkillCard(
    skill: SkillInfo = mockSkill,
    props: {
      showUninstall?: boolean
      showRemove?: boolean
      showAdd?: boolean
    } = {},
  ) {
    return mount(SkillCard, {
      props: { skill, ...props },
    })
  }

  describe("rendering", () => {
    it("renders skill name", () => {
      const wrapper = mountSkillCard()

      expect(wrapper.find(".name").text()).toBe("react-patterns")
    })

    it("renders skill description (truncated)", () => {
      const wrapper = mountSkillCard()

      expect(wrapper.find(".description").text()).toBe(
        "React best practices and patterns",
      )
    })

    it("truncates long descriptions", () => {
      const skillWithLongDesc: SkillInfo = {
        ...mockSkill,
        description:
          "This is a very long description that should be truncated because it exceeds the maximum length allowed for display",
      }
      const wrapper = mountSkillCard(skillWithLongDesc)

      const desc = wrapper.find(".description").text()
      expect(desc.length).toBeLessThanOrEqual(63) // 60 chars + "..."
      expect(desc.endsWith("...")).toBe(true)
    })

    it("renders source type badge", () => {
      const wrapper = mountSkillCard()

      expect(wrapper.find(".source-badge").text()).toBe("github")
    })

    it("handles skill with no description", () => {
      const skillWithoutDesc: SkillInfo = {
        ...mockSkill,
        description: "",
      }
      const wrapper = mountSkillCard(skillWithoutDesc)

      expect(wrapper.find(".description").exists()).toBe(false)
    })
  })

  describe("uninstall button", () => {
    it("hides uninstall button by default", () => {
      const wrapper = mountSkillCard()

      expect(wrapper.find(".uninstall-btn").exists()).toBe(false)
    })

    it("shows uninstall button when showUninstall is true", () => {
      const wrapper = mountSkillCard(mockSkill, { showUninstall: true })

      expect(wrapper.find(".uninstall-btn").exists()).toBe(true)
    })

    it("emits uninstall event with skill name", async () => {
      const wrapper = mountSkillCard(mockSkill, { showUninstall: true })

      await wrapper.find(".uninstall-btn").trigger("click")

      expect(wrapper.emitted("uninstall")).toBeTruthy()
      expect(wrapper.emitted("uninstall")![0]).toEqual(["react-patterns"])
    })

    it("shows 'Uninstall' label on the button", () => {
      const wrapper = mountSkillCard(mockSkill, { showUninstall: true })

      expect(wrapper.find(".uninstall-btn .btn-label").text()).toBe("Uninstall")
    })
  })

  describe("remove button", () => {
    it("hides remove button by default", () => {
      const wrapper = mountSkillCard()

      expect(wrapper.find(".remove-btn").exists()).toBe(false)
    })

    it("shows remove button when showRemove is true", () => {
      const wrapper = mountSkillCard(mockSkill, { showRemove: true })

      expect(wrapper.find(".remove-btn").exists()).toBe(true)
    })

    it("emits remove event with skill name", async () => {
      const wrapper = mountSkillCard(mockSkill, { showRemove: true })

      await wrapper.find(".remove-btn").trigger("click")

      expect(wrapper.emitted("remove")).toBeTruthy()
      expect(wrapper.emitted("remove")![0]).toEqual(["react-patterns"])
    })

    it("shows 'Remove' label on the button", () => {
      const wrapper = mountSkillCard(mockSkill, { showRemove: true })

      expect(wrapper.find(".remove-btn .btn-label").text()).toBe("Remove")
    })
  })

  describe("add button", () => {
    it("hides add button by default", () => {
      const wrapper = mountSkillCard()

      expect(wrapper.find(".add-btn").exists()).toBe(false)
    })

    it("shows add button when showAdd is true", () => {
      const wrapper = mountSkillCard(mockSkill, { showAdd: true })

      expect(wrapper.find(".add-btn").exists()).toBe(true)
    })

    it("emits add event with skill name", async () => {
      const wrapper = mountSkillCard(mockSkill, { showAdd: true })

      await wrapper.find(".add-btn").trigger("click")

      expect(wrapper.emitted("add")).toBeTruthy()
      expect(wrapper.emitted("add")![0]).toEqual(["react-patterns"])
    })
  })
})
