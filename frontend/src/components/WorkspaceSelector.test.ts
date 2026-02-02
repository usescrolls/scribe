import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import WorkspaceSelector from './WorkspaceSelector.vue'
import { mockAppService } from '../test/setup'
import type { WorkspaceInfo } from '../types/skill'

describe('WorkspaceSelector', () => {
  const mockWorkspaces: WorkspaceInfo[] = [
    {
      name: 'default',
      description: 'All skills',
      skills: ['skill-1', 'skill-2', 'skill-3'],
      isActive: true
    },
    {
      name: 'web-dev',
      description: 'Web development',
      skills: ['skill-1', 'skill-2'],
      isActive: false
    },
    {
      name: 'backend',
      description: 'Backend development',
      skills: ['skill-3'],
      isActive: false
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)
    mockAppService.SetActiveWorkspace.mockResolvedValue(undefined)
  })

  async function mountWorkspaceSelector() {
    const wrapper = mount(WorkspaceSelector)
    await flushPromises()
    return wrapper
  }

  describe('rendering', () => {
    it('renders a select element', async () => {
      const wrapper = await mountWorkspaceSelector()

      expect(wrapper.find('select').exists()).toBe(true)
    })

    it('renders all workspaces as options', async () => {
      const wrapper = await mountWorkspaceSelector()

      const options = wrapper.findAll('option')
      expect(options).toHaveLength(3)
    })

    it('shows workspace name and skill count', async () => {
      const wrapper = await mountWorkspaceSelector()

      const options = wrapper.findAll('option')
      expect(options[0].text()).toBe('default (3 skills)')
      expect(options[1].text()).toBe('web-dev (2 skills)')
      expect(options[2].text()).toBe('backend (1 skill)')
    })

    it('selects active workspace by default', async () => {
      const wrapper = await mountWorkspaceSelector()

      const select = wrapper.find('select')
      expect((select.element as HTMLSelectElement).value).toBe('default')
    })
  })

  describe('workspace switching', () => {
    it('calls SetActiveWorkspace on change', async () => {
      const wrapper = await mountWorkspaceSelector()

      const select = wrapper.find('select')
      await select.setValue('web-dev')

      expect(mockAppService.SetActiveWorkspace).toHaveBeenCalledWith('web-dev')
    })

    it('does not call API if selecting same workspace', async () => {
      const wrapper = await mountWorkspaceSelector()

      // Simulate selecting the already active workspace
      const select = wrapper.find('select')
      await select.setValue('default')

      // Should not be called since 'default' is already active
      expect(mockAppService.SetActiveWorkspace).not.toHaveBeenCalled()
    })
  })

  describe('loading state', () => {
    it('disables select while loading', async () => {
      // Make GetWorkspaces hang to simulate loading
      mockAppService.GetWorkspaces.mockImplementation(() => new Promise(() => {}))

      const wrapper = mount(WorkspaceSelector)

      const select = wrapper.find('select')
      expect((select.element as HTMLSelectElement).disabled).toBe(true)
    })
  })
})
