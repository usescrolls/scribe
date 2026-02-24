import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import SkillList from "./SkillList.vue"
import ConfirmDialog from "./ConfirmDialog.vue"
import SkillDetailModal from "./SkillDetailModal.vue"
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

    it("shows source avatar in group header", async () => {
      const wrapper = await mountSkillList()

      const avatars = wrapper.findAllComponents({ name: "SourceAvatar" })
      expect(avatars).toHaveLength(2)
      expect(avatars[0].props("source")).toBe("vercel-labs/skills")
      expect(avatars[0].props("sourceType")).toBe("github")
      expect(avatars[1].props("source")).toBe("local/path")
      expect(avatars[1].props("sourceType")).toBe("local")
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

  describe("skill detail modal", () => {
    it("opens detail modal on card detail event", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("# Content")

      const wrapper = await mountSkillList()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("detail", mockSkills[0])
      await flushPromises()

      const modal = wrapper.findComponent(SkillDetailModal)
      expect(modal.exists()).toBe(true)
      expect(modal.props("skill")).toEqual(mockSkills[0])
    })

    it("closes detail modal on close event", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("# Content")

      const wrapper = await mountSkillList()
      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("detail", mockSkills[0])
      await flushPromises()

      const modal = wrapper.findComponent(SkillDetailModal)
      await modal.vm.$emit("close")
      await flushPromises()

      expect(wrapper.findComponent(SkillDetailModal).exists()).toBe(false)
    })
  })

  describe("selection mode", () => {
    it("shows Select button in header", async () => {
      const wrapper = await mountSkillList()

      const selectBtn = wrapper.find(".header-actions button")
      expect(selectBtn.exists()).toBe(true)
      expect(selectBtn.text()).toBe("Select")
    })

    it("enters selection mode on Select click", async () => {
      const wrapper = await mountSkillList()

      await wrapper.find(".header-actions button").trigger("click")

      const buttons = wrapper.findAll(".header-actions button")
      expect(buttons.some((b) => b.text() === "Select All")).toBe(true)
      expect(buttons.some((b) => b.text() === "Cancel")).toBe(true)
    })

    it("passes selectable prop to skill cards in selection mode", async () => {
      const wrapper = await mountSkillList()

      await wrapper.find(".header-actions button").trigger("click")

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      cards.forEach((card) => {
        expect(card.props("selectable")).toBe(true)
      })
    })

    it("hides remove button in selection mode", async () => {
      const wrapper = await mountSkillList()

      await wrapper.find(".header-actions button").trigger("click")

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      cards.forEach((card) => {
        expect(card.props("showRemove")).toBe(false)
      })
    })

    it("exits selection mode on Cancel click", async () => {
      const wrapper = await mountSkillList()

      await wrapper.find(".header-actions button").trigger("click")

      const cancelBtn = wrapper
        .findAll(".header-actions button")
        .find((b) => b.text() === "Cancel")
      await cancelBtn!.trigger("click")

      const selectBtn = wrapper.find(".header-actions button")
      expect(selectBtn.text()).toBe("Select")
    })

    it("toggles individual skill selection", async () => {
      const wrapper = await mountSkillList()

      await wrapper.find(".header-actions button").trigger("click")

      const card = wrapper.findComponent({ name: "SkillCard" })
      await card.vm.$emit("toggle-select", "react-patterns")
      await wrapper.vm.$nextTick()

      const reactCard = wrapper
        .findAllComponents({ name: "SkillCard" })
        .find((c) => c.props("skill").name === "react-patterns")
      expect(reactCard?.props("selected")).toBe(true)
    })

    it("Select All selects all skills", async () => {
      const wrapper = await mountSkillList()

      await wrapper.find(".header-actions button").trigger("click")

      const selectAllBtn = wrapper
        .findAll(".header-actions button")
        .find((b) => b.text() === "Select All")
      await selectAllBtn!.trigger("click")

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      cards.forEach((card) => {
        expect(card.props("selected")).toBe(true)
      })
    })
  })

  describe("group selection", () => {
    it("shows group checkbox in selection mode", async () => {
      const wrapper = await mountSkillList()

      await wrapper.find(".header-actions button").trigger("click")

      expect(wrapper.find(".group-checkbox").exists()).toBe(true)
    })

    it("does not show group checkbox outside selection mode", async () => {
      const wrapper = await mountSkillList()

      expect(wrapper.find(".group-checkbox").exists()).toBe(false)
    })

    it("selects all skills in a group when group checkbox is clicked", async () => {
      // Use skills with 2 in same source to test group selection
      mockAppService.GetSkills.mockResolvedValue([
        mockSkills[0],
        {
          ...mockSkills[1],
          source: "vercel-labs/skills",
          sourceType: "github",
        },
      ])
      mockAppService.GetWorkspaces.mockResolvedValue([
        {
          name: "default",
          description: "",
          skills: ["react-patterns", "typescript-tips"],
          isActive: true,
        },
      ])

      const wrapper = await mountSkillList()

      await wrapper.find(".header-actions button").trigger("click")

      const checkbox = wrapper.find(".group-checkbox")
      await checkbox.trigger("click")

      const cards = wrapper.findAllComponents({ name: "SkillCard" })
      cards.forEach((card) => {
        expect(card.props("selected")).toBe(true)
      })
    })
  })

  describe("bulk remove from workspace", () => {
    async function enterSelectionAndSelect(
      wrapper: ReturnType<typeof mount>,
      skillNames: string[],
    ) {
      await wrapper.find(".header-actions button").trigger("click")
      for (const name of skillNames) {
        const card = wrapper
          .findAllComponents({ name: "SkillCard" })
          .find((c) => c.props("skill").name === name)
        await card!.vm.$emit("toggle-select", name)
      }
      await wrapper.vm.$nextTick()
    }

    it("shows bulk action bar when skills are selected", async () => {
      const wrapper = await mountSkillList()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      expect(wrapper.find(".bulk-action-bar").exists()).toBe(true)
      expect(wrapper.find(".bulk-count").text()).toBe("1 skill selected")
    })

    it("does not show bulk action bar when no skills selected", async () => {
      const wrapper = await mountSkillList()

      await wrapper.find(".header-actions button").trigger("click")

      expect(wrapper.find(".bulk-action-bar").exists()).toBe(false)
    })

    it("shows confirm dialog when remove button is clicked", async () => {
      const wrapper = await mountSkillList()

      await enterSelectionAndSelect(wrapper, [
        "react-patterns",
        "typescript-tips",
      ])

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      await removeBtn.trigger("click")
      await wrapper.vm.$nextTick()

      const dialogs = wrapper.findAllComponents(ConfirmDialog)
      const bulkDialog = dialogs.find(
        (d) => d.props("confirmLabel") === "Remove All",
      )
      expect(bulkDialog).toBeDefined()
      expect(bulkDialog!.props("message")).toContain("2 skills")
    })

    it("calls RemoveSkillFromWorkspace for each selected skill on confirm", async () => {
      mockAppService.RemoveSkillFromWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountSkillList()

      await enterSelectionAndSelect(wrapper, [
        "react-patterns",
        "typescript-tips",
      ])

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      await removeBtn.trigger("click")
      await wrapper.vm.$nextTick()

      const dialogs = wrapper.findAllComponents(ConfirmDialog)
      const bulkDialog = dialogs.find(
        (d) => d.props("confirmLabel") === "Remove All",
      )
      await bulkDialog!.vm.$emit("confirm")
      await flushPromises()

      expect(mockAppService.RemoveSkillFromWorkspace).toHaveBeenCalledWith(
        "react-patterns",
        "default",
      )
      expect(mockAppService.RemoveSkillFromWorkspace).toHaveBeenCalledWith(
        "typescript-tips",
        "default",
      )
      expect(mockAppService.RemoveSkillFromWorkspace).toHaveBeenCalledTimes(2)
    })

    it("exits selection mode after successful bulk remove", async () => {
      mockAppService.RemoveSkillFromWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountSkillList()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      await removeBtn.trigger("click")
      await wrapper.vm.$nextTick()

      const dialogs = wrapper.findAllComponents(ConfirmDialog)
      const bulkDialog = dialogs.find(
        (d) => d.props("confirmLabel") === "Remove All",
      )
      await bulkDialog!.vm.$emit("confirm")
      await flushPromises()

      const selectBtn = wrapper.find(".header-actions button")
      expect(selectBtn.text()).toBe("Select")
      expect(wrapper.find(".bulk-action-bar").exists()).toBe(false)
    })

    it("shows success toast after bulk remove", async () => {
      mockAppService.RemoveSkillFromWorkspace.mockResolvedValue(undefined)

      const wrapper = await mountSkillList()

      await enterSelectionAndSelect(wrapper, [
        "react-patterns",
        "typescript-tips",
      ])

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      await removeBtn.trigger("click")
      await wrapper.vm.$nextTick()

      const dialogs = wrapper.findAllComponents(ConfirmDialog)
      const bulkDialog = dialogs.find(
        (d) => d.props("confirmLabel") === "Remove All",
      )
      await bulkDialog!.vm.$emit("confirm")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("2 skills")
      expect(toast.props("message")).toContain("default")
      expect(toast.props("type")).toBe("success")
    })

    it("does not remove on cancel", async () => {
      const wrapper = await mountSkillList()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      await removeBtn.trigger("click")
      await wrapper.vm.$nextTick()

      const dialogs = wrapper.findAllComponents(ConfirmDialog)
      const bulkDialog = dialogs.find(
        (d) => d.props("confirmLabel") === "Remove All",
      )
      await bulkDialog!.vm.$emit("cancel")
      await flushPromises()

      expect(mockAppService.RemoveSkillFromWorkspace).not.toHaveBeenCalled()
    })

    it("shows error toast on bulk remove failure", async () => {
      mockAppService.RemoveSkillFromWorkspace.mockRejectedValue(
        new Error("Permission denied"),
      )

      const wrapper = await mountSkillList()

      await enterSelectionAndSelect(wrapper, ["react-patterns"])

      const removeBtn = wrapper.find(".bulk-action-bar .btn-danger")
      await removeBtn.trigger("click")
      await wrapper.vm.$nextTick()

      const dialogs = wrapper.findAllComponents(ConfirmDialog)
      const bulkDialog = dialogs.find(
        (d) => d.props("confirmLabel") === "Remove All",
      )
      await bulkDialog!.vm.$emit("confirm")
      await flushPromises()

      const toast = wrapper.findComponent({ name: "ToastNotification" })
      expect(toast.exists()).toBe(true)
      expect(toast.props("message")).toContain("Permission denied")
      expect(toast.props("type")).toBe("error")
    })
  })

  describe("source URL validation", () => {
    it("renders source as link for https URLs", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          ...mockSkills[0],
          sourceUrl: "https://github.com/vercel-labs/skills",
        },
      ])

      const wrapper = await mountSkillList()

      expect(wrapper.find(".group-source-link").exists()).toBe(true)
    })

    it("does not render source as link for SSH URLs", async () => {
      mockAppService.GetSkills.mockResolvedValue([
        {
          ...mockSkills[0],
          sourceUrl: "git@github.com:vercel-labs/skills.git",
        },
      ])

      const wrapper = await mountSkillList()

      expect(wrapper.find(".group-source-link").exists()).toBe(false)
      expect(wrapper.find(".group-source").exists()).toBe(true)
    })

    it("does not render source as link when sourceUrl is missing", async () => {
      const wrapper = await mountSkillList()

      expect(wrapper.find(".group-source-link").exists()).toBe(false)
    })
  })
})
