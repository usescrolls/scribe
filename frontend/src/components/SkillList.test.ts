import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import SkillList from "./SkillList.vue"
import ConfirmDialog from "./ConfirmDialog.vue"
import { mockAppService, mockBrowser } from "../test/setup"
import type { SkillInfo, WorkspaceInfo } from "../types/skill"

describe("SkillList", () => {
  const mockSkills: SkillInfo[] = [
    {
      name: "react-patterns",
      description: "React best practices",
      source: "vercel-labs/skills",
      sourceType: "github",
      installedAt: "2025-01-29T10:00:00Z",
      agents: ["claude-code", "cursor"],
    },
    {
      name: "typescript-tips",
      description: "TypeScript tips",
      source: "local/path",
      sourceType: "local",
      installedAt: "2025-01-28T10:00:00Z",
      agents: ["claude-code"],
    },
  ]

  const mockWorkspaces: WorkspaceInfo[] = [
    {
      name: "default",
      description: "Default workspace",
      skills: ["react-patterns", "typescript-tips"],
      isActive: true,
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.GetSkills.mockResolvedValue(mockSkills)
    mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)
  })

  async function mountSkillList() {
    const wrapper = mount(SkillList)
    await flushPromises()
    return wrapper
  }

  describe("rendering", () => {
    it("shows loading state initially", () => {
      mockAppService.GetSkills.mockReturnValue(new Promise(() => {}))
      mockAppService.GetWorkspaces.mockReturnValue(new Promise(() => {}))
      const wrapper = mount(SkillList)

      expect(wrapper.find(".loading").exists()).toBe(true)
      expect(wrapper.text()).toContain("Loading skills...")
    })

    it("shows error state on fetch failure", async () => {
      mockAppService.GetSkills.mockRejectedValue(new Error("Network error"))

      const wrapper = await mountSkillList()

      expect(wrapper.find(".error").exists()).toBe(true)
      expect(wrapper.text()).toContain("Network error")
    })

    it("shows retry button on error", async () => {
      mockAppService.GetSkills.mockRejectedValue(new Error("Network error"))

      const wrapper = await mountSkillList()

      expect(wrapper.find(".btn-secondary").exists()).toBe(true)
    })

    it("shows empty state when no skills in workspace", async () => {
      mockAppService.GetWorkspaces.mockResolvedValue([
        { name: "default", description: "", skills: [], isActive: true },
      ])

      const wrapper = await mountSkillList()

      expect(wrapper.findComponent({ name: "EmptyState" }).exists()).toBe(true)
    })

    it("renders skills in workspace", async () => {
      const wrapper = await mountSkillList()

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      expect(cards).toHaveLength(2)
    })

    it("shows skill count", async () => {
      const wrapper = await mountSkillList()

      expect(wrapper.find(".count").text()).toBe("2 skills in workspace")
    })

    it("groups skills by source", async () => {
      const wrapper = await mountSkillList()

      const groups = wrapper.findAll(".source-group")
      expect(groups).toHaveLength(2)
    })

    it("renders source as link when sourceUrl is present", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          ...mockSkills[0],
          sourceUrl: "https://github.com/vercel-labs/skills",
        },
        mockSkills[1],
      ])

      const wrapper = await mountSkillList()

      const links = wrapper.findAll(".group-source-link")
      expect(links).toHaveLength(1)
      expect(links[0].text()).toBe("vercel-labs/skills")

      const plainSources = wrapper.findAll(
        ".group-source:not(.group-source-link)",
      )
      expect(plainSources).toHaveLength(1)
      expect(plainSources[0].text()).toBe("local/path")
    })

    it("opens source URL in browser on click", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          ...mockSkills[0],
          sourceUrl: "https://github.com/vercel-labs/skills",
        },
      ])

      const wrapper = await mountSkillList()

      const link = wrapper.find(".group-source-link")
      await link.trigger("click")

      expect(mockBrowser.OpenURL).toHaveBeenCalledWith(
        "https://github.com/vercel-labs/skills",
      )
    })

    it("shows singular skill text for 1 skill", async () => {
      mockAppService.GetWorkspaces.mockResolvedValue([
        {
          name: "default",
          description: "",
          skills: ["react-patterns"],
          isActive: true,
        },
      ])

      const wrapper = await mountSkillList()

      expect(wrapper.find(".count").text()).toBe("1 skill in workspace")
    })
  })

  describe("remove from workspace", () => {
    it("passes show-remove prop to SkillCard", async () => {
      const wrapper = await mountSkillList()

      const card = wrapper.findComponent({ name: "SkillCard" })
      expect(card.props("showRemove")).toBe(true)
    })

    it("shows confirm dialog on remove", async () => {
      const wrapper = await mountSkillList()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("remove", "react-patterns")
      await flushPromises()

      const dialog = wrapper.findComponent(ConfirmDialog)
      expect(dialog.exists()).toBe(true)
      expect(dialog.props("title")).toBe("Remove from Workspace")
      expect(dialog.props("danger")).toBe(true)
    })

    it("calls RemoveSkillFromWorkspace on confirm", async () => {
      mockAppService.RemoveSkillFromWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountSkillList()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("remove", "react-patterns")
      await flushPromises()

      const dialog = wrapper.findComponent(ConfirmDialog)
      await dialog.vm.$emit("confirm")
      await flushPromises()

      expect(mockAppService.RemoveSkillFromWorkspace).toHaveBeenCalledWith(
        "react-patterns",
        "default",
      )
    })

    it("does not remove on cancel", async () => {
      const wrapper = await mountSkillList()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("remove", "react-patterns")
      await flushPromises()

      const dialog = wrapper.findComponent(ConfirmDialog)
      await dialog.vm.$emit("cancel")
      await flushPromises()

      expect(mockAppService.RemoveSkillFromWorkspace).not.toHaveBeenCalled()
    })

    it("shows error on remove failure", async () => {
      mockAppService.RemoveSkillFromWorkspace.mockRejectedValue(
        new Error("Permission denied"),
      )

      const wrapper = await mountSkillList()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("remove", "react-patterns")
      await flushPromises()

      const dialog = wrapper.findComponent(ConfirmDialog)
      await dialog.vm.$emit("confirm")
      await flushPromises()

      expect(wrapper.text()).toContain("Permission denied")
    })
  })
})
