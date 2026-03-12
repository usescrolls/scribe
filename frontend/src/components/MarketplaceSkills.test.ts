import { describe, it, expect, vi, beforeEach } from "vitest"
import { mount, flushPromises } from "@vue/test-utils"
import MarketplaceSkills from "./MarketplaceSkills.vue"
import { mockAppService, mockBrowser } from "../test/setup"
import type { MarketplaceResult, MarketplaceRepo } from "../types/skill"

const defaultProviders = [
  { id: "agenthub", displayName: "skills.sh" },
  { id: "github", displayName: "GitHub" },
]

const githubResults: MarketplaceResult = {
  repos: [
    {
      owner: "alice",
      name: "skills",
      fullName: "alice/skills",
      description: "A collection of useful skills",
      url: "https://github.com/alice/skills",
      avatarUrl: "https://avatars.githubusercontent.com/u/alice",
      stars: 42,
      skillCount: 3,
      provider: "github",
    },
    {
      owner: "bob",
      name: "tools",
      fullName: "bob/tools",
      description: "Handy tools",
      url: "https://github.com/bob/tools",
      avatarUrl: "https://avatars.githubusercontent.com/u/bob",
      stars: 1500,
      skillCount: 1,
      provider: "github",
    },
  ],
  totalCount: 2,
}

const agentHubResults: MarketplaceResult = {
  repos: [
    {
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
    },
    {
      owner: "dave",
      name: "code-review",
      fullName: "dave/review-tools",
      description: "Automated code review",
      url: "https://github.com/dave/review-tools",
      avatarUrl: "https://github.com/dave.png",
      stars: 0,
      skillCount: 1,
      provider: "agenthub",
      downloads: 250,
      verified: false,
      category: "workflow",
    },
  ],
  totalCount: 2,
}

const emptyResults: MarketplaceResult = { repos: [], totalCount: 0 }

describe("MarketplaceSkills", () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.GetMarketplaceProviders.mockResolvedValue(defaultProviders)
    // Default: auto-search on mount returns empty (agenthub is default provider)
    mockAppService.SearchMarketplace.mockResolvedValue(emptyResults)
  })

  function mountMarketplace() {
    return mount(MarketplaceSkills)
  }

  describe("provider tabs", () => {
    it("renders provider tabs from GetMarketplaceProviders", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      const tabs = wrapper.findAll(".provider-tab")
      expect(tabs).toHaveLength(2)
      expect(tabs[0].text()).toBe("skills.sh")
      expect(tabs[1].text()).toBe("GitHub")
    })

    it("marks agenthub tab as active by default", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      const tabs = wrapper.findAll(".provider-tab")
      expect(tabs[0].classes()).toContain("active")
      expect(tabs[1].classes()).not.toContain("active")
    })

    it("falls back to hardcoded providers if GetMarketplaceProviders fails", async () => {
      mockAppService.GetMarketplaceProviders.mockRejectedValue(
        new Error("network error"),
      )
      const wrapper = mountMarketplace()
      await flushPromises()

      const tabs = wrapper.findAll(".provider-tab")
      expect(tabs).toHaveLength(2)
      expect(tabs[0].text()).toBe("skills.sh")
      expect(tabs[1].text()).toBe("GitHub")
    })

    it("switches active tab on click", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      const tabs = wrapper.findAll(".provider-tab")
      await tabs[1].trigger("click")
      await flushPromises()

      expect(tabs[0].classes()).not.toContain("active")
      expect(tabs[1].classes()).toContain("active")
    })

    it("searches with new provider after switching tabs", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()
      mockAppService.SearchMarketplace.mockClear()

      mockAppService.SearchMarketplace.mockResolvedValue(githubResults)
      const tabs = wrapper.findAll(".provider-tab")
      await tabs[1].trigger("click")
      await flushPromises()

      expect(mockAppService.SearchMarketplace).toHaveBeenCalledWith(
        "github",
        "",
        1,
      )
    })

    it("clears query and results when switching providers", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      // Type a query
      await wrapper.find("input").setValue("test")

      // Switch to GitHub
      mockAppService.SearchMarketplace.mockResolvedValue(emptyResults)
      const tabs = wrapper.findAll(".provider-tab")
      await tabs[1].trigger("click")
      await flushPromises()

      expect((wrapper.find("input").element as HTMLInputElement).value).toBe("")
    })

    it("does nothing when clicking the already-active tab", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()
      mockAppService.SearchMarketplace.mockClear()

      const tabs = wrapper.findAll(".provider-tab")
      await tabs[0].trigger("click")
      await flushPromises()

      // Should not trigger a new search
      expect(mockAppService.SearchMarketplace).not.toHaveBeenCalled()
    })
  })

  describe("initial state", () => {
    it("shows search input and button", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.find("input").exists()).toBe(true)
      expect(wrapper.find(".search-btn").exists()).toBe(true)
      expect(wrapper.find(".search-btn").text()).toContain("Search")
    })

    it("auto-searches on mount with agenthub provider", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      mountMarketplace()
      await flushPromises()

      expect(mockAppService.SearchMarketplace).toHaveBeenCalledWith(
        "agenthub",
        "",
        1,
      )
    })

    it("shows results from auto-search", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.findAll(".repo-card")).toHaveLength(2)
    })
  })

  describe("search flow", () => {
    it("calls SearchMarketplace with active provider, query, and page on click", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()
      mockAppService.SearchMarketplace.mockClear()

      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      await wrapper.find("input").setValue("commit")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(mockAppService.SearchMarketplace).toHaveBeenCalledWith(
        "agenthub",
        "commit",
        1,
      )
    })

    it("calls SearchMarketplace on Enter key", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()
      mockAppService.SearchMarketplace.mockClear()

      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      await wrapper.find("input").setValue("review")
      await wrapper.find("input").trigger("keyup.enter")
      await flushPromises()

      expect(mockAppService.SearchMarketplace).toHaveBeenCalledWith(
        "agenthub",
        "review",
        1,
      )
    })

    it("uses github provider when GitHub tab is active", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      // Switch to GitHub
      mockAppService.SearchMarketplace.mockResolvedValue(emptyResults)
      const tabs = wrapper.findAll(".provider-tab")
      await tabs[1].trigger("click")
      await flushPromises()
      mockAppService.SearchMarketplace.mockClear()

      mockAppService.SearchMarketplace.mockResolvedValue(githubResults)
      await wrapper.find("input").setValue("commit")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(mockAppService.SearchMarketplace).toHaveBeenCalledWith(
        "github",
        "commit",
        1,
      )
    })

    it("shows loading state while searching", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      mockAppService.SearchMarketplace.mockReturnValue(new Promise(() => {}))
      await wrapper.find("input").setValue("test")
      await wrapper.find(".search-btn").trigger("click")

      expect(wrapper.find(".spinner-sm").exists()).toBe(true)
      expect(wrapper.find(".search-btn").text()).toContain("Searching")
      expect((wrapper.find("input").element as HTMLInputElement).disabled).toBe(
        true,
      )
    })

    it("shows error on search failure", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      mockAppService.SearchMarketplace.mockRejectedValue(
        new Error("API rate limit exceeded"),
      )
      await wrapper.find("input").setValue("test")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-error").exists()).toBe(true)
      expect(wrapper.text()).toContain("rate limit")
    })

    it("dismisses error on close button click", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()

      mockAppService.SearchMarketplace.mockRejectedValue(
        new Error("Network error"),
      )
      await wrapper.find("input").setValue("test")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".result-error").exists()).toBe(true)
      await wrapper.find(".dismiss-btn").trigger("click")
      expect(wrapper.find(".result-error").exists()).toBe(false)
    })
  })

  describe("results display", () => {
    async function mountWithGithubResults() {
      mockAppService.SearchMarketplace.mockResolvedValueOnce(emptyResults) // initial agenthub load
        .mockResolvedValueOnce(githubResults) // after switching to github
      const wrapper = mountMarketplace()
      await flushPromises()
      // Switch to GitHub
      const tabs = wrapper.findAll(".provider-tab")
      await tabs[1].trigger("click")
      await flushPromises()
      return wrapper
    }

    it("shows results count", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.find(".results-count").text()).toContain("2 results")
    })

    it("renders repo cards", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.findAll(".repo-card")).toHaveLength(2)
    })

    it("shows fullName as repo name for github provider", async () => {
      const wrapper = await mountWithGithubResults()

      const names = wrapper.findAll(".repo-name")
      expect(names[0].text()).toBe("alice/skills")
      expect(names[1].text()).toBe("bob/tools")
    })

    it("shows skill name (not fullName) as repo name for agenthub provider", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      const names = wrapper.findAll(".repo-name")
      expect(names[0].text()).toBe("commit-helper")
      expect(names[1].text()).toBe("code-review")
    })

    it("shows fullName as secondary source label for agenthub repos", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      const sources = wrapper.findAll(".repo-source")
      expect(sources).toHaveLength(2)
      expect(sources[0].text()).toBe("carol/skills-repo")
    })

    it("does not show source label for github repos", async () => {
      const wrapper = await mountWithGithubResults()

      expect(wrapper.findAll(".repo-source")).toHaveLength(0)
    })

    it("shows repo description", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      const descs = wrapper.findAll(".repo-desc")
      expect(descs[0].text()).toBe("AI-powered commit messages")
    })

    it("shows avatar image", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      const avatars = wrapper.findAll(".repo-avatar")
      expect(avatars).toHaveLength(2)
      expect(avatars[0].attributes("src")).toBe("https://github.com/carol.png")
    })

    it("shows stars badge for github repos with stars", async () => {
      const wrapper = await mountWithGithubResults()

      const badges = wrapper.findAll(".badge-stars")
      expect(badges).toHaveLength(2)
      expect(badges[0].text()).toBe("42")
      expect(badges[1].text()).toBe("1.5k")
      expect(badges[0].find("svg").exists()).toBe(true)
    })

    it("hides stars badge when stars is 0", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.findAll(".badge-stars")).toHaveLength(0)
    })

    it("shows install button on each card", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      const buttons = wrapper.findAll(".btn-primary")
      expect(buttons).toHaveLength(2)
      expect(buttons[0].text()).toBe("Install")
    })

    it("shows link button on each card", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.findAll(".btn-icon")).toHaveLength(2)
    })

    it("shows singular 'result' for count of 1", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue({
        repos: [agentHubResults.repos[0]],
        totalCount: 1,
      })

      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.find(".results-count").text()).toContain("1 result")
      expect(wrapper.find(".results-count").text()).not.toContain("1 results")
    })
  })

  describe("agenthub-specific display", () => {
    it("shows verified badge for verified repos", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      const cards = wrapper.findAll(".repo-card")
      // First repo is verified
      expect(cards[0].find(".verified-icon").exists()).toBe(true)
      // Second repo is not verified
      expect(cards[1].find(".verified-icon").exists()).toBe(false)
    })

    it("shows downloads badge with formatted count", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      const badges = wrapper.findAll(".badge-downloads")
      expect(badges).toHaveLength(2)
      expect(badges[0].text()).toBe("1.5k") // 1500
      expect(badges[1].text()).toBe("250") // 250
    })

    it("hides downloads badge when downloads is 0 or undefined", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue({
        repos: [
          {
            ...agentHubResults.repos[0],
            downloads: 0,
          } as MarketplaceRepo,
        ],
        totalCount: 1,
      })
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.findAll(".badge-downloads")).toHaveLength(0)
    })

    it("shows category badge for non-skill categories", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      const categoryBadges = wrapper.findAll(".badge-category")
      // First repo has category "skill" — should be hidden
      // Second repo has category "workflow" — should be shown
      expect(categoryBadges).toHaveLength(1)
      expect(categoryBadges[0].text()).toBe("workflow")
    })

    it("renders distinct cards for multiple skills from the same repo", async () => {
      const sameRepoSkills: MarketplaceResult = {
        repos: [
          {
            ...agentHubResults.repos[0],
            name: "skill-a",
            fullName: "carol/skills-repo",
          },
          {
            ...agentHubResults.repos[0],
            name: "skill-b",
            fullName: "carol/skills-repo",
          },
        ],
        totalCount: 2,
      }
      mockAppService.SearchMarketplace.mockResolvedValue(sameRepoSkills)
      const wrapper = mountMarketplace()
      await flushPromises()

      const cards = wrapper.findAll(".repo-card")
      expect(cards).toHaveLength(2)
      const names = wrapper.findAll(".repo-name")
      expect(names[0].text()).toBe("skill-a")
      expect(names[1].text()).toBe("skill-b")
    })

    it("hides category badge when category is 'skill'", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue({
        repos: [agentHubResults.repos[0]], // category: "skill"
        totalCount: 1,
      })
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.findAll(".badge-category")).toHaveLength(0)
    })
  })

  describe("stale response guards", () => {
    it("discards results from a previous provider if user switched mid-request", async () => {
      // Mount and let initial load complete
      mockAppService.SearchMarketplace.mockResolvedValue(emptyResults)
      const wrapper = mountMarketplace()
      await flushPromises()
      mockAppService.SearchMarketplace.mockClear()

      // Start a slow agenthub search
      let resolveSlowSearch: (v: MarketplaceResult) => void
      const slowSearch = new Promise<MarketplaceResult>((r) => {
        resolveSlowSearch = r
      })
      mockAppService.SearchMarketplace.mockReturnValueOnce(slowSearch)

      await wrapper.find("input").setValue("test")
      await wrapper.find(".search-btn").trigger("click")
      // Search is in-flight for agenthub

      // Switch to GitHub before agenthub responds
      mockAppService.SearchMarketplace.mockResolvedValue(githubResults)
      const tabs = wrapper.findAll(".provider-tab")
      await tabs[1].trigger("click")
      await flushPromises()

      // Now resolve the old agenthub search
      resolveSlowSearch!(agentHubResults)
      await flushPromises()

      // Should show GitHub results, not the stale agenthub ones
      const names = wrapper.findAll(".repo-name")
      expect(names[0].text()).toBe("alice/skills")
    })
  })

  describe("no results", () => {
    it("shows no results message when search returns empty", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()
      mockAppService.SearchMarketplace.mockClear()

      mockAppService.SearchMarketplace.mockResolvedValue(emptyResults)
      await wrapper.find("input").setValue("nonexistent-xyz")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".empty-state").exists()).toBe(true)
      expect(wrapper.text()).toContain("nonexistent-xyz")
    })
  })

  describe("install flow", () => {
    async function mountWithResults() {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()
      return wrapper
    }

    it("emits install-from-source with repo fullName on install click", async () => {
      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[0].trigger("click")

      expect(wrapper.emitted("install-from-source")).toEqual([
        ["carol/skills-repo"],
      ])
    })

    it("emits correct fullName for different repos", async () => {
      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[1].trigger("click")

      expect(wrapper.emitted("install-from-source")).toEqual([
        ["dave/review-tools"],
      ])
    })

    it("does not call InstallFromSource directly", async () => {
      const wrapper = await mountWithResults()
      const installBtns = wrapper.findAll(".btn-primary")
      await installBtns[0].trigger("click")
      await flushPromises()

      expect(mockAppService.InstallFromSource).not.toHaveBeenCalled()
    })
  })

  describe("link button", () => {
    it("opens repo URL in browser on link button click", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      const linkBtns = wrapper.findAll(".btn-icon")
      await linkBtns[0].trigger("click")

      expect(mockBrowser.OpenURL).toHaveBeenCalledWith(
        "https://github.com/carol/skills-repo",
      )
    })
  })

  describe("card click opens detail", () => {
    it("shows RepoReadmeModal on card click", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults)
      mockAppService.GetRepoReadme.mockResolvedValue("# README content")
      const wrapper = mountMarketplace()
      await flushPromises()

      const cards = wrapper.findAll(".repo-card")
      await cards[0].trigger("click")
      await flushPromises()

      expect(mockAppService.GetRepoReadme).toHaveBeenCalledWith(
        "carol",
        "skills-repo",
      )
    })
  })

  describe("pagination", () => {
    const manyResults: MarketplaceResult = {
      repos: agentHubResults.repos,
      totalCount: 90, // 3 pages at 30 per page
    }

    it("shows pagination controls when totalCount exceeds page size", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(manyResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.find(".pagination").exists()).toBe(true)
      expect(wrapper.find(".page-info").text()).toBe("1 / 3")
    })

    it("hides pagination controls when results fit one page", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(agentHubResults) // totalCount: 2
      const wrapper = mountMarketplace()
      await flushPromises()

      expect(wrapper.find(".pagination").exists()).toBe(false)
    })

    it("disables prev button on first page", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(manyResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      const prevBtn = wrapper.find(".pagination .page-btn")
      expect((prevBtn.element as HTMLButtonElement).disabled).toBe(true)
    })

    it("navigates to next page on next button click", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(manyResults)
      const wrapper = mountMarketplace()
      await flushPromises()
      mockAppService.SearchMarketplace.mockClear()

      mockAppService.SearchMarketplace.mockResolvedValue(manyResults)
      const nextBtn = wrapper.findAll(".pagination .page-btn")[1]
      await nextBtn.trigger("click")
      await flushPromises()

      expect(mockAppService.SearchMarketplace).toHaveBeenCalledWith(
        "agenthub",
        "",
        2,
      )
      expect(wrapper.find(".page-info").text()).toBe("2 / 3")
    })

    it("disables next button on last page", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(manyResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      // Navigate to page 3
      mockAppService.SearchMarketplace.mockResolvedValue(manyResults)
      const nextBtn = wrapper.findAll(".pagination .page-btn")[1]
      await nextBtn.trigger("click")
      await flushPromises()
      await nextBtn.trigger("click")
      await flushPromises()

      expect(wrapper.find(".page-info").text()).toBe("3 / 3")
      expect((nextBtn.element as HTMLButtonElement).disabled).toBe(true)
    })

    it("resets to page 1 on new search", async () => {
      mockAppService.SearchMarketplace.mockResolvedValue(manyResults)
      const wrapper = mountMarketplace()
      await flushPromises()

      // Go to page 2
      const nextBtn = wrapper.findAll(".pagination .page-btn")[1]
      await nextBtn.trigger("click")
      await flushPromises()
      expect(wrapper.find(".page-info").text()).toBe("2 / 3")

      // New search
      mockAppService.SearchMarketplace.mockClear()
      mockAppService.SearchMarketplace.mockResolvedValue(manyResults)
      await wrapper.find("input").setValue("test")
      await wrapper.find(".search-btn").trigger("click")
      await flushPromises()

      expect(wrapper.find(".page-info").text()).toBe("1 / 3")
      expect(mockAppService.SearchMarketplace).toHaveBeenCalledWith(
        "agenthub",
        "test",
        1,
      )
    })
  })

  describe("cleanup on unmount", () => {
    it("does not throw when unmounting", async () => {
      const wrapper = mountMarketplace()
      await flushPromises()
      expect(() => wrapper.unmount()).not.toThrow()
    })
  })
})
