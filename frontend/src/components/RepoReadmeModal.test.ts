import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import RepoReadmeModal from "./RepoReadmeModal.vue"
import { mockAppService } from "../test/setup"
import type { MarketplaceRepo } from "../types/skill"

const agentHubRepo: MarketplaceRepo = {
  owner: "carol",
  name: "commit-helper",
  fullName: "carol/skills-repo",
  description: "AI-powered commit messages",
  url: "https://github.com/carol/skills-repo",
  avatarUrl: "https://github.com/carol.png",
  stars: 0,
  skillCount: 1,
  provider: "agenthub",
  downloads: 1500,
  verified: true,
  category: "skill",
}

const githubRepo: MarketplaceRepo = {
  owner: "alice",
  name: "skills",
  fullName: "alice/skills",
  description: "A collection of useful skills",
  url: "https://github.com/alice/skills",
  avatarUrl: "https://avatars.githubusercontent.com/u/alice",
  stars: 42,
  skillCount: 3,
  provider: "github",
}

function mountModal(repo: MarketplaceRepo = agentHubRepo) {
  return mount(RepoReadmeModal, {
    props: { repo },
    global: { stubs: { Teleport: true } },
  })
}

describe("RepoReadmeModal", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.GetRepoReadme.mockResolvedValue("# Hello")
    mockAppService.GetSkillAudits.mockResolvedValue({ audits: [] })
  })

  describe("audit display for agenthub skills", () => {
    it("fetches audits on mount for agenthub repos", async () => {
      mountModal()
      await flushPromises()

      expect(mockAppService.GetSkillAudits).toHaveBeenCalledWith(
        "carol",
        "skills-repo",
        "commit-helper",
      )
    })

    it("does not fetch audits for github repos", async () => {
      mountModal(githubRepo)
      await flushPromises()

      expect(mockAppService.GetSkillAudits).not.toHaveBeenCalled()
    })

    it("does not render audits section for github repos", async () => {
      const w = mountModal(githubRepo)
      await flushPromises()

      expect(w.find(".audits-section").exists()).toBe(false)
    })

    it("shows loading state while fetching audits", async () => {
      mockAppService.GetSkillAudits.mockReturnValue(new Promise(() => {}))
      const w = mountModal()
      await flushPromises()

      expect(w.find(".audits-loading").exists()).toBe(true)
      expect(w.text()).toContain("Loading audits")
    })

    it("renders audit badges when audits are returned", async () => {
      mockAppService.GetSkillAudits.mockResolvedValue({
        audits: [
          { provider: "vibesafe", label: "VibeSafe", result: "Pass" },
          { provider: "snyk", label: "Snyk", result: "Warn" },
        ],
      })
      const w = mountModal()
      await flushPromises()

      const badges = w.findAll(".audit-badge")
      expect(badges).toHaveLength(2)
      expect(badges[0].text()).toContain("VibeSafe")
      expect(badges[0].text()).toContain("Pass")
      expect(badges[0].classes()).toContain("audit-pass")
      expect(badges[1].text()).toContain("Snyk")
      expect(badges[1].text()).toContain("Warn")
      expect(badges[1].classes()).toContain("audit-warn")
    })

    it("applies audit-fail class for failed audits", async () => {
      mockAppService.GetSkillAudits.mockResolvedValue({
        audits: [{ provider: "scanner", label: "Scanner", result: "Fail" }],
      })
      const w = mountModal()
      await flushPromises()

      const badge = w.find(".audit-badge")
      expect(badge.classes()).toContain("audit-fail")
    })

    it("shows audits title with shield icon when audits exist", async () => {
      mockAppService.GetSkillAudits.mockResolvedValue({
        audits: [{ provider: "vibesafe", label: "VibeSafe", result: "Pass" }],
      })
      const w = mountModal()
      await flushPromises()

      expect(w.find(".audits-title").exists()).toBe(true)
      expect(w.find(".audits-title").text()).toContain("Security Audits")
      expect(w.find(".audits-title svg").exists()).toBe(true)
    })

    it("hides audits list when no audits returned", async () => {
      mockAppService.GetSkillAudits.mockResolvedValue({ audits: [] })
      const w = mountModal()
      await flushPromises()

      expect(w.find(".audits-list").exists()).toBe(false)
      expect(w.find(".audits-loading").exists()).toBe(false)
    })

    it("gracefully handles audit fetch failure", async () => {
      mockAppService.GetSkillAudits.mockRejectedValue(
        new Error("network error"),
      )
      const w = mountModal()
      await flushPromises()

      // Should not show audits, but also not crash
      expect(w.find(".audits-list").exists()).toBe(false)
      expect(w.find(".audits-loading").exists()).toBe(false)
      // README should still load normally
      expect(w.find(".markdown-content").exists()).toBe(true)
    })
  })

  describe("readme loading", () => {
    it("fetches README on mount", async () => {
      mountModal()
      await flushPromises()

      expect(mockAppService.GetRepoReadme).toHaveBeenCalledWith(
        "carol",
        "skills-repo",
      )
    })

    it("renders markdown content", async () => {
      mockAppService.GetRepoReadme.mockResolvedValue("# Test Heading")
      const w = mountModal()
      await flushPromises()

      expect(w.find(".markdown-content").exists()).toBe(true)
    })
  })
})
