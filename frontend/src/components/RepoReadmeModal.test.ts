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

    it("renders audit rows when audits are returned", async () => {
      mockAppService.GetSkillAudits.mockResolvedValue({
        audits: [
          { provider: "vibesafe", label: "VibeSafe", result: "Pass" },
          { provider: "snyk", label: "Snyk", result: "Warn" },
        ],
        auditDetails: [],
      })
      const w = mountModal()
      await flushPromises()

      const rows = w.findAll(".audit-row-wrapper")
      expect(rows).toHaveLength(2)
      expect(rows[0].find(".audit-row-label").text()).toBe("VibeSafe")
      expect(rows[0].find(".audit-badge").text()).toBe("Pass")
      expect(rows[0].find(".audit-badge").classes()).toContain("audit-pass")
      expect(rows[1].find(".audit-row-label").text()).toBe("Snyk")
      expect(rows[1].find(".audit-badge").text()).toBe("Warn")
      expect(rows[1].find(".audit-badge").classes()).toContain("audit-warn")
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

      expect(w.find(".audits-list-container").exists()).toBe(false)
      expect(w.find(".audits-loading").exists()).toBe(false)
    })

    it("gracefully handles audit fetch failure", async () => {
      mockAppService.GetSkillAudits.mockRejectedValue(
        new Error("network error"),
      )
      const w = mountModal()
      await flushPromises()

      // Should not show audits, but also not crash
      expect(w.find(".audits-list-container").exists()).toBe(false)
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
