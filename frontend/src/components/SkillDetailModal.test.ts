import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import SkillDetailModal from "./SkillDetailModal.vue"
import { mockAppService, mockBrowser } from "../test/setup"
import type { SkillInfo } from "../types/skill"

describe("SkillDetailModal", () => {
  const mockSkill: SkillInfo = {
    name: "react-patterns",
    description: "React best practices and patterns",
    source: "vercel-labs/skills",
    sourceType: "github",
    installedAt: "2025-01-29T10:00:00Z",
    agents: ["claude-code"],
  }
  const prefixedSkill: SkillInfo = {
    ...mockSkill,
    name: "gitlab-nunomen-claude-skills--avoid-ai-tropes",
    displayName: "avoid-ai-tropes",
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  function mountModal(skill: SkillInfo = mockSkill) {
    return mount(SkillDetailModal, {
      props: { skill },
      global: {
        stubs: {
          Teleport: true,
          Transition: {
            setup(_, { slots }: { slots: any }) {
              return () => slots.default?.()
            },
          },
        },
      },
    })
  }

  describe("rendering", () => {
    it("shows skill name in header", () => {
      mockAppService.GetSkillContent.mockResolvedValue("")
      const wrapper = mountModal()

      expect(wrapper.find(".detail-title h3").text()).toBe("react-patterns")
    })

    it("prefers displayName in the header", () => {
      mockAppService.GetSkillContent.mockResolvedValue("")
      const wrapper = mountModal(prefixedSkill)

      expect(wrapper.find(".detail-title h3").text()).toBe("avoid-ai-tropes")
    })

    it("shows source type badge", () => {
      mockAppService.GetSkillContent.mockResolvedValue("")
      const wrapper = mountModal()

      expect(wrapper.find(".source-badge").text()).toBe("github")
    })

    it("shows description when available", () => {
      mockAppService.GetSkillContent.mockResolvedValue("")
      const wrapper = mountModal()

      expect(wrapper.find(".detail-description").text()).toBe(
        "React best practices and patterns",
      )
    })

    it("hides description when not available", () => {
      mockAppService.GetSkillContent.mockResolvedValue("")
      const wrapper = mountModal({ ...mockSkill, description: "" })

      expect(wrapper.find(".detail-description").exists()).toBe(false)
    })

    it("shows loading state initially", () => {
      mockAppService.GetSkillContent.mockReturnValue(new Promise(() => {}))
      const wrapper = mountModal()

      expect(wrapper.find(".loading").exists()).toBe(true)
      expect(wrapper.text()).toContain("Loading content...")
    })
  })

  describe("content loading", () => {
    it("calls GetSkillContent with skill name", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("# Hello")
      mountModal()
      await flushPromises()

      expect(mockAppService.GetSkillContent).toHaveBeenCalledWith(
        "react-patterns",
      )
    })

    it("uses the canonical name for content loading when displayName differs", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("# Hello")
      mountModal(prefixedSkill)
      await flushPromises()

      expect(mockAppService.GetSkillContent).toHaveBeenCalledWith(
        "gitlab-nunomen-claude-skills--avoid-ai-tropes",
      )
    })

    it("renders markdown content", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("# Hello World")
      const wrapper = mountModal()
      await flushPromises()

      const content = wrapper.find(".markdown-content")
      expect(content.exists()).toBe(true)
      expect(content.html()).toContain("<h1")
      expect(content.text()).toContain("Hello World")
    })

    it("shows error on load failure", async () => {
      mockAppService.GetSkillContent.mockRejectedValue(
        new Error("Skill not found"),
      )
      const wrapper = mountModal()
      await flushPromises()

      expect(wrapper.find(".error").exists()).toBe(true)
      expect(wrapper.text()).toContain("Skill not found")
    })

    it("shows empty state when content is empty", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("")
      const wrapper = mountModal()
      await flushPromises()

      expect(wrapper.find(".empty").exists()).toBe(true)
      expect(wrapper.text()).toContain("No content available")
    })
  })

  describe("link handling", () => {
    it("opens http links in system browser", async () => {
      mockAppService.GetSkillContent.mockResolvedValue(
        "[click me](https://example.com)",
      )
      const wrapper = mountModal()
      await flushPromises()

      const link = wrapper.find(".markdown-content a")
      expect(link.exists()).toBe(true)
      await link.trigger("click")

      expect(mockBrowser.OpenURL).toHaveBeenCalledWith("https://example.com")
    })

    it("does not call Browser.OpenURL for non-http links", async () => {
      mockAppService.GetSkillContent.mockResolvedValue(
        "[email](mailto:test@example.com)",
      )
      const wrapper = mountModal()
      await flushPromises()

      const link = wrapper.find(".markdown-content a")
      expect(link.exists()).toBe(true)
      await link.trigger("click")

      expect(mockBrowser.OpenURL).not.toHaveBeenCalled()
    })

    it("does not call Browser.OpenURL for relative links", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("[readme](./README.md)")
      const wrapper = mountModal()
      await flushPromises()

      const link = wrapper.find(".markdown-content a")
      expect(link.exists()).toBe(true)
      await link.trigger("click")

      expect(mockBrowser.OpenURL).not.toHaveBeenCalled()
    })

    it("does not call Browser.OpenURL for anchor links", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("[section](#section-id)")
      const wrapper = mountModal()
      await flushPromises()

      const link = wrapper.find(".markdown-content a")
      expect(link.exists()).toBe(true)
      await link.trigger("click")

      expect(mockBrowser.OpenURL).not.toHaveBeenCalled()
    })
  })

  describe("closing", () => {
    it("hides modal when close button is clicked", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("")
      const wrapper = mountModal()
      expect(wrapper.find(".detail-backdrop").exists()).toBe(true)

      await wrapper.find(".close-btn").trigger("click")
      await wrapper.vm.$nextTick()

      // visible becomes false, which removes the backdrop via v-if
      expect(wrapper.find(".detail-backdrop").exists()).toBe(false)
    })

    it("hides modal when backdrop is clicked", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("")
      const wrapper = mountModal()
      expect(wrapper.find(".detail-backdrop").exists()).toBe(true)

      await wrapper.find(".detail-backdrop").trigger("click")
      await wrapper.vm.$nextTick()

      expect(wrapper.find(".detail-backdrop").exists()).toBe(false)
    })

    it("hides modal on Escape key", async () => {
      mockAppService.GetSkillContent.mockResolvedValue("")
      const wrapper = mountModal()
      expect(wrapper.find(".detail-backdrop").exists()).toBe(true)

      document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }))
      await wrapper.vm.$nextTick()

      expect(wrapper.find(".detail-backdrop").exists()).toBe(false)
    })
  })
})
