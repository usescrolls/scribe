export interface SkillInfo {
  name: string
  displayName?: string
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
  isSystem?: boolean
  agents: string[]
}

export interface SourceGroupCheckResult {
  source: string
  hasUpdates: boolean
  updatedSkillNames: string[]
  newAvailableSkills?: DiscoveredSkill[]
  checkedAt: string
  error?: string
}

export interface UpdateResult {
  skillName: string
  updated: boolean
  removed?: boolean
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

export interface ProgressEvent {
  phase: "discover" | "install"
  step: string
  message: string
  detail?: string
}

export interface DiscoverResult {
  skills: DiscoveredSkill[]
  source: string
  sourceType: string
}

export interface DiscoveredSkill {
  name: string
  description: string
  alreadyInstalled: boolean
}

export interface MarketplaceRepo {
  owner: string
  name: string
  fullName: string
  description: string
  url: string
  avatarUrl: string
  stars: number
  skillCount: number
  provider: string
  downloads?: number
  verified?: boolean
  category?: string
}

export interface MarketplaceResult {
  repos: MarketplaceRepo[]
  totalCount: number
}

export interface MarketplaceProviderInfo {
  id: string
  displayName: string
}

export interface SkillAudit {
  provider: string
  label: string
  result: string
}

export interface AuditAlert {
  type: string
  severity: string
  file: string | null
  description: string
  confidence: number | null
}

export interface AuditDetail {
  provider: string
  result: string
  riskLevel: string | null
  analysisHtml: string | null
  analyzedAt: string | null
  alerts: AuditAlert[] | null
  metadata: Record<string, string> | null
}

export interface SkillAuditResult {
  audits: SkillAudit[]
  auditDetails: AuditDetail[]
}
