export interface SkillInfo {
  name: string
  description: string
  source: string
  sourceType: string
  sourceUrl?: string
  installedAt: string
  updatedAt?: string
  contentHash?: string
  commitHash?: string
  commitDate?: string
  isPrivate?: boolean
  agents: string[]
}

export interface UpdateResult {
  skillName: string
  updated: boolean
  oldHash?: string
  newHash?: string
  commitDate?: string
}

export interface WorkspaceInfo {
  name: string
  description: string
  skills: string[]
  isActive: boolean
}

export interface AgentStatus {
  id: string
  displayName: string
  installed: boolean
  skillCount: number
  globalSkillsDir: string
}

export interface ExistingSkillInfo {
  name: string
  path: string
  agentId: string
  agentName: string
  isGitRepo: boolean
}

export interface SkillConflict {
  name: string
  sources: ExistingSkillInfo[]
}

export interface InstallResult {
  success: boolean
  skillsCount: number
  skillNames: string[]
  errorMessage: string
}

export interface DiscoverResult {
  skills: DiscoveredSkill[]
  source: string
  sourceType: string
}

export interface DiscoveredSkill {
  name: string
  description: string
}
