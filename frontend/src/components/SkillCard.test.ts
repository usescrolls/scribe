import { describe, it, expect } from "vitest"
import { mount } from "@vue/test-utils"
import SkillCard from "./SkillCard.vue"
import type { SkillInfo, WorkspaceInfo } from "../types/skill"

describe("SkillCard", () => {
  const mockSkill: SkillInfo = {
    name: "react-patterns",
    description: "React best practices and patterns",
    source: "vercel-labs/skills",
    sourceType: "github",
    installedAt: "2025-01-29T10:00:00Z",
    agents: ["claude-code", "cursor"],
  }

  const mockWorkspaces: WorkspaceInfo[] = [
    {
      name: "default",
      description: "",
      skills: ["react-patterns"],
      isActive: true,
    },
    { name: "other", description: "", skills: [], isActive: false },
  ]

  function mountSkillCard(
    skill: SkillInfo = mockSkill,
    props: {
      showUninstall?: boolean
      showRemove?: boolean
      showAdd?: boolean
      showWorkspacePicker?: boolean
      skillWorkspaces?: string[]
      allWorkspaces?: WorkspaceInfo[]
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

  describe("detail emit", () => {
    it("emits detail event with skill on card click", async () => {
      const wrapper = mountSkillCard()

      await wrapper.find(".skill-card").trigger("click")

      expect(wrapper.emitted("detail")).toBeTruthy()
      expect(wrapper.emitted("detail")![0]).toEqual([mockSkill])
    })

    it("does not emit detail when clicking uninstall button", async () => {
      const wrapper = mountSkillCard(mockSkill, { showUninstall: true })

      await wrapper.find(".uninstall-btn").trigger("click")

      expect(wrapper.emitted("detail")).toBeFalsy()
      expect(wrapper.emitted("uninstall")).toBeTruthy()
    })

    it("does not emit detail when clicking remove button", async () => {
      const wrapper = mountSkillCard(mockSkill, { showRemove: true })

      await wrapper.find(".remove-btn").trigger("click")

      expect(wrapper.emitted("detail")).toBeFalsy()
      expect(wrapper.emitted("remove")).toBeTruthy()
    })

    it("does not emit detail when clicking add button", async () => {
      const wrapper = mountSkillCard(mockSkill, { showAdd: true })

      await wrapper.find(".add-btn").trigger("click")

      expect(wrapper.emitted("detail")).toBeFalsy()
      expect(wrapper.emitted("add")).toBeTruthy()
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

  describe("workspace badges", () => {
    it("shows workspace badges when skillWorkspaces is non-empty", () => {
      const wrapper = mountSkillCard(mockSkill, {
        skillWorkspaces: ["default", "other"],
      })

      const badges = wrapper.findAll(".ws-badge")
      expect(badges).toHaveLength(2)
      expect(badges[0].text()).toBe("default")
      expect(badges[1].text()).toBe("other")
    })

    it("hides workspace badges when skillWorkspaces is empty", () => {
      const wrapper = mountSkillCard(mockSkill, {
        skillWorkspaces: [],
      })

      expect(wrapper.find(".skill-meta").exists()).toBe(false)
    })
  })

  describe("workspace picker", () => {
    it("hides workspace picker by default", () => {
      const wrapper = mountSkillCard()

      expect(wrapper.find(".ws-picker-wrapper").exists()).toBe(false)
    })

    it("shows workspace picker button when showWorkspacePicker is true", () => {
      const wrapper = mountSkillCard(mockSkill, {
        showWorkspacePicker: true,
        allWorkspaces: mockWorkspaces,
      })

      expect(wrapper.find(".ws-picker-btn").exists()).toBe(true)
    })

    it("opens dropdown on picker button click", async () => {
      const wrapper = mountSkillCard(mockSkill, {
        showWorkspacePicker: true,
        allWorkspaces: mockWorkspaces,
      })

      expect(wrapper.find(".ws-picker-dropdown").exists()).toBe(false)

      await wrapper.find(".ws-picker-btn").trigger("click")

      expect(wrapper.find(".ws-picker-dropdown").exists()).toBe(true)
    })

    it("lists all workspaces in dropdown", async () => {
      const wrapper = mountSkillCard(mockSkill, {
        showWorkspacePicker: true,
        allWorkspaces: mockWorkspaces,
        skillWorkspaces: ["default"],
      })

      await wrapper.find(".ws-picker-btn").trigger("click")

      const items = wrapper.findAll(".ws-picker-item")
      expect(items).toHaveLength(2)
      expect(wrapper.findAll(".ws-picker-name")[0].text()).toBe("default")
      expect(wrapper.findAll(".ws-picker-name")[1].text()).toBe("other")
    })

    it("checks workspaces the skill belongs to", async () => {
      const wrapper = mountSkillCard(mockSkill, {
        showWorkspacePicker: true,
        allWorkspaces: mockWorkspaces,
        skillWorkspaces: ["default"],
      })

      await wrapper.find(".ws-picker-btn").trigger("click")

      const checkboxes = wrapper.findAll(
        '.ws-picker-item input[type="checkbox"]',
      )
      expect((checkboxes[0].element as HTMLInputElement).checked).toBe(true)
      expect((checkboxes[1].element as HTMLInputElement).checked).toBe(false)
    })

    it("emits add-to-workspace when checking a workspace", async () => {
      const wrapper = mountSkillCard(mockSkill, {
        showWorkspacePicker: true,
        allWorkspaces: mockWorkspaces,
        skillWorkspaces: [],
      })

      await wrapper.find(".ws-picker-btn").trigger("click")

      const checkboxes = wrapper.findAll(
        '.ws-picker-item input[type="checkbox"]',
      )
      await checkboxes[0].setValue(true)

      expect(wrapper.emitted("add-to-workspace")).toBeTruthy()
      expect(wrapper.emitted("add-to-workspace")![0]).toEqual([
        "react-patterns",
        "default",
      ])
    })

    it("emits remove-from-workspace when unchecking a workspace", async () => {
      const wrapper = mountSkillCard(mockSkill, {
        showWorkspacePicker: true,
        allWorkspaces: mockWorkspaces,
        skillWorkspaces: ["default"],
      })

      await wrapper.find(".ws-picker-btn").trigger("click")

      const checkboxes = wrapper.findAll(
        '.ws-picker-item input[type="checkbox"]',
      )
      await checkboxes[0].setValue(false)

      expect(wrapper.emitted("remove-from-workspace")).toBeTruthy()
      expect(wrapper.emitted("remove-from-workspace")![0]).toEqual([
        "react-patterns",
        "default",
      ])
    })

    it("closes dropdown on document click", async () => {
      const wrapper = mountSkillCard(mockSkill, {
        showWorkspacePicker: true,
        allWorkspaces: mockWorkspaces,
      })

      await wrapper.find(".ws-picker-btn").trigger("click")
      expect(wrapper.find(".ws-picker-dropdown").exists()).toBe(true)

      document.dispatchEvent(new Event("click"))
      await wrapper.vm.$nextTick()

      expect(wrapper.find(".ws-picker-dropdown").exists()).toBe(false)
    })

    it("toggles dropdown closed on second button click", async () => {
      const wrapper = mountSkillCard(mockSkill, {
        showWorkspacePicker: true,
        allWorkspaces: mockWorkspaces,
      })

      await wrapper.find(".ws-picker-btn").trigger("click")
      expect(wrapper.find(".ws-picker-dropdown").exists()).toBe(true)

      await wrapper.find(".ws-picker-btn").trigger("click")
      expect(wrapper.find(".ws-picker-dropdown").exists()).toBe(false)
    })
  })
})
