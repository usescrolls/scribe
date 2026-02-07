import { vi, beforeEach } from "vitest"

// Mock AppService from Wails bindings
export const mockAppService = {
  GetSkills: vi.fn(),
  GetSkillCount: vi.fn(),
  RemoveSkill: vi.fn(),
  GetWorkspaces: vi.fn(),
  GetActiveWorkspaceName: vi.fn(),
  SetActiveWorkspace: vi.fn(),
  CreateWorkspace: vi.fn(),
  DeleteWorkspace: vi.fn(),
  AddSkillToWorkspace: vi.fn(),
  RemoveSkillFromWorkspace: vi.fn(),
  GetAgentStatus: vi.fn(),
  GetInstalledAgentCount: vi.fn(),
  GetTotalAgentCount: vi.fn(),
  GetVersion: vi.fn(),
  GetPlugins: vi.fn(),
  UninstallPlugin: vi.fn(),
  IsOnboardingCompleted: vi.fn(),
  CompleteOnboarding: vi.fn(),
  DetectExistingSkills: vi.fn(),
  DetectSkillConflicts: vi.fn(),
  ImportAllExistingSkills: vi.fn(),
  DeleteAllExistingSkills: vi.fn(),
  ResolveSkillConflict: vi.fn(),
  InstallDemoSkill: vi.fn(),
  InstallFromSource: vi.fn(),
}

// Mock the scribe bindings module
vi.mock("../bindings/scribe", () => ({
  AppService: mockAppService,
}))

// Mock Events from @wailsio/runtime
export const mockEvents = {
  On: vi.fn(() => vi.fn()), // Returns unsubscribe function
  Emit: vi.fn(),
}

vi.mock("@wailsio/runtime", () => ({
  Events: mockEvents,
  Dialogs: {
    Question: vi.fn(),
  },
}))

// Reset all mocks before each test
beforeEach(() => {
  vi.clearAllMocks()
})
