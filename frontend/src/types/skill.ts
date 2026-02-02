export interface SkillInfo {
  name: string
  description: string
  source: string
  sourceType: string
  installedAt: string
  agents: string[]
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
