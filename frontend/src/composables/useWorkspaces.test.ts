import { describe, it, expect, vi, beforeEach } from "vitest"
import { flushPromises } from "@vue/test-utils"
import {
  useWorkspaces,
  showSwitchInfoModal,
  switchInfoWorkspaceName,
  dismissSwitchInfo,
} from "./useWorkspaces"
import { mockAppService } from "../test/setup"
import type { WorkspaceInfo } from "../types/skill"

const SKIP_SWITCH_INFO_KEY = "scribe-skip-workspace-switch-info"

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key]
    }),
    clear: vi.fn(() => {
      store = {}
    }),
  }
})()

Object.defineProperty(globalThis, "localStorage", { value: localStorageMock })

describe("useWorkspaces", () => {
  const mockWorkspaces: WorkspaceInfo[] = [
    {
      name: "default",
      description: "All installed skills",
      skills: ["react-patterns", "typescript-tips", "go-patterns"],
      isActive: true,
    },
    {
      name: "web-dev",
      description: "Web development skills",
      skills: ["react-patterns", "typescript-tips"],
      isActive: false,
    },
    {
      name: "backend",
      description: "Backend skills",
      skills: ["go-patterns"],
      isActive: false,
    },
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    localStorageMock.clear()
    showSwitchInfoModal.value = false
    switchInfoWorkspaceName.value = ""
    mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)
  })

  describe("fetchWorkspaces", () => {
    it("fetches workspaces and sets active", async () => {
      const { workspaces, activeWorkspace, fetchWorkspaces } = useWorkspaces()

      await fetchWorkspaces()
      await flushPromises()

      expect(mockAppService.GetWorkspaces).toHaveBeenCalled()
      expect(workspaces.value).toEqual(mockWorkspaces)
      expect(activeWorkspace.value).toBe("default")
    })

    it("sets loading state during fetch", async () => {
      const { loading, fetchWorkspaces } = useWorkspaces()

      const fetchPromise = fetchWorkspaces()
      expect(loading.value).toBe(true)

      await fetchPromise
      expect(loading.value).toBe(false)
    })

    it("handles fetch errors", async () => {
      mockAppService.GetWorkspaces.mockRejectedValue(new Error("Network error"))

      const { error, fetchWorkspaces } = useWorkspaces()
      await fetchWorkspaces()

      expect(error.value).toBe("Network error")
    })
  })

  describe("switchWorkspace", () => {
    it("calls SetActiveWorkspace and refreshes", async () => {
      mockAppService.SetActiveWorkspace.mockResolvedValue(undefined)

      // After switching, the mock should return web-dev as active
      const workspacesAfterSwitch = mockWorkspaces.map((ws) => ({
        ...ws,
        isActive: ws.name === "web-dev",
      }))
      mockAppService.GetWorkspaces.mockResolvedValueOnce(mockWorkspaces) // initial fetch
        .mockResolvedValueOnce(workspacesAfterSwitch) // after switch

      const { switchWorkspace, activeWorkspace, fetchWorkspaces } =
        useWorkspaces()
      await fetchWorkspaces()

      const result = await switchWorkspace("web-dev")

      expect(mockAppService.SetActiveWorkspace).toHaveBeenCalledWith("web-dev")
      expect(result).toBe(true)
      expect(activeWorkspace.value).toBe("web-dev")
    })

    it("returns false and sets error on failure", async () => {
      mockAppService.SetActiveWorkspace.mockRejectedValue(
        new Error("Workspace not found"),
      )

      const { switchWorkspace, error, fetchWorkspaces } = useWorkspaces()
      await fetchWorkspaces()

      const result = await switchWorkspace("invalid")

      expect(result).toBe(false)
      expect(error.value).toBe("Workspace not found")
    })
  })

  describe("createWorkspace", () => {
    it("creates workspace and refreshes list", async () => {
      mockAppService.CreateWorkspace.mockResolvedValue(undefined)

      const { createWorkspace, fetchWorkspaces } = useWorkspaces()
      await fetchWorkspaces()

      const result = await createWorkspace(
        "my-workspace",
        "My custom workspace",
      )

      expect(mockAppService.CreateWorkspace).toHaveBeenCalledWith(
        "my-workspace",
        "My custom workspace",
      )
      expect(result).toBe(true)
    })

    it("handles empty description", async () => {
      mockAppService.CreateWorkspace.mockResolvedValue(undefined)

      const { createWorkspace } = useWorkspaces()
      await createWorkspace("minimal")

      expect(mockAppService.CreateWorkspace).toHaveBeenCalledWith("minimal", "")
    })

    it("returns false on error", async () => {
      mockAppService.CreateWorkspace.mockRejectedValue(
        new Error("Name already exists"),
      )

      const { createWorkspace, error } = useWorkspaces()
      const result = await createWorkspace("existing")

      expect(result).toBe(false)
      expect(error.value).toBe("Name already exists")
    })
  })

  describe("deleteWorkspace", () => {
    it("deletes workspace and refreshes list", async () => {
      mockAppService.DeleteWorkspace.mockResolvedValue(undefined)

      const { deleteWorkspace, fetchWorkspaces } = useWorkspaces()
      await fetchWorkspaces()

      const result = await deleteWorkspace("web-dev")

      expect(mockAppService.DeleteWorkspace).toHaveBeenCalledWith("web-dev")
      expect(result).toBe(true)
    })

    it("returns false and sets error on failure", async () => {
      mockAppService.DeleteWorkspace.mockRejectedValue(
        new Error("Cannot delete default workspace"),
      )

      const { deleteWorkspace, error } = useWorkspaces()
      const result = await deleteWorkspace("default")

      expect(result).toBe(false)
      expect(error.value).toBe("Cannot delete default workspace")
    })
  })

  describe("switch info modal", () => {
    it("shows info modal after successful switch", async () => {
      mockAppService.SetActiveWorkspace.mockResolvedValue(undefined)
      mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)

      const { switchWorkspace, fetchWorkspaces } = useWorkspaces()
      await fetchWorkspaces()

      await switchWorkspace("web-dev")

      expect(showSwitchInfoModal.value).toBe(true)
      expect(switchInfoWorkspaceName.value).toBe("web-dev")
    })

    it("does not show info modal when localStorage preference is set", async () => {
      localStorageMock.getItem.mockReturnValue("true")
      mockAppService.SetActiveWorkspace.mockResolvedValue(undefined)
      mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)

      const { switchWorkspace, fetchWorkspaces } = useWorkspaces()
      await fetchWorkspaces()

      await switchWorkspace("web-dev")

      expect(showSwitchInfoModal.value).toBe(false)
    })

    it("does not show info modal when switch fails", async () => {
      mockAppService.SetActiveWorkspace.mockRejectedValue(new Error("Failed"))

      const { switchWorkspace, fetchWorkspaces } = useWorkspaces()
      await fetchWorkspaces()

      await switchWorkspace("web-dev")

      expect(showSwitchInfoModal.value).toBe(false)
    })

    it("dismissSwitchInfo hides modal", () => {
      showSwitchInfoModal.value = true

      dismissSwitchInfo(false)

      expect(showSwitchInfoModal.value).toBe(false)
    })

    it("dismissSwitchInfo saves preference when dontShowAgain is true", () => {
      showSwitchInfoModal.value = true

      dismissSwitchInfo(true)

      expect(showSwitchInfoModal.value).toBe(false)
      expect(localStorageMock.setItem).toHaveBeenCalledWith(
        SKIP_SWITCH_INFO_KEY,
        "true",
      )
    })

    it("dismissSwitchInfo does not save preference when dontShowAgain is false", () => {
      showSwitchInfoModal.value = true

      dismissSwitchInfo(false)

      expect(localStorageMock.setItem).not.toHaveBeenCalled()
    })
  })

  describe("initial state", () => {
    it("starts with default active workspace", () => {
      const { activeWorkspace } = useWorkspaces()
      expect(activeWorkspace.value).toBe("default")
    })

    it("starts with loading true", () => {
      const { loading } = useWorkspaces()
      expect(loading.value).toBe(true)
    })
  })
})
