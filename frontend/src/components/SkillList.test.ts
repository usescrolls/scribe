import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import { Dialogs } from "@wailsio/runtime"
import SkillList from "./SkillList.vue"
import { mockAppService } from "../test/setup"
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

    it("calls RemoveSkillFromWorkspace on confirm", async () => {
      vi.mocked(Dialogs.Question).mockResolvedValue("Remove")
      mockAppService.RemoveSkillFromWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountSkillList()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("remove", "react-patterns")
      await flushPromises()

      expect(Dialogs.Question).toHaveBeenCalledWith(
        expect.objectContaining({ Title: "Remove from Workspace" }),
      )
      expect(mockAppService.RemoveSkillFromWorkspace).toHaveBeenCalledWith(
        "react-patterns",
        "default",
      )
    })

    it("does not remove on cancel", async () => {
      vi.mocked(Dialogs.Question).mockResolvedValue("Cancel")

      const wrapper = await mountSkillList()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("remove", "react-patterns")
      await flushPromises()

      expect(mockAppService.RemoveSkillFromWorkspace).not.toHaveBeenCalled()
    })

    it("shows error on remove failure", async () => {
      vi.mocked(Dialogs.Question).mockResolvedValue("Remove")
      mockAppService.RemoveSkillFromWorkspace.mockRejectedValue(
        new Error("Permission denied"),
      )

      const wrapper = await mountSkillList()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("remove", "react-patterns")
      await flushPromises()

      expect(wrapper.text()).toContain("Permission denied")
    })
  })
})
