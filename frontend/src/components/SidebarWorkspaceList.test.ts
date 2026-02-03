import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import SidebarWorkspaceList from './SidebarWorkspaceList.vue'
import { mockAppService } from '../test/setup'
import type { WorkspaceInfo } from '../types/skill'

const SKIP_CONFIRM_KEY = 'scribe-skip-workspace-switch-confirm'

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value }),
    removeItem: vi.fn((key: string) => { delete store[key] }),
    clear: vi.fn(() => { store = {} })
  }
})()

Object.defineProperty(global, 'localStorage', { value: localStorageMock })

describe('SidebarWorkspaceList', () => {
  const mockWorkspaces: WorkspaceInfo[] = [
    {
      name: 'default',
      description: 'Default workspace',
      skills: ['skill1', 'skill2', 'skill3'],
      isActive: true
    },
    {
      name: 'work',
      description: 'Work projects',
      skills: ['skill4'],
      isActive: false
    },
    {
      name: 'personal',
      description: 'Personal projects',
      skills: [],
      isActive: false
    }
  ]

  beforeEach(() => {
    vi.clearAllMocks()
    localStorageMock.clear()
    mockAppService.GetWorkspaces.mockResolvedValue(mockWorkspaces)
    mockAppService.SetActiveWorkspace.mockResolvedValue(undefined)
    mockAppService.CreateWorkspace.mockResolvedValue(undefined)
    mockAppService.DeleteWorkspace.mockResolvedValue(undefined)
  })

  async function mountComponent() {
    const wrapper = mount(SidebarWorkspaceList)
    await flushPromises()
    return wrapper
  }

  describe('rendering', () => {
    it('renders panel header with title', async () => {
      const wrapper = await mountComponent()

      expect(wrapper.find('.panel-header h3').text()).toBe('Workspaces')
    })

    it('shows total count in summary', async () => {
      const wrapper = await mountComponent()

      expect(wrapper.find('.summary').text()).toBe('3 total')
    })

    it('renders all workspaces', async () => {
      const wrapper = await mountComponent()

      const items = wrapper.findAll('.workspace-item')
      expect(items).toHaveLength(3)
    })

    it('shows checkmark for active workspace', async () => {
      const wrapper = await mountComponent()

      const checkmarks = wrapper.findAll('.checkmark')
      expect(checkmarks).toHaveLength(1)
    })

    it('shows empty circle for inactive workspaces', async () => {
      const wrapper = await mountComponent()

      const emptyCircles = wrapper.findAll('.empty')
      expect(emptyCircles).toHaveLength(2)
    })

    it('shows skill count for each workspace', async () => {
      const wrapper = await mountComponent()

      const skillCounts = wrapper.findAll('.skill-count')
      expect(skillCounts[0].text()).toBe('3 skills')
      expect(skillCounts[1].text()).toBe('1 skill')
      expect(skillCounts[2].text()).toBe('0 skills')
    })

    it('applies active class to active workspace', async () => {
      const wrapper = await mountComponent()

      const activeItem = wrapper.findAll('.workspace-item')[0]
      expect(activeItem.classes()).toContain('active')
    })

    it('shows delete button for non-default workspaces', async () => {
      const wrapper = await mountComponent()

      const deleteButtons = wrapper.findAll('.delete-btn')
      expect(deleteButtons).toHaveLength(2) // work and personal, not default
    })

    it('does not show delete button for default workspace', async () => {
      const wrapper = await mountComponent()

      const firstItem = wrapper.findAll('.workspace-item')[0]
      expect(firstItem.find('.delete-btn').exists()).toBe(false)
    })
  })

  describe('add workspace', () => {
    it('shows add button in header', async () => {
      const wrapper = await mountComponent()

      expect(wrapper.find('.add-btn').exists()).toBe(true)
      expect(wrapper.find('.add-btn').text()).toBe('+')
    })

    it('toggles add mode on button click', async () => {
      const wrapper = await mountComponent()

      // Initially no form
      expect(wrapper.find('.add-workspace-form').exists()).toBe(false)

      // Click to show form
      await wrapper.find('.add-btn').trigger('click')
      expect(wrapper.find('.add-workspace-form').exists()).toBe(true)
      expect(wrapper.find('.add-btn').text()).toBe('×')

      // Click to hide form
      await wrapper.find('.add-btn').trigger('click')
      expect(wrapper.find('.add-workspace-form').exists()).toBe(false)
    })

    it('creates workspace on form submit', async () => {
      const wrapper = await mountComponent()

      await wrapper.find('.add-btn').trigger('click')

      const input = wrapper.find('.add-workspace-form input')
      await input.setValue('new-workspace')
      await wrapper.find('.create-btn').trigger('click')
      await flushPromises()

      expect(mockAppService.CreateWorkspace).toHaveBeenCalledWith('new-workspace', '')
    })

    it('creates workspace on Enter key', async () => {
      const wrapper = await mountComponent()

      await wrapper.find('.add-btn').trigger('click')

      const input = wrapper.find('.add-workspace-form input')
      await input.setValue('new-workspace')
      await input.trigger('keyup.enter')
      await flushPromises()

      expect(mockAppService.CreateWorkspace).toHaveBeenCalledWith('new-workspace', '')
    })

    it('cancels add mode on Escape key', async () => {
      const wrapper = await mountComponent()

      await wrapper.find('.add-btn').trigger('click')
      expect(wrapper.find('.add-workspace-form').exists()).toBe(true)

      const input = wrapper.find('.add-workspace-form input')
      await input.trigger('keyup.escape')

      expect(wrapper.find('.add-workspace-form').exists()).toBe(false)
    })

    it('disables create button when input is empty', async () => {
      const wrapper = await mountComponent()

      await wrapper.find('.add-btn').trigger('click')

      const createBtn = wrapper.find('.create-btn')
      expect(createBtn.attributes('disabled')).toBeDefined()
    })

    it('does not create workspace with empty name', async () => {
      const wrapper = await mountComponent()

      await wrapper.find('.add-btn').trigger('click')
      await wrapper.find('.create-btn').trigger('click')

      expect(mockAppService.CreateWorkspace).not.toHaveBeenCalled()
    })

    it('closes form after successful creation', async () => {
      const wrapper = await mountComponent()

      await wrapper.find('.add-btn').trigger('click')

      const input = wrapper.find('.add-workspace-form input')
      await input.setValue('new-workspace')
      await wrapper.find('.create-btn').trigger('click')
      await flushPromises()

      expect(wrapper.find('.add-workspace-form').exists()).toBe(false)
    })
  })

  describe('delete workspace', () => {
    it('shows delete confirmation on delete button click', async () => {
      const wrapper = await mountComponent()

      const deleteBtn = wrapper.findAll('.delete-btn')[0]
      await deleteBtn.trigger('click')

      expect(wrapper.find('.delete-confirm').exists()).toBe(true)
      expect(wrapper.find('.confirm-message').text()).toContain('work')
    })

    it('deletes workspace on confirm', async () => {
      const wrapper = await mountComponent()

      const deleteBtn = wrapper.findAll('.delete-btn')[0]
      await deleteBtn.trigger('click')

      await wrapper.find('.delete-confirm-btn').trigger('click')
      await flushPromises()

      expect(mockAppService.DeleteWorkspace).toHaveBeenCalledWith('work')
    })

    it('cancels deletion on cancel button', async () => {
      const wrapper = await mountComponent()

      const deleteBtn = wrapper.findAll('.delete-btn')[0]
      await deleteBtn.trigger('click')

      await wrapper.find('.cancel-btn').trigger('click')

      expect(wrapper.find('.delete-confirm').exists()).toBe(false)
      expect(mockAppService.DeleteWorkspace).not.toHaveBeenCalled()
    })

    it('shows note about skills remaining installed', async () => {
      const wrapper = await mountComponent()

      const deleteBtn = wrapper.findAll('.delete-btn')[0]
      await deleteBtn.trigger('click')

      expect(wrapper.find('.confirm-note').text()).toContain('Skills will remain installed')
    })
  })

  describe('switch workspace', () => {
    it('shows confirmation dialog when clicking inactive workspace', async () => {
      const wrapper = await mountComponent()

      const workItem = wrapper.findAll('.workspace-item')[1] // 'work' workspace
      await workItem.trigger('click')

      expect(wrapper.find('.confirm-dialog').exists()).toBe(true)
      expect(wrapper.find('.confirm-message').text()).toContain('work')
    })

    it('does not show confirmation when clicking active workspace', async () => {
      const wrapper = await mountComponent()

      const activeItem = wrapper.findAll('.workspace-item')[0] // 'default' workspace (active)
      await activeItem.trigger('click')

      expect(wrapper.find('.confirm-dialog').exists()).toBe(false)
    })

    it('switches workspace on confirm', async () => {
      const wrapper = await mountComponent()

      const workItem = wrapper.findAll('.workspace-item')[1]
      await workItem.trigger('click')

      await wrapper.find('.confirm-btn').trigger('click')
      await flushPromises()

      expect(mockAppService.SetActiveWorkspace).toHaveBeenCalledWith('work')
    })

    it('cancels switch on cancel button', async () => {
      const wrapper = await mountComponent()

      const workItem = wrapper.findAll('.workspace-item')[1]
      await workItem.trigger('click')

      await wrapper.find('.cancel-btn').trigger('click')

      expect(wrapper.find('.confirm-dialog').exists()).toBe(false)
      expect(mockAppService.SetActiveWorkspace).not.toHaveBeenCalled()
    })

    it('shows "don\'t ask again" checkbox', async () => {
      const wrapper = await mountComponent()

      const workItem = wrapper.findAll('.workspace-item')[1]
      await workItem.trigger('click')

      expect(wrapper.find('.dont-ask-label').exists()).toBe(true)
    })

    it('saves preference to localStorage when checkbox is checked', async () => {
      const wrapper = await mountComponent()

      const workItem = wrapper.findAll('.workspace-item')[1]
      await workItem.trigger('click')

      const checkbox = wrapper.find('.dont-ask-label input[type="checkbox"]')
      await checkbox.setValue(true)

      await wrapper.find('.confirm-btn').trigger('click')
      await flushPromises()

      expect(localStorageMock.setItem).toHaveBeenCalledWith(SKIP_CONFIRM_KEY, 'true')
    })

    it('skips confirmation when localStorage preference is set', async () => {
      localStorageMock.getItem.mockReturnValue('true')

      const wrapper = await mountComponent()

      const workItem = wrapper.findAll('.workspace-item')[1]
      await workItem.trigger('click')
      await flushPromises()

      // Should switch directly without showing dialog
      expect(wrapper.find('.confirm-dialog').exists()).toBe(false)
      expect(mockAppService.SetActiveWorkspace).toHaveBeenCalledWith('work')
    })
  })

  describe('singular/plural handling', () => {
    it('uses singular "skill" for count of 1', async () => {
      const wrapper = await mountComponent()

      const skillCounts = wrapper.findAll('.skill-count')
      expect(skillCounts[1].text()).toBe('1 skill') // 'work' has 1 skill
    })

    it('uses plural "skills" for count of 0', async () => {
      const wrapper = await mountComponent()

      const skillCounts = wrapper.findAll('.skill-count')
      expect(skillCounts[2].text()).toBe('0 skills') // 'personal' has 0 skills
    })

    it('uses plural "skills" for count > 1', async () => {
      const wrapper = await mountComponent()

      const skillCounts = wrapper.findAll('.skill-count')
      expect(skillCounts[0].text()).toBe('3 skills') // 'default' has 3 skills
    })
  })
})
